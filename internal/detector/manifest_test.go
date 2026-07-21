package detector

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/daviPeter07/forgepath/internal/project"
)

func TestDetectTechnologies(t *testing.T) {
	tests := []struct {
		name       string
		files      []string
		technology project.Technology
		markers    []string
	}{
		{name: "Go", files: []string{"go.mod"}, technology: project.TechnologyGo, markers: []string{"go.mod"}},
		{name: "PHP", files: []string{"composer.json"}, technology: project.TechnologyPHP, markers: []string{"composer.json"}},
		{name: "Java with Maven", files: []string{"pom.xml"}, technology: project.TechnologyJava, markers: []string{"pom.xml"}},
		{name: "Java with Gradle", files: []string{"build.gradle"}, technology: project.TechnologyJava, markers: []string{"build.gradle"}},
		{name: "Java with Gradle Kotlin", files: []string{"build.gradle.kts"}, technology: project.TechnologyJava, markers: []string{"build.gradle.kts"}},
		{name: "Java with all markers", files: []string{"pom.xml", "build.gradle", "build.gradle.kts"}, technology: project.TechnologyJava, markers: []string{"pom.xml", "build.gradle", "build.gradle.kts"}},
		{name: "Python with pyproject", files: []string{"pyproject.toml"}, technology: project.TechnologyPython, markers: []string{"pyproject.toml"}},
		{name: "Python with requirements", files: []string{"requirements.txt"}, technology: project.TechnologyPython, markers: []string{"requirements.txt"}},
		{name: "Python with Pipfile", files: []string{"Pipfile"}, technology: project.TechnologyPython, markers: []string{"Pipfile"}},
		{name: "TypeScript", files: []string{"package.json", "tsconfig.json"}, technology: project.TechnologyTypeScript, markers: []string{"package.json", "tsconfig.json"}},
		{name: "JavaScript", files: []string{"package.json"}, technology: project.TechnologyJavaScript, markers: []string{"package.json"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			createFiles(t, dir, tt.files...)

			result, found, err := Detect(dir)
			if err != nil {
				t.Fatalf("Detect() error = %v", err)
			}
			if !found {
				t.Fatal("Detect() found = false, want true")
			}
			if result.Technology != tt.technology {
				t.Fatalf("Detect() technology = %q, want %q", result.Technology, tt.technology)
			}
			if !reflect.DeepEqual(result.Markers, tt.markers) {
				t.Fatalf("Detect() markers = %v, want %v", result.Markers, tt.markers)
			}
		})
	}
}

func TestDetectReturnsFirstMatchingRule(t *testing.T) {
	tests := []struct {
		name       string
		files      []string
		technology project.Technology
	}{
		{name: "Go before PHP", files: []string{"go.mod", "composer.json"}, technology: project.TechnologyGo},
		{name: "TypeScript before JavaScript", files: []string{"package.json", "tsconfig.json"}, technology: project.TechnologyTypeScript},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			createFiles(t, dir, tt.files...)

			result, found, err := Detect(dir)
			if err != nil {
				t.Fatalf("Detect() error = %v", err)
			}
			if !found {
				t.Fatal("Detect() found = false, want true")
			}
			if result.Technology != tt.technology {
				t.Fatalf("Detect() technology = %q, want %q", result.Technology, tt.technology)
			}
		})
	}
}

func TestDetectWithoutMarker(t *testing.T) {
	result, found, err := Detect(t.TempDir())
	if err != nil {
		t.Fatalf("Detect() error = %v", err)
	}
	if found {
		t.Fatalf("Detect() found = true, result = %+v", result)
	}
	if !reflect.DeepEqual(result, Result{}) {
		t.Fatalf("Detect() result = %+v, want zero value", result)
	}
}

func TestDetectIgnoresMarkerDirectory(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "go.mod"), 0o755); err != nil {
		t.Fatal(err)
	}

	_, found, err := Detect(dir)
	if err != nil {
		t.Fatalf("Detect() error = %v", err)
	}
	if found {
		t.Fatal("Detect() found = true, want false")
	}
}

func TestDetectReturnsFilesystemError(t *testing.T) {
	tests := []struct {
		name string
		path func(t *testing.T) string
	}{
		{name: "invalid path", path: func(_ *testing.T) string { return string([]byte{'\x00'}) }},
		{name: "missing path", path: func(t *testing.T) string { return filepath.Join(t.TempDir(), "missing") }},
		{name: "regular file", path: func(t *testing.T) string {
			path := filepath.Join(t.TempDir(), "file")
			if err := os.WriteFile(path, nil, 0o644); err != nil {
				t.Fatal(err)
			}
			return path
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, found, err := Detect(tt.path(t))
			if err == nil {
				t.Fatal("Detect() error = nil, want filesystem error")
			}
			if found {
				t.Fatal("Detect() found = true, want false")
			}
		})
	}
}

func createFiles(t *testing.T, dir string, names ...string) {
	t.Helper()
	for _, name := range names {
		if err := os.WriteFile(filepath.Join(dir, name), nil, 0o644); err != nil {
			t.Fatalf("create marker %q: %v", name, err)
		}
	}
}
