package cli

import (
	"fmt"
	"path/filepath"

	appconfig "github.com/daviPeter07/forgepath/internal/config"
	"github.com/spf13/cobra"
)

type configPathFunc func() (string, error)

func newConfigCommand(configPath configPathFunc) *cobra.Command {
	command := &cobra.Command{
		Use:   "config",
		Short: "Manage ForgePath configuration",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}
	command.AddCommand(&cobra.Command{
		Use:   "init",
		Short: "Create the default configuration file",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			path, err := configPath()
			if err != nil {
				return err
			}
			if err := appconfig.Init(path); err != nil {
				return fmt.Errorf("initialize config %q: %w", path, err)
			}
			_, err = fmt.Fprintln(cmd.OutOrStdout(), path)
			return err
		},
	})
	return command
}

func resolveConfigPath(configured string) (string, error) {
	if configured == "" {
		return appconfig.DefaultPath()
	}
	return filepath.Abs(configured)
}
