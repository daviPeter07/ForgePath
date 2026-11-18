package tui

import (
	"fmt"
	"image/color"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/daviPeter07/forgepath/internal/icon"
	"github.com/daviPeter07/forgepath/internal/project"
)

func restoreTheme(t *testing.T) {
	t.Helper()
	t.Cleanup(func() { applyTheme(ThemeDark) })
}

func rgb(c color.Color) string {
	r, g, b, _ := c.RGBA()
	return fmt.Sprintf("%d;%d;%d", r>>8, g>>8, b>>8)
}

func TestValidateTheme(t *testing.T) {
	for _, name := range []string{"", ThemeAuto, ThemeDark, ThemeLight, ThemeDracula} {
		if err := ValidateTheme(name); err != nil {
			t.Errorf("ValidateTheme(%q) error = %v", name, err)
		}
	}
	err := ValidateTheme("solarized")
	if err == nil || !strings.Contains(err.Error(), "dracula") {
		t.Fatalf("ValidateTheme(solarized) error = %v, want the list of themes", err)
	}
}

func TestExplicitThemeIsAppliedImmediately(t *testing.T) {
	restoreTheme(t)
	model := NewModelWithOptions([]project.Project{{Name: "api", Technology: project.TechnologyGo}}, Options{Icons: icon.ModeASCII, Theme: ThemeDracula})
	if rgb(palette.primary) != rgb(themes[ThemeDracula].primary) {
		t.Fatalf("primary = %s, want Dracula", rgb(palette.primary))
	}
	if !strings.Contains(model.list.Title, rgb(themes[ThemeDracula].primary)) {
		t.Fatalf("title was not drawn with the theme: %q", model.list.Title)
	}

	updated, _ := model.Update(tea.BackgroundColorMsg{Color: color.White})
	if rgb(palette.primary) != rgb(themes[ThemeDracula].primary) {
		t.Fatal("an explicit theme must ignore the terminal background")
	}
	_ = updated
}

func TestAutoThemeFollowsTerminalBackground(t *testing.T) {
	restoreTheme(t)
	model := NewModelWithOptions([]project.Project{{Name: "api", Technology: project.TechnologyGo}}, Options{Icons: icon.ModeASCII})
	if model.Init() == nil {
		t.Fatal("Init() = nil, want the background color request")
	}

	updated, _ := model.Update(tea.BackgroundColorMsg{Color: color.White})
	model = updated.(Model)
	if rgb(palette.text) != rgb(themes[ThemeLight].text) {
		t.Fatalf("text = %s on a white terminal, want the light theme", rgb(palette.text))
	}
	if !strings.Contains(model.list.Title, rgb(themes[ThemeLight].primary)) {
		t.Fatalf("title was not redrawn with the light theme: %q", model.list.Title)
	}

	updated, _ = model.Update(tea.BackgroundColorMsg{Color: color.Black})
	if rgb(palette.text) != rgb(themes[ThemeDark].text) {
		t.Fatal("a dark terminal should switch back to the dark theme")
	}
	_ = updated
}

func TestEveryThemeDefinesEveryColor(t *testing.T) {
	for name, theme := range themes {
		for field, value := range map[string]color.Color{
			"primary": theme.primary, "bright": theme.bright, "surface": theme.surface, "text": theme.text,
			"muted": theme.muted, "favorite": theme.favorite, "ink": theme.ink, "danger": theme.danger,
			"success": theme.success, "changed": theme.changed, "border": theme.border,
		} {
			if value == nil {
				t.Errorf("theme %s has no %s color", name, field)
			}
		}
	}
}
