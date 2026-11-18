package cli

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"

	appconfig "github.com/daviPeter07/forgepath/internal/config"
	"github.com/daviPeter07/forgepath/internal/project"
	appstate "github.com/daviPeter07/forgepath/internal/state"
)

// writeWorkspaceConfig saves a config whose catalog contains the given workspaces.
func writeWorkspaceConfig(t *testing.T, editor string, workspaces ...string) configPathFunc {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.json")
	configuration := appconfig.Default()
	configuration.Workspaces = workspaces
	configuration.Editor.Executable = editor
	if err := appconfig.Save(path, configuration); err != nil {
		t.Fatal(err)
	}
	return func() (string, error) { return path, nil }
}

func TestRootListShowsProjectsFromEveryConfiguredWorkspace(t *testing.T) {
	work, clients := t.TempDir(), t.TempDir()
	createCLIProject(t, work, "api", "go.mod")
	createCLIProject(t, clients, "storefront", "package.json")
	configPath := writeWorkspaceConfig(t, "", work, clients)
	path, _ := configPath()

	var output bytes.Buffer
	command := NewRootCommand(&output, &bytes.Buffer{})
	command.SetArgs([]string{"list", "--config", path, "--cache", t.TempDir()})
	if err := command.Execute(); err != nil {
		t.Fatalf("list error = %v", err)
	}
	for _, want := range []string{filepath.Join(work, "api"), filepath.Join(clients, "storefront")} {
		if !strings.Contains(output.String(), want) {
			t.Fatalf("list output = %q, want %q", output.String(), want)
		}
	}
}

func TestResolveProjectSearchesConfiguredWorkspaces(t *testing.T) {
	work, clients := t.TempDir(), t.TempDir()
	createCLIProject(t, clients, "storefront", "package.json")
	configPath := writeWorkspaceConfig(t, "", work, clients)

	found, err := resolveProject(context.Background(), "storefront", nil, configPath)
	if err != nil {
		t.Fatalf("resolveProject() error = %v", err)
	}
	if found.Path != filepath.Join(clients, "storefront") {
		t.Fatalf("resolved %q, want the project in the second workspace", found.Path)
	}
}

func TestResolveProjectReportsAmbiguousNames(t *testing.T) {
	work, clients := t.TempDir(), t.TempDir()
	createCLIProject(t, work, "api", "go.mod")
	createCLIProject(t, clients, "api", "go.mod")
	configPath := writeWorkspaceConfig(t, "", work, clients)

	_, err := resolveProject(context.Background(), "api", nil, configPath)
	if err == nil {
		t.Fatal("resolveProject() error = nil, want ambiguity error")
	}
	for _, want := range []string{filepath.Join(work, "api"), filepath.Join(clients, "api"), "workspace"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("error = %q, want it to mention %q", err, want)
		}
	}

	found, err := resolveProject(context.Background(), "api", []string{clients}, configPath)
	if err != nil || found.Path != filepath.Join(clients, "api") {
		t.Fatalf("explicit workspace resolved %q, %v", found.Path, err)
	}
}

func TestResolveProjectNotFoundNamesSearchedWorkspaces(t *testing.T) {
	work := t.TempDir()
	configPath := writeWorkspaceConfig(t, "", work)

	_, err := resolveProject(context.Background(), "missing", nil, configPath)
	if err == nil || !strings.Contains(err.Error(), work) {
		t.Fatalf("error = %v, want it to name %q", err, work)
	}
}

func TestOpenFindsProjectInConfiguredWorkspaceFromAnyDirectory(t *testing.T) {
	work := t.TempDir()
	createCLIProject(t, work, "api", "go.mod")
	configPath := writeWorkspaceConfig(t, "configured-editor", work)

	var openedPath, openedEditor string
	command := newOpenCommand(func(_ context.Context, path, editor string) error {
		openedPath, openedEditor = path, editor
		return nil
	}, configPath)
	command.SetArgs([]string{"api"})
	if err := command.Execute(); err != nil {
		t.Fatalf("open error = %v", err)
	}
	if openedPath != filepath.Join(work, "api") || openedEditor != "configured-editor" {
		t.Fatalf("opened %q with %q", openedPath, openedEditor)
	}
}

func TestFavoriteAddFindsProjectInConfiguredWorkspace(t *testing.T) {
	work := t.TempDir()
	createCLIProject(t, work, "api", "go.mod")
	configPath := writeWorkspaceConfig(t, "", work)
	statePath := filepath.Join(t.TempDir(), "state.json")

	command := newFavoriteCommand(func() (string, error) { return statePath, nil }, configPath)
	command.SetArgs([]string{"add", "api"})
	if err := command.Execute(); err != nil {
		t.Fatalf("favorite add error = %v", err)
	}
	value, err := (appstate.Store{Path: statePath}).Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(value.Favorites) != 1 || !sameWorkspace(value.Favorites[0], filepath.Join(work, "api")) {
		t.Fatalf("favorites = %q", value.Favorites)
	}
}

func TestWorkspaceAddExplainsFoldersWithoutProjects(t *testing.T) {
	tests := []struct {
		name    string
		prepare func(t *testing.T, folder string) string
		hint    string
	}{
		{
			name:    "empty folder",
			prepare: func(t *testing.T, folder string) string { return folder },
			hint:    "no projects",
		},
		{
			name: "folder is a project itself",
			prepare: func(t *testing.T, folder string) string {
				createCLIProject(t, folder, "api", "go.mod")
				return filepath.Join(folder, "api")
			},
			hint: "parent folder",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workspace := tt.prepare(t, t.TempDir())
			configPath := writeWorkspaceConfig(t, "")
			var output, errors bytes.Buffer
			command := newWorkspaceCommand(configPath)
			command.SetOut(&output)
			command.SetErr(&errors)
			command.SetArgs([]string{"add", workspace})
			if err := command.Execute(); err != nil {
				t.Fatalf("workspace add error = %v", err)
			}
			if !strings.Contains(errors.String(), tt.hint) {
				t.Fatalf("stderr = %q, want hint %q", errors.String(), tt.hint)
			}
		})
	}
}

func TestWorkspaceAddReportsProjectCount(t *testing.T) {
	workspace := t.TempDir()
	createCLIProject(t, workspace, "api", "go.mod")
	createCLIProject(t, workspace, "web", "package.json")
	configPath := writeWorkspaceConfig(t, "")

	var output, errors bytes.Buffer
	command := newWorkspaceCommand(configPath)
	command.SetOut(&output)
	command.SetErr(&errors)
	command.SetArgs([]string{"add", workspace})
	if err := command.Execute(); err != nil {
		t.Fatalf("workspace add error = %v", err)
	}
	if strings.TrimSpace(output.String()) != workspace {
		t.Fatalf("stdout = %q, want only the workspace path", output.String())
	}
	if !strings.Contains(errors.String(), "2 projects") {
		t.Fatalf("stderr = %q, want project count", errors.String())
	}
}

func TestFavoriteTogglerPersistsState(t *testing.T) {
	statePath := filepath.Join(t.TempDir(), "state.json")
	toggle := favoriteToggler(func() (string, error) { return statePath, nil })
	target := project.Project{Name: "api", Path: t.TempDir()}

	favorite, err := toggle(target)
	if err != nil || !favorite {
		t.Fatalf("first toggle = %t, %v; want favorite", favorite, err)
	}
	target.Favorite = true
	favorite, err = toggle(target)
	if err != nil || favorite {
		t.Fatalf("second toggle = %t, %v; want not favorite", favorite, err)
	}
	value, err := (appstate.Store{Path: statePath}).Load()
	if err != nil || len(value.Favorites) != 0 {
		t.Fatalf("favorites = %q, %v; want empty", value.Favorites, err)
	}
}

func TestPickerHintOnlyWhenCatalogIsEmpty(t *testing.T) {
	empty := writeWorkspaceConfig(t, "")
	missing := func() (string, error) { return filepath.Join(t.TempDir(), "missing.json"), nil }
	configured := writeWorkspaceConfig(t, "", t.TempDir())

	if hint := pickerHint(nil, empty); !strings.Contains(hint, "fg workspace add") {
		t.Fatalf("hint with empty catalog = %q", hint)
	}
	if hint := pickerHint(nil, missing); !strings.Contains(hint, "fg workspace add") {
		t.Fatalf("hint without config file = %q", hint)
	}
	if hint := pickerHint(nil, configured); hint != "" {
		t.Fatalf("hint with workspaces = %q, want none", hint)
	}
	if hint := pickerHint([]string{t.TempDir()}, empty); hint != "" {
		t.Fatalf("hint with explicit workspace = %q, want none", hint)
	}
	if hint := pickerHint(nil, nil); hint != "" {
		t.Fatalf("hint without config = %q, want none", hint)
	}
}
