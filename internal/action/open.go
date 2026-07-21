package action

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

func OpenEditor(ctx context.Context, path, editor string) error {
	if editor == "" {
		return fmt.Errorf("editor executable cannot be empty")
	}
	command, err := editorCommand(runtime.GOOS, editor, path)
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
	if goos == "windows" {
		extension := strings.ToLower(filepath.Ext(editor))
		if extension == ".cmd" || extension == ".bat" {
			return nil, fmt.Errorf("batch editor launcher %q is not supported; use the editor .exe", editor)
		}
	}

	executable, err := exec.LookPath(editor)
	if err != nil {
		return nil, err
	}
	if goos == "windows" {
		extension := strings.ToLower(filepath.Ext(executable))
		if extension == ".cmd" || extension == ".bat" {
			return nil, fmt.Errorf("batch editor launcher %q is not supported; use the editor .exe", executable)
		}
	}
	return exec.Command(executable, path), nil
}

func startEditor(ctx context.Context, command *exec.Cmd) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := command.Start(); err != nil {
		return err
	}

	finished := make(chan error, 1)
	go func() {
		finished <- command.Wait()
	}()

	timer := time.NewTimer(250 * time.Millisecond)
	defer timer.Stop()
	select {
	case err := <-finished:
		return err
	case <-ctx.Done():
		_ = command.Process.Kill()
		<-finished
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
