package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	appconfig "github.com/daviPeter07/forgepath/internal/config"
	"github.com/spf13/cobra"
)

func newWorkspaceCommand(configPath configPathFunc) *cobra.Command {
	command := &cobra.Command{
		Use:   "workspace",
		Short: "Manage folders that contain your projects",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}
	command.AddCommand(newWorkspaceAddCommand(configPath))
	command.AddCommand(newWorkspaceRemoveCommand(configPath))
	command.AddCommand(newWorkspaceListCommand(configPath))
	return command
}

func newWorkspaceAddCommand(configPath configPathFunc) *cobra.Command {
	return &cobra.Command{
		Use:   "add [path]",
		Short: "Add a folder to the global project catalog",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			workspace, err := workspaceFrom(args)
			if err != nil {
				return err
			}
			info, err := os.Stat(workspace)
			if err != nil {
				return err
			}
			if !info.IsDir() {
				return fmt.Errorf("workspace %q is not a directory", workspace)
			}
			if resolved, resolveErr := filepath.EvalSymlinks(workspace); resolveErr == nil {
				workspace = resolved
			}

			path, err := configPath()
			if err != nil {
				return err
			}
			if err := appconfig.Update(path, func(configuration *appconfig.Config) error {
				for _, configured := range configuration.Workspaces {
					if sameWorkspace(configured, workspace) {
						return nil
					}
				}
				configuration.Workspaces = append(configuration.Workspaces, workspace)
				return nil
			}); err != nil {
				return fmt.Errorf("save config %q: %w", path, err)
			}
			_, err = fmt.Fprintln(cmd.OutOrStdout(), workspace)
			return err
		},
	}
}

func newWorkspaceRemoveCommand(configPath configPathFunc) *cobra.Command {
	return &cobra.Command{
		Use:   "remove <path>",
		Short: "Remove a folder from the global project catalog",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			workspace, err := filepath.Abs(args[0])
			if err != nil {
				return err
			}
			if resolved, resolveErr := filepath.EvalSymlinks(workspace); resolveErr == nil {
				workspace = resolved
			}
			path, err := configPath()
			if err != nil {
				return err
			}
			removed := false
			if err := appconfig.Update(path, func(configuration *appconfig.Config) error {
				kept := configuration.Workspaces[:0]
				for _, configured := range configuration.Workspaces {
					if sameWorkspace(configured, workspace) {
						removed = true
						continue
					}
					kept = append(kept, configured)
				}
				if !removed {
					return fmt.Errorf("workspace %q is not configured", workspace)
				}
				configuration.Workspaces = kept
				return nil
			}); err != nil {
				if !removed {
					return err
				}
				return fmt.Errorf("save config %q: %w", path, err)
			}
			return nil
		},
	}
}

func newWorkspaceListCommand(configPath configPathFunc) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List folders in the global project catalog",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			_, configuration, err := editableConfiguration(configPath)
			if err != nil {
				return err
			}
			for _, workspace := range configuration.Workspaces {
				if _, err := fmt.Fprintln(cmd.OutOrStdout(), workspace); err != nil {
					return err
				}
			}
			return nil
		},
	}
}

func configuredWorkspaces(args []string, configPath configPathFunc) ([]string, error) {
	if len(args) <= 1 || configPath == nil {
		workspace, err := workspaceFrom(args)
		if err != nil {
			return nil, err
		}
		return []string{workspace}, nil
	}
	_, configuration, err := editableConfiguration(configPath)
	if err != nil {
		return nil, err
	}
	if len(configuration.Workspaces) > 0 {
		return append([]string(nil), configuration.Workspaces...), nil
	}
	workspace, err := workspaceFrom(nil)
	if err != nil {
		return nil, err
	}
	return []string{workspace}, nil
}

func editableConfiguration(configPath configPathFunc) (string, appconfig.Config, error) {
	path, err := configPath()
	if err != nil {
		return "", appconfig.Config{}, err
	}
	configuration, err := appconfig.Load(path)
	if os.IsNotExist(err) {
		return path, appconfig.Default(), nil
	}
	if err != nil {
		return "", appconfig.Config{}, err
	}
	return path, configuration, nil
}

func sameWorkspace(left, right string) bool {
	left = filepath.Clean(left)
	right = filepath.Clean(right)
	if runtime.GOOS == "windows" {
		return strings.EqualFold(left, right)
	}
	return left == right
}
