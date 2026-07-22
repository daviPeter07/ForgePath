package ide

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/daviPeter07/forgepath/internal/project"
)

func TestDiscoverFindsOnlyInstalledEditors(t *testing.T) {
	directory := t.TempDir()
	code := filepath.Join(directory, "code.exe")
	phpstorm := filepath.Join(directory, "phpstorm64.exe")
	for _, path := range []string{code, phpstorm} {
		if err := os.WriteFile(path, nil, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	paths := map[string]string{"code": code, "phpstorm": phpstorm}
	found := discover(finder{
		goos: "windows",
		lookPath: func(command string) (string, error) {
			if path := paths[command]; path != "" {
				return path, nil
			}
			return "", errors.New("not found")
		},
		glob:   filepath.Glob,
		stat:   os.Stat,
		getenv: func(string) string { return filepath.Join(directory, "missing") },
	})

	ids := make([]string, len(found))
	for index, editor := range found {
		ids[index] = editor.ID
	}
	if !reflect.DeepEqual(ids, []string{"phpstorm", "vscode"}) {
		t.Fatalf("installed editor IDs = %q, want phpstorm and vscode", ids)
	}
}

func TestDiscoverRejectsWindowsBatchLaunchers(t *testing.T) {
	found := discover(finder{
		goos:     "windows",
		lookPath: func(string) (string, error) { return `C:\\tools\\code.cmd`, nil },
		glob:     func(string) ([]string, error) { return nil, nil },
		stat:     func(string) (os.FileInfo, error) { return nil, os.ErrNotExist },
		getenv:   func(string) string { return "" },
	})
	if len(found) != 0 {
		t.Fatalf("discover() = %+v, want no batch launchers", found)
	}
}

func TestFindExecutableRejectsNonExecutableUnixFile(t *testing.T) {
	path := "/opt/editor/bin/editor"
	mode := os.FileMode(0o644)
	system := finder{
		goos:     "linux",
		lookPath: func(string) (string, error) { return "", errors.New("not found") },
		glob:     func(string) ([]string, error) { return []string{path}, nil },
		stat:     func(string) (os.FileInfo, error) { return fakeFileInfo{mode: mode}, nil },
	}
	if executable := findExecutable(system, candidate{paths: []string{path}}); executable != "" {
		t.Fatalf("findExecutable() = %q, want non-executable file rejected", executable)
	}
	mode = 0o755
	if executable := findExecutable(system, candidate{paths: []string{path}}); executable != path {
		t.Fatalf("findExecutable() = %q, want %q", executable, path)
	}
}

type fakeFileInfo struct{ mode os.FileMode }

func (fakeFileInfo) Name() string           { return "editor" }
func (fakeFileInfo) Size() int64            { return 0 }
func (info fakeFileInfo) Mode() os.FileMode { return info.mode }
func (fakeFileInfo) ModTime() time.Time     { return time.Time{} }
func (fakeFileInfo) IsDir() bool            { return false }
func (fakeFileInfo) Sys() any               { return nil }

func TestRankPrefersTechnologySpecificIDE(t *testing.T) {
	installed := []IDE{
		{ID: "vscode", Name: "Visual Studio Code", Technologies: allTechnologies},
		{ID: "phpstorm", Name: "PhpStorm", Technologies: []project.Technology{project.TechnologyPHP}},
		{ID: "webstorm", Name: "WebStorm", Technologies: []project.Technology{project.TechnologyTypeScript}},
	}

	php := Rank(installed, project.TechnologyPHP)
	if php[0].ID != "phpstorm" || php[1].ID != "vscode" {
		t.Fatalf("PHP ranking = %q, want phpstorm then vscode", []string{php[0].ID, php[1].ID})
	}
	typescript := Rank(installed, project.TechnologyTypeScript)
	if typescript[0].ID != "webstorm" || typescript[1].ID != "vscode" {
		t.Fatalf("TypeScript ranking = %q, want webstorm then vscode", []string{typescript[0].ID, typescript[1].ID})
	}
}

func TestRankOmitsIncompatibleEditors(t *testing.T) {
	installed := []IDE{{ID: "webstorm", Name: "WebStorm", Technologies: []project.Technology{project.TechnologyTypeScript}}}
	if ranked := Rank(installed, project.TechnologyPHP); len(ranked) != 0 {
		t.Fatalf("PHP ranking = %+v, want no incompatible editors", ranked)
	}
}

func TestSupports(t *testing.T) {
	editor := IDE{Technologies: []project.Technology{project.TechnologyGo, project.TechnologyRust}}
	if !editor.Supports(project.TechnologyGo) {
		t.Fatal("Supports(Go) = false, want true")
	}
	if editor.Supports(project.TechnologyPHP) {
		t.Fatal("Supports(PHP) = true, want false")
	}
}
