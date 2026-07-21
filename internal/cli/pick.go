package cli

import (
	"fmt"

	"github.com/daviPeter07/forgepath/internal/icon"
	"github.com/daviPeter07/forgepath/internal/scanner"
	"github.com/daviPeter07/forgepath/internal/tui"
	"github.com/spf13/cobra"
)

func newPickCommand(statePaths ...statePathFunc) *cobra.Command {
	var statePath statePathFunc
	if len(statePaths) > 0 {
		statePath = statePaths[0]
	}
	var iconMode string
	command := &cobra.Command{
		Use:   "pick [workspace]",
		Short: "Select a project and print its path",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			icons, err := icon.ParseMode(iconMode)
			if err != nil {
				return err
			}
			workspace, err := workspaceFrom(args)
			if err != nil {
				return err
			}
			projects, err := scanner.Scan(workspace)
			if err != nil {
				return err
			}
			decorateProjectsBestEffort(cmd, statePath, projects)
			if len(projects) == 0 {
				return fmt.Errorf("no projects found in %q", workspace)
			}

			selected, found, err := tui.Select(cmd.Context(), projects, icons, cmd.InOrStdin(), cmd.ErrOrStderr())
			if err != nil {
				return err
			}
			if !found {
				return nil
			}
			if _, err = fmt.Fprintln(cmd.OutOrStdout(), selected.Path); err != nil {
				return err
			}
			recordRecentBestEffort(cmd, statePath, selected.Path)
			return nil
		},
	}
	command.Flags().Bool("print-path", false, "print only the selected project path")
	command.Flags().StringVar(&iconMode, "icons", string(icon.ModeASCII), "icon mode: ascii or nerd-font")

	return command
}
