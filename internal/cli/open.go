package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/daviPeter07/forgepath/internal/project"
	"github.com/daviPeter07/forgepath/internal/scanner"
	"github.com/spf13/cobra"
)

type openEditorFunc func(context.Context, string, string) error
type openFolderFunc func(context.Context, string) error

func newOpenCommand(openEditor openEditorFunc) *cobra.Command {
	editor := os.Getenv("FORGEPATH_EDITOR")
	command := &cobra.Command{
		Use:   "open <project> [workspace]",
		Short: "Open a project in an editor",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if editor == "" {
				return fmt.Errorf("editor is required: use --editor or FORGEPATH_EDITOR")
			}
			found, err := resolveProject(cmd.Context(), args[0], args[1:])
			if err != nil {
				return err
			}
			if err := openEditor(cmd.Context(), found.Path, editor); err != nil {
				return fmt.Errorf("open %q in editor %q: %w", found.Name, editor, err)
			}
			return nil
		},
	}
	command.Flags().StringVar(&editor, "editor", editor, "editor executable")
	return command
}

func newRevealCommand(openFolder openFolderFunc) *cobra.Command {
	return &cobra.Command{
		Use:   "reveal <project> [workspace]",
		Short: "Reveal a project in the file manager",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			found, err := resolveProject(cmd.Context(), args[0], args[1:])
			if err != nil {
				return err
			}
			if err := openFolder(cmd.Context(), found.Path); err != nil {
				return fmt.Errorf("reveal %q: %w", found.Name, err)
			}
			return nil
		},
	}
}

func resolveProject(ctx context.Context, name string, workspaceArguments []string) (project.Project, error) {
	if err := ctx.Err(); err != nil {
		return project.Project{}, err
	}
	workspace, err := workspaceFrom(workspaceArguments)
	if err != nil {
		return project.Project{}, err
	}
	workspace, err = filepath.Abs(workspace)
	if err != nil {
		return project.Project{}, err
	}
	projects, err := scanner.Scan(workspace)
	if err != nil {
		return project.Project{}, err
	}
	for _, found := range projects {
		if found.Name == name {
			if err := ctx.Err(); err != nil {
				return project.Project{}, err
			}
			return found, nil
		}
	}
	return project.Project{}, fmt.Errorf("project %q not found in %q", name, workspace)
}
