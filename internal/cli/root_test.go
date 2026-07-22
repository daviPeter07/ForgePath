package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	appconfig "github.com/daviPeter07/forgepath/internal/config"
)

func TestListCommand(t *testing.T) {
	workspace := t.TempDir()
	createCLIProject(t, workspace, "web", "package.json", "tsconfig.json")
	createCLIProject(t, workspace, "api", "go.mod")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	command := NewRootCommand(&stdout, &stderr)
	command.SetArgs([]string{"list", workspace})

	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	want := strings.Join([]string{
		"api\tGo\t" + filepath.Join(workspace, "api"),
		"web\tTypeScript\t" + filepath.Join(workspace, "web"),
		"",
	}, "\n")
	if stdout.String() != want {
		t.Fatalf("stdout = %q, want %q", stdout.String(), want)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestListCommandUsesCurrentDirectory(t *testing.T) {
	workspace := t.TempDir()
	createCLIProject(t, workspace, "app", "composer.json")

	originalDirectory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(workspace); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(originalDirectory); err != nil {
			t.Errorf("restore working directory: %v", err)
		}
	})

	var stdout bytes.Buffer
	command := NewRootCommand(&stdout, &bytes.Buffer{})
	command.SetArgs([]string{"list"})

	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(stdout.String(), "app\tPHP\t"+filepath.Join(workspace, "app")) {
		t.Fatalf("stdout = %q, want project from current directory", stdout.String())
	}
}

func TestListCommandReturnsErrors(t *testing.T) {
	command := NewRootCommand(&bytes.Buffer{}, &bytes.Buffer{})
	command.SetArgs([]string{"list", filepath.Join(t.TempDir(), "missing")})

	if err := command.Execute(); err == nil {
		t.Fatal("Execute() error = nil, want error")
	}
}

func TestRootCommandOpensSelector(t *testing.T) {
	workspace := t.TempDir()
	createCLIProject(t, workspace, "app", "go.mod")
	originalDirectory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(workspace); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(originalDirectory); err != nil {
			t.Errorf("restore working directory: %v", err)
		}
	})

	var stdout bytes.Buffer
	command := NewRootCommand(&stdout, &bytes.Buffer{})
	command.SetIn(strings.NewReader("c"))
	command.SetArgs([]string{"--icons", "nerd-font", "--refresh"})

	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	want := filepath.Join(workspace, "app") + "\n"
	if stdout.String() != want {
		t.Fatalf("stdout = %q, want %q", stdout.String(), want)
	}
}

func TestRootCommandUsesAutomaticIconModeByDefault(t *testing.T) {
	command := NewRootCommand(&bytes.Buffer{}, &bytes.Buffer{})
	flag := command.Flags().Lookup("icons")
	if flag == nil || flag.DefValue != "auto" {
		t.Fatalf("icons default = %v, want auto", flag)
	}
}

func TestRootCommandCancellation(t *testing.T) {
	workspace := t.TempDir()
	createCLIProject(t, workspace, "app", "go.mod")
	originalDirectory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(workspace); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(originalDirectory); err != nil {
			t.Errorf("restore working directory: %v", err)
		}
	})

	var stdout bytes.Buffer
	command := NewRootCommand(&stdout, &bytes.Buffer{})
	command.SetIn(strings.NewReader("q"))
	command.SetArgs([]string{})
	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty after cancellation", stdout.String())
	}
}

func TestRootCommandUsesConfiguredWorkspaceFromAnyDirectory(t *testing.T) {
	workspace := t.TempDir()
	createCLIProject(t, workspace, "global-app", "go.mod")
	configPath := filepath.Join(t.TempDir(), "config.json")
	configuration := appconfig.Default()
	configuration.Workspaces = []string{workspace}
	if err := appconfig.Save(configPath, configuration); err != nil {
		t.Fatal(err)
	}

	originalDirectory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(originalDirectory); err != nil {
			t.Errorf("restore working directory: %v", err)
		}
	})

	var stdout bytes.Buffer
	command := NewRootCommand(&stdout, &bytes.Buffer{})
	command.SetIn(strings.NewReader("c"))
	command.SetArgs([]string{"--config", configPath, "--icons", "ascii"})
	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	want := filepath.Join(workspace, "global-app") + "\n"
	if stdout.String() != want {
		t.Fatalf("stdout = %q, want %q", stdout.String(), want)
	}
}

func TestRootCommandReturnsErrorWithoutProjects(t *testing.T) {
	originalDirectory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(originalDirectory); err != nil {
			t.Errorf("restore working directory: %v", err)
		}
	})

	command := NewRootCommand(&bytes.Buffer{}, &bytes.Buffer{})
	command.SetArgs([]string{})
	if err := command.Execute(); err == nil {
		t.Fatal("Execute() error = nil, want no-projects error")
	}
}

func TestRootHelpDoesNotStartSelector(t *testing.T) {
	var stdout bytes.Buffer
	command := NewRootCommand(&stdout, &bytes.Buffer{})
	command.SetArgs([]string{"--help"})

	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(stdout.String(), "Available Commands:") {
		t.Fatalf("stdout = %q, want help", stdout.String())
	}
}

func createCLIProject(t *testing.T, workspace, name string, markers ...string) {
	t.Helper()
	dir := filepath.Join(workspace, name)
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, marker := range markers {
		if err := os.WriteFile(filepath.Join(dir, marker), nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
}
