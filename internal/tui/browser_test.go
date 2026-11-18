package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/daviPeter07/forgepath/internal/icon"
	"github.com/daviPeter07/forgepath/internal/ide"
	"github.com/daviPeter07/forgepath/internal/project"
)

var (
	enterKey     = tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter})
	downKey      = tea.KeyPressMsg(tea.Key{Code: tea.KeyDown})
	backspaceKey = tea.KeyPressMsg(tea.Key{Code: tea.KeyBackspace})
	escapeKey    = tea.KeyPressMsg(tea.Key{Code: tea.KeyEscape})
)

func letterKey(letter rune) tea.KeyPressMsg {
	return tea.KeyPressMsg(tea.Key{Text: string(letter), Code: letter})
}

func press(t *testing.T, model Model, keys ...tea.KeyPressMsg) Model {
	t.Helper()
	for _, pressed := range keys {
		updated, _ := model.Update(pressed)
		model = updated.(Model)
	}
	return model
}

// selectItem moves the cursor to the list item whose FilterValue starts with name.
func selectItem(t *testing.T, model Model, name string) Model {
	t.Helper()
	for index, item := range model.list.Items() {
		if strings.HasPrefix(item.FilterValue(), name) {
			model.list.Select(index)
			return model
		}
	}
	t.Fatalf("item %q not found in %d items", name, len(model.list.Items()))
	return model
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// monorepo creates app/ with a nested JavaScript project in app/web and a plain app/docs directory.
func monorepo(t *testing.T) (root, web string, model Model) {
	t.Helper()
	root = filepath.Join(t.TempDir(), "app")
	web = filepath.Join(root, "web")
	writeFile(t, filepath.Join(root, "go.mod"), "module example.com/app\n")
	writeFile(t, filepath.Join(web, "package.json"), "{}")
	if err := os.MkdirAll(filepath.Join(root, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}
	model = NewModel([]project.Project{{Name: "app", Path: root, Technology: project.TechnologyGo}}, icon.ModeASCII)
	return root, web, model
}

func TestBackFromNestedProjectReturnsToParentDirectory(t *testing.T) {
	root, web, model := monorepo(t)

	model = press(t, model, enterKey)
	model = selectItem(t, model, "web")
	if _, ok := model.list.SelectedItem().(projectItem); !ok {
		t.Fatalf("web should be detected as a nested project, got %T", model.list.SelectedItem())
	}
	model = press(t, model, enterKey)
	if !samePath(model.currentPath, web) {
		t.Fatalf("currentPath = %q, want %q", model.currentPath, web)
	}

	model = press(t, model, backspaceKey)
	if model.mode != directoryScreen || !samePath(model.currentPath, root) {
		t.Fatalf("after back: mode=%v path=%q, want directory screen at %q", model.mode, model.currentPath, root)
	}
	if !samePath(model.currentProject.Path, root) {
		t.Fatalf("currentProject = %q, want the enclosing project %q", model.currentProject.Path, root)
	}

	model = press(t, model, backspaceKey)
	if model.mode != projectScreen {
		t.Fatalf("second back mode = %v, want project list", model.mode)
	}
}

func TestConfirmInDirectoryUsesDisplayedDirectory(t *testing.T) {
	root, _, model := monorepo(t)

	model = press(t, model, enterKey)
	model = selectItem(t, model, "web") // highlight the nested project, but stay in app/
	updated, command := model.Update(letterKey('c'))
	model = updated.(Model)

	if command == nil {
		t.Fatal("confirm command = nil, want quit")
	}
	selected, found := model.Selection()
	if !found || !samePath(selected.Path, root) {
		t.Fatalf("Selection() = %q, %t, want displayed directory %q", selected.Path, found, root)
	}
}

func TestTitleShowsBreadcrumbAndStaysConsistent(t *testing.T) {
	root, _, model := monorepo(t)
	if err := os.MkdirAll(filepath.Join(root, "docs", "guides"), 0o755); err != nil {
		t.Fatal(err)
	}
	initial := model.list.Title

	model = press(t, model, enterKey)
	model = selectItem(t, model, "docs")
	model = press(t, model, enterKey)
	model = selectItem(t, model, "guides")
	model = press(t, model, enterKey)
	if title := model.list.Title; !strings.Contains(title, "app › docs › guides") {
		t.Fatalf("title = %q, want breadcrumb app › docs › guides", title)
	}

	model = press(t, model, backspaceKey, backspaceKey, backspaceKey)
	if model.mode != projectScreen {
		t.Fatalf("mode = %v, want project list", model.mode)
	}
	if model.list.Title != initial {
		t.Fatalf("title after returning = %q, want initial %q", model.list.Title, initial)
	}
}

func TestHelpOnlyShowsKeysThatWorkOnTheScreen(t *testing.T) {
	helpText := func(model Model) string { return ansi.Strip(model.footerView()) }
	projectRoot := t.TempDir()
	model := NewModelWithOptions([]project.Project{{Name: "app", Path: projectRoot, Technology: project.TechnologyGo}}, Options{
		Icons: icon.ModeASCII,
		IDEs:  []ide.IDE{{ID: "zed", Name: "Zed", Executable: "zed", Technologies: []project.Technology{project.TechnologyGo}}},
	})

	projects := helpText(model)
	if strings.Contains(projects, "back") {
		t.Fatalf("project list help = %q, back does nothing there", projects)
	}
	if !strings.Contains(projects, "open IDE") || !strings.Contains(projects, "compose") {
		t.Fatalf("project list help = %q, want IDE and compose actions", projects)
	}

	editors := helpText(press(t, model, letterKey('o')))
	if !strings.Contains(editors, "enter open") || strings.Contains(editors, "compose") || !strings.Contains(editors, "back") {
		t.Fatalf("IDE chooser help = %q", editors)
	}

	compose := helpText(press(t, model, letterKey('d')))
	if !strings.Contains(compose, "enter generate") || strings.Contains(compose, "open IDE") || strings.Contains(compose, "cd here") {
		t.Fatalf("compose help = %q", compose)
	}

	directory := helpText(press(t, model, enterKey))
	if !strings.Contains(directory, "back") || !strings.Contains(directory, "cd here") {
		t.Fatalf("directory help = %q", directory)
	}
}

func TestEscapeReturnsFromSubScreens(t *testing.T) {
	_, _, model := monorepo(t)
	model = press(t, model, enterKey, escapeKey)
	if model.mode != projectScreen || model.Cancelled() {
		t.Fatalf("esc from directory: mode=%v cancelled=%t", model.mode, model.Cancelled())
	}
}

func TestOpenIDEWithoutInstalledEditorsShowsError(t *testing.T) {
	model := NewModel([]project.Project{{Name: "app", Path: t.TempDir(), Technology: project.TechnologyGo}}, icon.ModeASCII)
	updated, command := model.Update(letterKey('o'))
	model = updated.(Model)
	if model.mode != projectScreen {
		t.Fatalf("mode = %v, want to stay on project list", model.mode)
	}
	if command == nil {
		t.Fatal("command = nil, want status message")
	}
	if status := ansi.Strip(model.list.View()); !strings.Contains(status, "no supported IDE") {
		t.Fatalf("view = %q, want error status", status)
	}
}

func TestStartPathBrowsesDirectoryWhenNoProjectsWereFound(t *testing.T) {
	workspace := t.TempDir()
	writeFile(t, filepath.Join(workspace, "service", "go.mod"), "module service\n")
	if err := os.MkdirAll(filepath.Join(workspace, "notes"), 0o755); err != nil {
		t.Fatal(err)
	}

	model := NewModelWithOptions(nil, Options{Icons: icon.ModeASCII, StartPath: workspace})
	if model.mode != directoryScreen || !samePath(model.currentPath, workspace) {
		t.Fatalf("mode/path = %v/%q, want directory screen at workspace", model.mode, model.currentPath)
	}
	kinds := map[string]string{}
	for _, item := range model.list.Items() {
		switch item := item.(type) {
		case projectItem:
			kinds[item.project.Name] = "project"
		case directoryItem:
			kinds[item.directory.Name] = "directory"
		}
	}
	if kinds["service"] != "project" || kinds["notes"] != "directory" {
		t.Fatalf("items = %v, want service as project and notes as directory", kinds)
	}
}

func TestReadDirectoriesSkipsFilesAndGeneratedFolders(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"src", "node_modules", ".git", "__pycache__"} {
		if err := os.MkdirAll(filepath.Join(root, name), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	writeFile(t, filepath.Join(root, "README.md"), "")

	directories, err := readDirectories(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(directories) != 1 || directories[0].Name != "src" || directories[0].Path != filepath.Join(root, "src") {
		t.Fatalf("readDirectories() = %+v, want only src", directories)
	}
	if _, err := readDirectories(filepath.Join(root, "missing")); err == nil {
		t.Fatal("readDirectories(missing) error = nil")
	}
}

func TestFavoriteStarIsHighlighted(t *testing.T) {
	item := projectItem{
		project: project.Project{Name: "forgepath", Path: "/work/forgepath", Technology: project.TechnologyGo, Favorite: true},
		icons:   icon.ModeGraphics,
	}
	for _, compact := range []bool{false, true} {
		delegate := projectDelegate{compact: compact}
		model := list.New([]list.Item{testItem("selected"), item}, delegate, 80, 20)
		var output strings.Builder
		delegate.Render(&output, model, 1, item)
		if !strings.Contains(output.String(), "★") || !strings.Contains(output.String(), "250;204;21") {
			t.Fatalf("compact=%t: favorite star missing or not colored: %q", compact, output.String())
		}
	}
}

type testItem string

func (item testItem) FilterValue() string { return string(item) }

func TestBreadcrumb(t *testing.T) {
	root := filepath.Join(string(filepath.Separator), "work", "app")
	long := filepath.Join(string(filepath.Separator), strings.Repeat("deep", 20), "folder")
	tests := []struct {
		name, base, path, want string
	}{
		{"project root", root, root, "app"},
		{"nested", root, filepath.Join(root, "src", "api"), "app › src › api"},
		{"deep path keeps root and tail", root, filepath.Join(root, "a", "b", "c", "d"), "app › … › c › d"},
		{"outside root shows path", root, filepath.Join(string(filepath.Separator), "work"), filepath.Join(string(filepath.Separator), "work")},
		{"no root", "", root, root},
		{"long path keeps the end", "", long, "…" + long[len(long)-maxCrumbWidth+1:]},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := breadcrumb(tt.base, tt.path); got != tt.want {
				t.Fatalf("breadcrumb(%q, %q) = %q, want %q", tt.base, tt.path, got, tt.want)
			}
		})
	}
}
