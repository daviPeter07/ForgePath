package scanner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/daviPeter07/forgepath/internal/detector"
	gitinfo "github.com/daviPeter07/forgepath/internal/git"
	"github.com/daviPeter07/forgepath/internal/project"
)

var ignoredDirectories = map[string]struct{}{
	".git":         {},
	".idea":        {},
	".vscode":      {},
	"node_modules": {},
	"vendor":       {},
	".next":        {},
	"dist":         {},
	"build":        {},
	"target":       {},
	".venv":        {},
	"venv":         {},
}

func Scan(workspace string) ([]project.Project, error) {
	info, err := os.Stat(workspace)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("workspace %q is not a directory", workspace)
	}

	entries, err := os.ReadDir(workspace)
	if err != nil {
		return nil, err
	}

	projects := make([]project.Project, 0)
	for _, entry := range entries {
		if !entry.IsDir() || shouldIgnore(entry.Name()) {
			continue
		}

		path := filepath.Join(workspace, entry.Name())
		result, found, err := detector.Detect(path)
		if err != nil {
			return nil, fmt.Errorf("detect project %q: %w", path, err)
		}
		if !found {
			continue
		}

		foundProject := project.Project{
			Name:            entry.Name(),
			Path:            path,
			Technology:      result.Technology,
			Markers:         result.Markers,
			Frameworks:      result.Frameworks,
			PackageManagers: result.PackageManagers,
			HasDocker:       result.HasDocker,
		}
		projects = append(projects, foundProject)
	}
	enrichWithGit(projects)

	sort.Slice(projects, func(i, j int) bool {
		return projects[i].Name < projects[j].Name
	})

	return projects, nil
}

func enrichWithGit(projects []project.Project) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var waitGroup sync.WaitGroup
	limit := make(chan struct{}, 8)
	for index := range projects {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			select {
			case limit <- struct{}{}:
				defer func() { <-limit }()
			case <-ctx.Done():
				return
			}

			if info, repository := gitinfo.InspectContext(ctx, projects[index].Path); repository {
				projects[index].GitBranch = info.Branch
				projects[index].GitDirty = info.Dirty
				projects[index].GitStatusKnown = info.StatusKnown
			}
		}()
	}
	waitGroup.Wait()
}

func shouldIgnore(name string) bool {
	if strings.HasPrefix(name, ".") {
		return true
	}
	_, ignored := ignoredDirectories[name]
	return ignored
}
