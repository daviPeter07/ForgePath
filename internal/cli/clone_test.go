package cli

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestParseRepository(t *testing.T) {
	tests := []struct {
		spec, url, name string
	}{
		{"daviPeter07/forgepath", "https://github.com/daviPeter07/forgepath.git", "forgepath"},
		{"owner/repo.git", "https://github.com/owner/repo.git", "repo"},
		{"https://gitlab.com/group/sub/app.git", "https://gitlab.com/group/sub/app.git", "app"},
		{"https://github.com/owner/site/", "https://github.com/owner/site/", "site"},
		{"git@github.com:owner/api.git", "git@github.com:owner/api.git", "api"},
		{"ssh://git@host.example/owner/worker", "ssh://git@host.example/owner/worker", "worker"},
		{"file:///srv/git/tools.git", "file:///srv/git/tools.git", "tools"},
	}
	for _, tt := range tests {
		url, name, err := parseRepository(tt.spec)
		if err != nil || url != tt.url || name != tt.name {
			t.Errorf("parseRepository(%q) = %q, %q, %v; want %q, %q", tt.spec, url, name, err, tt.url, tt.name)
		}
	}
	for _, invalid := range []string{"", "repo", "-upload-pack=evil/x", "owner/", "/repo", "owner/repo/extra", "ftp://host/repo", "owner/.."} {
		if _, _, err := parseRepository(invalid); err == nil {
			t.Errorf("parseRepository(%q) error = nil, want rejection", invalid)
		}
	}
}

type recordedClone struct{ url, destination string }

func cloneCommandFor(t *testing.T, configPath configPathFunc, calls *[]recordedClone) (*cobra.Command, *bytes.Buffer, *bytes.Buffer) {
	t.Helper()
	command := newCloneCommand(configPath, func(_ context.Context, url, destination string, _, _ io.Writer) error {
		*calls = append(*calls, recordedClone{url, destination})
		if err := os.MkdirAll(destination, 0o755); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(destination, "go.mod"), []byte("module x"), 0o644)
	})
	var output, errors bytes.Buffer
	command.SetOut(&output)
	command.SetErr(&errors)
	return command, &output, &errors
}

func TestCloneIntoFirstConfiguredWorkspace(t *testing.T) {
	work, clients := t.TempDir(), t.TempDir()
	var calls []recordedClone
	command, output, errors := cloneCommandFor(t, writeWorkspaceConfig(t, "", work, clients), &calls)
	command.SetArgs([]string{"owner/api"})
	if err := command.Execute(); err != nil {
		t.Fatalf("clone error = %v", err)
	}
	want := filepath.Join(work, "api")
	if len(calls) != 1 || calls[0].url != "https://github.com/owner/api.git" || calls[0].destination != want {
		t.Fatalf("clone calls = %+v, want %s", calls, want)
	}
	if strings.TrimSpace(output.String()) != want {
		t.Fatalf("stdout = %q, want the new project path", output.String())
	}
	if !strings.Contains(errors.String(), "Go project") {
		t.Fatalf("stderr = %q, want the detected technology", errors.String())
	}
}

func TestCloneWorkspaceFlagAndName(t *testing.T) {
	work, other := t.TempDir(), t.TempDir()
	var calls []recordedClone
	command, _, _ := cloneCommandFor(t, writeWorkspaceConfig(t, "", work), &calls)
	command.SetArgs([]string{"owner/api", "--workspace", other, "--name", "api-v2"})
	if err := command.Execute(); err != nil {
		t.Fatalf("clone error = %v", err)
	}
	if len(calls) != 1 || calls[0].destination != filepath.Join(other, "api-v2") {
		t.Fatalf("clone calls = %+v", calls)
	}
}

func TestCloneRefusesExistingFolder(t *testing.T) {
	work := t.TempDir()
	if err := os.Mkdir(filepath.Join(work, "api"), 0o755); err != nil {
		t.Fatal(err)
	}
	var calls []recordedClone
	command, _, _ := cloneCommandFor(t, writeWorkspaceConfig(t, "", work), &calls)
	command.SetArgs([]string{"owner/api"})
	err := command.Execute()
	if err == nil || !strings.Contains(err.Error(), "already exists") || len(calls) != 0 {
		t.Fatalf("error = %v, calls = %+v; want a refusal before cloning", err, calls)
	}
}

func TestCloneRunsGit(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}
	root := t.TempDir()
	source := filepath.Join(root, "source")
	for _, args := range [][]string{
		{"init", "-q", "-b", "main", source},
		{"-C", source, "-c", "user.name=t", "-c", "user.email=t@example.com", "-c", "commit.gpgsign=false", "commit", "-q", "--allow-empty", "-m", "init"},
	} {
		if output, err := exec.Command("git", args...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, output)
		}
	}
	destination := filepath.Join(root, "workspace", "copy")
	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := gitClone(context.Background(), "file://"+source, destination, io.Discard, io.Discard); err != nil {
		t.Fatalf("gitClone() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(destination, ".git")); err != nil {
		t.Fatalf("clone has no .git: %v", err)
	}
}
