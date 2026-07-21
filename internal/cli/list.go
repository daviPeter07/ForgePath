package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/daviPeter07/forgepath/internal/scanner"
	"github.com/spf13/cobra"
)

func newListCommand(statePaths ...statePathFunc) *cobra.Command {
	var statePath statePathFunc
	if len(statePaths) > 0 {
		statePath = statePaths[0]
	}
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
			decorateProjectsBestEffort(cmd, statePath, projects)

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
	var workspace string
	if len(args) == 1 {
		workspace = args[0]
	} else {
		current, err := os.Getwd()
		if err != nil {
			return "", err
		}
		workspace = current
	}
	return filepath.Abs(workspace)
}
