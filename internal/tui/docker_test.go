package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/daviPeter07/forgepath/internal/icon"
	"github.com/daviPeter07/forgepath/internal/project"
)

func dockerModel(t *testing.T) (string, Model) {
	t.Helper()
	root := t.TempDir()
	model := NewModel([]project.Project{{Name: "api", Path: root, Technology: project.TechnologyGo}}, icon.ModeASCII)
	return root, model
}

func TestDockerScreenTargetsSelectedProject(t *testing.T) {
	root, model := dockerModel(t)
	model = press(t, model, letterKey('d'))

	if model.mode != dockerScreen {
		t.Fatalf("mode = %v, want docker screen", model.mode)
	}
	if !strings.Contains(model.list.Title, "api") {
		t.Fatalf("title = %q, want the target project name", model.list.Title)
	}
	items := model.list.Items()
	if len(items) == 0 {
		t.Fatal("docker screen has no options")
	}
	for _, raw := range items {
		item, ok := raw.(dockerItem)
		if !ok {
			t.Fatalf("item %T, want dockerItem", raw)
		}
		if item.projectPath != root {
			t.Fatalf("option %q targets %q, want %q", item.label, item.projectPath, root)
		}
	}
	if view := ansi.Strip(model.View().Content); !strings.Contains(view, "DOCKER") {
		t.Fatalf("view = %q, want docker badges", view)
	}
}

func TestDockerGenerationWritesComposeAndReturns(t *testing.T) {
	root, model := dockerModel(t)
	model = press(t, model, letterKey('d'), enterKey)

	if model.mode != projectScreen {
		t.Fatalf("mode after generating = %v, want project list", model.mode)
	}
	content, err := os.ReadFile(filepath.Join(root, "docker-compose.yml"))
	if err != nil {
		t.Fatalf("compose file was not written: %v", err)
	}
	if !strings.Contains(string(content), "services:") {
		t.Fatalf("compose content = %q", content)
	}
	selected, ok := model.list.SelectedItem().(projectItem)
	if !ok || !selected.project.HasDocker {
		t.Fatalf("project after generation = %+v, want HasDocker", model.list.SelectedItem())
	}
	if status := ansi.Strip(model.View().Content); !strings.Contains(status, "Generated docker-compose.yml") {
		t.Fatalf("view = %q, want success status", status)
	}
}

func TestDockerGenerationNeverReplacesExistingComposeFiles(t *testing.T) {
	for _, existing := range []string{"docker-compose.yml", "docker-compose.yaml", "compose.yaml", "compose.yml"} {
		t.Run(existing, func(t *testing.T) {
			root, model := dockerModel(t)
			writeFile(t, filepath.Join(root, existing), "original")

			model = press(t, model, letterKey('d'), enterKey)

			content, err := os.ReadFile(filepath.Join(root, existing))
			if err != nil || string(content) != "original" {
				t.Fatalf("%s = %q, %v; want untouched", existing, content, err)
			}
			if existing != "docker-compose.yml" {
				if _, err := os.Stat(filepath.Join(root, "docker-compose.yml")); err == nil {
					t.Fatalf("docker-compose.yml was created next to %s", existing)
				}
			}
			if model.mode != dockerScreen {
				t.Fatalf("mode = %v, want to stay on the docker screen", model.mode)
			}
			if status := ansi.Strip(model.View().Content); !strings.Contains(status, existing+" already exists") {
				t.Fatalf("view = %q, want already-exists error", status)
			}
		})
	}
}

func TestDockerTemplatesAreValidModernCompose(t *testing.T) {
	for _, option := range generateDockerOptions("/project", project.TechnologyGo) {
		t.Run(option.label, func(t *testing.T) {
			if strings.Contains(option.compose, "version:") {
				t.Fatalf("template uses the obsolete top-level version key:\n%s", option.compose)
			}
			if !strings.HasPrefix(option.compose, "services:\n") {
				t.Fatalf("template must start with services:\n%s", option.compose)
			}
			if strings.Contains(option.compose, "\t") {
				t.Fatal("YAML template contains tabs")
			}
			for _, line := range strings.Split(option.compose, "\n") {
				image, isImage := strings.CutPrefix(strings.TrimSpace(line), "image:")
				if isImage && !strings.Contains(image, ":") {
					t.Fatalf("image without a pinned tag: %q", line)
				}
			}
		})
	}
}

func TestDockerScreenRequiresATarget(t *testing.T) {
	model := NewModel(nil, icon.ModeASCII)
	model = press(t, model, letterKey('d'))
	if model.mode != projectScreen {
		t.Fatalf("mode = %v, want project list when nothing is selected", model.mode)
	}
}

func TestBackFromDockerScreenRestoresDirectory(t *testing.T) {
	root, _, model := monorepo(t)
	model = press(t, model, enterKey)
	model = selectItem(t, model, "docs")
	model = press(t, model, letterKey('d'))
	if model.mode != dockerScreen {
		t.Fatalf("mode = %v, want docker screen", model.mode)
	}
	model = press(t, model, backspaceKey)
	if model.mode != directoryScreen || !samePath(model.currentPath, root) {
		t.Fatalf("mode/path = %v/%q, want directory %q", model.mode, model.currentPath, root)
	}
}
