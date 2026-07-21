package detector

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/daviPeter07/forgepath/internal/project"
)

func TestDetectProjectMetadata(t *testing.T) {
	tests := []struct {
		name       string
		files      map[string]string
		frameworks []project.Framework
		managers   []project.PackageManager
		hasDocker  bool
	}{
		{
			name: "Next.js with pnpm and Docker",
			files: map[string]string{
				"package.json":   `{"dependencies":{"next":"latest","react":"latest"},"packageManager":"pnpm@10.0.0"}`,
				"tsconfig.json":  `{}`,
				"pnpm-lock.yaml": "",
				"Dockerfile":     "FROM node:alpine",
			},
			frameworks: []project.Framework{project.FrameworkNextJS},
			managers:   []project.PackageManager{project.PackageManagerPNPM},
			hasDocker:  true,
		},
		{
			name: "Laravel with Composer",
			files: map[string]string{
				"composer.json": `{"require":{"laravel/framework":"^12.0"}}`,
			},
			frameworks: []project.Framework{project.FrameworkLaravel},
			managers:   []project.PackageManager{project.PackageManagerComposer},
		},
		{
			name: "Spring Boot with Gradle",
			files: map[string]string{
				"build.gradle.kts": `plugins { id("org.springframework.boot") version "4.0.0" }`,
			},
			frameworks: []project.Framework{project.FrameworkSpringBoot},
			managers:   []project.PackageManager{project.PackageManagerGradle},
		},
		{
			name: "FastAPI with Poetry and Compose",
			files: map[string]string{
				"pyproject.toml": "[project]\ndependencies = [\"fastapi\"]",
				"poetry.lock":    "",
				"compose.yaml":   "services: {}",
			},
			frameworks: []project.Framework{project.FrameworkFastAPI},
			managers:   []project.PackageManager{project.PackageManagerPoetry},
			hasDocker:  true,
		},
		{
			name: "Go modules",
			files: map[string]string{
				"go.mod": "module example.com/app",
			},
			managers: []project.PackageManager{project.PackageManagerGoModules},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			writeProjectFiles(t, dir, tt.files)

			result, found, err := Detect(dir)
			if err != nil {
				t.Fatalf("Detect() error = %v", err)
			}
			if !found {
				t.Fatal("Detect() found = false, want true")
			}
			if !reflect.DeepEqual(result.Frameworks, tt.frameworks) {
				t.Fatalf("Detect() frameworks = %q, want %q", result.Frameworks, tt.frameworks)
			}
			if !reflect.DeepEqual(result.PackageManagers, tt.managers) {
				t.Fatalf("Detect() package managers = %q, want %q", result.PackageManagers, tt.managers)
			}
			if result.HasDocker != tt.hasDocker {
				t.Fatalf("Detect() has Docker = %t, want %t", result.HasDocker, tt.hasDocker)
			}
		})
	}
}

func TestDetectNodePackageManagers(t *testing.T) {
	tests := []struct {
		name        string
		packageJSON string
		marker      string
		managers    []project.PackageManager
	}{
		{name: "unknown without evidence", packageJSON: `{}`},
		{name: "npm lock", packageJSON: `{}`, marker: "package-lock.json", managers: []project.PackageManager{project.PackageManagerNPM}},
		{name: "pnpm lock", packageJSON: `{}`, marker: "pnpm-lock.yaml", managers: []project.PackageManager{project.PackageManagerPNPM}},
		{name: "Yarn lock", packageJSON: `{}`, marker: "yarn.lock", managers: []project.PackageManager{project.PackageManagerYarn}},
		{name: "Bun lock", packageJSON: `{}`, marker: "bun.lock", managers: []project.PackageManager{project.PackageManagerBun}},
		{name: "multiple evidence", packageJSON: `{"packageManager":"yarn@4.9.1"}`, marker: "package-lock.json", managers: []project.PackageManager{project.PackageManagerYarn, project.PackageManagerNPM}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			files := map[string]string{"package.json": tt.packageJSON}
			if tt.marker != "" {
				files[tt.marker] = ""
			}
			writeProjectFiles(t, dir, files)

			result, found, err := Detect(dir)
			if err != nil || !found {
				t.Fatalf("Detect() found = %t, error = %v", found, err)
			}
			if !reflect.DeepEqual(result.PackageManagers, tt.managers) {
				t.Fatalf("Detect() package managers = %q, want %q", result.PackageManagers, tt.managers)
			}
		})
	}
}

func TestDetectMultiEcosystemMetadata(t *testing.T) {
	dir := t.TempDir()
	writeProjectFiles(t, dir, map[string]string{
		"composer.json":  `{"require":{"laravel/framework":"^12.0"}}`,
		"package.json":   `{"dependencies":{"vue":"latest"}}`,
		"pnpm-lock.yaml": "",
	})

	result, found, err := Detect(dir)
	if err != nil || !found {
		t.Fatalf("Detect() found = %t, error = %v", found, err)
	}
	if result.Technology != project.TechnologyPHP {
		t.Fatalf("Detect() technology = %q, want PHP", result.Technology)
	}
	wantFrameworks := []project.Framework{project.FrameworkLaravel, project.FrameworkVue}
	if !reflect.DeepEqual(result.Frameworks, wantFrameworks) {
		t.Fatalf("Detect() frameworks = %q, want %q", result.Frameworks, wantFrameworks)
	}
	wantManagers := []project.PackageManager{project.PackageManagerComposer, project.PackageManagerPNPM}
	if !reflect.DeepEqual(result.PackageManagers, wantManagers) {
		t.Fatalf("Detect() package managers = %q, want %q", result.PackageManagers, wantManagers)
	}
}

func TestDetectAvoidsFrameworkSubstringFalsePositives(t *testing.T) {
	tests := []map[string]string{
		{"requirements.txt": "fastapi-users==14.0.0\n# fastapi==1.0.0"},
		{"pyproject.toml": "[project]\ndescription = \"client for FastAPI\""},
		{"build.gradle": "// id 'org.springframework.boot'\n/* id 'org.springframework.boot' */\nplugins {}"},
	}

	for _, files := range tests {
		dir := t.TempDir()
		writeProjectFiles(t, dir, files)
		result, found, err := Detect(dir)
		if err != nil || !found {
			t.Fatalf("Detect() found = %t, error = %v", found, err)
		}
		if len(result.Frameworks) != 0 {
			t.Fatalf("Detect() frameworks = %q, want none", result.Frameworks)
		}
	}
}

func TestDetectMultipleNodeFrameworks(t *testing.T) {
	dir := t.TempDir()
	writeProjectFiles(t, dir, map[string]string{
		"package.json": `{"dependencies":{"next":"latest","react":"latest","express":"latest"}}`,
	})

	result, found, err := Detect(dir)
	if err != nil || !found {
		t.Fatalf("Detect() found = %t, error = %v", found, err)
	}
	want := []project.Framework{project.FrameworkNextJS, project.FrameworkExpress}
	if !reflect.DeepEqual(result.Frameworks, want) {
		t.Fatalf("Detect() frameworks = %q, want %q", result.Frameworks, want)
	}
}

func TestDetectPreservesTechnologyWithInvalidMetadata(t *testing.T) {
	dir := t.TempDir()
	writeProjectFiles(t, dir, map[string]string{"package.json": "not JSON"})

	result, found, err := Detect(dir)
	if err != nil || !found {
		t.Fatalf("Detect() found = %t, error = %v", found, err)
	}
	if result.Technology != project.TechnologyJavaScript {
		t.Fatalf("Detect() technology = %q, want JavaScript", result.Technology)
	}
	if len(result.Frameworks) != 0 || len(result.PackageManagers) != 0 {
		t.Fatalf("Detect() metadata = %+v, want none", result)
	}
}

func TestDetectDockerOnlyProject(t *testing.T) {
	dir := t.TempDir()
	writeProjectFiles(t, dir, map[string]string{"compose.yaml": "services: {}"})

	result, found, err := Detect(dir)
	if err != nil || !found {
		t.Fatalf("Detect() found = %t, error = %v", found, err)
	}
	if result.Technology != project.TechnologyDocker || !result.HasDocker {
		t.Fatalf("Detect() result = %+v, want Docker project", result)
	}
}

func TestDetectIgnoresDockerMarkerDirectory(t *testing.T) {
	dir := t.TempDir()
	writeProjectFiles(t, dir, map[string]string{"go.mod": "module example.com/app"})
	if err := os.Mkdir(filepath.Join(dir, "Dockerfile"), 0o755); err != nil {
		t.Fatal(err)
	}

	result, found, err := Detect(dir)
	if err != nil || !found {
		t.Fatalf("Detect() found = %t, error = %v", found, err)
	}
	if result.HasDocker {
		t.Fatal("Detect() has Docker = true, want false")
	}
}

func writeProjectFiles(t *testing.T, dir string, files map[string]string) {
	t.Helper()
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatalf("write %q: %v", name, err)
		}
	}
}
