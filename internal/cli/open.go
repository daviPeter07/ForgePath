package cli

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	appconfig "github.com/daviPeter07/forgepath/internal/config"
	"github.com/daviPeter07/forgepath/internal/project"
	"github.com/daviPeter07/forgepath/internal/scanner"
	"github.com/spf13/cobra"
)

type openEditorFunc func(context.Context, string, string) error
type openFolderFunc func(context.Context, string) error

func newOpenCommand(openEditor openEditorFunc, configPath configPathFunc, statePaths ...statePathFunc) *cobra.Command {
	var statePath statePathFunc
	if len(statePaths) > 0 {
		statePath = statePaths[0]
	}
	editor := os.Getenv("FORGEPATH_EDITOR")
	command := &cobra.Command{
		Use:   "open <project> [workspace]",
		Short: "Open a project in an editor",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if editor == "" {
				path, err := configPath()
				if err != nil {
					return err
				}
				configuration, err := appconfig.Load(path)
				if err != nil {
					return fmt.Errorf("editor is required: use --editor, FORGEPATH_EDITOR, or configure editor.executable: %w", err)
				}
				editor = configuration.Editor.Executable
				if editor == "" {
					return fmt.Errorf("editor is required: use --editor, FORGEPATH_EDITOR, or configure editor.executable")
				}
			}
			found, err := resolveProject(cmd.Context(), args[0], args[1:], configPath)
			if err != nil {
				return err
			}
			if err := openEditor(cmd.Context(), found.Path, editor); err != nil {
				return fmt.Errorf("open %q in editor %q: %w", found.Name, editor, err)
			}
			recordRecentBestEffort(cmd, statePath, found.Path)
			return nil
		},
	}
	command.Flags().StringVar(&editor, "editor", editor, "editor executable")
	return command
}

func newRevealCommand(openFolder openFolderFunc, configPath configPathFunc, statePaths ...statePathFunc) *cobra.Command {
	var statePath statePathFunc
	if len(statePaths) > 0 {
		statePath = statePaths[0]
	}
	return &cobra.Command{
		Use:   "reveal <project> [workspace]",
		Short: "Reveal a project in the file manager",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			found, err := resolveProject(cmd.Context(), args[0], args[1:], configPath)
			if err != nil {
				return err
			}
			if err := openFolder(cmd.Context(), found.Path); err != nil {
				return fmt.Errorf("reveal %q: %w", found.Name, err)
			}
			recordRecentBestEffort(cmd, statePath, found.Path)
			return nil
		},
	}
}

// resolveProject finds a project by name in the workspaces chosen by
// configuredWorkspaces. A name found in more than one workspace is an error
// that lists the candidates, so the user can pass the workspace explicitly.
func resolveProject(ctx context.Context, name string, workspaceArguments []string, configPath configPathFunc) (project.Project, error) {
	if err := ctx.Err(); err != nil {
		return project.Project{}, err
	}
	workspaces, err := configuredWorkspaces(workspaceArguments, configPath)
	if err != nil {
		return project.Project{}, err
	}
	var matches []project.Project
	for _, workspace := range workspaces {
		projects, err := scanner.Scan(workspace)
		if err != nil {
			if len(workspaces) == 1 {
				return project.Project{}, err
			}
			continue
		}
		for _, found := range projects {
			if found.Name == name {
				matches = append(matches, found)
			}
		}
	}
	if err := ctx.Err(); err != nil {
		return project.Project{}, err
	}
	switch len(matches) {
	case 0:
		return project.Project{}, fmt.Errorf("project %q not found in %s", name, quotedList(workspaces))
	case 1:
		return matches[0], nil
	default:
		paths := make([]string, len(matches))
		for index, found := range matches {
			paths[index] = found.Path
		}
		return project.Project{}, fmt.Errorf("project %q exists in more than one workspace (%s); pass the workspace as the next argument", name, quotedList(paths))
	}
}

func quotedList(values []string) string {
	quoted := make([]string, len(values))
	for index, value := range values {
		quoted[index] = strconv.Quote(value)
	}
	return strings.Join(quoted, ", ")
}
