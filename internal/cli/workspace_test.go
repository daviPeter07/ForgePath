package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	appconfig "github.com/daviPeter07/forgepath/internal/config"
)

func TestWorkspaceAddListAndRemove(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "forgepath", "config.json")
	workspace := t.TempDir()
	path := func() (string, error) { return configPath, nil }

	var output bytes.Buffer
	add := newWorkspaceCommand(path)
	add.SetOut(&output)
	add.SetArgs([]string{"add", workspace})
	if err := add.Execute(); err != nil {
		t.Fatalf("workspace add error = %v", err)
	}
	absolute, err := filepath.Abs(workspace)
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(output.String()) != absolute {
		t.Fatalf("workspace add output = %q, want %q", output.String(), absolute)
	}

	configuration, err := appconfig.Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(configuration.Workspaces) != 1 || !sameWorkspace(configuration.Workspaces[0], absolute) {
		t.Fatalf("Workspaces = %q, want %q", configuration.Workspaces, absolute)
	}

	output.Reset()
	list := newWorkspaceCommand(path)
	list.SetOut(&output)
	list.SetArgs([]string{"list"})
	if err := list.Execute(); err != nil {
		t.Fatalf("workspace list error = %v", err)
	}
	if strings.TrimSpace(output.String()) != absolute {
		t.Fatalf("workspace list output = %q, want %q", output.String(), absolute)
	}

	remove := newWorkspaceCommand(path)
	remove.SetArgs([]string{"remove", workspace})
	if err := remove.Execute(); err != nil {
		t.Fatalf("workspace remove error = %v", err)
	}
	configuration, err = appconfig.Load(configPath)
	if err != nil {
		t.Fatalf("Load() after remove error = %v", err)
	}
	if len(configuration.Workspaces) != 0 {
		t.Fatalf("Workspaces after remove = %q, want empty", configuration.Workspaces)
	}
}

func TestConfiguredWorkspacesUsesConfiguration(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	want := []string{t.TempDir(), t.TempDir()}
	configuration := appconfig.Default()
	configuration.Workspaces = want
	if err := appconfig.Save(path, configuration); err != nil {
		t.Fatal(err)
	}

	got, err := configuredWorkspaces(nil, func() (string, error) { return path, nil })
	if err != nil {
		t.Fatalf("configuredWorkspaces() error = %v", err)
	}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Fatalf("configuredWorkspaces() = %q, want %q", got, want)
	}
}

func TestConfiguredWorkspacesFallsBackToCurrentDirectory(t *testing.T) {
	current, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	got, err := configuredWorkspaces(nil, func() (string, error) {
		return filepath.Join(t.TempDir(), "missing.json"), nil
	})
	if err != nil {
		t.Fatalf("configuredWorkspaces() error = %v", err)
	}
	if len(got) != 1 || !sameWorkspace(got[0], current) {
		t.Fatalf("configuredWorkspaces() = %q, want current directory %q", got, current)
	}
}

func TestRootListUsesAllConfiguredWorkspaces(t *testing.T) {
	first := t.TempDir()
	second := t.TempDir()
	createCLIProject(t, first, "api", "go.mod")
	createCLIProject(t, second, "web", "package.json", "tsconfig.json")
	configPath := filepath.Join(t.TempDir(), "config.json")
	configuration := appconfig.Default()
	configuration.Workspaces = []string{first, second}
	if err := appconfig.Save(configPath, configuration); err != nil {
		t.Fatal(err)
	}

	var output bytes.Buffer
	command := NewRootCommand(&output, &bytes.Buffer{})
	command.SetArgs([]string{"--config", configPath, "list"})
	if err := command.Execute(); err != nil {
		t.Fatalf("list configured workspaces error = %v", err)
	}
	if !strings.Contains(output.String(), filepath.Join(first, "api")) || !strings.Contains(output.String(), filepath.Join(second, "web")) {
		t.Fatalf("list output = %q, want projects from both workspaces", output.String())
	}
}

func TestRootListFailsWhenAllConfiguredWorkspacesFail(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	configuration := appconfig.Default()
	configuration.Workspaces = []string{
		filepath.Join(t.TempDir(), "missing-one"),
		filepath.Join(t.TempDir(), "missing-two"),
	}
	if err := appconfig.Save(configPath, configuration); err != nil {
		t.Fatal(err)
	}

	command := NewRootCommand(&bytes.Buffer{}, &bytes.Buffer{})
	command.SetArgs([]string{"--config", configPath, "list"})
	if err := command.Execute(); err == nil {
		t.Fatal("list error = nil, want all configured workspace failures")
	}
}
