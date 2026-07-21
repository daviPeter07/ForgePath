package cli

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestOpenCommandResolvesProject(t *testing.T) {
	workspace := t.TempDir()
	createCLIProject(t, workspace, "project with spaces", "go.mod")

	var receivedPath string
	var receivedEditor string
	command := newOpenCommand(func(_ context.Context, path, editor string) error {
		receivedPath = path
		receivedEditor = editor
		return nil
	}, unusedConfigPath)
	command.SetArgs([]string{"project with spaces", workspace, "--editor", "test-editor"})

	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if receivedPath != filepath.Join(workspace, "project with spaces") {
		t.Fatalf("open path = %q, want project path", receivedPath)
	}
	if receivedEditor != "test-editor" {
		t.Fatalf("editor = %q, want test-editor", receivedEditor)
	}
}

func TestRevealCommandResolvesProject(t *testing.T) {
	workspace := t.TempDir()
	createCLIProject(t, workspace, "app", "package.json")

	var receivedPath string
	command := newRevealCommand(func(_ context.Context, path string) error {
		receivedPath = path
		return nil
	})
	command.SetArgs([]string{"app", workspace})

	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if receivedPath != filepath.Join(workspace, "app") {
		t.Fatalf("reveal path = %q, want project path", receivedPath)
	}
}

func TestOpenCommandReturnsMissingProjectError(t *testing.T) {
	command := newOpenCommand(func(context.Context, string, string) error {
		t.Fatal("open editor called for missing project")
		return nil
	}, unusedConfigPath)
	command.SetArgs([]string{"missing", t.TempDir()})

	if err := command.Execute(); err == nil {
		t.Fatal("Execute() error = nil, want missing project error")
	}
}

func TestOpenCommandUsesEditorEnvironment(t *testing.T) {
	t.Setenv("FORGEPATH_EDITOR", "configured-editor")
	workspace := t.TempDir()
	createCLIProject(t, workspace, "app", "go.mod")

	var receivedEditor string
	command := newOpenCommand(func(_ context.Context, _, editor string) error {
		receivedEditor = editor
		return nil
	}, unusedConfigPath)
	command.SetArgs([]string{"app", workspace})

	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if receivedEditor != "configured-editor" {
		t.Fatalf("editor = %q, want configured-editor", receivedEditor)
	}
}

func TestOpenCommandUsesConfiguredEditor(t *testing.T) {
	t.Setenv("FORGEPATH_EDITOR", "")
	workspace := t.TempDir()
	createCLIProject(t, workspace, "app", "go.mod")
	configPath := writeCLIConfig(t, `{"editor":{"executable":"configured-editor.exe"}}`)

	var receivedEditor string
	command := newOpenCommand(func(_ context.Context, _, editor string) error {
		receivedEditor = editor
		return nil
	}, func() (string, error) { return configPath, nil })
	command.SetArgs([]string{"app", workspace})

	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if receivedEditor != "configured-editor.exe" {
		t.Fatalf("editor = %q, want configured editor", receivedEditor)
	}
}

func TestResolveProjectUsesCurrentDirectory(t *testing.T) {
	workspace := t.TempDir()
	createCLIProject(t, workspace, "app", "go.mod")
	original, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(workspace); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(original); err != nil {
			t.Errorf("restore working directory: %v", err)
		}
	})

	found, err := resolveProject(context.Background(), "app", nil)
	if err != nil {
		t.Fatalf("resolveProject() error = %v", err)
	}
	if found.Path != filepath.Join(workspace, "app") {
		t.Fatalf("resolved path = %q, want project path", found.Path)
	}
}

func unusedConfigPath() (string, error) {
	return "", errors.New("config path should not be used")
}
