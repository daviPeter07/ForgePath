package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/daviPeter07/forgepath/internal/action"
	"github.com/daviPeter07/forgepath/internal/catalog"
	"github.com/daviPeter07/forgepath/internal/icon"
	"github.com/spf13/cobra"
)

func NewRootCommand(out, errOut io.Writer) *cobra.Command {
	configuredPath := os.Getenv("FORGEPATH_CONFIG")
	configuredStatePath := os.Getenv("FORGEPATH_STATE")
	configuredCachePath := os.Getenv("FORGEPATH_CACHE")
	var rootIconMode string
	var rootRefresh bool
	command := &cobra.Command{
		Use:           "fg",
		Short:         "Discover software projects in a workspace",
		SilenceErrors: true,
		SilenceUsage:  true,
		Args:          cobra.NoArgs,
	}
	command.SetOut(out)
	command.SetErr(errOut)
	command.PersistentFlags().StringVar(&configuredPath, "config", configuredPath, "configuration file path")
	command.PersistentFlags().StringVar(&configuredStatePath, "state", configuredStatePath, "state file path")
	command.PersistentFlags().StringVar(&configuredCachePath, "cache", configuredCachePath, "project cache directory")
	command.Flags().StringVar(&rootIconMode, "icons", string(icon.ModeAuto), "icon mode: auto, graphics, ascii, or nerd-font")
	command.Flags().BoolVar(&rootRefresh, "refresh", false, "ignore and rebuild the project cache")
	configPath := func() (string, error) {
		return resolveConfigPath(configuredPath)
	}
	statePath := func() (string, error) {
		return resolveStatePath(configuredStatePath)
	}
	cachePath := func() (string, error) {
		return resolveCachePath(configuredCachePath)
	}
	scan := func(workspace string, refresh bool) (catalog.Result, error) {
		path, err := cachePath()
		if err != nil {
			result, scanErr := directProjectScan(workspace, refresh)
			result.Warning = fmt.Errorf("resolve cache directory: %w", err)
			return result, scanErr
		}
		return (catalog.Store{Directory: path}).Scan(workspace, refresh)
	}
	command.RunE = func(cmd *cobra.Command, _ []string) error {
		icons, err := icon.ParseMode(rootIconMode)
		if err != nil {
			return err
		}
		return runPicker(cmd, nil, icons, rootRefresh, scan, statePath, configPath)
	}
	command.AddCommand(newConfiguredListCommand(scan, statePath, configPath))
	command.AddCommand(newConfiguredPickCommand(scan, statePath, configPath))
	command.AddCommand(newScanCommand(scan))
	command.AddCommand(newOpenCommand(action.OpenEditor, configPath, statePath))
	command.AddCommand(newRevealCommand(action.OpenFolder, statePath))
	command.AddCommand(newRunCommand(configPath, action.RunCommand, statePath))
	command.AddCommand(newConfigCommand(configPath))
	command.AddCommand(newWorkspaceCommand(configPath))
	command.AddCommand(newFavoriteCommand(statePath))
	command.AddCommand(newRecentCommand(statePath))

	return command
}
