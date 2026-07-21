package catalog

import (
	"crypto/sha256"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/daviPeter07/forgepath/internal/project"
)

func TestStoreCachesScans(t *testing.T) {
	workspace := t.TempDir()
	now := time.Date(2026, 7, 21, 20, 0, 0, 0, time.UTC)
	scans := 0
	store := Store{
		Directory: t.TempDir(),
		MaxAge:    time.Minute,
		Now:       func() time.Time { return now },
		Scanner: func(path string) ([]project.Project, error) {
			scans++
			return []project.Project{{Name: "app", Path: filepath.Join(path, "app")}}, nil
		},
	}

	first, err := store.Scan(workspace, false)
	if err != nil {
		t.Fatal(err)
	}
	if first.Hit || first.Warning != nil || scans != 1 {
		t.Fatalf("first scan = %+v, scans = %d", first, scans)
	}
	second, err := store.Scan(workspace, false)
	if err != nil {
		t.Fatal(err)
	}
	if !second.Hit || scans != 1 || len(second.Projects) != 1 {
		t.Fatalf("second scan = %+v, scans = %d", second, scans)
	}

	if _, err := store.Scan(workspace, true); err != nil {
		t.Fatal(err)
	}
	if scans != 2 {
		t.Fatalf("scans after refresh = %d, want 2", scans)
	}
}

func TestStoreInvalidatesExpiredAndChangedWorkspace(t *testing.T) {
	workspace := t.TempDir()
	now := time.Date(2026, 7, 21, 20, 0, 0, 0, time.UTC)
	scans := 0
	store := Store{
		Directory: t.TempDir(),
		MaxAge:    time.Minute,
		Now:       func() time.Time { return now },
		Scanner: func(string) ([]project.Project, error) {
			scans++
			return []project.Project{}, nil
		},
	}
	if _, err := store.Scan(workspace, false); err != nil {
		t.Fatal(err)
	}

	now = now.Add(2 * time.Minute)
	if _, err := store.Scan(workspace, false); err != nil {
		t.Fatal(err)
	}
	if scans != 2 {
		t.Fatalf("expired cache scans = %d, want 2", scans)
	}

	if _, err := store.Scan(workspace, false); err != nil {
		t.Fatal(err)
	}
	modified := time.Now().Add(2 * time.Second)
	if err := os.Chtimes(workspace, modified, modified); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Scan(workspace, false); err != nil {
		t.Fatal(err)
	}
	if scans != 3 {
		t.Fatalf("changed workspace scans = %d, want 3", scans)
	}
}

func TestStoreInvalidatesChangedProjectManifest(t *testing.T) {
	workspace := t.TempDir()
	projectDirectory := filepath.Join(workspace, "app")
	if err := os.Mkdir(projectDirectory, 0o755); err != nil {
		t.Fatal(err)
	}
	scans := 0
	store := Store{
		Directory: t.TempDir(),
		MaxAge:    time.Minute,
		Scanner: func(string) ([]project.Project, error) {
			scans++
			return []project.Project{}, nil
		},
	}
	if _, err := store.Scan(workspace, false); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(projectDirectory, "package.json"), []byte(`{"dependencies":{"react":"latest"}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Scan(workspace, false); err != nil {
		t.Fatal(err)
	}
	if scans != 2 {
		t.Fatalf("scans = %d, want manifest change to invalidate cache", scans)
	}
}

func TestOlderConcurrentScanCannotOverwriteChangedWorkspace(t *testing.T) {
	workspace := t.TempDir()
	projectDirectory := filepath.Join(workspace, "app")
	if err := os.Mkdir(projectDirectory, 0o755); err != nil {
		t.Fatal(err)
	}
	firstStarted := make(chan struct{})
	releaseFirst := make(chan struct{})
	calls := 0
	store := Store{
		Directory: t.TempDir(),
		MaxAge:    time.Minute,
		Scanner: func(string) ([]project.Project, error) {
			calls++
			if calls == 1 {
				close(firstStarted)
				<-releaseFirst
				return []project.Project{{Name: "old"}}, nil
			}
			return []project.Project{{Name: "new"}}, nil
		},
	}

	firstResult := make(chan Result, 1)
	firstError := make(chan error, 1)
	go func() {
		result, err := store.Scan(workspace, true)
		firstResult <- result
		firstError <- err
	}()
	<-firstStarted
	if err := os.WriteFile(filepath.Join(projectDirectory, "go.mod"), []byte("module example.com/app"), 0o644); err != nil {
		t.Fatal(err)
	}
	newer, err := store.Scan(workspace, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(newer.Projects) != 1 || newer.Projects[0].Name != "new" {
		t.Fatalf("newer scan = %+v", newer)
	}
	close(releaseFirst)
	older := <-firstResult
	if err := <-firstError; err != nil {
		t.Fatal(err)
	}
	if older.Warning == nil {
		t.Fatal("older scan warning = nil, want changed-workspace warning")
	}

	cached, err := store.Scan(workspace, false)
	if err != nil {
		t.Fatal(err)
	}
	if !cached.Hit || cached.Projects[0].Name != "new" {
		t.Fatalf("cached result = %+v, want newer scan", cached)
	}
}

func TestStoreRecoversFromCorruptCache(t *testing.T) {
	workspace := t.TempDir()
	store := Store{Directory: t.TempDir(), Scanner: func(string) ([]project.Project, error) {
		return []project.Project{{Name: "app"}}, nil
	}}
	path, err := canonicalWorkspace(workspace)
	if err != nil {
		t.Fatal(err)
	}
	cachePath := store.cachePath(path)
	if err := os.MkdirAll(filepath.Dir(cachePath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(cachePath, []byte("broken"), 0o600); err != nil {
		t.Fatal(err)
	}

	result, err := store.Scan(workspace, false)
	if err != nil {
		t.Fatal(err)
	}
	if result.Hit || len(result.Projects) != 1 || result.Warning == nil {
		t.Fatalf("result = %+v, want recovered scan with warning", result)
	}
}

func TestStoreReturnsScanErrorsAndOnlyWarnsForCacheWrites(t *testing.T) {
	workspace := t.TempDir()
	cacheFile := filepath.Join(t.TempDir(), "not-a-directory")
	if err := os.WriteFile(cacheFile, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	store := Store{Directory: cacheFile, Scanner: func(string) ([]project.Project, error) {
		return []project.Project{{Name: "app"}}, nil
	}}
	result, err := store.Scan(workspace, false)
	if err != nil {
		t.Fatal(err)
	}
	if result.Warning == nil || len(result.Projects) != 1 {
		t.Fatalf("result = %+v, want projects with cache warning", result)
	}

	wantErr := errors.New("scan failed")
	store.Directory = t.TempDir()
	store.Scanner = func(string) ([]project.Project, error) { return nil, wantErr }
	if _, err := store.Scan(workspace, true); !errors.Is(err, wantErr) {
		t.Fatalf("Scan() error = %v, want scan error", err)
	}
}

func TestCacheHitRefreshesGitStatus(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}
	workspace := t.TempDir()
	projectPath := filepath.Join(workspace, "app")
	if err := os.Mkdir(projectPath, 0o755); err != nil {
		t.Fatal(err)
	}
	runCatalogGit(t, projectPath, "init", "-b", "main")
	runCatalogGit(t, projectPath, "config", "user.name", "ForgePath Tests")
	runCatalogGit(t, projectPath, "config", "user.email", "forgepath@example.com")
	if err := os.WriteFile(filepath.Join(projectPath, "go.mod"), []byte("module example.com/app"), 0o644); err != nil {
		t.Fatal(err)
	}
	runCatalogGit(t, projectPath, "add", "go.mod")
	runCatalogGit(t, projectPath, "commit", "-m", "initial")

	store := Store{Directory: t.TempDir(), Scanner: func(string) ([]project.Project, error) {
		return []project.Project{{Name: "app", Path: projectPath, Technology: project.TechnologyGo}}, nil
	}}
	if _, err := store.Scan(workspace, false); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(projectPath, "untracked.txt"), nil, 0o644); err != nil {
		t.Fatal(err)
	}

	result, err := store.Scan(workspace, false)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Hit || !result.Projects[0].GitDirty || result.Projects[0].GitBranch != "main" {
		t.Fatalf("cached Git metadata = %+v", result)
	}
}

func TestCleanupOnlyRemovesCacheFiles(t *testing.T) {
	directory := t.TempDir()
	store := Store{Directory: directory, Now: func() time.Time { return time.Now().UTC() }}
	old := time.Now().Add(-maxCacheLife - time.Hour)
	userFile := filepath.Join(directory, "package.json")
	cacheFile := filepath.Join(directory, strings.Repeat("a", sha256.Size*2)+".json")
	for _, path := range []string{userFile, cacheFile} {
		if err := os.WriteFile(path, nil, 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.Chtimes(path, old, old); err != nil {
			t.Fatal(err)
		}
	}

	store.cleanup()
	if _, err := os.Stat(userFile); err != nil {
		t.Fatalf("user JSON was removed: %v", err)
	}
	if _, err := os.Stat(cacheFile); !os.IsNotExist(err) {
		t.Fatalf("old cache still exists: %v", err)
	}
}

func TestWriteRejectsOversizedEntry(t *testing.T) {
	store := Store{Directory: t.TempDir()}
	cached := entry{Version: cacheVersion, Projects: []cachedProject{{Name: strings.Repeat("x", maxCacheSize)}}}
	if err := store.write(filepath.Join(store.Directory, strings.Repeat("a", 64)+".json"), cached); err == nil {
		t.Fatal("write() error = nil, want size limit error")
	}
}

func runCatalogGit(t *testing.T, directory string, arguments ...string) {
	t.Helper()
	args := append([]string{"-C", directory}, arguments...)
	if output, err := exec.Command("git", args...).CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", arguments, err, output)
	}
}
