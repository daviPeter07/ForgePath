package icon

import (
	"testing"

	"github.com/charmbracelet/colorprofile"
	"github.com/daviPeter07/forgepath/internal/project"
)

func TestLabelSupportsEveryTechnology(t *testing.T) {
	technologies := []project.Technology{
		project.TechnologyTypeScript,
		project.TechnologyJavaScript,
		project.TechnologyPython,
		project.TechnologyGo,
		project.TechnologyJava,
		project.TechnologyPHP,
		project.TechnologyDocker,
		project.TechnologyRust,
		project.TechnologyRuby,
		project.TechnologySwift,
		project.TechnologyElixir,
	}

	for _, technology := range technologies {
		if label := Label(technology, ModeASCII); label == "" {
			t.Fatalf("ASCII Label(%q) is empty", technology)
		}
		if label := Label(technology, ModeNerdFont); label == "" {
			t.Fatalf("Nerd Font Label(%q) is empty", technology)
		}
	}
}

func TestParseMode(t *testing.T) {
	tests := []struct {
		value string
		mode  Mode
	}{
		{value: "ascii", mode: ModeASCII},
		{value: "ASCII", mode: ModeASCII},
		{value: "auto", mode: ModeAuto},
		{value: "graphics", mode: ModeGraphics},
		{value: "nerd-font", mode: ModeNerdFont},
	}

	for _, tt := range tests {
		mode, err := ParseMode(tt.value)
		if err != nil {
			t.Fatalf("ParseMode(%q) error = %v", tt.value, err)
		}
		if mode != tt.mode {
			t.Fatalf("ParseMode(%q) = %q, want %q", tt.value, mode, tt.mode)
		}
	}

	if _, err := ParseMode("emoji"); err == nil {
		t.Fatal("ParseMode(emoji) error = nil, want error")
	}
}

func TestResolveModePreservesExplicitMode(t *testing.T) {
	if got := ResolveMode(ModeGraphics, nil); got != ModeGraphics {
		t.Fatalf("ResolveMode(graphics) = %q", got)
	}
}

func TestResolveModeUsesASCIIFallbackWhenColorIsDisabled(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	if got := ResolveMode(ModeAuto, nil); got != ModeASCII {
		t.Fatalf("ResolveMode(auto) = %q, want ascii", got)
	}
}

func TestModeForProfileUsesGraphicsOnlyForTrueColor(t *testing.T) {
	if got := modeForProfile(colorprofile.TrueColor); got != ModeGraphics {
		t.Fatalf("modeForProfile(TrueColor) = %q", got)
	}
	if got := modeForProfile(colorprofile.ANSI256); got != ModeASCII {
		t.Fatalf("modeForProfile(ANSI256) = %q", got)
	}
}
