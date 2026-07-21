package cli

import (
	"fmt"
	"path/filepath"

	"github.com/daviPeter07/forgepath/internal/catalog"
	"github.com/daviPeter07/forgepath/internal/project"
	"github.com/daviPeter07/forgepath/internal/scanner"
	"github.com/spf13/cobra"
)

type cachePathFunc func() (string, error)
type scanProjectsFunc func(string, bool) (catalog.Result, error)

func resolveCachePath(configured string) (string, error) {
	if configured == "" {
		return catalog.DefaultDirectory()
	}
	return filepath.Abs(configured)
}

func directProjectScan(workspace string, _ bool) (catalog.Result, error) {
	projects, err := scanner.Scan(workspace)
	return catalog.Result{Projects: projects}, err
}

func scanProjects(cmd *cobra.Command, scan scanProjectsFunc, workspace string, refresh bool) ([]project.Project, error) {
	result, err := scan(workspace, refresh)
	if err != nil {
		return nil, err
	}
	if result.Warning != nil {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "warning: project cache unavailable: %v\n", result.Warning)
	}
	return result.Projects, nil
}
