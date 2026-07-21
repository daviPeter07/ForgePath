package tui

import (
	"bytes"
	"context"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/daviPeter07/forgepath/internal/project"
)

func TestModelSelectsProject(t *testing.T) {
	model := testModel()

	updated, _ := model.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyDown}))
	model = updated.(Model)
	updated, command := model.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter}))
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
	projects := []project.Project{
		{Name: "api", Technology: project.TechnologyGo, Markers: []string{"go.mod"}},
		{Name: "web", Technology: project.TechnologyTypeScript, Markers: []string{"package.json", "tsconfig.json"}},
	}
	input := strings.NewReader("/web\r\r")

	selected, found, err := Select(context.Background(), projects, input, &bytes.Buffer{})
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

	_, _, err := Select(ctx, []project.Project{{Name: "api"}}, strings.NewReader(""), &bytes.Buffer{})
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
	if !strings.Contains(view, "filter") || !strings.Contains(view, "select") {
		t.Fatalf("View() missing filter/select help: %q", view)
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
	})
}
