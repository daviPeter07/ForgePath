package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/daviPeter07/forgepath/internal/scanner"
	"github.com/spf13/cobra"
)

func newPickCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "pick [workspace]",
		Short: "Select a project and print its path",
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
			if len(projects) == 0 {
				return fmt.Errorf("no projects found in %q", workspace)
			}

			for index, found := range projects {
				if _, err := fmt.Fprintf(cmd.ErrOrStderr(), "%d. %s (%s)\n", index+1, found.Name, found.Technology); err != nil {
					return err
				}
			}
			if _, err := fmt.Fprint(cmd.ErrOrStderr(), "Select a project (Enter to cancel): "); err != nil {
				return err
			}

			selection, err := bufio.NewReader(cmd.InOrStdin()).ReadString('\n')
			if err != nil && !errors.Is(err, io.EOF) {
				return err
			}
			selection = strings.TrimSpace(selection)
			if selection == "" {
				return nil
			}

			selected, err := strconv.Atoi(selection)
			if err != nil || selected < 1 || selected > len(projects) {
				return fmt.Errorf("invalid selection %q", selection)
			}

			_, err = fmt.Fprintln(cmd.OutOrStdout(), projects[selected-1].Path)
			return err
		},
	}
	command.Flags().Bool("print-path", false, "print only the selected project path")

	return command
}
