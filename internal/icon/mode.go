package icon

import (
	"io"
	"os"

	"github.com/charmbracelet/colorprofile"
)

func ResolveMode(mode Mode, output io.Writer) Mode {
	if mode != ModeAuto {
		return mode
	}
	if os.Getenv("NO_COLOR") != "" || os.Getenv("TERM") == "dumb" {
		return ModeASCII
	}
	return modeForProfile(colorprofile.Detect(output, os.Environ()))
}

// modeForProfile draws logos only where they can be shown in full color.
func modeForProfile(profile colorprofile.Profile) Mode {
	if profile == colorprofile.TrueColor {
		return ModeGraphics
	}
	return ModeASCII
}
