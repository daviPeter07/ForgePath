package cli

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/daviPeter07/forgepath/internal/catalog"
	"github.com/daviPeter07/forgepath/internal/project"
)

func TestListRefreshBypassesCache(t *testing.T) {
	workspace := t.TempDir()
	var refreshed bool
	command := newListCommandWithScanner(func(path string, refresh bool) (catalog.Result, error) {
		refreshed = refresh
		return catalog.Result{Projects: []project.Project{{Name: "app", Path: filepath.Join(path, "app"), Technology: project.TechnologyGo}}}, nil
	})
	command.SetOut(&bytes.Buffer{})
	command.SetArgs([]string{workspace, "--refresh"})

	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !refreshed {
		t.Fatal("refresh = false, want true")
	}
}

func TestScanCommandAlwaysRefreshes(t *testing.T) {
	workspace := t.TempDir()
	var refreshed bool
	var stdout bytes.Buffer
	command := newScanCommand(func(_ string, refresh bool) (catalog.Result, error) {
		refreshed = refresh
		return catalog.Result{Projects: []project.Project{{Name: "one"}, {Name: "two"}}}, nil
	})
	command.SetOut(&stdout)
	command.SetArgs([]string{workspace})

	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !refreshed {
		t.Fatal("refresh = false, want true")
	}
	if strings.TrimSpace(stdout.String()) != "2 projects cached" {
		t.Fatalf("stdout = %q", stdout.String())
	}
}

func TestCacheWarningsUseStderr(t *testing.T) {
	workspace := t.TempDir()
	warning := bytes.ErrTooLarge
	var stderr bytes.Buffer
	command := newListCommandWithScanner(func(string, bool) (catalog.Result, error) {
		return catalog.Result{Warning: warning}, nil
	})
	command.SetOut(&bytes.Buffer{})
	command.SetErr(&stderr)
	command.SetArgs([]string{workspace})

	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !strings.Contains(stderr.String(), "warning:") {
		t.Fatalf("stderr = %q, want warning", stderr.String())
	}
}

func TestRootUsesCacheEnvironmentAsFlagDefault(t *testing.T) {
	configured := filepath.Join(t.TempDir(), "cache")
	t.Setenv("FORGEPATH_CACHE", configured)
	command := NewRootCommand(&bytes.Buffer{}, &bytes.Buffer{})

	flag := command.PersistentFlags().Lookup("cache")
	if flag == nil || flag.Value.String() != configured {
		t.Fatalf("cache flag = %v, want environment path %q", flag, configured)
	}
}
