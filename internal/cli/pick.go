package cli

import (
	"fmt"

	"github.com/daviPeter07/forgepath/internal/scanner"
	"github.com/daviPeter07/forgepath/internal/tui"
	"github.com/spf13/cobra"
)

func newPickCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "pick [workspace]",
		Short: "Select a project and print its path",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			workspace, err := workspaceFrom(args)
			if err != nil {
				return err
			}

			projects, err := scanner.Scan(workspace)
			if err != nil {
				return err
			}
			if len(projects) == 0 {
				return fmt.Errorf("no projects found in %q", workspace)
			}

			selected, found, err := tui.Select(cmd.Context(), projects, cmd.InOrStdin(), cmd.ErrOrStderr())
			if err != nil {
				return err
			}
			if !found {
				return nil
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), selected.Path)
			return err
		},
	}
	command.Flags().Bool("print-path", false, "print only the selected project path")

	return command
}
