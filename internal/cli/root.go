package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/daviPeter07/forgepath/internal/action"
	"github.com/daviPeter07/forgepath/internal/catalog"
	"github.com/spf13/cobra"
)

func NewRootCommand(out, errOut io.Writer) *cobra.Command {
	var configuredPath string
	var configuredStatePath string
	configuredCachePath := os.Getenv("FORGEPATH_CACHE")
	command := &cobra.Command{
		Use:           "forgepath",
		Short:         "Discover software projects in a workspace",
		SilenceErrors: true,
		SilenceUsage:  true,
		Args:          cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}
	command.SetOut(out)
	command.SetErr(errOut)
	command.PersistentFlags().StringVar(&configuredPath, "config", "", "configuration file path")
	command.PersistentFlags().StringVar(&configuredStatePath, "state", "", "state file path")
	command.PersistentFlags().StringVar(&configuredCachePath, "cache", configuredCachePath, "project cache directory")
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
	command.AddCommand(newListCommandWithScanner(scan, statePath))
	command.AddCommand(newPickCommandWithScanner(scan, statePath))
	command.AddCommand(newScanCommand(scan))
	command.AddCommand(newOpenCommand(action.OpenEditor, configPath, statePath))
	command.AddCommand(newRevealCommand(action.OpenFolder, statePath))
	command.AddCommand(newRunCommand(configPath, action.RunCommand, statePath))
	command.AddCommand(newConfigCommand(configPath))
	command.AddCommand(newFavoriteCommand(statePath))
	command.AddCommand(newRecentCommand(statePath))

	return command
}
