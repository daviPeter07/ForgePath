package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestInspectRepository(t *testing.T) {
	requireGit(t)
	dir := t.TempDir()
	runGit(t, dir, "init", "-b", "main")
	runGit(t, dir, "config", "user.name", "ForgePath Tests")
	runGit(t, dir, "config", "user.email", "forgepath@example.com")
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/app"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, dir, "add", "go.mod")
	runGit(t, dir, "commit", "-m", "initial")

	info, found := Inspect(dir)
	if !found {
		t.Fatal("Inspect() found = false, want true")
	}
	if info.Branch != "main" || info.Dirty || !info.StatusKnown {
		t.Fatalf("Inspect() = %+v, want clean main branch", info)
	}

	if err := os.WriteFile(filepath.Join(dir, "untracked.txt"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	info, found = Inspect(dir)
	if !found || !info.Dirty || !info.StatusKnown {
		t.Fatalf("Inspect() = %+v, found = %t, want dirty repository", info, found)
	}
}

func TestInspectNonRepository(t *testing.T) {
	requireGit(t)
	if _, found := Inspect(t.TempDir()); found {
		t.Fatal("Inspect() found = true, want false")
	}
}

func TestInspectDetachedHead(t *testing.T) {
	requireGit(t)
	dir := t.TempDir()
	runGit(t, dir, "init", "-b", "main")
	runGit(t, dir, "config", "user.name", "ForgePath Tests")
	runGit(t, dir, "config", "user.email", "forgepath@example.com")
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/app"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, dir, "add", "go.mod")
	runGit(t, dir, "commit", "-m", "initial")
	runGit(t, dir, "checkout", "--detach")

	info, found := Inspect(dir)
	if !found || len(info.Branch) <= len("detached@") || info.Branch[:len("detached@")] != "detached@" {
		t.Fatalf("Inspect() = %+v, found = %t, want detached commit", info, found)
	}
}

func requireGit(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}
}

func runGit(t *testing.T, dir string, arguments ...string) {
	t.Helper()
	args := append([]string{"-C", dir}, arguments...)
	if output, err := exec.Command("git", args...).CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", arguments, err, output)
	}
}
