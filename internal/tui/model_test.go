package tui

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/daviPeter07/forgepath/internal/icon"
	"github.com/daviPeter07/forgepath/internal/ide"
	"github.com/daviPeter07/forgepath/internal/project"
)

func TestModelSelectsProject(t *testing.T) {
	model := testModel()

	updated, _ := model.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyDown}))
	model = updated.(Model)
	updated, command := model.Update(tea.KeyPressMsg(tea.Key{Text: "c", Code: 'c'}))
	model = updated.(Model)

	if command == nil {
		t.Fatal("select command = nil, want quit command")
	}
	selected, found := model.Selection()
	if !found {
		t.Fatal("Selection() found = false, want true")
	}
	if selected.Name != "web" {
		t.Fatalf("Selection().Name = %q, want web", selected.Name)
	}
}

func TestModelCancels(t *testing.T) {
	tests := []tea.Key{
		{Text: "q", Code: 'q'},
		{Code: tea.KeyEscape},
		{Code: 'c', Mod: tea.ModCtrl},
	}

	for _, pressed := range tests {
		model := testModel()
		updated, command := model.Update(tea.KeyPressMsg(pressed))
		model = updated.(Model)

		if command == nil {
			t.Fatalf("cancel command for %q = nil, want quit command", pressed.String())
		}
		if !model.Cancelled() {
			t.Fatalf("Cancelled() for %q = false, want true", pressed.String())
		}
		if _, found := model.Selection(); found {
			t.Fatalf("Selection() for %q found = true, want false", pressed.String())
		}
	}
}

func TestModelFiltersProjects(t *testing.T) {
	model := testModel()
	model.list.SetFilterText("web")

	items := model.list.VisibleItems()
	if len(items) != 1 {
		t.Fatalf("len(VisibleItems()) = %d, want 1", len(items))
	}
	if items[0].(projectItem).project.Name != "web" {
		t.Fatalf("filtered project = %q, want web", items[0].(projectItem).project.Name)
	}
}

func TestSelectFiltersAndSelectsProject(t *testing.T) {
	workspace := t.TempDir()
	apiPath := filepath.Join(workspace, "api")
	webPath := filepath.Join(workspace, "web")
	if err := os.Mkdir(apiPath, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(webPath, 0o755); err != nil {
		t.Fatal(err)
	}
	projects := []project.Project{
		{Name: "api", Path: apiPath, Technology: project.TechnologyGo, Markers: []string{"go.mod"}},
		{Name: "web", Path: webPath, Technology: project.TechnologyTypeScript, Markers: []string{"package.json", "tsconfig.json"}},
	}
	input := strings.NewReader("/web\r\rc")

	selected, found, err := Select(context.Background(), projects, icon.ModeASCII, input, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("Select() error = %v", err)
	}
	if !found {
		t.Fatal("Select() found = false, want true")
	}
	if selected.Name != "web" {
		t.Fatalf("Select().Name = %q, want web", selected.Name)
	}
}

func TestSelectHonorsCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, _, err := Select(ctx, []project.Project{{Name: "api"}}, icon.ModeASCII, strings.NewReader(""), &bytes.Buffer{})
	if err == nil {
		t.Fatal("Select() error = nil, want context cancellation error")
	}
}

func TestModelHandlesWindowResize(t *testing.T) {
	model := testModel()
	updated, _ := model.Update(tea.WindowSizeMsg{Width: 100, Height: 36})
	model = updated.(Model)

	if model.list.Width() != 100 || model.list.Height() != 36 {
		t.Fatalf("list size = %dx%d, want 100x36", model.list.Width(), model.list.Height())
	}
}

func TestModelViewIncludesHelp(t *testing.T) {
	view := testModel().View().Content
	if !strings.Contains(view, "filter") || !strings.Contains(view, "browse") || !strings.Contains(view, "cd here") {
		t.Fatalf("View() missing filter/browser help: %q", view)
	}
}

func TestModelBrowsesDirectoriesWithoutQuitting(t *testing.T) {
	projectRoot := t.TempDir()
	source := filepath.Join(projectRoot, "src")
	nested := filepath.Join(source, "internal")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	model := NewModel([]project.Project{{Name: "app", Path: projectRoot, Technology: project.TechnologyGo}}, icon.ModeASCII)

	updated, command := model.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
	model = updated.(Model)
	if model.hasSelection || model.mode != directoryScreen || !samePath(model.currentPath, projectRoot) {
		t.Fatalf("enter project mode/path/command = %v, %q, %v", model.mode, model.currentPath, command)
	}
	updated, command = model.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
	model = updated.(Model)
	if model.hasSelection || !samePath(model.currentPath, source) {
		t.Fatalf("enter directory path/command = %q, %v", model.currentPath, command)
	}
	updated, _ = model.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyBackspace}))
	model = updated.(Model)
	if !samePath(model.currentPath, projectRoot) {
		t.Fatalf("back directory = %q, want %q", model.currentPath, projectRoot)
	}
}

func TestModelSuggestsInstalledIDEAndReturnsToBrowser(t *testing.T) {
	projectRoot := t.TempDir()
	installed := []ide.IDE{
		{ID: "vscode", Name: "Visual Studio Code", Executable: "code", Technologies: []project.Technology{project.TechnologyPHP}},
		{ID: "phpstorm", Name: "PhpStorm", Executable: "phpstorm", Technologies: []project.Technology{project.TechnologyPHP}},
	}
	var openedPath, openedEditor string
	model := NewModelWithOptions([]project.Project{{Name: "app", Path: projectRoot, Technology: project.TechnologyPHP}}, Options{
		Icons: icon.ModeASCII,
		IDEs:  installed,
		OpenEditor: func(_ context.Context, path string, _ project.Project, editor ide.IDE) error {
			openedPath = path
			openedEditor = editor.ID
			return nil
		},
	})

	updated, _ := model.Update(tea.KeyPressMsg(tea.Key{Text: "o", Code: 'o'}))
	model = updated.(Model)
	if model.mode != editorScreen {
		t.Fatalf("mode = %v, want editor screen", model.mode)
	}
	first, ok := model.list.SelectedItem().(editorItem)
	if !ok || first.editor.ID != "phpstorm" || !first.recommended {
		t.Fatalf("first editor = %+v, want recommended PhpStorm", first)
	}
	updated, command := model.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
	model = updated.(Model)
	if command == nil {
		t.Fatal("open editor command = nil")
	}
	message := command()
	updated, _ = model.Update(message)
	model = updated.(Model)
	if openedPath != projectRoot || openedEditor != "phpstorm" {
		t.Fatalf("opened path/editor = %q/%q", openedPath, openedEditor)
	}
	if model.mode != projectScreen {
		t.Fatalf("mode after opening = %v, want projects", model.mode)
	}
}

func TestModelIgnoresStaleEditorCompletion(t *testing.T) {
	projectRoot := t.TempDir()
	installed := []ide.IDE{{ID: "vscode", Name: "Visual Studio Code", Executable: "code", Technologies: []project.Technology{project.TechnologyGo}}}
	model := NewModelWithOptions([]project.Project{{Name: "app", Path: projectRoot, Technology: project.TechnologyGo}}, Options{
		Icons: icon.ModeASCII,
		IDEs:  installed,
		OpenEditor: func(_ context.Context, _ string, _ project.Project, _ ide.IDE) error {
			return nil
		},
	})
	updated, _ := model.Update(tea.KeyPressMsg(tea.Key{Text: "o", Code: 'o'}))
	model = updated.(Model)
	updated, command := model.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyBackspace}))
	model = updated.(Model)
	if model.mode != projectScreen {
		t.Fatalf("mode after leaving editor chooser = %v, want projects", model.mode)
	}
	updated, _ = model.Update(command())
	model = updated.(Model)
	if model.mode != projectScreen {
		t.Fatalf("stale completion changed mode to %v", model.mode)
	}
}

func TestModelCanConfirmPathFromEditorChooser(t *testing.T) {
	projectRoot := t.TempDir()
	model := NewModelWithOptions([]project.Project{{Name: "app", Path: projectRoot, Technology: project.TechnologyGo}}, Options{
		Icons: icon.ModeASCII,
		IDEs:  []ide.IDE{{ID: "vscode", Name: "Visual Studio Code", Executable: "code", Technologies: []project.Technology{project.TechnologyGo}}},
	})
	updated, _ := model.Update(tea.KeyPressMsg(tea.Key{Text: "o", Code: 'o'}))
	model = updated.(Model)
	updated, command := model.Update(tea.KeyPressMsg(tea.Key{Text: "c", Code: 'c'}))
	model = updated.(Model)
	if command == nil {
		t.Fatal("confirm command = nil, want quit")
	}
	selected, found := model.Selection()
	if !found || selected.Path != projectRoot {
		t.Fatalf("Selection() = %+v, %t, want project root", selected, found)
	}
}

func TestSafeTerminalTextReplacesControlCharacters(t *testing.T) {
	got := safeTerminalText("folder\x1b]52;c;payload\a\nname")
	if strings.ContainsAny(got, "\x1b\a\n") {
		t.Fatalf("safeTerminalText() retained control characters: %q", got)
	}
}

func TestPortableViewDoesNotRequireNerdFont(t *testing.T) {
	view := testModel().View().Content
	if !strings.Contains(view, "[GO]") {
		t.Fatalf("View() missing portable Go badge: %q", view)
	}
	if strings.Contains(view, icon.Label(project.TechnologyGo, icon.ModeNerdFont)) {
		t.Fatalf("View() unexpectedly contains a Nerd Font glyph: %q", view)
	}
}

func TestModelOnlyRendersProjectsVisibleInViewport(t *testing.T) {
	projects := make([]project.Project, 40)
	for index := range projects {
		projects[index] = project.Project{
			Name:       fmt.Sprintf("project-%02d", index),
			Path:       filepath.Join("workspace", fmt.Sprintf("project-%02d", index)),
			Technology: project.TechnologyGo,
		}
	}
	model := NewModel(projects, icon.ModeASCII)
	updated, _ := model.Update(tea.WindowSizeMsg{Width: 80, Height: 12})
	view := updated.(Model).View().Content
	if strings.Contains(view, "project-39") {
		t.Fatalf("View() rendered a project outside the viewport: %q", view)
	}
}

func TestProjectDelegateKeepsFixedHeightForLongContent(t *testing.T) {
	item := projectItem{project: project.Project{
		Name:       strings.Repeat("very-long-project-name-", 5),
		Path:       filepath.Join("workspace-with-a-long-name", "project"),
		Technology: project.TechnologyTypeScript,
		Frameworks: []project.Framework{project.FrameworkNextJS, project.FrameworkReact, project.FrameworkVue},
		GitBranch:  strings.Repeat("feature/long-branch-", 4),
	}, icons: icon.ModeNerdFont}
	model := list.New([]list.Item{item}, projectDelegate{}, 28, 10)
	var output bytes.Buffer
	(projectDelegate{}).Render(&output, model, 0, item)

	if height := lipgloss.Height(output.String()); height > 2 {
		t.Fatalf("rendered item height = %d, want at most 2: %q", height, output.String())
	}
}

func TestProjectItemDescriptionIncludesMetadata(t *testing.T) {
	item := projectItem{project: project.Project{
		Technology:      project.TechnologyPHP,
		Frameworks:      []project.Framework{project.FrameworkLaravel, project.FrameworkVue},
		PackageManagers: []project.PackageManager{project.PackageManagerComposer, project.PackageManagerPNPM},
		HasDocker:       true,
		GitBranch:       "main",
		GitDirty:        true,
		GitStatusKnown:  true,
	}}

	want := "PHP | Laravel | Vue.js | Composer | pnpm | Docker | main*"
	if item.Description() != want {
		t.Fatalf("Description() = %q, want %q", item.Description(), want)
	}
}

func TestProjectItemTitleUsesConfiguredIcons(t *testing.T) {
	project := project.Project{Name: "forgepath", Technology: project.TechnologyGo}
	ascii := projectItem{project: project, icons: icon.ModeASCII}.Title()
	nerdFont := projectItem{project: project, icons: icon.ModeNerdFont}.Title()

	if ascii != "[GO] forgepath" {
		t.Fatalf("ASCII Title() = %q, want %q", ascii, "[GO] forgepath")
	}
	if nerdFont == ascii || !strings.HasSuffix(nerdFont, " forgepath") {
		t.Fatalf("Nerd Font Title() = %q, want distinct icon", nerdFont)
	}
}

func TestFavoriteProjectTitle(t *testing.T) {
	favorite := project.Project{Name: "forgepath", Technology: project.TechnologyGo, Favorite: true}

	ascii := projectItem{project: favorite, icons: icon.ModeASCII}.Title()
	if ascii != "[F] [GO] forgepath" {
		t.Fatalf("ASCII favorite title = %q", ascii)
	}
	nerd := projectItem{project: favorite, icons: icon.ModeNerdFont}.Title()
	if !strings.HasPrefix(nerd, " ") {
		t.Fatalf("Nerd Font favorite title = %q", nerd)
	}
}

func TestProjectItemDescriptionMarksUnknownGitStatus(t *testing.T) {
	item := projectItem{project: project.Project{
		Technology: project.TechnologyGo,
		GitBranch:  "main",
	}}

	if item.Description() != "Go | main?" {
		t.Fatalf("Description() = %q, want %q", item.Description(), "Go | main?")
	}
}

func TestDockerProjectDescriptionDoesNotRepeatDocker(t *testing.T) {
	item := projectItem{project: project.Project{
		Technology: project.TechnologyDocker,
		HasDocker:  true,
	}}

	if item.Description() != "Docker" {
		t.Fatalf("Description() = %q, want Docker", item.Description())
	}
}

func testModel() Model {
	return NewModel([]project.Project{
		{Name: "api", Technology: project.TechnologyGo, Markers: []string{"go.mod"}},
		{Name: "web", Technology: project.TechnologyTypeScript, Markers: []string{"package.json", "tsconfig.json"}},
	}, icon.ModeASCII)
}
