package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	appconfig "github.com/daviPeter07/forgepath/internal/config"
	"github.com/daviPeter07/forgepath/internal/project"
)

func TestRunPlannerPrefersConfiguredCommand(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "Cargo.toml"), []byte("[package]"), 0o644); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "config.json")
	configuration := appconfig.Default()
	configuration.Projects["api"] = appconfig.ProjectConfig{Command: []string{"make", "dev"}}
	if err := appconfig.Save(path, configuration); err != nil {
		t.Fatal(err)
	}
	plan := runPlanner(func() (string, error) { return path, nil })

	configured, ok := plan(project.Project{Name: "api", Path: dir, Technology: project.TechnologyRust})
	if !ok || !configured.Configured || strings.Join(configured.Command, " ") != "make dev" {
		t.Fatalf("configured plan = %+v, %t", configured, ok)
	}
	detected, ok := plan(project.Project{Name: "other", Path: dir, Technology: project.TechnologyRust})
	if !ok || detected.Configured || strings.Join(detected.Command, " ") != "cargo run" {
		t.Fatalf("detected plan = %+v, %t", detected, ok)
	}
	if _, ok := plan(project.Project{Name: "empty", Path: t.TempDir(), Technology: project.TechnologyPython}); ok {
		t.Fatal("a project without a command should have no plan")
	}
	if fallback, ok := runPlanner(nil)(project.Project{Name: "api", Path: dir}); !ok || fallback.Configured {
		t.Fatalf("without config the plan should be detected: %+v, %t", fallback, ok)
	}
}
