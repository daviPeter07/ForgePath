package detector

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/daviPeter07/forgepath/internal/project"
)

type rule struct {
	technology   project.Technology
	required     []string
	alternatives []string
	excluded     []string
}

var rules = []rule{
	{technology: project.TechnologyGo, required: []string{"go.mod"}},
	{technology: project.TechnologyPHP, required: []string{"composer.json"}},
	{technology: project.TechnologyJava, alternatives: []string{"pom.xml", "build.gradle", "build.gradle.kts"}},
	{technology: project.TechnologyPython, alternatives: []string{"pyproject.toml", "requirements.txt", "Pipfile"}},
	{technology: project.TechnologyTypeScript, required: []string{"package.json", "tsconfig.json"}},
	{technology: project.TechnologyJavaScript, required: []string{"package.json"}, excluded: []string{"tsconfig.json"}},
	{technology: project.TechnologyDocker, alternatives: []string{"Dockerfile", "compose.yaml", "compose.yml", "docker-compose.yml", "docker-compose.yaml"}},
}

func Detect(path string) (Result, bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return Result{}, false, err
	}
	if !info.IsDir() {
		return Result{}, false, fmt.Errorf("project path %q is not a directory", path)
	}

	for _, candidate := range rules {
		markers, matched, err := match(path, candidate)
		if err != nil {
			return Result{}, false, err
		}
		if matched {
			metadata := detectMetadata(path)
			return Result{
				Technology:      candidate.technology,
				Markers:         markers,
				Frameworks:      metadata.frameworks,
				PackageManagers: metadata.packageManagers,
				HasDocker:       metadata.hasDocker,
			}, true, nil
		}
	}

	return Result{}, false, nil
}

func match(path string, candidate rule) ([]string, bool, error) {
	markers := make([]string, 0, len(candidate.required)+len(candidate.alternatives))

	for _, marker := range candidate.required {
		exists, err := isFile(filepath.Join(path, marker))
		if err != nil {
			return nil, false, err
		}
		if !exists {
			return nil, false, nil
		}
		markers = append(markers, marker)
	}

	for _, marker := range candidate.excluded {
		exists, err := isFile(filepath.Join(path, marker))
		if err != nil {
			return nil, false, err
		}
		if exists {
			return nil, false, nil
		}
	}

	foundAlternative := len(candidate.alternatives) == 0
	for _, marker := range candidate.alternatives {
		exists, err := isFile(filepath.Join(path, marker))
		if err != nil {
			return nil, false, err
		}
		if exists {
			foundAlternative = true
			markers = append(markers, marker)
		}
	}

	return markers, foundAlternative, nil
}

func isFile(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, err
	}

	return info.Mode().IsRegular(), nil
}
