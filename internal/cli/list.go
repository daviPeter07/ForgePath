package cli

import (
	"fmt"
	"os"

	"github.com/daviPeter07/forgepath/internal/scanner"
	"github.com/spf13/cobra"
)

func newListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list [workspace]",
		Short: "List projects found in a workspace",
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

			for _, found := range projects {
				if _, err := fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\t%s\n", found.Name, found.Technology, found.Path); err != nil {
					return err
				}
			}

			return nil
		},
	}
}

func workspaceFrom(args []string) (string, error) {
	if len(args) == 1 {
		return args[0], nil
	}
	return os.Getwd()
}
