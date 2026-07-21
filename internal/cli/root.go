package cli

import (
	"io"

	"github.com/daviPeter07/forgepath/internal/action"
	"github.com/spf13/cobra"
)

func NewRootCommand(out, errOut io.Writer) *cobra.Command {
	var configuredPath string
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
	configPath := func() (string, error) {
		return resolveConfigPath(configuredPath)
	}
	command.AddCommand(newListCommand())
	command.AddCommand(newPickCommand())
	command.AddCommand(newOpenCommand(action.OpenEditor, configPath))
	command.AddCommand(newRevealCommand(action.OpenFolder))
	command.AddCommand(newRunCommand(configPath, action.RunCommand))
	command.AddCommand(newConfigCommand(configPath))

	return command
}
