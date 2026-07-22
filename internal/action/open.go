package action

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func OpenEditor(ctx context.Context, path, editor string) error {
	return OpenEditorWithArguments(ctx, path, editor, nil)
}

func OpenEditorWithArguments(ctx context.Context, path, editor string, arguments []string) error {
	if editor == "" {
		return fmt.Errorf("editor executable cannot be empty")
	}
	command, err := editorCommandWithArguments(runtime.GOOS, editor, arguments, path)
	if err != nil {
		return err
	}
	return startEditor(ctx, command)
}

func OpenFolder(ctx context.Context, path string) error {
	executable, arguments, err := folderCommand(runtime.GOOS, path)
	if err != nil {
		return err
	}
	return exec.CommandContext(ctx, executable, arguments...).Run()
}

func folderCommand(goos, path string) (string, []string, error) {
	switch goos {
	case "windows":
		return "explorer.exe", []string{path}, nil
	case "darwin":
		return "open", []string{path}, nil
	case "linux":
		return "xdg-open", []string{path}, nil
	default:
		return "", nil, fmt.Errorf("opening folders is not supported on %s", goos)
	}
}

func editorCommand(goos, editor, path string) (*exec.Cmd, error) {
	return editorCommandWithArguments(goos, editor, nil, path)
}

func editorCommandWithArguments(goos, editor string, arguments []string, path string) (*exec.Cmd, error) {
	executable, err := safeExecutable(goos, editor)
	if err != nil {
		return nil, err
	}
	arguments = append(append([]string(nil), arguments...), path)
	return exec.Command(executable, arguments...), nil
}

func safeExecutable(goos, executable string) (string, error) {
	if goos == "windows" {
		extension := strings.ToLower(filepath.Ext(executable))
		if extension == ".cmd" || extension == ".bat" {
			return "", fmt.Errorf("batch launcher %q is not supported; use an .exe", executable)
		}
	}

	resolved, err := exec.LookPath(executable)
	if err != nil {
		return "", err
	}
	if goos == "windows" {
		extension := strings.ToLower(filepath.Ext(resolved))
		if extension == ".cmd" || extension == ".bat" {
			return "", fmt.Errorf("batch launcher %q is not supported; use an .exe", resolved)
		}
	}
	return resolved, nil
}

func startEditor(ctx context.Context, command *exec.Cmd) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	prepareCommand(command, false)
	if err := command.Start(); err != nil {
		return err
	}
	return command.Process.Release()
}
