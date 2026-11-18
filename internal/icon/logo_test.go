package icon

import (
	"fmt"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
	"github.com/daviPeter07/forgepath/internal/project"
)

func TestNormalizeArcFlagsSplitsCompactFlags(t *testing.T) {
	tests := map[string]string{
		// rx ry rotation large-arc sweep x y: "017.82" is 0, 1, 7.82
		`<path d="M1 2a15.56 15.56 0 017.82 4.5"/>`:     `<path d="M1 2a15.56 15.56 0 0 1 7.82 4.5"/>`,
		`<path d="M0 0A5 5 0 1 1 10 10a1 1 0 00.5.5"/>`: `<path d="M0 0A5 5 0 1 1 10 10a1 1 0 0 0 .5 .5"/>`,
		`<path d="M0 0h10v10z"/>`:                       `<path d="M0 0h10v10z"/>`,
	}
	for input, want := range tests {
		if got := normalizeArcFlags(input); got != want {
			t.Errorf("normalizeArcFlags(%q)\n got %q\nwant %q", input, got, want)
		}
	}
}

func TestLogosKeepTheirInnerShapes(t *testing.T) {
	// TypeScript's "TS" and Swift's bird are white cut-outs drawn with compact
	// arc flags; they disappeared before the flags were normalized.
	for _, technology := range []project.Technology{project.TechnologyTypeScript, project.TechnologySwift} {
		rendered, err := GraphicSized(technology, 24)
		if err != nil {
			t.Fatal(err)
		}
		if !hasLightPixel(rendered) {
			t.Fatalf("%s logo has no light pixels; its inner shape is missing", technology)
		}
	}
}

func TestDarkLogosStayVisibleOnDarkTerminals(t *testing.T) {
	rendered, err := GraphicSized(project.TechnologyRust, 16)
	if err != nil {
		t.Fatal(err)
	}
	for _, code := range strings.Split(rendered, "38;2;")[1:] {
		var r, g, b int
		if _, err := fmtSscanRGB(code, &r, &g, &b); err == nil && r+g+b < 120 {
			t.Fatalf("Rust logo still draws near-black pixels (%d,%d,%d)", r, g, b)
		}
	}
}

func TestGraphicSizedDimensions(t *testing.T) {
	for _, size := range []int{8, 16, 20} {
		rendered, err := GraphicSized(project.TechnologyGo, size)
		if err != nil {
			t.Fatal(err)
		}
		if width, height := lipgloss.Width(rendered), lipgloss.Height(rendered); width != size || height != size/2 {
			t.Fatalf("GraphicSized(%d) = %dx%d cells, want %dx%d", size, width, height, size, size/2)
		}
	}
	if _, err := GraphicSized(project.TechnologyGo, 1); err == nil {
		t.Fatal("GraphicSized(1) error = nil, want a minimum size error")
	}
}

func TestAutoModeUsesLogosOnTrueColorTerminals(t *testing.T) {
	if mode := modeForProfile(colorprofile.TrueColor); mode != ModeGraphics {
		t.Fatalf("modeForProfile(TrueColor) = %q, want graphics", mode)
	}
	for _, profile := range []colorprofile.Profile{colorprofile.ANSI256, colorprofile.ANSI, colorprofile.Ascii} {
		if mode := modeForProfile(profile); mode != ModeASCII {
			t.Fatalf("modeForProfile(%v) = %q, want ascii", profile, mode)
		}
	}
}

func hasLightPixel(rendered string) bool {
	for _, code := range strings.Split(rendered, "8;2;")[1:] {
		var r, g, b int
		if _, err := fmtSscanRGB(code, &r, &g, &b); err == nil && r > 220 && g > 220 && b > 220 {
			return true
		}
	}
	return false
}

func fmtSscanRGB(code string, r, g, b *int) (int, error) {
	return fmt.Sscanf(code, "%d;%d;%d", r, g, b)
}
