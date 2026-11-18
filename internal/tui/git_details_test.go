package tui

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/daviPeter07/forgepath/internal/git"
	"github.com/daviPeter07/forgepath/internal/icon"
	"github.com/daviPeter07/forgepath/internal/project"
)

// runCommands executes a command tree and feeds every resulting message back
// into the model, the way the Bubble Tea runtime would.
func runCommands(t *testing.T, model Model, command tea.Cmd) Model {
	t.Helper()
	pending := []tea.Cmd{command}
	for steps := 0; len(pending) > 0 && steps < 50; steps++ {
		next := pending[0]
		pending = pending[1:]
		if next == nil {
			continue
		}
		message := next()
		if batch, ok := message.(tea.BatchMsg); ok {
			pending = append(pending, batch...)
			continue
		}
		if _, ok := message.(gitDetailsMsg); !ok {
			continue // only the git loads matter here; skip timers and the like
		}
		updated, follow := model.Update(message)
		model = updated.(Model)
		pending = append(pending, follow)
	}
	return model
}

func gitModel(t *testing.T, load func(context.Context, string) (git.Details, error)) (Model, tea.Cmd) {
	t.Helper()
	projects := []project.Project{
		{Name: "api", Path: "/work/api", Technology: project.TechnologyGo, GitBranch: "main", GitStatusKnown: true},
		{Name: "notes", Path: "/work/notes", Technology: project.TechnologyPython},
	}
	model := NewModelWithOptions(projects, Options{Icons: icon.ModeASCII, GitDetails: load, Now: func() time.Time { return fixedNow }})
	updated, command := model.Update(tea.WindowSizeMsg{Width: 120, Height: 32})
	return updated.(Model), command
}

func TestInspectorLoadsGitDetailsForHighlightedProject(t *testing.T) {
	var calls []string
	model, command := gitModel(t, func(_ context.Context, path string) (git.Details, error) {
		calls = append(calls, path)
		return git.Details{
			Upstream: "origin/main", Ahead: 2, Behind: 1,
			Commits: []git.Commit{
				{Hash: "a1b2c3d", Subject: "fix login redirect", When: fixedNow.Add(-2 * time.Hour)},
				{Hash: "9f8e7d6", Subject: "add tests", When: fixedNow.Add(-26 * time.Hour)},
			},
		}, nil
	})

	if plain := ansi.Strip(model.View().Content); !strings.Contains(plain, "loading") {
		t.Fatalf("inspector should say the details are loading:\n%s", plain)
	}
	model = runCommands(t, model, command)

	plain := ansi.Strip(model.View().Content)
	for _, want := range []string{"↑2 to push", "↓1 to pull", "origin/main", "a1b2c3d fix login redirect", "2h ago", "9f8e7d6 add tests", "1d ago"} {
		if !strings.Contains(plain, want) {
			t.Fatalf("inspector is missing %q:\n%s", want, plain)
		}
	}
	if strings.Join(calls, ",") != "/work/api" {
		t.Fatalf("loads = %q, want one load for the highlighted project", calls)
	}

	// Moving away and back reuses the cache.
	updated, command := model.Update(downKey)
	model = runCommands(t, updated.(Model), command)
	updated, command = model.Update(tea.KeyPressMsg(tea.Key{Code: tea.KeyUp}))
	model = runCommands(t, updated.(Model), command)
	if strings.Join(calls, ",") != "/work/api" {
		t.Fatalf("loads = %q; projects outside Git must not load and the cache must be reused", calls)
	}
}

func TestInspectorShowsSyncedAndUntrackedBranches(t *testing.T) {
	tests := []struct {
		details git.Details
		want    string
	}{
		{git.Details{Upstream: "origin/main"}, "in sync with origin/main"},
		{git.Details{}, "no upstream branch"},
	}
	for _, tt := range tests {
		model, command := gitModel(t, func(context.Context, string) (git.Details, error) { return tt.details, nil })
		model = runCommands(t, model, command)
		if plain := ansi.Strip(model.View().Content); !strings.Contains(plain, tt.want) {
			t.Fatalf("inspector is missing %q:\n%s", tt.want, plain)
		}
	}
}

func TestInspectorHidesGitDetailsThatFailedToLoad(t *testing.T) {
	model, command := gitModel(t, func(context.Context, string) (git.Details, error) { return git.Details{}, errors.New("timeout") })
	model = runCommands(t, model, command)
	plain := ansi.Strip(model.View().Content)
	if strings.Contains(plain, "loading") || strings.Contains(plain, "commits") {
		t.Fatalf("failed details should disappear quietly:\n%s", plain)
	}
}

func TestGitDetailsAreNotLoadedOnNarrowTerminals(t *testing.T) {
	called := false
	model := NewModelWithOptions([]project.Project{{Name: "api", Path: "/work/api", Technology: project.TechnologyGo, GitBranch: "main"}},
		Options{Icons: icon.ModeASCII, GitDetails: func(context.Context, string) (git.Details, error) { called = true; return git.Details{}, nil }})
	updated, command := model.Update(tea.WindowSizeMsg{Width: 80, Height: 30})
	runCommands(t, updated.(Model), command)
	if called {
		t.Fatal("details were loaded although the inspector is hidden")
	}
}

func TestOneLongInspectorLineDoesNotTruncateTheOthers(t *testing.T) {
	model, command := gitModel(t, func(context.Context, string) (git.Details, error) {
		return git.Details{Upstream: "origin/" + strings.Repeat("very-long-branch-", 6), Ahead: 3, Behind: 2}, nil
	})
	model.options.Icons = icon.ModeGraphics
	model = runCommands(t, model, command)
	truncated := 0
	for _, line := range strings.Split(ansi.Strip(model.View().Content), "\n") {
		if strings.Contains(line, "…") {
			truncated++
		}
	}
	if truncated > 1 {
		t.Fatalf("%d lines were truncated; only the long remote line should be:\n%s", truncated, ansi.Strip(model.View().Content))
	}
}
