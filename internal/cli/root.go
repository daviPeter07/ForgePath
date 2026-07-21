package cli

import (
	"io"

	"github.com/spf13/cobra"
)

func NewRootCommand(out, errOut io.Writer) *cobra.Command {
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
	command.AddCommand(newListCommand())
	command.AddCommand(newPickCommand())

	return command
}
