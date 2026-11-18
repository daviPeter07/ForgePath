package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

const matchColor = "38;2;250;204;21" // palette.favorite, also used for matches

func renderVisibleItem(model Model, index int) string {
	var output strings.Builder
	projectDelegate{}.Render(&output, model.list, index, model.list.VisibleItems()[index])
	return output.String()
}

func TestFilterHighlightsMatchedCharacters(t *testing.T) {
	model := testModel()
	model.list.SetFilterText("wb")

	if len(model.list.VisibleItems()) != 1 {
		t.Fatalf("visible = %d, want only web", len(model.list.VisibleItems()))
	}
	row := renderVisibleItem(model, 0)
	if !strings.Contains(row, matchColor) {
		t.Fatalf("matched characters are not highlighted: %q", row)
	}
	if !strings.Contains(ansi.Strip(row), "web") {
		t.Fatalf("highlighting broke the name: %q", ansi.Strip(row))
	}
}

func TestRowsWithoutFilterHaveNoHighlight(t *testing.T) {
	row := renderVisibleItem(testModel(), 0)
	if strings.Contains(row, matchColor) {
		t.Fatalf("unfiltered row is highlighted: %q", row)
	}
}

func TestFolderRowsHighlightMatches(t *testing.T) {
	_, _, model := monorepo(t)
	model = press(t, model, enterKey)
	model.list.SetFilterText("doc")
	if len(model.list.VisibleItems()) != 1 {
		t.Fatalf("visible = %d, want only docs", len(model.list.VisibleItems()))
	}
	if row := renderVisibleItem(model, 0); !strings.Contains(row, matchColor) {
		t.Fatalf("folder match is not highlighted: %q", row)
	}
}

func TestHighlightMatchesOffsets(t *testing.T) {
	got := ansi.Strip(highlightMatches("api", []int{0, 2, 7}, 0, newRowStyles(false).name))
	if got != "api" {
		t.Fatalf("highlightMatches changed the text: %q", got)
	}
	plain := highlightMatches("api", nil, 0, newRowStyles(false).name)
	if strings.Contains(plain, matchColor) {
		t.Fatal("no matches should render without highlight")
	}
	shifted := highlightMatches("api", []int{4}, 4, newRowStyles(false).name)
	if !strings.Contains(shifted, matchColor) {
		t.Fatal("offset matches were not applied")
	}
}
