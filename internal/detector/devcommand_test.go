package detector

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/daviPeter07/forgepath/internal/project"
)

func TestDevCommand(t *testing.T) {
	tests := []struct {
		name     string
		files    map[string]string
		project  project.Project
		expected []string
	}{
		{
			name:     "pnpm dev script",
			files:    map[string]string{"package.json": `{"scripts":{"dev":"next dev","start":"next start"}}`},
			project:  project.Project{Technology: project.TechnologyTypeScript, PackageManagers: []project.PackageManager{project.PackageManagerPNPM}},
			expected: []string{"pnpm", "dev"},
		},
		{
			name:     "npm falls back to start",
			files:    map[string]string{"package.json": `{"scripts":{"start":"node index.js"}}`},
			project:  project.Project{Technology: project.TechnologyJavaScript, PackageManagers: []project.PackageManager{project.PackageManagerNPM}},
			expected: []string{"npm", "run", "start"},
		},
		{
			name:     "npm is the default manager",
			files:    map[string]string{"package.json": `{"scripts":{"dev":"vite"}}`},
			project:  project.Project{Technology: project.TechnologyJavaScript},
			expected: []string{"npm", "run", "dev"},
		},
		{
			name:     "yarn and bun",
			files:    map[string]string{"package.json": `{"scripts":{"dev":"vite"}}`},
			project:  project.Project{Technology: project.TechnologyJavaScript, PackageManagers: []project.PackageManager{project.PackageManagerBun}},
			expected: []string{"bun", "run", "dev"},
		},
		{
			name:     "Laravel serves with artisan before the node scripts",
			files:    map[string]string{"artisan": "", "composer.json": "{}", "package.json": `{"scripts":{"dev":"vite"}}`},
			project:  project.Project{Technology: project.TechnologyPHP, Frameworks: []project.Framework{project.FrameworkLaravel}},
			expected: []string{"php", "artisan", "serve"},
		},
		{
			name:     "Go main package at the root",
			files:    map[string]string{"go.mod": "module app", "main.go": "package main"},
			project:  project.Project{Technology: project.TechnologyGo},
			expected: []string{"go", "run", "."},
		},
		{
			name:     "Go single command",
			files:    map[string]string{"go.mod": "module app", "cmd/api/main.go": "package main"},
			project:  project.Project{Technology: project.TechnologyGo},
			expected: []string{"go", "run", "./cmd/api"},
		},
		{
			name:    "Go library has nothing to run",
			files:   map[string]string{"go.mod": "module app", "lib.go": "package lib"},
			project: project.Project{Technology: project.TechnologyGo},
		},
		{
			name:     "Rust",
			files:    map[string]string{"Cargo.toml": "[package]"},
			project:  project.Project{Technology: project.TechnologyRust},
			expected: []string{"cargo", "run"},
		},
		{
			name:     "Django",
			files:    map[string]string{"manage.py": "", "requirements.txt": "django"},
			project:  project.Project{Technology: project.TechnologyPython},
			expected: []string{"python", "manage.py", "runserver"},
		},
		{
			name:     "FastAPI app in main.py",
			files:    map[string]string{"main.py": "from fastapi import FastAPI\napp = FastAPI()\n", "requirements.txt": "fastapi"},
			project:  project.Project{Technology: project.TechnologyPython, Frameworks: []project.Framework{project.FrameworkFastAPI}},
			expected: []string{"uvicorn", "main:app", "--reload"},
		},
		{
			name:     "Phoenix",
			files:    map[string]string{"mix.exs": "{:phoenix, \"~> 1.7\"}"},
			project:  project.Project{Technology: project.TechnologyElixir},
			expected: []string{"mix", "phx.server"},
		},
		{
			name:     "Rails",
			files:    map[string]string{"Gemfile": "gem 'rails'", "bin/rails": ""},
			project:  project.Project{Technology: project.TechnologyRuby},
			expected: []string{"ruby", "bin/rails", "server"},
		},
		{
			name:     "Docker Compose",
			files:    map[string]string{"compose.yaml": "services: {}"},
			project:  project.Project{Technology: project.TechnologyDocker, HasDocker: true},
			expected: []string{"docker", "compose", "up"},
		},
		{
			name:    "nothing recognizable",
			files:   map[string]string{"README.md": ""},
			project: project.Project{Technology: project.TechnologyPython},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			for name, content := range tt.files {
				path := filepath.Join(dir, filepath.FromSlash(name))
				if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			tt.project.Path = dir
			if got := DevCommand(tt.project); !reflect.DeepEqual(got, tt.expected) {
				t.Fatalf("DevCommand() = %q, want %q", got, tt.expected)
			}
		})
	}
}
