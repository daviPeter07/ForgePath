package cli

import (
	"fmt"

	appstate "github.com/daviPeter07/forgepath/internal/state"
	"github.com/spf13/cobra"
)

func newFavoriteCommand(statePath statePathFunc) *cobra.Command {
	command := &cobra.Command{
		Use:   "favorite",
		Short: "Manage favorite projects",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}
	command.AddCommand(newSetFavoriteCommand("add", true, statePath))
	command.AddCommand(newSetFavoriteCommand("remove", false, statePath))
	command.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List favorite project paths",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			path, err := statePath()
			if err != nil {
				return err
			}
			value, err := (appstate.Store{Path: path}).Load()
			if err != nil {
				return err
			}
			for _, favorite := range value.Favorites {
				if _, err := fmt.Fprintln(cmd.OutOrStdout(), favorite); err != nil {
					return err
				}
			}
			return nil
		},
	})
	return command
}

func newSetFavoriteCommand(use string, favorite bool, statePath statePathFunc) *cobra.Command {
	return &cobra.Command{
		Use:   use + " <project> [workspace]",
		Short: use + " a favorite project",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			found, err := resolveProject(cmd.Context(), args[0], args[1:])
			if err != nil {
				return err
			}
			path, err := statePath()
			if err != nil {
				return err
			}
			if err := (appstate.Store{Path: path}).SetFavorite(found.Path, favorite); err != nil {
				return fmt.Errorf("update favorite %q: %w", found.Name, err)
			}
			return nil
		},
	}
}
