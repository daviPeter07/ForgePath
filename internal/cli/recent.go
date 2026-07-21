package cli

import (
	"fmt"
	"time"

	appstate "github.com/daviPeter07/forgepath/internal/state"
	"github.com/spf13/cobra"
)

func newRecentCommand(statePath statePathFunc) *cobra.Command {
	return &cobra.Command{
		Use:   "recent",
		Short: "List recently used projects",
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
			for _, recent := range value.Recent {
				if _, err := fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\n", recent.OpenedAt.Format(time.RFC3339), recent.Path); err != nil {
					return err
				}
			}
			return nil
		},
	}
}
