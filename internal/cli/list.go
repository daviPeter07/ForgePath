package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func newListCommand(statePaths ...statePathFunc) *cobra.Command {
	return newListCommandWithScanner(directProjectScan, statePaths...)
}

func newListCommandWithScanner(scan scanProjectsFunc, statePaths ...statePathFunc) *cobra.Command {
	var statePath statePathFunc
	if len(statePaths) > 0 {
		statePath = statePaths[0]
	}
	var refresh bool
	command := &cobra.Command{
		Use:   "list [workspace]",
		Short: "List projects found in a workspace",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			workspace, err := workspaceFrom(args)
			if err != nil {
				return err
			}
			projects, err := scanProjects(cmd, scan, workspace, refresh)
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
	command.Flags().BoolVar(&refresh, "refresh", false, "ignore and rebuild the project cache")
	return command
}

func newConfiguredListCommand(scan scanProjectsFunc, statePath statePathFunc, configPath configPathFunc) *cobra.Command {
	command := newListCommandWithScanner(scan, statePath)
	command.RunE = func(cmd *cobra.Command, args []string) error {
		workspaces, err := configuredWorkspaces(args, configPath)
		if err != nil {
			return err
		}
		refresh, err := cmd.Flags().GetBool("refresh")
		if err != nil {
			return err
		}
		projects, err := scanConfiguredProjects(cmd, scan, workspaces, refresh)
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
	}
	return command
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
