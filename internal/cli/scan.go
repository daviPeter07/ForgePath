package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newScanCommand(scan scanProjectsFunc) *cobra.Command {
	return &cobra.Command{
		Use:   "scan [workspace]",
		Short: "Scan a workspace and rebuild its project cache",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			workspace, err := workspaceFrom(args)
			if err != nil {
				return err
			}
			result, err := scan(workspace, true)
			if err != nil {
				return err
			}
			if result.Warning != nil {
				return fmt.Errorf("scan completed but cache could not be updated: %w", result.Warning)
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "%d projects cached\n", len(result.Projects))
			return err
		},
	}
}
