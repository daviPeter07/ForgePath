package scanner

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/daviPeter07/forgepath/internal/detector"
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

		projects = append(projects, project.Project{
			Name:       entry.Name(),
			Path:       path,
			Technology: result.Technology,
			Markers:    result.Markers,
		})
	}

	sort.Slice(projects, func(i, j int) bool {
		return projects[i].Name < projects[j].Name
	})

	return projects, nil
}

func shouldIgnore(name string) bool {
	if strings.HasPrefix(name, ".") {
		return true
	}
	_, ignored := ignoredDirectories[name]
	return ignored
}
