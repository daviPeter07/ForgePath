package scanner

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/daviPeter07/forgepath/internal/project"
)

func TestScanFindsAndSortsProjects(t *testing.T) {
	workspace := t.TempDir()
	createProject(t, workspace, "zeta", "package.json")
	createProject(t, workspace, "alpha", "go.mod")
	createProject(t, workspace, "middle", "pyproject.toml")
	createProject(t, workspace, "unknown", "README.md")
	createProject(t, workspace, ".hidden", "go.mod")
	createProject(t, workspace, "node_modules", "go.mod")
	if err := os.WriteFile(filepath.Join(workspace, "regular-file"), nil, 0o644); err != nil {
		t.Fatal(err)
	}

	projects, err := Scan(workspace)
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	want := []project.Project{
		{Name: "alpha", Path: filepath.Join(workspace, "alpha"), Technology: project.TechnologyGo, Markers: []string{"go.mod"}},
		{Name: "middle", Path: filepath.Join(workspace, "middle"), Technology: project.TechnologyPython, Markers: []string{"pyproject.toml"}},
		{Name: "zeta", Path: filepath.Join(workspace, "zeta"), Technology: project.TechnologyJavaScript, Markers: []string{"package.json"}},
	}
	if !reflect.DeepEqual(projects, want) {
		t.Fatalf("Scan() = %+v, want %+v", projects, want)
	}
}

func TestScanEmptyWorkspace(t *testing.T) {
	projects, err := Scan(t.TempDir())
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}
	if projects == nil {
		t.Fatal("Scan() = nil, want empty slice")
	}
	if len(projects) != 0 {
		t.Fatalf("len(Scan()) = %d, want 0", len(projects))
	}
}

func TestScanRejectsInvalidWorkspace(t *testing.T) {
	t.Run("missing", func(t *testing.T) {
		_, err := Scan(filepath.Join(t.TempDir(), "missing"))
		if err == nil {
			t.Fatal("Scan() error = nil, want error")
		}
	})

	t.Run("file", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "workspace.txt")
		if err := os.WriteFile(path, nil, 0o644); err != nil {
			t.Fatal(err)
		}
		_, err := Scan(path)
		if err == nil {
			t.Fatal("Scan() error = nil, want error")
		}
	})
}

func createProject(t *testing.T, workspace, name string, markers ...string) {
	t.Helper()
	dir := filepath.Join(workspace, name)
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatalf("create project %q: %v", name, err)
	}
	for _, marker := range markers {
		if err := os.WriteFile(filepath.Join(dir, marker), nil, 0o644); err != nil {
			t.Fatalf("create marker %q: %v", marker, err)
		}
	}
}
