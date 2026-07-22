package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"
)

func TestInitAndLoad(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "config.json")
	if err := Init(path); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	configuration, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if configuration.Projects == nil {
		t.Fatal("Load().Projects = nil, want initialized map")
	}

	if err := Init(path); !errors.Is(err, os.ErrExist) {
		t.Fatalf("second Init() error = %v, want os.ErrExist", err)
	}
}

func TestLoadProjectCommand(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	content := `{
  "editor": {"executable": "code.exe"},
  "projects": {"app": {"command": ["pnpm", "dev"]}}
}`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	configuration, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if configuration.Editor.Executable != "code.exe" {
		t.Fatalf("editor = %q, want code.exe", configuration.Editor.Executable)
	}
	want := []string{"pnpm", "dev"}
	if !reflect.DeepEqual(configuration.Projects["app"].Command, want) {
		t.Fatalf("command = %q, want %q", configuration.Projects["app"].Command, want)
	}
}

func TestSaveWorkspaces(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "config.json")
	want := []string{filepath.Join(t.TempDir(), "projects"), filepath.Join(t.TempDir(), "work")}
	configuration := Default()
	configuration.Workspaces = want

	if err := Save(path, configuration); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if !reflect.DeepEqual(loaded.Workspaces, want) {
		t.Fatalf("Load().Workspaces = %q, want %q", loaded.Workspaces, want)
	}

	configuration.Workspaces = want[:1]
	if err := Save(path, configuration); err != nil {
		t.Fatalf("second Save() error = %v", err)
	}
	loaded, err = Load(path)
	if err != nil {
		t.Fatalf("Load() after replacement error = %v", err)
	}
	if !reflect.DeepEqual(loaded.Workspaces, want[:1]) {
		t.Fatalf("replaced Workspaces = %q, want %q", loaded.Workspaces, want[:1])
	}
}

func TestUpdateSerializesConcurrentChanges(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	const updates = 12
	var wait sync.WaitGroup
	for index := range updates {
		wait.Add(1)
		go func() {
			defer wait.Done()
			workspace := filepath.Join(t.TempDir(), fmt.Sprintf("workspace-%d", index))
			if err := Update(path, func(configuration *Config) error {
				configuration.Workspaces = append(configuration.Workspaces, workspace)
				return nil
			}); err != nil {
				t.Errorf("Update() error = %v", err)
			}
		}()
	}
	wait.Wait()

	configuration, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(configuration.Workspaces) != updates {
		t.Fatalf("len(Workspaces) = %d, want %d", len(configuration.Workspaces), updates)
	}
}

func TestLoadRejectsInvalidConfiguration(t *testing.T) {
	tests := []string{
		`{"unknown": true}`,
		`{"workspaces": ["relative/path"]}`,
		`{"projects": {"app": {"command": []}}}`,
		`{} {}`,
		`null`,
		`[]`,
	}

	for _, content := range tests {
		path := filepath.Join(t.TempDir(), "config.json")
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := Load(path); err == nil {
			t.Fatalf("Load(%s) error = nil, want error", content)
		}
	}
}
