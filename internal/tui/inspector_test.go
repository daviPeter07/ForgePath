package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/daviPeter07/forgepath/internal/icon"
	"github.com/daviPeter07/forgepath/internal/ide"
	"github.com/daviPeter07/forgepath/internal/project"
)

func inspectorProjects(root string) []project.Project {
	return []project.Project{
		{
			Name: "api", Path: filepath.Join(root, "api"), Technology: project.TechnologyGo,
			Markers: []string{"go.mod"}, PackageManagers: []project.PackageManager{project.PackageManagerGoModules},
			GitBranch: "main", GitDirty: true, GitStatusKnown: true, LastOpened: fixedNow.Add(-2 * time.Hour),
		},
		{
			Name: "web", Path: filepath.Join(root, "web"), Technology: project.TechnologyTypeScript,
			Markers: []string{"package.json", "tsconfig.json"}, Frameworks: []project.Framework{project.FrameworkNextJS},
			PackageManagers: []project.PackageManager{project.PackageManagerPNPM},
			GitBranch:       "feat/ui", GitStatusKnown: true,
		},
	}
}

func sizedModel(t *testing.T, projects []project.Project, options Options, width, height int) Model {
	t.Helper()
	if options.Now == nil {
		options.Now = func() time.Time { return fixedNow }
	}
	model := NewModelWithOptions(projects, options)
	updated, _ := model.Update(tea.WindowSizeMsg{Width: width, Height: height})
	return updated.(Model)
}

func assertFitsWindow(t *testing.T, view string, width, height int) {
	t.Helper()
	lines := strings.Split(view, "\n")
	if len(lines) != height {
		t.Fatalf("view has %d lines, want %d", len(lines), height)
	}
	for index, line := range lines {
		if got := ansi.StringWidth(line); got > width {
			t.Fatalf("line %d is %d columns wide, want at most %d: %q", index, got, width, ansi.Strip(line))
		}
	}
}

func TestWideTerminalShowsInspectorForHighlightedProject(t *testing.T) {
	root := t.TempDir()
	model := sizedModel(t, inspectorProjects(root), Options{
		Icons: icon.ModeASCII,
		IDEs:  []ide.IDE{{ID: "goland", Name: "GoLand", Executable: "goland", Technologies: []project.Technology{project.TechnologyGo}}},
	}, 120, 32)

	view := model.View().Content
	assertFitsWindow(t, view, 120, 32)
	plain := ansi.Strip(view)
	for _, want := range []string{"main ● uncommitted changes", "Go Modules", "2h ago", "GoLand", "go.mod", filepath.Base(root)} {
		if !strings.Contains(plain, want) {
			t.Fatalf("inspector is missing %q:\n%s", want, plain)
		}
	}

	model = press(t, model, downKey)
	plain = ansi.Strip(model.View().Content)
	for _, want := range []string{"feat/ui ✓ clean", "Next.js", "package.json, tsconfig.json", "not opened yet"} {
		if !strings.Contains(plain, want) {
			t.Fatalf("inspector did not follow the cursor, missing %q:\n%s", want, plain)
		}
	}
}

func TestNarrowTerminalHidesInspector(t *testing.T) {
	model := sizedModel(t, inspectorProjects(t.TempDir()), Options{Icons: icon.ModeASCII}, 90, 30)
	view := model.View().Content
	assertFitsWindow(t, view, 90, 30)
	if strings.Contains(ansi.Strip(view), "found by") {
		t.Fatalf("narrow view shows the inspector:\n%s", ansi.Strip(view))
	}
}

func TestInspectorDrawsLogoInGraphicsMode(t *testing.T) {
	graphics := sizedModel(t, inspectorProjects(t.TempDir()), Options{Icons: icon.ModeGraphics}, 120, 32)
	view := graphics.View().Content
	assertFitsWindow(t, view, 120, 32)
	if !strings.Contains(view, "▀") {
		t.Fatal("graphics inspector has no logo pixels")
	}

	ascii := sizedModel(t, inspectorProjects(t.TempDir()), Options{Icons: icon.ModeASCII}, 120, 32)
	if strings.Contains(ascii.View().Content, "▀") {
		t.Fatal("ASCII inspector drew logo pixels")
	}
}

func TestFolderInspectorListsContents(t *testing.T) {
	root, _, model := monorepo(t)
	writeFile(t, filepath.Join(root, "docs", "README.md"), "# docs")
	if err := os.MkdirAll(filepath.Join(root, "docs", "guides"), 0o755); err != nil {
		t.Fatal(err)
	}
	updated, _ := model.Update(tea.WindowSizeMsg{Width: 120, Height: 30})
	model = press(t, updated.(Model), enterKey)
	model = selectItem(t, model, "docs")

	plain := ansi.Strip(model.View().Content)
	guides, readme := strings.Index(plain, "guides/"), strings.Index(plain, "README.md")
	if guides < 0 || readme < 0 {
		t.Fatalf("folder inspector is missing entries:\n%s", plain)
	}
	if guides > readme {
		t.Fatal("folders should be listed before files")
	}
}

func TestDockerInspectorPreviewsTheTemplate(t *testing.T) {
	model := sizedModel(t, inspectorProjects(t.TempDir()), Options{Icons: icon.ModeASCII}, 120, 34)
	model = press(t, model, letterKey('d'))
	plain := ansi.Strip(model.View().Content)
	for _, want := range []string{"services:", "postgres:15-alpine", "docker-compose.yml"} {
		if !strings.Contains(plain, want) {
			t.Fatalf("docker inspector is missing %q:\n%s", want, plain)
		}
	}
}

func TestEditorInspectorShowsCommand(t *testing.T) {
	model := sizedModel(t, inspectorProjects(t.TempDir()), Options{
		Icons: icon.ModeASCII,
		IDEs:  []ide.IDE{{ID: "goland", Name: "GoLand", Executable: "/opt/goland/bin/goland", Technologies: []project.Technology{project.TechnologyGo}}},
	}, 120, 30)
	model = press(t, model, letterKey('o'))
	plain := ansi.Strip(model.View().Content)
	if !strings.Contains(plain, "/opt/goland/bin/goland") || !strings.Contains(plain, "Suggested for Go") {
		t.Fatalf("editor inspector is missing details:\n%s", plain)
	}
}

func TestShortenHome(t *testing.T) {
	tests := []struct{ path, home, want string }{
		{"/home/dev/projects/api", "/home/dev", "~/projects/api"},
		{"/home/dev", "/home/dev", "~"},
		{"/home/developer/x", "/home/dev", "/home/developer/x"},
		{"/srv/api", "", "/srv/api"},
	}
	for _, tt := range tests {
		if got := shortenHome(tt.path, tt.home); got != tt.want {
			t.Errorf("shortenHome(%q, %q) = %q, want %q", tt.path, tt.home, got, tt.want)
		}
	}
}
