package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/daviPeter07/forgepath/internal/project"
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
	command.SetIn(strings.NewReader("\x1b[Bc"))
	command.SetArgs([]string{"pick", workspace, "--print-path", "--icons", "nerd-font"})

	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	want := filepath.Join(workspace, "web app") + "\n"
	if stdout.String() != want {
		t.Fatalf("stdout = %q, want %q", stdout.String(), want)
	}
	if stderr.Len() == 0 {
		t.Fatal("stderr is empty, want TUI rendering")
	}
}

func TestPickCommandCancellation(t *testing.T) {
	workspace := t.TempDir()
	createCLIProject(t, workspace, "app", "go.mod")

	var stdout bytes.Buffer
	command := NewRootCommand(&stdout, &bytes.Buffer{})
	command.SetIn(strings.NewReader("q"))
	command.SetArgs([]string{"pick", workspace})

	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
}

func TestPickCommandRejectsInvalidIconMode(t *testing.T) {
	command := NewRootCommand(&bytes.Buffer{}, &bytes.Buffer{})
	command.SetArgs([]string{"pick", "--icons", "emoji"})

	if err := command.Execute(); err == nil {
		t.Fatal("Execute() error = nil, want invalid icon mode error")
	}
}

func TestProjectRootForPathUsesContainingProject(t *testing.T) {
	root := filepath.Join(t.TempDir(), "project")
	nested := filepath.Join(root, "internal", "service")
	projects := []project.Project{{Name: "project", Path: root}}
	if got := projectRootForPath(projects, nested); got != root {
		t.Fatalf("projectRootForPath() = %q, want %q", got, root)
	}
	outside := filepath.Join(t.TempDir(), "other")
	if got := projectRootForPath(projects, outside); got != outside {
		t.Fatalf("outside projectRootForPath() = %q, want unchanged", got)
	}
}

func makeDirectory(path string) error {
	return os.Mkdir(path, 0o755)
}
