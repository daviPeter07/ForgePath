package cli

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestRunCommandUsesConfiguredArguments(t *testing.T) {
	workspace := t.TempDir()
	createCLIProject(t, workspace, "app", "go.mod")
	configPath := writeCLIConfig(t, `{"projects":{"app":{"command":["go","run","."]}}}`)

	var receivedPath string
	var receivedArguments []string
	command := newRunCommand(
		func() (string, error) { return configPath, nil },
		func(_ context.Context, path string, arguments []string, _ io.Reader, _, _ io.Writer) error {
			receivedPath = path
			receivedArguments = arguments
			return nil
		},
	)
	command.SetArgs([]string{"app", workspace})

	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if receivedPath != filepath.Join(workspace, "app") {
		t.Fatalf("run path = %q, want project path", receivedPath)
	}
	want := []string{"go", "run", "."}
	if !reflect.DeepEqual(receivedArguments, want) {
		t.Fatalf("run arguments = %q, want %q", receivedArguments, want)
	}
}

func TestRunCommandRequiresProjectConfiguration(t *testing.T) {
	workspace := t.TempDir()
	createCLIProject(t, workspace, "app", "go.mod")
	configPath := writeCLIConfig(t, `{}`)
	command := newRunCommand(
		func() (string, error) { return configPath, nil },
		func(context.Context, string, []string, io.Reader, io.Writer, io.Writer) error {
			t.Fatal("run called without project configuration")
			return nil
		},
	)
	command.SetArgs([]string{"app", workspace})

	if err := command.Execute(); err == nil {
		t.Fatal("Execute() error = nil, want missing command error")
	}
}

func writeCLIConfig(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}
