package scanner

import (
	"os"
	"os/exec"
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
		{Name: "alpha", Path: filepath.Join(workspace, "alpha"), Technology: project.TechnologyGo, Markers: []string{"go.mod"}, PackageManagers: []project.PackageManager{project.PackageManagerGoModules}},
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

func TestScanPropagatesProjectMetadata(t *testing.T) {
	workspace := t.TempDir()
	dir := filepath.Join(workspace, "full-stack")
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	files := map[string]string{
		"composer.json":  `{"require":{"laravel/framework":"^12.0"}}`,
		"package.json":   `{"dependencies":{"vue":"latest"}}`,
		"pnpm-lock.yaml": "",
		"Dockerfile":     "FROM scratch",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	projects, err := Scan(workspace)
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}
	if len(projects) != 1 {
		t.Fatalf("len(Scan()) = %d, want 1", len(projects))
	}
	wantFrameworks := []project.Framework{project.FrameworkLaravel, project.FrameworkVue}
	if !reflect.DeepEqual(projects[0].Frameworks, wantFrameworks) {
		t.Fatalf("frameworks = %q, want %q", projects[0].Frameworks, wantFrameworks)
	}
	wantManagers := []project.PackageManager{project.PackageManagerComposer, project.PackageManagerPNPM}
	if !reflect.DeepEqual(projects[0].PackageManagers, wantManagers) {
		t.Fatalf("package managers = %q, want %q", projects[0].PackageManagers, wantManagers)
	}
	if !projects[0].HasDocker {
		t.Fatal("HasDocker = false, want true")
	}
}

func TestScanIncludesGitStatus(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}
	workspace := t.TempDir()
	dir := filepath.Join(workspace, "repository")
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	runScannerGit(t, dir, "init", "-b", "main")
	runScannerGit(t, dir, "config", "user.name", "ForgePath Tests")
	runScannerGit(t, dir, "config", "user.email", "forgepath@example.com")
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/app"), 0o644); err != nil {
		t.Fatal(err)
	}
	runScannerGit(t, dir, "add", "go.mod")
	runScannerGit(t, dir, "commit", "-m", "initial")
	if err := os.WriteFile(filepath.Join(dir, "change.txt"), nil, 0o644); err != nil {
		t.Fatal(err)
	}

	projects, err := Scan(workspace)
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}
	if len(projects) != 1 || projects[0].GitBranch != "main" || !projects[0].GitDirty || !projects[0].GitStatusKnown {
		t.Fatalf("Scan() = %+v, want dirty main repository", projects)
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

func runScannerGit(t *testing.T, dir string, arguments ...string) {
	t.Helper()
	args := append([]string{"-C", dir}, arguments...)
	if output, err := exec.Command("git", args...).CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", arguments, err, output)
	}
}
