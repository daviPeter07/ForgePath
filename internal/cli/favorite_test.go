package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	appstate "github.com/daviPeter07/forgepath/internal/state"
)

func TestFavoriteCommands(t *testing.T) {
	workspace := t.TempDir()
	createCLIProject(t, workspace, "app", "go.mod")
	stateFile := filepath.Join(t.TempDir(), "state.json")
	statePath := func() (string, error) { return stateFile, nil }

	add := newFavoriteCommand(statePath)
	add.SetArgs([]string{"add", "app", workspace})
	if err := add.Execute(); err != nil {
		t.Fatalf("favorite add error = %v", err)
	}

	var stdout bytes.Buffer
	list := newFavoriteCommand(statePath)
	list.SetOut(&stdout)
	list.SetArgs([]string{"list"})
	if err := list.Execute(); err != nil {
		t.Fatalf("favorite list error = %v", err)
	}
	wantPath := filepath.Join(workspace, "app")
	if strings.TrimSpace(stdout.String()) != wantPath {
		t.Fatalf("favorite list = %q, want %q", stdout.String(), wantPath)
	}

	remove := newFavoriteCommand(statePath)
	remove.SetArgs([]string{"remove", "app", workspace})
	if err := remove.Execute(); err != nil {
		t.Fatalf("favorite remove error = %v", err)
	}
	value, err := (appstate.Store{Path: stateFile}).Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(value.Favorites) != 0 {
		t.Fatalf("favorites = %q, want empty", value.Favorites)
	}
}

func TestSuccessfulOpenRecordsRecentProject(t *testing.T) {
	workspace := t.TempDir()
	createCLIProject(t, workspace, "app", "go.mod")
	stateFile := filepath.Join(t.TempDir(), "state.json")
	statePath := func() (string, error) { return stateFile, nil }

	command := newOpenCommand(
		func(context.Context, string, string) error { return nil },
		unusedConfigPath,
		statePath,
	)
	command.SetArgs([]string{"app", workspace, "--editor", "test-editor"})
	if err := command.Execute(); err != nil {
		t.Fatalf("open error = %v", err)
	}

	value, err := (appstate.Store{Path: stateFile}).Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(value.Recent) != 1 || value.Recent[0].Path != filepath.Join(workspace, "app") {
		t.Fatalf("recent = %+v, want opened project", value.Recent)
	}
}

func TestSuccessfulOpenIgnoresBrokenRecentState(t *testing.T) {
	workspace := t.TempDir()
	createCLIProject(t, workspace, "app", "go.mod")
	stateFile := filepath.Join(t.TempDir(), "state.json")
	if err := os.WriteFile(stateFile, []byte("broken"), 0o600); err != nil {
		t.Fatal(err)
	}

	var stderr bytes.Buffer
	command := newOpenCommand(
		func(context.Context, string, string) error { return nil },
		unusedConfigPath,
		func() (string, error) { return stateFile, nil },
	)
	command.SetErr(&stderr)
	command.SetArgs([]string{"app", workspace, "--editor", "test-editor"})
	if err := command.Execute(); err != nil {
		t.Fatalf("open error = %v, want successful action", err)
	}
	if !strings.Contains(stderr.String(), "warning:") {
		t.Fatalf("stderr = %q, want state warning", stderr.String())
	}
}

func TestRecentCommand(t *testing.T) {
	stateFile := filepath.Join(t.TempDir(), "state.json")
	store := appstate.Store{Path: stateFile}
	if err := store.RecordRecent(filepath.Join("projects", "app")); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	command := newRecentCommand(func() (string, error) { return stateFile, nil })
	command.SetOut(&stdout)
	if err := command.Execute(); err != nil {
		t.Fatalf("recent error = %v", err)
	}
	if !strings.Contains(stdout.String(), filepath.Join("projects", "app")) {
		t.Fatalf("recent output = %q, want project path", stdout.String())
	}
}
