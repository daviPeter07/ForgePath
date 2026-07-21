package cli

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	appconfig "github.com/daviPeter07/forgepath/internal/config"
)

func TestConfigInitCommand(t *testing.T) {
	path := filepath.Join(t.TempDir(), "forgepath", "config.json")
	var stdout bytes.Buffer
	command := newConfigCommand(func() (string, error) { return path, nil })
	command.SetOut(&stdout)
	command.SetArgs([]string{"init"})

	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if strings.TrimSpace(stdout.String()) != path {
		t.Fatalf("stdout = %q, want config path", stdout.String())
	}
	if _, err := appconfig.Load(path); err != nil {
		t.Fatalf("Load() initialized config error = %v", err)
	}
}

func TestRootConfigFlag(t *testing.T) {
	path := filepath.Join(t.TempDir(), "custom.json")
	command := NewRootCommand(&bytes.Buffer{}, &bytes.Buffer{})
	command.SetArgs([]string{"config", "init", "--config", path})

	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if _, err := appconfig.Load(path); err != nil {
		t.Fatalf("Load() custom config error = %v", err)
	}
}
