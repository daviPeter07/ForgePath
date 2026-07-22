package icon

import (
	"fmt"
	"strings"

	"github.com/daviPeter07/forgepath/internal/project"
)

type Mode string

const (
	ModeAuto     Mode = "auto"
	ModeGraphics Mode = "graphics"
	ModeASCII    Mode = "ascii"
	ModeNerdFont Mode = "nerd-font"
)

var asciiLabels = map[project.Technology]string{
	project.TechnologyTypeScript: "[TS]",
	project.TechnologyJavaScript: "[JS]",
	project.TechnologyPython:     "[PY]",
	project.TechnologyGo:         "[GO]",
	project.TechnologyJava:       "[JV]",
	project.TechnologyPHP:        "[PHP]",
	project.TechnologyDocker:     "[DK]",
	project.TechnologyRust:       "[RS]",
	project.TechnologyRuby:       "[RB]",
	project.TechnologySwift:      "[SW]",
	project.TechnologyElixir:     "[EX]",
}

var nerdFontLabels = map[project.Technology]string{
	project.TechnologyTypeScript: "",
	project.TechnologyJavaScript: "",
	project.TechnologyPython:     "",
	project.TechnologyGo:         "",
	project.TechnologyJava:       "",
	project.TechnologyPHP:        "",
	project.TechnologyDocker:     "",
	project.TechnologyRust:       "",
	project.TechnologyRuby:       "",
	project.TechnologySwift:      "",
	project.TechnologyElixir:     "",
}

func ParseMode(value string) (Mode, error) {
	switch Mode(strings.ToLower(strings.TrimSpace(value))) {
	case ModeAuto:
		return ModeAuto, nil
	case ModeGraphics:
		return ModeGraphics, nil
	case ModeASCII:
		return ModeASCII, nil
	case ModeNerdFont:
		return ModeNerdFont, nil
	default:
		return "", fmt.Errorf("invalid icon mode %q: use auto, graphics, ascii, or nerd-font", value)
	}
}

func Label(technology project.Technology, mode Mode) string {
	if mode == ModeNerdFont {
		return nerdFontLabels[technology]
	}
	return asciiLabels[technology]
}
