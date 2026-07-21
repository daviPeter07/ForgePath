package cli

import (
	"fmt"
	"path/filepath"

	"github.com/daviPeter07/forgepath/internal/project"
	appstate "github.com/daviPeter07/forgepath/internal/state"
	"github.com/spf13/cobra"
)

type statePathFunc func() (string, error)

func resolveStatePath(configured string) (string, error) {
	if configured == "" {
		return appstate.DefaultPath()
	}
	return filepath.Abs(configured)
}

func decorateProjects(statePath statePathFunc, projects []project.Project) error {
	if statePath == nil {
		return nil
	}
	path, err := statePath()
	if err != nil {
		return err
	}
	_, err = (appstate.Store{Path: path}).Decorate(projects)
	return err
}

func recordRecent(statePath statePathFunc, path string) error {
	if statePath == nil {
		return nil
	}
	storePath, err := statePath()
	if err != nil {
		return err
	}
	return (appstate.Store{Path: storePath}).RecordRecent(path)
}

func decorateProjectsBestEffort(cmd *cobra.Command, statePath statePathFunc, projects []project.Project) {
	if err := decorateProjects(statePath, projects); err != nil {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "warning: could not load favorites and recent projects: %v\n", err)
	}
}

func recordRecentBestEffort(cmd *cobra.Command, statePath statePathFunc, path string) {
	if err := recordRecent(statePath, path); err != nil {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "warning: could not record recent project: %v\n", err)
	}
}
