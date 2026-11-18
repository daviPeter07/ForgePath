package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/daviPeter07/forgepath/internal/icon"
	"github.com/daviPeter07/forgepath/internal/project"
)

func changesModel(t *testing.T, width int) Model {
	t.Helper()
	return sizedModel(t, []project.Project{
		{Name: "api", Path: "/work/api", Technology: project.TechnologyGo, GitBranch: "main", GitDirty: true, GitStatusKnown: true},
		{Name: "docs", Path: "/work/docs", Technology: project.TechnologyPython},
		{Name: "web", Path: "/work/web", Technology: project.TechnologyTypeScript, GitBranch: "main", GitStatusKnown: true},
		{Name: "worker", Path: "/work/worker", Technology: project.TechnologyGo, GitBranch: "fix", GitDirty: true, GitStatusKnown: true},
	}, Options{Icons: icon.ModeASCII}, width, 30)
}

func TestBangShowsOnlyProjectsWithChanges(t *testing.T) {
	model := changesModel(t, 96)
	model = press(t, model, letterKey('!'))

	if names := strings.Join(visibleProjectNames(model), ","); names != "api,worker" {
		t.Fatalf("visible = %s, want only projects with uncommitted changes", names)
	}
	if !strings.Contains(model.list.Title, "With changes") {
		t.Fatalf("title = %q, want the filter named in the breadcrumb", ansi.Strip(model.list.Title))
	}
	if footer := ansi.Strip(model.footerView()); !strings.Contains(footer, "! show all") {
		t.Fatalf("footer = %q, want the way back", footer)
	}

	model = press(t, model, letterKey('!'))
	if names := strings.Join(visibleProjectNames(model), ","); names != "api,docs,web,worker" {
		t.Fatalf("visible after second ! = %s, want every project", names)
	}
	if footer := ansi.Strip(model.footerView()); !strings.Contains(footer, "! changes only") {
		t.Fatalf("footer = %q, want the filter action", footer)
	}
}

func TestChangesFilterExplainsEmptyResult(t *testing.T) {
	model := sizedModel(t, []project.Project{
		{Name: "web", Path: "/work/web", Technology: project.TechnologyTypeScript, GitBranch: "main", GitStatusKnown: true},
	}, Options{Icons: icon.ModeASCII}, 120, 30)
	model = press(t, model, letterKey('!'))
	if plain := ansi.Strip(model.View().Content); !strings.Contains(plain, "Every project is committed") {
		t.Fatalf("empty filter result is not explained:\n%s", plain)
	}
}

func TestChangesFilterOnlyOnProjectList(t *testing.T) {
	_, _, model := monorepo(t)
	model = press(t, model, enterKey, letterKey('!'))
	if model.changesOnly {
		t.Fatal("! toggled the filter from the directory browser")
	}
}

func TestChangesFilterSurvivesFavoriteToggle(t *testing.T) {
	model := changesModel(t, 96)
	model.options.ToggleFavorite = func(p project.Project) (bool, error) { return !p.Favorite, nil }
	model = press(t, model, letterKey('!'), downKey, letterKey('f'))
	if names := strings.Join(visibleProjectNames(model), ","); names != "worker,api" {
		t.Fatalf("visible = %s, want the favorite first and the filter kept", names)
	}
}

func TestFooterNeverWrapsOnNarrowTerminals(t *testing.T) {
	model := changesModel(t, 60)
	model.options.ToggleFavorite = func(p project.Project) (bool, error) { return true, nil }
	assertFitsWindow(t, model.View().Content, 60, 30)
}
