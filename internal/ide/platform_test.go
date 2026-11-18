package ide

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/daviPeter07/forgepath/internal/project"
)

func candidateByID(candidates []candidate, id string) (candidate, bool) {
	for _, found := range candidates {
		if found.ID == id {
			return found, true
		}
	}
	return candidate{}, false
}

func TestDarwinCandidatesIncludeApplicationsAndXcode(t *testing.T) {
	candidates := platformCandidates("darwin", func(name string) string {
		if name == "HOME" {
			return "/Users/dev"
		}
		return ""
	})

	xcode, ok := candidateByID(candidates, "xcode")
	if !ok {
		t.Fatal("darwin candidates missing Xcode")
	}
	if !reflect.DeepEqual(xcode.requires, []string{"/Applications/Xcode.app"}) || !reflect.DeepEqual(xcode.Arguments, []string{"-a", "Xcode"}) {
		t.Fatalf("Xcode candidate = %+v", xcode)
	}

	goland, _ := candidateByID(candidates, "goland")
	if len(goland.paths) != 3 || !strings.Contains(goland.paths[2], filepath.Join("JetBrains", "Toolbox")) {
		t.Fatalf("GoLand paths = %q, want /Applications, ~/Applications and Toolbox", goland.paths)
	}
	vscode, _ := candidateByID(candidates, "vscode")
	if len(vscode.paths) != 2 || !strings.HasPrefix(vscode.paths[1], "/Users/dev") {
		t.Fatalf("VS Code paths = %q, want only the two Applications folders", vscode.paths)
	}
}

func TestWindowsCandidatesUseEnvironmentFolders(t *testing.T) {
	environment := map[string]string{"LOCALAPPDATA": `C:\Users\dev\AppData\Local`, "ProgramFiles": `C:\Program Files`}
	candidates := platformCandidates("windows", func(name string) string { return environment[name] })

	vscode, _ := candidateByID(candidates, "vscode")
	if len(vscode.paths) != 2 || !strings.HasPrefix(vscode.paths[0], environment["LOCALAPPDATA"]) || !strings.HasPrefix(vscode.paths[1], environment["ProgramFiles"]) {
		t.Fatalf("VS Code paths = %q", vscode.paths)
	}
	if _, ok := candidateByID(candidates, "xcode"); ok {
		t.Fatal("Xcode must only be offered on macOS")
	}
}

func TestDiscoverFindsFlatpakEditorOnLinux(t *testing.T) {
	home := t.TempDir()
	installed := filepath.Join(home, ".local", "share", "flatpak", "app", "com.visualstudio.code")
	if err := os.MkdirAll(installed, 0o755); err != nil {
		t.Fatal(err)
	}

	found := discover(finder{
		goos: "linux",
		lookPath: func(command string) (string, error) {
			if command == "flatpak" {
				return "/usr/bin/flatpak", nil
			}
			return "", errors.New("not found")
		},
		glob:   filepath.Glob,
		stat:   os.Stat,
		getenv: func(name string) string { return map[string]string{"HOME": home}[name] },
	})

	if len(found) != 1 || found[0].ID != "vscode" {
		t.Fatalf("discover() = %+v, want only the Flatpak VS Code", found)
	}
	if found[0].Executable != "/usr/bin/flatpak" || !reflect.DeepEqual(found[0].Arguments, []string{"run", "com.visualstudio.code"}) {
		t.Fatalf("Flatpak editor = %+v", found[0])
	}
}

func TestDiscoverSkipsFlatpakEditorThatIsNotInstalled(t *testing.T) {
	found := discover(finder{
		goos: "linux",
		lookPath: func(command string) (string, error) {
			if command == "flatpak" {
				return "/usr/bin/flatpak", nil
			}
			return "", errors.New("not found")
		},
		glob:   func(string) ([]string, error) { return nil, nil },
		stat:   func(string) (os.FileInfo, error) { return nil, os.ErrNotExist },
		getenv: func(string) string { return "/home/dev" },
	})
	if len(found) != 0 {
		t.Fatalf("discover() = %+v, want nothing when no Flatpak app is installed", found)
	}
}

func TestAnyRequirementMetExpandsGlobs(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "apps", "IDEA-U"), 0o755); err != nil {
		t.Fatal(err)
	}
	system := finder{glob: filepath.Glob, stat: os.Stat}

	if !anyRequirementMet(system, nil) {
		t.Fatal("no requirements should be satisfied")
	}
	if !anyRequirementMet(system, []string{filepath.Join(root, "missing"), filepath.Join(root, "apps", "IDEA-*")}) {
		t.Fatal("glob requirement was not matched")
	}
	if anyRequirementMet(system, []string{filepath.Join(root, "apps", "PyCharm-*")}) {
		t.Fatal("unmatched glob reported as installed")
	}
}

func TestRankWithoutTechnologyKeepsEveryEditorAlphabetically(t *testing.T) {
	installed := []IDE{
		{ID: "zed", Name: "Zed", Technologies: allTechnologies},
		{ID: "goland", Name: "GoLand", Technologies: []project.Technology{project.TechnologyGo}},
		{ID: "cursor", Name: "Cursor", Technologies: allTechnologies},
	}
	ranked := Rank(installed, "")
	names := make([]string, len(ranked))
	for index, editor := range ranked {
		names[index] = editor.Name
	}
	if !reflect.DeepEqual(names, []string{"Cursor", "GoLand", "Zed"}) {
		t.Fatalf("Rank(\"\") = %q, want all editors alphabetically", names)
	}
}

func TestFindExecutablePrefersNewestToolboxVersion(t *testing.T) {
	root := t.TempDir()
	var created []string
	for _, version := range []string{"2024.3", "2025.2"} {
		path := filepath.Join(root, "GoLand", version, "bin", "goland")
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, nil, 0o755); err != nil {
			t.Fatal(err)
		}
		created = append(created, path)
	}
	system := finder{
		goos:     "linux",
		lookPath: func(string) (string, error) { return "", errors.New("not found") },
		glob:     filepath.Glob,
		stat:     os.Stat,
	}
	got := findExecutable(system, candidate{paths: []string{filepath.Join(root, "GoLand", "*", "bin", "goland")}})
	if got != created[1] {
		t.Fatalf("findExecutable() = %q, want newest %q", got, created[1])
	}
}
