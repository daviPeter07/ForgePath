package action

import (
	"context"
	"errors"
	"os/exec"
	"reflect"
	"runtime"
	"testing"
)

func TestFolderCommand(t *testing.T) {
	tests := []struct {
		goos       string
		executable string
	}{
		{goos: "windows", executable: "explorer.exe"},
		{goos: "darwin", executable: "open"},
		{goos: "linux", executable: "xdg-open"},
	}

	for _, tt := range tests {
		executable, arguments, err := folderCommand(tt.goos, "path with spaces")
		if err != nil {
			t.Fatalf("folderCommand(%q) error = %v", tt.goos, err)
		}
		if executable != tt.executable {
			t.Fatalf("folderCommand(%q) executable = %q, want %q", tt.goos, executable, tt.executable)
		}
		if !reflect.DeepEqual(arguments, []string{"path with spaces"}) {
			t.Fatalf("folderCommand(%q) arguments = %q", tt.goos, arguments)
		}
	}
}

func TestFolderCommandRejectsUnsupportedPlatform(t *testing.T) {
	if _, _, err := folderCommand("plan9", "project"); err == nil {
		t.Fatal("folderCommand() error = nil, want unsupported platform error")
	}
}

func TestOpenEditorRejectsEmptyExecutable(t *testing.T) {
	if err := OpenEditor(context.Background(), "project", ""); err == nil {
		t.Fatal("OpenEditor() error = nil, want error")
	}
}

func TestEditorCommandRejectsWindowsBatchLauncher(t *testing.T) {
	if _, err := editorCommand("windows", "code.cmd", "project"); err == nil {
		t.Fatal("editorCommand() error = nil, want batch launcher error")
	}
}

func TestEditorCommandIncludesEditorArgumentsBeforePath(t *testing.T) {
	goExecutable, err := exec.LookPath("go")
	if err != nil {
		t.Fatal(err)
	}
	command, err := editorCommandWithArguments(runtime.GOOS, goExecutable, []string{"tool"}, "project")
	if err != nil {
		t.Fatalf("editorCommandWithArguments() error = %v", err)
	}
	if len(command.Args) < 3 || command.Args[len(command.Args)-2] != "tool" || command.Args[len(command.Args)-1] != "project" {
		t.Fatalf("command args = %q, want editor args before project path", command.Args)
	}
}

func TestOpenEditorHonorsCancelledContext(t *testing.T) {
	goExecutable, err := exec.LookPath("go")
	if err != nil {
		t.Skip("go executable is not available")
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := OpenEditor(ctx, "project", goExecutable); !errors.Is(err, context.Canceled) {
		t.Fatalf("OpenEditor() error = %v, want context cancellation", err)
	}
}
