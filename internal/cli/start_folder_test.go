package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func chdir(t *testing.T, dir string) {
	t.Helper()
	original, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(original) })
}

func runRoot(t *testing.T, input string, args ...string) (string, error) {
	t.Helper()
	var stdout bytes.Buffer
	command := NewRootCommand(&stdout, &bytes.Buffer{})
	command.SetIn(strings.NewReader(input))
	command.SetArgs(append(args, "--icons", "ascii", "--cache", t.TempDir()))
	err := command.Execute()
	return stdout.String(), err
}

func TestRootStartsInTheCurrentFolder(t *testing.T) {
	work := t.TempDir()
	createCLIProject(t, work, "api", "go.mod")
	configPath := writeWorkspaceConfig(t, "", t.TempDir()) // a catalog elsewhere must not win
	path, _ := configPath()
	chdir(t, work)

	out, err := runRoot(t, "c", "--config", path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(out) != work {
		t.Fatalf("fg + c printed %q, want the current folder %q", out, work)
	}
}

func TestRootOpensTheGivenFolder(t *testing.T) {
	target := t.TempDir()
	createCLIProject(t, target, "only", "go.mod")
	chdir(t, t.TempDir())

	out, err := runRoot(t, "\rc", target)
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(out) != filepath.Join(target, "only") {
		t.Fatalf("fg <folder> + enter + c printed %q, want %q", out, filepath.Join(target, "only"))
	}
}

func TestRootRejectsMissingFolder(t *testing.T) {
	_, err := runRoot(t, "q", filepath.Join(t.TempDir(), "missing"))
	if err == nil {
		t.Fatal("fg <missing folder> error = nil")
	}
	file := filepath.Join(t.TempDir(), "file.txt")
	if err := os.WriteFile(file, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := runRoot(t, "q", file); err == nil || !strings.Contains(err.Error(), "not a directory") {
		t.Fatalf("fg <file> error = %v, want not a directory", err)
	}
}

func TestRootProjectsFlagOpensTheProjectList(t *testing.T) {
	catalog := t.TempDir()
	createCLIProject(t, catalog, "catalog-app", "go.mod")
	configPath := writeWorkspaceConfig(t, "", catalog)
	path, _ := configPath()
	chdir(t, t.TempDir())

	out, err := runRoot(t, "c", "--projects", "--config", path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(out) != filepath.Join(catalog, "catalog-app") {
		t.Fatalf("fg --projects + c printed %q, want the catalog project", out)
	}
}

func TestPickAcceptsFolderAndProjects(t *testing.T) {
	for _, args := range [][]string{{"--help"}, {"pick", "--help"}} {
		var output bytes.Buffer
		command := NewRootCommand(&output, &bytes.Buffer{})
		command.SetArgs(args)
		if err := command.Execute(); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(output.String(), "--projects") {
			t.Fatalf("%v help does not list --projects:\n%s", args, output.String())
		}
	}
}
