package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPickCommandPrintsOnlySelectedPath(t *testing.T) {
	workspace := filepath.Join(t.TempDir(), "workspace with spaces")
	if err := makeDirectory(workspace); err != nil {
		t.Fatal(err)
	}
	createCLIProject(t, workspace, "api service", "go.mod")
	createCLIProject(t, workspace, "web app", "package.json")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	command := NewRootCommand(&stdout, &stderr)
	command.SetIn(strings.NewReader("2\n"))
	command.SetArgs([]string{"pick", workspace, "--print-path"})

	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	want := filepath.Join(workspace, "web app") + "\n"
	if stdout.String() != want {
		t.Fatalf("stdout = %q, want %q", stdout.String(), want)
	}
	if !strings.Contains(stderr.String(), "web app (JavaScript)") {
		t.Fatalf("stderr = %q, want visual project list", stderr.String())
	}
}

func TestPickCommandCancellation(t *testing.T) {
	workspace := t.TempDir()
	createCLIProject(t, workspace, "app", "go.mod")

	var stdout bytes.Buffer
	command := NewRootCommand(&stdout, &bytes.Buffer{})
	command.SetIn(strings.NewReader("\n"))
	command.SetArgs([]string{"pick", workspace})

	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
}

func TestPickCommandRejectsInvalidSelection(t *testing.T) {
	workspace := t.TempDir()
	createCLIProject(t, workspace, "app", "go.mod")

	command := NewRootCommand(&bytes.Buffer{}, &bytes.Buffer{})
	command.SetIn(strings.NewReader("2\n"))
	command.SetArgs([]string{"pick", workspace})

	if err := command.Execute(); err == nil {
		t.Fatal("Execute() error = nil, want invalid selection error")
	}
}

func makeDirectory(path string) error {
	return os.Mkdir(path, 0o755)
}
