package cli

import (
	"io"

	"github.com/daviPeter07/forgepath/internal/action"
	"github.com/spf13/cobra"
)

func NewRootCommand(out, errOut io.Writer) *cobra.Command {
	var configuredPath string
	var configuredStatePath string
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
	configPath := func() (string, error) {
		return resolveConfigPath(configuredPath)
	}
	statePath := func() (string, error) {
		return resolveStatePath(configuredStatePath)
	}
	command.AddCommand(newListCommand(statePath))
	command.AddCommand(newPickCommand(statePath))
	command.AddCommand(newOpenCommand(action.OpenEditor, configPath, statePath))
	command.AddCommand(newRevealCommand(action.OpenFolder, statePath))
	command.AddCommand(newRunCommand(configPath, action.RunCommand, statePath))
	command.AddCommand(newConfigCommand(configPath))
	command.AddCommand(newFavoriteCommand(statePath))
	command.AddCommand(newRecentCommand(statePath))

	return command
}
