# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

ForgePath (`fg`) is a Go terminal app (Cobra CLI + Bubble Tea v2 TUI) that scans configured workspaces, detects projects by technology, and lets the user browse them, open them in an installed IDE, run their dev command, or `cd` into them through a shell wrapper. It is also a Go study project: `docs/phases.md` (in Portuguese) records the incremental phases and a guiding rule: don't introduce interfaces or abstractions until there is a real second implementation or a test needs one.

## Commands

```bash
go run ./cmd/fg                 # run the TUI
go build -o fg ./cmd/fg         # build (fg.exe on Windows; binaries are gitignored)
go test ./...                   # all tests
go test ./internal/catalog -run TestName   # single package / single test
go vet ./...
gofmt -l .                      # CI fails if this prints anything
```

`.gitattributes` forces LF. Without it, a global `core.autocrlf=true` checks files out as CRLF, which makes `gofmt -l` fail and breaks `source shell/forgepath.sh`.

CI (`.github/workflows/ci.yml`) runs gofmt check, `go vet`, `go test`, plus syntax/behavior checks of `shell/forgepath.sh` (bash) and `shell/forgepath.ps1` (pwsh) for `fg back`.

## Architecture

Flow: `cmd/fg/main.go` → `cli.NewRootCommand` → scan workspaces → decorate with favorites/recents → `tui.SelectWithOptions` → print path / launch IDE / run command.

- **`internal/cli`**: Cobra commands. `root.go` wires all subcommands and resolves the three persistent paths (`--config`/`FORGEPATH_CONFIG`, `--state`/`FORGEPATH_STATE`, `--cache`/`FORGEPATH_CACHE`). Dependencies are injected as function values (`scanProjectsFunc`, `statePathFunc`, `configPathFunc`, `action.*` funcs) so tests can stub them. Several commands have a plain constructor and a `newConfigured...` variant used by the root. State/cache failures are reported as `warning:` on stderr and never abort the command (`*BestEffort` helpers).
- **`internal/scanner`**: scans only the **direct children** of a workspace (not recursive), calls `detector.Detect` on each, then enriches with Git info (`EnrichGit`).
- **`internal/detector`**: `manifest.go` has an ordered `rules` list; the **first matching rule wins** (e.g. TypeScript requires `package.json`+`tsconfig.json` and must come before JavaScript). `metadata.go` reads manifests for frameworks and package managers.
- **`internal/catalog`**: per-workspace JSON cache in `os.UserCacheDir()/forgepath`. A cache entry is valid for 30s **and** only while the workspace fingerprint matches (mtimes plus content hashes of the `fingerprintFiles` list). Git status is never cached; it is re-read on every hit. If the workspace changes mid-scan, the result is returned but not cached.
- **`internal/config`** (user `config.json`: workspaces, editor, per-project `command` arrays) and **`internal/state`** (favorites, recents): both use a `.lock` file with OS-specific locking (`lock_unix.go`/`lock_windows.go`/`lock_other.go`) plus atomic temp-file replace (`replace_*.go`). Mutations go through `Update`/`update(mutate func)`; keep that pattern.
- **`internal/action`**: process execution. Commands are always argument arrays passed to `exec.Command` and are never run via `sh -c`/`cmd /c`/`powershell -Command`. On Windows, `.cmd`/`.bat` launchers are rejected (`safeExecutable`). `RunCommand` kills the process tree on context cancel (SIGINT), escalating after 3s (`process_unix.go`/`process_windows.go`).
- **`internal/tui`**: one Bubble Tea `Model` with screen modes (projects → directory browser → IDE chooser → docker compose generator in `docker.go`). All screens reuse a single `list.Model` with different item types rendered by `projectDelegate.Render`. Keys: Enter, `c` (confirm cd), `o` (IDE), `d` (docker), Backspace/← (back), `/`, `q`.
  - Switch screens through `setScreen`. It sets the breadcrumb header, the per-screen help keys (`screenKeys`) and the delegate.
  - `parentProjects` is a stack of enclosing projects, so Back out of a nested project returns to its parent folder.
  - `actionTarget` decides what `o`/`d` act on.
  - Any `*Model` method that changes `m.list` (status messages, items) must use a pointer receiver. With a value receiver the change lands on a copy and is lost; this once hid every error message.
  - Row rendering: inner lipgloss styles end with a reset, so the selected row's background has to be carried into every segment (`newRowStyles`). `rowContainer` keeps selected and unselected rows the same width.
- **`internal/ide`**: discovers installed IDEs (PATH, JetBrains Toolbox, Snap, Flatpak, `/Applications`, Windows install dirs) and ranks them per technology.
- **`internal/icon`** + **`public/icons`**: Devicon SVGs are `go:embed`ded (`public/icons/assets.go`) and rasterized to ANSI half-block truecolor (`oksvg`/`rasterx`). Modes: `auto`, `graphics`, `ascii`, `nerd-font`.

### Cross-cutting changes

- **Adding a technology** touches: `project` constants, `detector` rules (order matters) and extension fallbacks, `catalog.fingerprintFiles`/`contentFingerprintFiles`, `icon` labels/graphics plus an SVG in `public/icons`, `tui.technologyColor`, `ide` ranking, and the README tables.
- **Adding or renaming a subcommand**: the shell wrappers (`shell/forgepath.sh`, `shell/forgepath.ps1`) hard-code the list of subcommands and value-taking flags (`--config|--state|--cache|--icons`) to decide between dispatching to `fg` and running `fg pick --print-path`. Update both scripts **and** the copies embedded in `README.md` and `README.pt-BR.md`.
- Platform-specific code uses build-tagged file pairs (`*_unix.go` / `*_windows.go`). Many functions take `goos` as a parameter so Windows behavior can be tested on Linux.

## Testing notes

- `internal/cli/main_test.go` sets `FORGEPATH_CACHE/CONFIG/STATE` to a temp dir in `TestMain` so tests never touch the real user files. Do the same for any new package that resolves default paths.
- Tests build their fixture workspaces in `t.TempDir()`. The placeholder dirs in `testdata/workspace/` are not referenced by any test.

## Docs

The README is bilingual as two files: `README.md` (English) and `README.pt-BR.md` (Portuguese). Keep both in sync. The README's module table lists a `platform` package that does not exist; platform logic lives in the build-tagged files above.
