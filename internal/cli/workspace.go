package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	appconfig "github.com/daviPeter07/forgepath/internal/config"
	"github.com/daviPeter07/forgepath/internal/detector"
	"github.com/daviPeter07/forgepath/internal/scanner"
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
			if _, err := fmt.Fprintln(cmd.OutOrStdout(), workspace); err != nil {
				return err
			}
			describeWorkspace(cmd.ErrOrStderr(), workspace)
			return nil
		},
	}
}

// describeWorkspace tells the user what ForgePath will find in a new
// workspace. Only direct subfolders are scanned, so adding a project folder
// itself (instead of the folder that contains it) is a common mistake.
func describeWorkspace(out io.Writer, workspace string) {
	projects, err := scanner.Scan(workspace)
	if err != nil {
		return
	}
	switch {
	case len(projects) == 1:
		_, _ = fmt.Fprintln(out, "1 project found")
	case len(projects) > 1:
		_, _ = fmt.Fprintf(out, "%d projects found\n", len(projects))
	default:
		if _, isProject, err := detector.Detect(workspace); err == nil && isProject {
			_, _ = fmt.Fprintf(out, "warning: %q is a project itself; ForgePath lists the projects inside a workspace, so add its parent folder instead: fg workspace add %q\n", workspace, filepath.Dir(workspace))
			return
		}
		_, _ = fmt.Fprintf(out, "warning: no projects found directly inside %q; a workspace is the folder that contains your project folders\n", workspace)
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

// configuredWorkspaces picks the folders to scan: an explicit workspace
// argument, otherwise every folder in the catalog (fg workspace add),
// otherwise the current directory.
func configuredWorkspaces(args []string, configPath configPathFunc) ([]string, error) {
	if len(args) == 1 || configPath == nil {
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
