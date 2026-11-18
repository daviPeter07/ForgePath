package tui

import (
	"errors"
	"strings"
	"testing"
	"time"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/daviPeter07/forgepath/internal/icon"
	"github.com/daviPeter07/forgepath/internal/project"
)

var fixedNow = time.Date(2026, time.November, 18, 15, 0, 0, 0, time.UTC)

func favoriteModel(t *testing.T, toggle func(project.Project) (bool, error)) Model {
	t.Helper()
	model := NewModelWithOptions([]project.Project{
		{Name: "api", Path: "/work/api", Technology: project.TechnologyGo},
		{Name: "web", Path: "/work/web", Technology: project.TechnologyTypeScript},
	}, Options{Icons: icon.ModeASCII, ToggleFavorite: toggle, Now: func() time.Time { return fixedNow }})
	updated, _ := model.Update(tea.WindowSizeMsg{Width: 96, Height: 30})
	return updated.(Model)
}

func visibleProjectNames(model Model) []string {
	var names []string
	for _, item := range model.list.VisibleItems() {
		if project, ok := item.(projectItem); ok {
			names = append(names, project.project.Name)
		}
	}
	return names
}

func TestFavoriteKeyPinsProjectToTheTop(t *testing.T) {
	var toggled []string
	model := favoriteModel(t, func(target project.Project) (bool, error) {
		toggled = append(toggled, target.Name)
		return !target.Favorite, nil
	})

	model = press(t, model, downKey, letterKey('f'))

	if strings.Join(toggled, ",") != "web" {
		t.Fatalf("toggled = %q, want web", toggled)
	}
	if names := strings.Join(visibleProjectNames(model), ","); names != "web,api" {
		t.Fatalf("order = %s, want favorite first", names)
	}
	selected, ok := model.list.SelectedItem().(projectItem)
	if !ok || selected.project.Name != "web" || !selected.project.Favorite {
		t.Fatalf("selected = %+v, want the cursor to follow the favorite", model.list.SelectedItem())
	}
	if view := ansi.Strip(model.View().Content); !strings.Contains(view, "web added to favorites") {
		t.Fatalf("view = %q, want confirmation", view)
	}

	model = press(t, model, letterKey('f'))
	if names := strings.Join(visibleProjectNames(model), ","); names != "api,web" {
		t.Fatalf("order after unfavorite = %s, want alphabetical", names)
	}
	if view := ansi.Strip(model.View().Content); !strings.Contains(view, "web removed from favorites") {
		t.Fatalf("view = %q, want removal confirmation", view)
	}
}

func TestFavoriteFailureKeepsListAndShowsError(t *testing.T) {
	model := favoriteModel(t, func(project.Project) (bool, error) { return false, errors.New("state is locked") })
	model = press(t, model, downKey, letterKey('f'))

	if names := strings.Join(visibleProjectNames(model), ","); names != "api,web" {
		t.Fatalf("order = %s, want unchanged", names)
	}
	if view := ansi.Strip(model.View().Content); !strings.Contains(view, "state is locked") {
		t.Fatalf("view = %q, want error", view)
	}
}

func TestFavoriteKeyIsIgnoredOutsideTheProjectList(t *testing.T) {
	_, _, model := monorepo(t)
	called := false
	model.options.ToggleFavorite = func(project.Project) (bool, error) { called = true; return true, nil }
	model = press(t, model, enterKey, letterKey('f'))
	if called {
		t.Fatal("f toggled a favorite from the directory browser")
	}
}

func TestRelativeTime(t *testing.T) {
	tests := []struct {
		ago  time.Duration
		want string
	}{
		{20 * time.Second, "just now"},
		{5 * time.Minute, "5m ago"},
		{3 * time.Hour, "3h ago"},
		{50 * time.Hour, "2d ago"},
		{15 * 24 * time.Hour, "2w ago"},
		{60 * 24 * time.Hour, "Sep 19"},
		{400 * 24 * time.Hour, "Oct 14, 2025"},
	}
	for _, tt := range tests {
		if got := relativeTime(fixedNow.Add(-tt.ago), fixedNow); got != tt.want {
			t.Errorf("relativeTime(-%s) = %q, want %q", tt.ago, got, tt.want)
		}
	}
	if got := relativeTime(time.Time{}, fixedNow); got != "" {
		t.Errorf("relativeTime(zero) = %q, want empty", got)
	}
}

func TestSubtitleShowsRecencyAndHighlightsChanges(t *testing.T) {
	item := projectItem{
		project: project.Project{
			Name: "api", Path: "/work/api", Technology: project.TechnologyGo,
			GitBranch: "main", GitDirty: true, GitStatusKnown: true,
			LastOpened: fixedNow.Add(-2 * time.Hour),
		},
		now: fixedNow,
	}
	subtitle := renderSubtitle(item, newRowStyles(false))
	plain := ansi.Strip(subtitle)
	for _, want := range []string{"Go", "main ●", "work", "2h ago"} {
		if !strings.Contains(plain, want) {
			t.Fatalf("subtitle = %q, want %q", plain, want)
		}
	}
	if !strings.Contains(subtitle, "251;191;36") {
		t.Fatalf("dirty marker is not highlighted: %q", subtitle)
	}
	if !strings.Contains(subtitle, "34;211;238") {
		t.Fatalf("technology is not in its color: %q", subtitle)
	}
}

func TestFooterGroupsNavigationAndActions(t *testing.T) {
	model := favoriteModel(t, func(project.Project) (bool, error) { return true, nil })
	lines := strings.Split(ansi.Strip(model.View().Content), "\n")
	if len(lines) != 30 {
		t.Fatalf("view has %d lines, want the window height 30", len(lines))
	}
	navigation, actions := lines[len(lines)-2], lines[len(lines)-1]
	for _, want := range []string{"move", "filter", "quit"} {
		if !strings.Contains(navigation, want) {
			t.Fatalf("navigation line = %q, want %q", navigation, want)
		}
	}
	for _, want := range []string{"browse", "cd here", "open IDE", "compose", "favorite"} {
		if !strings.Contains(actions, want) {
			t.Fatalf("actions line = %q, want %q", actions, want)
		}
	}
	if strings.Contains(navigation+actions, "…") {
		t.Fatalf("footer is truncated:\n%s\n%s", navigation, actions)
	}
}

func TestFooterHidesFavoriteWithoutStateAndShowsFilterKeys(t *testing.T) {
	model := testModel()
	footer := ansi.Strip(model.footerView())
	if strings.Contains(footer, "favorite") {
		t.Fatalf("footer = %q, favorite needs a state file", footer)
	}

	model = press(t, model, letterKey('/'))
	footer = ansi.Strip(model.footerView())
	if !strings.Contains(footer, "apply") || !strings.Contains(footer, "cancel") {
		t.Fatalf("footer while filtering = %q, want apply/cancel", footer)
	}
}

func TestHintIsShownOnStart(t *testing.T) {
	model := NewModelWithOptions(nil, Options{Icons: icon.ModeASCII, Hint: "fg workspace add <folder>"})
	if model.Init() == nil {
		t.Fatal("Init() = nil, want the hint timer")
	}
	if view := ansi.Strip(model.View().Content); !strings.Contains(view, "fg workspace add <folder>") {
		t.Fatalf("view = %q, want hint", view)
	}
}

func TestActionKeysDoNotPageTheList(t *testing.T) {
	model := favoriteModel(t, nil)
	model = press(t, model, letterKey('d'))
	if model.mode != dockerScreen {
		t.Fatalf("mode = %v, want docker", model.mode)
	}
	before := model.list.Index()
	model = press(t, model, letterKey('d'), letterKey('o'), letterKey('f'))
	if model.mode != dockerScreen || model.list.Index() != before || model.list.Paginator.Page != 0 {
		t.Fatalf("action keys changed the docker screen: mode=%v index=%d page=%d", model.mode, model.list.Index(), model.list.Paginator.Page)
	}
}

var _ list.Item = projectItem{}
