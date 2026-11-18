package tui

import (
	"fmt"
	"image/color"
	"sort"
	"strings"

	"charm.land/lipgloss/v2"
)

// Theme names accepted by --theme and the config file.
const (
	// ThemeAuto picks dark or light from the terminal background.
	ThemeAuto    = "auto"
	ThemeDark    = "dark"
	ThemeLight   = "light"
	ThemeDracula = "dracula"
)

type colorPalette struct {
	primary  color.Color // brand chip, selection border, footer labels
	bright   color.Color // breadcrumb and key names
	surface  color.Color // selected row background
	text     color.Color
	muted    color.Color
	favorite color.Color // favorite star and filter matches
	ink      color.Color // text drawn on colored badges
	danger   color.Color
	success  color.Color
	changed  color.Color // uncommitted changes
	border   color.Color // inspector frame
}

var themes = map[string]colorPalette{
	ThemeDark: {
		primary:  lipgloss.Color("#A855F7"),
		bright:   lipgloss.Color("#D8B4FE"),
		surface:  lipgloss.Color("#241532"),
		text:     lipgloss.Color("#F7F2FF"),
		muted:    lipgloss.Color("#A99DB8"),
		favorite: lipgloss.Color("#FACC15"),
		ink:      lipgloss.Color("#17111F"),
		danger:   lipgloss.Color("#EF4444"),
		success:  lipgloss.Color("#22C55E"),
		changed:  lipgloss.Color("#FBBF24"),
		border:   lipgloss.Color("#4C3A63"),
	},
	// light keeps the same purple identity with darker tones that stay
	// readable on white backgrounds.
	ThemeLight: {
		primary:  lipgloss.Color("#7E22CE"),
		bright:   lipgloss.Color("#6B21A8"),
		surface:  lipgloss.Color("#F3E8FF"),
		text:     lipgloss.Color("#1F1430"),
		muted:    lipgloss.Color("#6B5E7B"),
		favorite: lipgloss.Color("#B45309"),
		ink:      lipgloss.Color("#17111F"),
		danger:   lipgloss.Color("#DC2626"),
		success:  lipgloss.Color("#15803D"),
		changed:  lipgloss.Color("#C2410C"),
		border:   lipgloss.Color("#D8B4FE"),
	},
	// dracula follows the Dracula specification (draculatheme.com).
	ThemeDracula: {
		primary:  lipgloss.Color("#BD93F9"),
		bright:   lipgloss.Color("#FF79C6"),
		surface:  lipgloss.Color("#44475A"),
		text:     lipgloss.Color("#F8F8F2"),
		muted:    lipgloss.Color("#8B95C9"),
		favorite: lipgloss.Color("#F1FA8C"),
		ink:      lipgloss.Color("#282A36"),
		danger:   lipgloss.Color("#FF5555"),
		success:  lipgloss.Color("#50FA7B"),
		changed:  lipgloss.Color("#FFB86C"),
		border:   lipgloss.Color("#6272A4"),
	},
}

// ValidateTheme accepts an empty name (auto) or one of the theme names.
func ValidateTheme(name string) error {
	if name == "" || name == ThemeAuto {
		return nil
	}
	if _, ok := themes[name]; ok {
		return nil
	}
	names := []string{ThemeAuto}
	for theme := range themes {
		names = append(names, theme)
	}
	sort.Strings(names[1:])
	return fmt.Errorf("unknown theme %q: use %s", name, strings.Join(names, ", "))
}

func applyTheme(name string) {
	if theme, ok := themes[name]; ok {
		palette = theme
	}
}

func autoTheme(name string) bool {
	return name == "" || name == ThemeAuto
}
