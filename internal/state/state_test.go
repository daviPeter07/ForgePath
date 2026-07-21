package state

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/daviPeter07/forgepath/internal/project"
)

func TestStoreFavoritesAndRecent(t *testing.T) {
	openedAt := time.Date(2026, 7, 21, 20, 0, 0, 0, time.UTC)
	projectsDirectory := t.TempDir()
	webPath := filepath.Join(projectsDirectory, "web")
	apiPath := filepath.Join(projectsDirectory, "api")
	store := Store{
		Path: filepath.Join(t.TempDir(), "forgepath", "state.json"),
		Now:  func() time.Time { return openedAt },
	}

	if err := store.SetFavorite(webPath, true); err != nil {
		t.Fatal(err)
	}
	if err := store.SetFavorite(apiPath, true); err != nil {
		t.Fatal(err)
	}
	if err := store.SetFavorite(webPath, true); err != nil {
		t.Fatal(err)
	}
	if err := store.RecordRecent(webPath); err != nil {
		t.Fatal(err)
	}

	value, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	wantFavorites := []string{apiPath, webPath}
	sort.Strings(wantFavorites)
	if !reflect.DeepEqual(value.Favorites, wantFavorites) {
		t.Fatalf("favorites = %q, want %q", value.Favorites, wantFavorites)
	}
	if len(value.Recent) != 1 || value.Recent[0].Path != webPath || !value.Recent[0].OpenedAt.Equal(openedAt) {
		t.Fatalf("recent = %+v, want recorded project", value.Recent)
	}

	if err := store.SetFavorite(webPath, false); err != nil {
		t.Fatal(err)
	}
	value, err = store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(value.Favorites, []string{apiPath}) {
		t.Fatalf("favorites after removal = %q", value.Favorites)
	}
}

func TestStoreLimitsAndDeduplicatesRecent(t *testing.T) {
	now := time.Date(2026, 7, 21, 20, 0, 0, 0, time.UTC)
	store := Store{Path: filepath.Join(t.TempDir(), "state.json"), Now: func() time.Time { return now }}
	for index := 0; index < maxRecent+5; index++ {
		path := filepath.Join("projects", string(rune('a'+index)))
		if err := store.RecordRecent(path); err != nil {
			t.Fatal(err)
		}
	}
	if err := store.RecordRecent(filepath.Join("projects", "z")); err != nil {
		t.Fatal(err)
	}

	value, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(value.Recent) != maxRecent {
		t.Fatalf("len(recent) = %d, want %d", len(value.Recent), maxRecent)
	}
	wantRecent, err := normalizePath(filepath.Join("projects", "z"))
	if err != nil {
		t.Fatal(err)
	}
	if value.Recent[0].Path != wantRecent {
		t.Fatalf("most recent = %q, want projects/z", value.Recent[0].Path)
	}
}

func TestStoreSerializesConcurrentUpdates(t *testing.T) {
	store := Store{Path: filepath.Join(t.TempDir(), "state.json")}
	projectsDirectory := t.TempDir()

	var waitGroup sync.WaitGroup
	errors := make(chan error, 12)
	for index := 0; index < 12; index++ {
		path := filepath.Join(projectsDirectory, string(rune('a'+index)))
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			errors <- store.SetFavorite(path, true)
		}()
	}
	waitGroup.Wait()
	close(errors)
	for err := range errors {
		if err != nil {
			t.Fatalf("concurrent SetFavorite() error = %v", err)
		}
	}

	value, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(value.Favorites) != 12 {
		t.Fatalf("len(favorites) = %d, want 12", len(value.Favorites))
	}
}

func TestNativeLockIsReleasedWhenHandleCloses(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json.lock")
	first, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	acquired, err := tryLockFile(first)
	if err != nil || !acquired {
		t.Fatalf("first lock acquired = %t, error = %v", acquired, err)
	}
	if err := first.Close(); err != nil {
		t.Fatal(err)
	}

	second, err := os.OpenFile(path, os.O_RDWR, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()
	acquired, err = tryLockFile(second)
	if err != nil || !acquired {
		t.Fatalf("second lock acquired = %t, error = %v", acquired, err)
	}
	_ = unlockFile(second)
}

func TestStoreDecoratesAndSortsProjects(t *testing.T) {
	store := Store{Path: filepath.Join(t.TempDir(), "state.json")}
	value := defaultState()
	value.Favorites = []string{"/projects/zeta"}
	value.Recent = []Recent{{Path: "/projects/middle", OpenedAt: time.Now().UTC()}}
	if err := store.Save(value); err != nil {
		t.Fatal(err)
	}

	projects := []project.Project{
		{Name: "alpha", Path: "/projects/alpha"},
		{Name: "middle", Path: "/projects/middle"},
		{Name: "zeta", Path: "/projects/zeta"},
	}
	if _, err := store.Decorate(projects); err != nil {
		t.Fatal(err)
	}
	if projects[0].Name != "zeta" || !projects[0].Favorite {
		t.Fatalf("first project = %+v, want favorite zeta", projects[0])
	}
	if projects[1].Name != "middle" || projects[1].LastOpened.IsZero() {
		t.Fatalf("second project = %+v, want recent middle", projects[1])
	}
}

func TestStoreDecoratesProjectThroughSymlink(t *testing.T) {
	root := t.TempDir()
	realProject := filepath.Join(root, "real-project")
	if err := os.Mkdir(realProject, 0o755); err != nil {
		t.Fatal(err)
	}
	alias := filepath.Join(root, "project-alias")
	if err := os.Symlink(realProject, alias); err != nil {
		t.Skipf("symlinks are unavailable: %v", err)
	}

	store := Store{Path: filepath.Join(t.TempDir(), "state.json")}
	if err := store.SetFavorite(alias, true); err != nil {
		t.Fatal(err)
	}
	projects := []project.Project{{Name: "app", Path: alias}}
	if _, err := store.Decorate(projects); err != nil {
		t.Fatal(err)
	}
	if !projects[0].Favorite {
		t.Fatal("Favorite = false for project accessed through symlink")
	}
}
