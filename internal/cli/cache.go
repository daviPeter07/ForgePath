package cli

import (
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

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

func scanConfiguredProjects(cmd *cobra.Command, scan scanProjectsFunc, workspaces []string, refresh bool) ([]project.Project, error) {
	var projects []project.Project
	var failures []error
	successful := 0
	seen := make(map[string]struct{})
	for _, workspace := range workspaces {
		found, err := scanProjects(cmd, scan, workspace, refresh)
		if err != nil {
			failure := fmt.Errorf("scan workspace %q: %w", workspace, err)
			if len(workspaces) == 1 {
				return nil, failure
			}
			failures = append(failures, failure)
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "warning: %v\n", failure)
			continue
		}
		successful++
		for _, candidate := range found {
			key := filepath.Clean(candidate.Path)
			if runtime.GOOS == "windows" {
				key = strings.ToLower(key)
			}
			if _, exists := seen[key]; exists {
				continue
			}
			seen[key] = struct{}{}
			projects = append(projects, candidate)
		}
	}
	if successful == 0 && len(failures) > 0 {
		return nil, errors.Join(failures...)
	}
	return projects, nil
}
