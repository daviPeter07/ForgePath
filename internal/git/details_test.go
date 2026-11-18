package git

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func commitFile(t *testing.T, dir, name, message string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(message), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, dir, "add", name)
	runGit(t, dir, "-c", "user.name=ForgePath Tests", "-c", "user.email=forgepath@example.com", "-c", "commit.gpgsign=false", "commit", "-q", "-m", message)
}

func TestLoadDetailsListsRecentCommitsNewestFirst(t *testing.T) {
	requireGit(t)
	dir := t.TempDir()
	runGit(t, dir, "init", "-q", "-b", "main")
	for index, message := range []string{"first", "second", "third", "fourth"} {
		commitFile(t, dir, "file"+string(rune('a'+index)), message)
	}

	details, err := LoadDetails(context.Background(), dir)
	if err != nil {
		t.Fatalf("LoadDetails() error = %v", err)
	}
	if len(details.Commits) != RecentCommits {
		t.Fatalf("commits = %d, want %d", len(details.Commits), RecentCommits)
	}
	var subjects []string
	for _, commit := range details.Commits {
		subjects = append(subjects, commit.Subject)
		if len(commit.Hash) < 7 || commit.When.IsZero() || time.Since(commit.When) > time.Hour {
			t.Fatalf("commit = %+v, want a short hash and a recent time", commit)
		}
	}
	if strings.Join(subjects, ",") != "fourth,third,second" {
		t.Fatalf("subjects = %q, want newest first", subjects)
	}
	if details.Upstream != "" {
		t.Fatalf("Upstream = %q, want none for a repository without a remote", details.Upstream)
	}
}

func TestLoadDetailsCountsAheadAndBehind(t *testing.T) {
	requireGit(t)
	root := t.TempDir()
	remote := filepath.Join(root, "remote.git")
	runGit(t, root, "init", "-q", "--bare", "-b", "main", remote)

	seed := filepath.Join(root, "seed")
	runGit(t, root, "clone", "-q", remote, seed)
	runGit(t, seed, "checkout", "-q", "-b", "main")
	commitFile(t, seed, "a", "base")
	runGit(t, seed, "push", "-q", "origin", "main")

	work := filepath.Join(root, "work")
	runGit(t, root, "clone", "-q", remote, work)

	commitFile(t, seed, "b", "from someone else")
	runGit(t, seed, "push", "-q", "origin", "main")
	commitFile(t, work, "c", "mine one")
	commitFile(t, work, "d", "mine two")
	runGit(t, work, "fetch", "-q")

	details, err := LoadDetails(context.Background(), work)
	if err != nil {
		t.Fatalf("LoadDetails() error = %v", err)
	}
	if details.Upstream != "origin/main" || details.Ahead != 2 || details.Behind != 1 {
		t.Fatalf("details = upstream %q ahead %d behind %d, want origin/main 2/1", details.Upstream, details.Ahead, details.Behind)
	}
}

func TestLoadDetailsRejectsNonRepository(t *testing.T) {
	requireGit(t)
	if _, err := LoadDetails(context.Background(), t.TempDir()); err == nil {
		t.Fatal("LoadDetails(non-repository) error = nil")
	}
}

func TestLoadDetailsOfEmptyRepository(t *testing.T) {
	requireGit(t)
	dir := t.TempDir()
	runGit(t, dir, "init", "-q", "-b", "main")
	details, err := LoadDetails(context.Background(), dir)
	if err != nil {
		t.Fatalf("LoadDetails(empty) error = %v", err)
	}
	if len(details.Commits) != 0 {
		t.Fatalf("commits = %+v, want none", details.Commits)
	}
}
