package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	appconfig "github.com/daviPeter07/forgepath/internal/config"
	"github.com/spf13/cobra"
)

type runCommandFunc func(context.Context, string, []string, io.Reader, io.Writer, io.Writer) error

func newRunCommand(configPath configPathFunc, runCommand runCommandFunc) *cobra.Command {
	return &cobra.Command{
		Use:   "run <project> [workspace]",
		Short: "Run a configured project command",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := configPath()
			if err != nil {
				return err
			}
			configuration, err := appconfig.Load(path)
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					return fmt.Errorf("load config %q (run forgepath config init to create it): %w", path, err)
				}
				return fmt.Errorf("load config %q: %w", path, err)
			}
			projectConfiguration, configured := configuration.Projects[args[0]]
			if !configured {
				return fmt.Errorf("project %q has no configured command", args[0])
			}

			found, err := resolveProject(cmd.Context(), args[0], args[1:])
			if err != nil {
				return err
			}
			if err := runCommand(
				cmd.Context(),
				found.Path,
				projectConfiguration.Command,
				cmd.InOrStdin(),
				cmd.OutOrStdout(),
				cmd.ErrOrStderr(),
			); err != nil {
				return fmt.Errorf("run command for %q: %w", found.Name, err)
			}
			return nil
		},
	}
}
