package cli

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	appconfig "github.com/daviPeter07/forgepath/internal/config"
	"github.com/spf13/cobra"
)

func themeCommand(t *testing.T, args ...string) *cobra.Command {
	t.Helper()
	command := &cobra.Command{Use: "test", RunE: func(*cobra.Command, []string) error { return nil }}
	command.Flags().String("theme", "", "")
	command.SetArgs(args)
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}
	return command
}

func themeConfig(t *testing.T, theme string) configPathFunc {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.json")
	configuration := appconfig.Default()
	configuration.Theme = theme
	if err := appconfig.Save(path, configuration); err != nil {
		t.Fatal(err)
	}
	return func() (string, error) { return path, nil }
}

func TestPickerThemePrefersFlagOverConfig(t *testing.T) {
	theme, err := pickerTheme(themeCommand(t, "--theme", "dracula"), themeConfig(t, "light"))
	if err != nil || theme != "dracula" {
		t.Fatalf("pickerTheme() = %q, %v; want the flag", theme, err)
	}
	theme, err = pickerTheme(themeCommand(t), themeConfig(t, "light"))
	if err != nil || theme != "light" {
		t.Fatalf("pickerTheme() = %q, %v; want the config theme", theme, err)
	}
	theme, err = pickerTheme(themeCommand(t), nil)
	if err != nil || theme != "" {
		t.Fatalf("pickerTheme() = %q, %v; want auto", theme, err)
	}
}

func TestPickerThemeRejectsUnknownNames(t *testing.T) {
	if _, err := pickerTheme(themeCommand(t, "--theme", "neon"), nil); err == nil || !strings.Contains(err.Error(), "neon") {
		t.Fatalf("flag error = %v", err)
	}
	if _, err := pickerTheme(themeCommand(t), themeConfig(t, "neon")); err == nil || !strings.Contains(err.Error(), "neon") {
		t.Fatalf("config error = %v", err)
	}
}

func TestRootAndPickAcceptTheme(t *testing.T) {
	for _, args := range [][]string{{"--help"}, {"pick", "--help"}} {
		var output bytes.Buffer
		command := NewRootCommand(&output, &bytes.Buffer{})
		command.SetArgs(args)
		if err := command.Execute(); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(output.String(), "--theme") {
			t.Fatalf("%v help does not list --theme:\n%s", args, output.String())
		}
	}
}
