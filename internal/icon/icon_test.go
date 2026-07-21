package icon

import (
	"testing"

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
