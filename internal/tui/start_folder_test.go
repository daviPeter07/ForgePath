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

// startFolderModel opens the browser in <tmp>/home/work, with a catalog
// project elsewhere.
func startFolderModel(t *testing.T) (home, work string, model Model) {
	t.Helper()
	home = filepath.Join(t.TempDir(), "home")
	work = filepath.Join(home, "work")
	for _, dir := range []string{filepath.Join(work, "notes"), filepath.Join(home, "music")} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	writeFile(t, filepath.Join(work, "api", "go.mod"), "module api\n")
	model = NewModelWithOptions([]project.Project{{Name: "catalog-app", Path: "/srv/catalog-app", Technology: project.TechnologyGo}},
		Options{Icons: icon.ModeASCII, StartPath: work})
	model.home = home
	return home, work, model
}

func TestStartPathOpensFolderEvenWithCatalogProjects(t *testing.T) {
	_, work, model := startFolderModel(t)
	if model.mode != directoryScreen || !samePath(model.currentPath, work) {
		t.Fatalf("mode/path = %v/%q, want the start folder", model.mode, model.currentPath)
	}
	kinds := map[string]bool{}
	for _, item := range model.list.Items() {
		switch item := item.(type) {
		case projectItem:
			kinds[item.project.Name] = true
		case directoryItem:
			kinds[item.directory.Name] = false
		}
	}
	if isProject, ok := kinds["api"]; !ok || !isProject {
		t.Fatalf("items = %v, want api detected as a project", kinds)
	}
	if isProject, ok := kinds["notes"]; !ok || isProject {
		t.Fatalf("items = %v, want notes as a folder", kinds)
	}
}

func TestBackClimbsAboveTheStartFolder(t *testing.T) {
	home, _, model := startFolderModel(t)
	model = press(t, model, backspaceKey)
	if model.mode != directoryScreen || !samePath(model.currentPath, home) {
		t.Fatalf("back from the start folder = %v/%q, want its parent %q", model.mode, model.currentPath, home)
	}
	if title := ansi.Strip(model.list.Title); !strings.Contains(title, "~") {
		t.Fatalf("title = %q, want the folder shown relative to home", title)
	}
	model = selectItem(t, model, "music")
	model = press(t, model, enterKey)
	if !samePath(model.currentPath, filepath.Join(home, "music")) {
		t.Fatalf("path = %q, want to enter a sibling folder", model.currentPath)
	}
}

func TestPTogglesBetweenFolderAndProjects(t *testing.T) {
	_, work, model := startFolderModel(t)
	model = press(t, model, enterKey) // into api (a project)
	api := model.currentPath

	if footer := ansi.Strip(model.footerView()); !strings.Contains(footer, "p projects") {
		t.Fatalf("folder footer = %q, want p projects", footer)
	}
	model = press(t, model, letterKey('p'))
	if model.mode != projectScreen || strings.Join(visibleProjectNames(model), ",") != "catalog-app" {
		t.Fatalf("p did not open the project list: mode %v, %v", model.mode, visibleProjectNames(model))
	}
	if footer := ansi.Strip(model.footerView()); !strings.Contains(footer, "p back to folder") {
		t.Fatalf("project footer = %q, want the way back", footer)
	}

	model = press(t, model, letterKey('p'))
	if model.mode != directoryScreen || !samePath(model.currentPath, api) {
		t.Fatalf("p did not return to %q: %v/%q", api, model.mode, model.currentPath)
	}
	// The browsing context survives the round trip: back leaves api for work.
	model = press(t, model, backspaceKey)
	if !samePath(model.currentPath, work) {
		t.Fatalf("back = %q, want %q", model.currentPath, work)
	}
}

func TestPWithoutFolderHistoryDoesNothing(t *testing.T) {
	model := testModel()
	model = press(t, model, letterKey('p'))
	if model.mode != projectScreen {
		t.Fatalf("mode = %v, want to stay on the project list", model.mode)
	}
	if footer := ansi.Strip(model.footerView()); strings.Contains(footer, "back to folder") {
		t.Fatalf("footer = %q, there is no folder to return to", footer)
	}
}
