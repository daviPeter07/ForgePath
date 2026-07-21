package catalog

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/daviPeter07/forgepath/internal/project"
	"github.com/daviPeter07/forgepath/internal/scanner"
)

const (
	cacheVersion = 1
	defaultAge   = 30 * time.Second
	maxCacheSize = 10 << 20
	maxFileHash  = 2 << 20
	maxTotalHash = 8 << 20
	maxCacheLife = 30 * 24 * time.Hour
)

type ScanFunc func(string) ([]project.Project, error)

type Store struct {
	Directory string
	MaxAge    time.Duration
	Now       func() time.Time
	Scanner   ScanFunc
}

type Result struct {
	Projects []project.Project
	Hit      bool
	Warning  error
}

type entry struct {
	Version             int             `json:"version"`
	Workspace           string          `json:"workspace"`
	ScannedAt           time.Time       `json:"scannedAt"`
	WorkspaceModifiedAt time.Time       `json:"workspaceModifiedAt"`
	Fingerprint         string          `json:"fingerprint"`
	Projects            []cachedProject `json:"projects"`
}

type cachedProject struct {
	Name            string                   `json:"name"`
	Path            string                   `json:"path"`
	Technology      project.Technology       `json:"technology"`
	Markers         []string                 `json:"markers,omitempty"`
	Frameworks      []project.Framework      `json:"frameworks,omitempty"`
	PackageManagers []project.PackageManager `json:"packageManagers,omitempty"`
	HasDocker       bool                     `json:"hasDocker,omitempty"`
}

var fingerprintFiles = []string{
	"go.mod", "composer.json", "pom.xml", "build.gradle", "build.gradle.kts",
	"pyproject.toml", "requirements.txt", "Pipfile", "package.json", "tsconfig.json",
	"Dockerfile", "compose.yaml", "compose.yml", "docker-compose.yml", "docker-compose.yaml",
	"bun.lock", "bun.lockb", "pnpm-lock.yaml", "yarn.lock", "package-lock.json",
	"uv.lock", "poetry.lock",
}

var contentFingerprintFiles = map[string]struct{}{
	"go.mod": {}, "composer.json": {}, "pom.xml": {}, "build.gradle": {}, "build.gradle.kts": {},
	"pyproject.toml": {}, "requirements.txt": {}, "Pipfile": {}, "package.json": {}, "tsconfig.json": {},
}

func DefaultDirectory() (string, error) {
	directory, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(directory, "forgepath"), nil
}

func (store Store) Scan(workspace string, refresh bool) (Result, error) {
	workspace, err := canonicalWorkspace(workspace)
	if err != nil {
		return Result{}, err
	}
	info, err := os.Stat(workspace)
	if err != nil {
		return Result{}, err
	}
	if !info.IsDir() {
		return Result{}, fmt.Errorf("workspace %q is not a directory", workspace)
	}

	cachePath := store.cachePath(workspace)
	fingerprint, fingerprintWarning := workspaceFingerprint(workspace, info)
	var readWarning error
	if !refresh && fingerprintWarning == nil {
		cached, readErr := store.read(cachePath)
		if readErr == nil && store.valid(cached, workspace, info.ModTime(), fingerprint) {
			projects := projectsFromCache(cached.Projects)
			scanner.EnrichGit(projects)
			return Result{Projects: projects, Hit: true}, nil
		}
		if readErr != nil && !os.IsNotExist(readErr) {
			readWarning = readErr
		}
	}

	scan := store.Scanner
	if scan == nil {
		scan = scanner.Scan
	}
	projects, err := scan(workspace)
	if err != nil {
		return Result{}, err
	}
	latestInfo, err := os.Stat(workspace)
	if err != nil {
		return Result{}, err
	}
	latestFingerprint, latestFingerprintWarning := workspaceFingerprint(workspace, latestInfo)
	if fingerprintWarning != nil || latestFingerprintWarning != nil {
		return Result{Projects: projects, Warning: errors.Join(fingerprintWarning, latestFingerprintWarning)}, nil
	}
	if fingerprint != latestFingerprint {
		return Result{Projects: projects, Warning: fmt.Errorf("workspace changed during scan; result was not cached")}, nil
	}
	cached := entry{
		Version:             cacheVersion,
		Workspace:           workspace,
		ScannedAt:           store.now(),
		WorkspaceModifiedAt: latestInfo.ModTime(),
		Fingerprint:         latestFingerprint,
		Projects:            projectsToCache(projects),
	}
	writeWarning := store.write(cachePath, cached)
	return Result{Projects: projects, Warning: errors.Join(readWarning, writeWarning)}, nil
}

func (store Store) read(path string) (entry, error) {
	file, err := os.Open(path)
	if err != nil {
		return entry{}, err
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, maxCacheSize+1))
	if err != nil {
		return entry{}, err
	}
	if len(data) > maxCacheSize {
		return entry{}, fmt.Errorf("cache %q exceeds %d bytes", path, maxCacheSize)
	}
	var cached entry
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&cached); err != nil {
		return entry{}, err
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return entry{}, fmt.Errorf("cache %q contains multiple JSON values", path)
		}
		return entry{}, err
	}
	return cached, nil
}

func (store Store) write(path string, cached entry) error {
	data, err := json.Marshal(cached)
	if err != nil {
		return err
	}
	if len(data) > maxCacheSize {
		return fmt.Errorf("cache entry requires %d bytes, limit is %d", len(data), maxCacheSize)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	temporary, err := os.CreateTemp(filepath.Dir(path), ".projects-*.json")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := temporary.Chmod(0o600); err != nil {
		temporary.Close()
		return err
	}
	if _, err := temporary.Write(append(data, '\n')); err != nil {
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
	if err := replaceFile(temporaryPath, path); err != nil {
		return err
	}
	store.cleanup()
	return nil
}

func (store Store) valid(cached entry, workspace string, modifiedAt time.Time, fingerprint string) bool {
	if cached.Version != cacheVersion || cached.Workspace != workspace || !cached.WorkspaceModifiedAt.Equal(modifiedAt) || cached.Fingerprint != fingerprint {
		return false
	}
	age := store.now().Sub(cached.ScannedAt)
	return age >= 0 && age <= store.maxAge()
}

func (store Store) cleanup() {
	entries, err := os.ReadDir(store.Directory)
	if err != nil {
		return
	}
	cutoff := store.now().Add(-maxCacheLife)
	for _, cached := range entries {
		if cached.IsDir() || !isCacheFileName(cached.Name()) {
			continue
		}
		info, err := cached.Info()
		if err == nil && info.ModTime().Before(cutoff) {
			_ = os.Remove(filepath.Join(store.Directory, cached.Name()))
		}
	}
}

func isCacheFileName(name string) bool {
	if filepath.Ext(name) != ".json" {
		return false
	}
	encoded := strings.TrimSuffix(name, ".json")
	if len(encoded) != sha256.Size*2 {
		return false
	}
	_, err := hex.DecodeString(encoded)
	return err == nil
}

func (store Store) cachePath(workspace string) string {
	hash := sha256.Sum256([]byte(workspace))
	return filepath.Join(store.Directory, hex.EncodeToString(hash[:])+".json")
}

func (store Store) maxAge() time.Duration {
	if store.MaxAge > 0 {
		return store.MaxAge
	}
	return defaultAge
}

func (store Store) now() time.Time {
	if store.Now != nil {
		return store.Now().UTC()
	}
	return time.Now().UTC()
}

func canonicalWorkspace(path string) (string, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	if resolved, resolveErr := filepath.EvalSymlinks(absolute); resolveErr == nil {
		absolute = resolved
	}
	return filepath.Clean(absolute), nil
}

func workspaceFingerprint(workspace string, workspaceInfo os.FileInfo) (string, error) {
	hash := sha256.New()
	remainingContent := int64(maxTotalHash)
	_, _ = fmt.Fprintf(hash, "%s|%d|%d\n", workspace, workspaceInfo.ModTime().UnixNano(), workspaceInfo.Mode())
	entries, err := os.ReadDir(workspace)
	if err != nil {
		return "", err
	}
	for _, directory := range entries {
		if !directory.IsDir() || strings.HasPrefix(directory.Name(), ".") || ignoredDirectory(directory.Name()) {
			continue
		}
		_, _ = fmt.Fprintln(hash, directory.Name())
		for _, name := range fingerprintFiles {
			path := filepath.Join(workspace, directory.Name(), filepath.FromSlash(name))
			info, err := os.Stat(path)
			if err != nil {
				if os.IsNotExist(err) {
					continue
				}
				continue
			}
			if !info.Mode().IsRegular() {
				continue
			}
			_, _ = fmt.Fprintf(hash, "%s|%d|%d|%d\n", name, info.Size(), info.ModTime().UnixNano(), info.Mode())
			_, hashContent := contentFingerprintFiles[name]
			if hashContent && info.Size() <= maxFileHash && info.Size() <= remainingContent {
				file, err := os.Open(path)
				if err != nil {
					continue
				}
				limit := int64(maxFileHash)
				if remainingContent < limit {
					limit = remainingContent
				}
				content, readErr := io.ReadAll(io.LimitReader(file, limit+1))
				_ = file.Close()
				if readErr != nil || int64(len(content)) > limit {
					continue
				}
				_, _ = hash.Write(content)
				remainingContent -= int64(len(content))
			}
		}
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func ignoredDirectory(name string) bool {
	ignored := map[string]struct{}{
		"build": {}, "dist": {}, "node_modules": {}, "target": {}, "vendor": {}, "venv": {},
	}
	_, found := ignored[name]
	return found
}

func projectsToCache(projects []project.Project) []cachedProject {
	cached := make([]cachedProject, len(projects))
	for index, found := range projects {
		cached[index] = cachedProject{
			Name: found.Name, Path: found.Path, Technology: found.Technology, Markers: found.Markers,
			Frameworks: found.Frameworks, PackageManagers: found.PackageManagers, HasDocker: found.HasDocker,
		}
	}
	return cached
}

func projectsFromCache(cached []cachedProject) []project.Project {
	projects := make([]project.Project, len(cached))
	for index, found := range cached {
		projects[index] = project.Project{
			Name: found.Name, Path: found.Path, Technology: found.Technology, Markers: found.Markers,
			Frameworks: found.Frameworks, PackageManagers: found.PackageManagers, HasDocker: found.HasDocker,
		}
	}
	return projects
}
