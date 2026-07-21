package state

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/daviPeter07/forgepath/internal/project"
)

const (
	currentVersion = 1
	maxRecent      = 20
	lockTimeout    = 2 * time.Second
)

type State struct {
	Version   int      `json:"version"`
	Favorites []string `json:"favorites"`
	Recent    []Recent `json:"recent"`
}

type Recent struct {
	Path     string    `json:"path"`
	OpenedAt time.Time `json:"openedAt"`
}

type Store struct {
	Path string
	Now  func() time.Time
}

func DefaultPath() (string, error) {
	directory, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(directory, "forgepath", "state.json"), nil
}

func (store Store) Load() (State, error) {
	data, err := os.ReadFile(store.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return defaultState(), nil
		}
		return State{}, err
	}
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 || trimmed[0] != '{' {
		return State{}, fmt.Errorf("decode state %q: root must be a JSON object", store.Path)
	}

	var value State
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		return State{}, fmt.Errorf("decode state %q: %w", store.Path, err)
	}
	var extra any
	if err := decoder.Decode(&extra); err == nil {
		return State{}, fmt.Errorf("decode state %q: multiple JSON values", store.Path)
	} else if !errors.Is(err, io.EOF) {
		return State{}, fmt.Errorf("decode state %q: %w", store.Path, err)
	}
	if value.Version != currentVersion {
		return State{}, fmt.Errorf("state %q has unsupported version %d", store.Path, value.Version)
	}
	return value, nil
}

func (store Store) Save(value State) error {
	value.Version = currentVersion
	if err := os.MkdirAll(filepath.Dir(store.Path), 0o755); err != nil {
		return err
	}

	temporary, err := os.CreateTemp(filepath.Dir(store.Path), ".state-*.json")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := temporary.Chmod(0o600); err != nil {
		temporary.Close()
		return err
	}

	encoder := json.NewEncoder(temporary)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(value); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Sync(); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	return replaceFile(temporaryPath, store.Path)
}

func (store Store) SetFavorite(path string, favorite bool) error {
	path, err := normalizePath(path)
	if err != nil {
		return err
	}
	return store.update(func(value *State) {
		index := pathIndex(value.Favorites, path)
		if favorite && index < 0 {
			value.Favorites = append(value.Favorites, path)
			sort.Strings(value.Favorites)
		}
		if !favorite && index >= 0 {
			value.Favorites = append(value.Favorites[:index], value.Favorites[index+1:]...)
		}
	})
}

func (store Store) RecordRecent(path string) error {
	path, err := normalizePath(path)
	if err != nil {
		return err
	}
	return store.update(func(value *State) {
		recent := Recent{Path: path, OpenedAt: store.now()}
		updated := []Recent{recent}
		for _, existing := range value.Recent {
			if pathKey(existing.Path) != pathKey(path) && len(updated) < maxRecent {
				updated = append(updated, existing)
			}
		}
		value.Recent = updated
	})
}

func (store Store) Decorate(projects []project.Project) (State, error) {
	value, err := store.Load()
	if err != nil {
		return State{}, err
	}

	favorites := make(map[string]struct{}, len(value.Favorites))
	for _, path := range value.Favorites {
		favorites[pathKey(path)] = struct{}{}
	}
	recent := make(map[string]time.Time, len(value.Recent))
	recentRank := make(map[string]int, len(value.Recent))
	for rank, entry := range value.Recent {
		key := pathKey(entry.Path)
		recent[key] = entry.OpenedAt
		recentRank[key] = rank
	}
	for index := range projects {
		key := pathKey(projects[index].Path)
		_, projects[index].Favorite = favorites[key]
		projects[index].LastOpened = recent[key]
	}
	sort.SliceStable(projects, func(i, j int) bool {
		if projects[i].Favorite != projects[j].Favorite {
			return projects[i].Favorite
		}
		if !projects[i].LastOpened.Equal(projects[j].LastOpened) {
			return projects[i].LastOpened.After(projects[j].LastOpened)
		}
		iRank, iRecent := recentRank[pathKey(projects[i].Path)]
		jRank, jRecent := recentRank[pathKey(projects[j].Path)]
		if iRecent != jRecent {
			return iRecent
		}
		if iRecent && iRank != jRank {
			return iRank < jRank
		}
		return projects[i].Name < projects[j].Name
	})
	return value, nil
}

func defaultState() State {
	return State{Version: currentVersion, Favorites: []string{}, Recent: []Recent{}}
}

func (store Store) now() time.Time {
	if store.Now != nil {
		return store.Now().UTC()
	}
	return time.Now().UTC()
}

func (store Store) update(mutate func(*State)) error {
	unlock, err := store.lock()
	if err != nil {
		return err
	}
	defer unlock()

	value, err := store.Load()
	if err != nil {
		return err
	}
	mutate(&value)
	return store.Save(value)
}

func (store Store) lock() (func(), error) {
	if err := os.MkdirAll(filepath.Dir(store.Path), 0o755); err != nil {
		return nil, err
	}
	lockPath := store.Path + ".lock"
	file, err := os.OpenFile(lockPath, os.O_RDWR|os.O_CREATE, 0o600)
	if err != nil {
		return nil, err
	}
	deadline := time.Now().Add(lockTimeout)
	for {
		acquired, err := tryLockFile(file)
		if err != nil {
			file.Close()
			return nil, err
		}
		if acquired {
			return func() {
				_ = unlockFile(file)
				_ = file.Close()
			}, nil
		}
		if time.Now().After(deadline) {
			file.Close()
			return nil, fmt.Errorf("state %q is locked by another process", store.Path)
		}
		time.Sleep(25 * time.Millisecond)
	}
}

func normalizePath(path string) (string, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	if resolved, resolveErr := filepath.EvalSymlinks(absolute); resolveErr == nil {
		absolute = resolved
	}
	return filepath.Clean(absolute), nil
}

func pathKey(path string) string {
	cleaned, err := normalizePath(path)
	if err != nil {
		cleaned = filepath.Clean(path)
	}
	if runtime.GOOS == "windows" {
		return strings.ToLower(cleaned)
	}
	return cleaned
}

func pathIndex(values []string, target string) int {
	targetKey := pathKey(target)
	for index, value := range values {
		if pathKey(value) == targetKey {
			return index
		}
	}
	return -1
}
