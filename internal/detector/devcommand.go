package detector

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/daviPeter07/forgepath/internal/project"
)

// DevCommand suggests the command that starts a project's development
// environment, as an argument list, or nil when nothing obvious applies.
// Framework-specific servers win over generic package scripts, and Docker
// Compose is the last resort. Suggestions are never run without the user
// confirming them.
func DevCommand(found project.Project) []string {
	path := found.Path
	switch {
	case metadataFileExists(filepath.Join(path, "artisan")):
		return []string{"php", "artisan", "serve"}
	case metadataFileExists(filepath.Join(path, "manage.py")):
		return []string{"python", "manage.py", "runserver"}
	case metadataFileExists(filepath.Join(path, "bin", "rails")):
		return []string{"ruby", "bin/rails", "server"}
	case fileContains(filepath.Join(path, "mix.exs"), ":phoenix"):
		return []string{"mix", "phx.server"}
	case fileContains(filepath.Join(path, "main.py"), "FastAPI("):
		return []string{"uvicorn", "main:app", "--reload"}
	}
	if command := packageScript(path, found.PackageManagers); command != nil {
		return command
	}
	if metadataFileExists(filepath.Join(path, "go.mod")) {
		if command := goRunTarget(path); command != nil {
			return command
		}
	}
	if metadataFileExists(filepath.Join(path, "Cargo.toml")) {
		return []string{"cargo", "run"}
	}
	if hasAnyMetadataFile(path, []string{"compose.yaml", "compose.yml", "docker-compose.yml", "docker-compose.yaml"}) {
		return []string{"docker", "compose", "up"}
	}
	return nil
}

// packageScript runs the "dev" script, else "start", with the project's
// package manager (npm when none was detected).
func packageScript(path string, managers []project.PackageManager) []string {
	data, ok := readMetadataFile(filepath.Join(path, "package.json"))
	if !ok {
		return nil
	}
	var manifest struct {
		Scripts map[string]string `json:"scripts"`
	}
	if json.Unmarshal(data, &manifest) != nil {
		return nil
	}
	script := ""
	for _, candidate := range []string{"dev", "start"} {
		if _, ok := manifest.Scripts[candidate]; ok {
			script = candidate
			break
		}
	}
	if script == "" {
		return nil
	}
	manager := project.PackageManagerNPM
	for _, preferred := range []project.PackageManager{project.PackageManagerPNPM, project.PackageManagerYarn, project.PackageManagerBun, project.PackageManagerNPM} {
		if containsManager(managers, preferred) {
			manager = preferred
			break
		}
	}
	switch manager {
	case project.PackageManagerPNPM:
		return []string{"pnpm", script}
	case project.PackageManagerYarn:
		return []string{"yarn", script}
	case project.PackageManagerBun:
		return []string{"bun", "run", script}
	default:
		return []string{"npm", "run", script}
	}
}

// goRunTarget runs the main package at the root, or the only command under
// cmd/. A library or several commands have no single obvious target.
func goRunTarget(path string) []string {
	if fileContains(filepath.Join(path, "main.go"), "package main") {
		return []string{"go", "run", "."}
	}
	entries, err := os.ReadDir(filepath.Join(path, "cmd"))
	if err != nil {
		return nil
	}
	var commands []string
	for _, entry := range entries {
		if entry.IsDir() && fileContains(filepath.Join(path, "cmd", entry.Name(), "main.go"), "package main") {
			commands = append(commands, entry.Name())
		}
	}
	if len(commands) != 1 {
		return nil
	}
	return []string{"go", "run", "./cmd/" + commands[0]}
}

func fileContains(path, text string) bool {
	data, ok := readMetadataFile(path)
	return ok && strings.Contains(string(data), text)
}

func containsManager(managers []project.PackageManager, wanted project.PackageManager) bool {
	for _, manager := range managers {
		if manager == wanted {
			return true
		}
	}
	return false
}
