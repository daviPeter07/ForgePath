package action

import (
	"context"
	"os/exec"
	"reflect"
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

func TestOpenEditorReportsImmediateFailure(t *testing.T) {
	goExecutable, err := exec.LookPath("go")
	if err != nil {
		t.Skip("go executable is not available")
	}

	if err := OpenEditor(context.Background(), "not-a-go-command", goExecutable); err == nil {
		t.Fatal("OpenEditor() error = nil, want editor process error")
	}
}
