package icon

import (
	"image"
	"io"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/daviPeter07/forgepath/internal/project"
	iconassets "github.com/daviPeter07/forgepath/public/icons"
)

func TestGraphicRendersEveryTechnologyAsset(t *testing.T) {
	for technology := range graphicFiles {
		rendered, err := Graphic(technology)
		if err != nil {
			t.Fatalf("Graphic(%q) error = %v", technology, err)
		}
		if !strings.Contains(rendered, "38;2;") {
			t.Fatalf("Graphic(%q) has no truecolor ANSI output", technology)
		}
		blank := renderANSI(image.NewRGBA(image.Rect(0, 0, graphicWidth, graphicHeight)))
		if rendered == blank {
			t.Fatalf("Graphic(%q) rendered only transparent background pixels", technology)
		}
		if width, height := lipgloss.Width(rendered), lipgloss.Height(rendered); width != graphicWidth || height != graphicHeight/2 {
			t.Fatalf("Graphic(%q) size = %dx%d, want %dx%d", technology, width, height, graphicWidth, graphicHeight/2)
		}
	}
}

func TestGraphicAssetsMatchTheirTechnology(t *testing.T) {
	expectedColors := map[project.Technology]string{
		project.TechnologyTypeScript: "#007acc",
		project.TechnologyJavaScript: "#f0db4f",
		project.TechnologyPython:     "#5a9fd4",
		project.TechnologyGo:         "#00acd7",
		project.TechnologyJava:       "#0074bd",
		project.TechnologyPHP:        "#777bb3",
		project.TechnologyDocker:     "#00aada",
		project.TechnologyRuby:       "#fb7655",
		project.TechnologySwift:      "#f05138",
		project.TechnologyElixir:     "#8d67af",
	}
	for technology, expectedColor := range expectedColors {
		file, err := iconassets.Files.Open(graphicFiles[technology])
		if err != nil {
			t.Fatal(err)
		}
		content, err := io.ReadAll(file)
		file.Close()
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(strings.ToLower(string(content)), expectedColor) {
			t.Fatalf("asset for %q does not contain brand color %s", technology, expectedColor)
		}
	}
}

func TestGraphicRejectsUnknownTechnology(t *testing.T) {
	if _, err := Graphic(project.Technology("Unknown")); err == nil {
		t.Fatal("Graphic(Unknown) error = nil")
	}
}
