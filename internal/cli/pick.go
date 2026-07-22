package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/daviPeter07/forgepath/internal/action"
	"github.com/daviPeter07/forgepath/internal/icon"
	"github.com/daviPeter07/forgepath/internal/ide"
	"github.com/daviPeter07/forgepath/internal/project"
	"github.com/daviPeter07/forgepath/internal/tui"
	"github.com/spf13/cobra"
)

func newPickCommand(statePaths ...statePathFunc) *cobra.Command {
	return newPickCommandWithScanner(directProjectScan, statePaths...)
}

func newPickCommandWithScanner(scan scanProjectsFunc, statePaths ...statePathFunc) *cobra.Command {
	var statePath statePathFunc
	if len(statePaths) > 0 {
		statePath = statePaths[0]
	}
	var iconMode string
	var refresh bool
	command := &cobra.Command{
		Use:   "pick [workspace]",
		Short: "Select a project and print its path",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			icons, err := icon.ParseMode(iconMode)
			if err != nil {
				return err
			}
			return runPicker(cmd, args, icons, refresh, scan, statePath)
		},
	}
	command.Flags().Bool("print-path", false, "print only the selected project path")
	command.Flags().StringVar(&iconMode, "icons", string(icon.ModeAuto), "icon mode: auto, graphics, ascii, or nerd-font")
	command.Flags().BoolVar(&refresh, "refresh", false, "ignore and rebuild the project cache")

	return command
}

func newConfiguredPickCommand(scan scanProjectsFunc, statePath statePathFunc, configPath configPathFunc) *cobra.Command {
	command := newPickCommandWithScanner(scan, statePath)
	command.RunE = func(cmd *cobra.Command, args []string) error {
		icons, err := icon.ParseMode(cmd.Flag("icons").Value.String())
		if err != nil {
			return err
		}
		refresh, err := cmd.Flags().GetBool("refresh")
		if err != nil {
			return err
		}
		return runPicker(cmd, args, icons, refresh, scan, statePath, configPath)
	}
	return command
}

func runPicker(cmd *cobra.Command, args []string, icons icon.Mode, refresh bool, scan scanProjectsFunc, statePath statePathFunc, configPaths ...configPathFunc) error {
	icons = icon.ResolveMode(icons, cmd.ErrOrStderr())
	var configPath configPathFunc
	if len(configPaths) > 0 {
		configPath = configPaths[0]
	}
	workspaces, err := configuredWorkspaces(args, configPath)
	if err != nil {
		return err
	}
	projects, err := scanConfiguredProjects(cmd, scan, workspaces, refresh)
	if err != nil {
		return err
	}
	decorateProjectsBestEffort(cmd, statePath, projects)

	startPath := ""
	if len(workspaces) > 0 {
		startPath = workspaces[0]
	}

	selected, found, err := tui.SelectWithOptions(cmd.Context(), projects, tui.Options{
		Icons:     icons,
		IDEs:      ide.Discover(),
		StartPath: startPath,
		OpenEditor: func(ctx context.Context, path string, selectedProject project.Project, editor ide.IDE) error {
			if err := action.OpenEditorWithArguments(ctx, path, editor.Executable, editor.Arguments); err != nil {
				return err
			}
			recordRecentBestEffort(cmd, statePath, selectedProject.Path)
			return nil
		},
	}, cmd.InOrStdin(), cmd.ErrOrStderr())
	if err != nil {
		return err
	}
	if !found {
		return nil
	}
	printPath := false
	if flag, err := cmd.Flags().GetBool("print-path"); err == nil && flag {
		printPath = true
	} else if !isTerminalWriter(cmd.OutOrStdout()) {
		printPath = true
	}

	if printPath {
		if _, err = fmt.Fprintln(cmd.OutOrStdout(), selected.Path); err != nil {
			return err
		}
	} else {
		shell := os.Getenv("SHELL")
		if shell == "" {
			if runtime.GOOS == "windows" {
				shell = os.Getenv("COMSPEC")
				if shell == "" {
					shell = "powershell.exe"
				}
			} else {
				shell = "bash"
			}
		}

		fmt.Fprintf(cmd.ErrOrStderr(), "Changing directory to %s...\n", selected.Path)
		shellCmd := exec.CommandContext(cmd.Context(), shell)
		shellCmd.Dir = selected.Path
		shellCmd.Stdin = os.Stdin
		shellCmd.Stdout = os.Stdout
		shellCmd.Stderr = os.Stderr

		if err := shellCmd.Run(); err != nil {
			return err
		}
	}
	recordRecentBestEffort(cmd, statePath, projectRootForPath(projects, selected.Path))
	return nil
}

func isTerminalWriter(out io.Writer) bool {
	file, ok := out.(*os.File)
	if !ok {
		return false
	}
	info, err := file.Stat()
	return err == nil && (info.Mode()&os.ModeCharDevice) != 0
}

func projectRootForPath(projects []project.Project, path string) string {
	best := path
	bestLength := 0
	for _, candidate := range projects {
		relative, err := filepath.Rel(candidate.Path, path)
		if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			continue
		}
		if len(candidate.Path) > bestLength {
			best = candidate.Path
			bestLength = len(candidate.Path)
		}
	}
	return best
}
