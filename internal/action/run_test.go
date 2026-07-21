package action

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRunCommandUsesProjectDirectory(t *testing.T) {
	goExecutable, err := exec.LookPath("go")
	if err != nil {
		t.Skip("go executable is not available")
	}
	dir := t.TempDir()
	modulePath := filepath.Join(dir, "go.mod")
	if err := os.WriteFile(modulePath, []byte("module example.com/app"), 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err = RunCommand(context.Background(), dir, []string{goExecutable, "env", "GOMOD"}, strings.NewReader(""), &stdout, &stderr)
	if err != nil {
		t.Fatalf("RunCommand() error = %v, stderr = %q", err, stderr.String())
	}
	if strings.TrimSpace(stdout.String()) != modulePath {
		t.Fatalf("RunCommand() stdout = %q, want %q", stdout.String(), modulePath)
	}
}

func TestRunCommandRejectsInvalidCommands(t *testing.T) {
	if err := RunCommand(context.Background(), t.TempDir(), nil, nil, nil, nil); err == nil {
		t.Fatal("RunCommand() empty command error = nil")
	}
	if _, err := safeExecutable("windows", "pnpm.cmd"); err == nil {
		t.Fatal("safeExecutable() batch command error = nil")
	}
}

func TestRunCommandHonorsCancelledContext(t *testing.T) {
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("FORGEPATH_RUN_HELPER", "1")
	ctx, cancel := context.WithCancel(context.Background())
	time.AfterFunc(100*time.Millisecond, cancel)

	err = RunCommand(ctx, t.TempDir(), []string{executable, "-test.run=^TestRunCommandHelperProcess$"}, nil, nil, nil)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("RunCommand() error = %v, want context.Canceled", err)
	}
}

func TestRunCommandDoesNotStartWithCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := RunCommand(ctx, t.TempDir(), []string{"missing-command"}, nil, nil, nil)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("RunCommand() error = %v, want context.Canceled before executable lookup", err)
	}
}

func TestRunCommandHelperProcess(t *testing.T) {
	if os.Getenv("FORGEPATH_RUN_HELPER") != "1" {
		return
	}
	time.Sleep(30 * time.Second)
}
