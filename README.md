# ForgePath

ForgePath is an interactive terminal application for discovering, navigating, and managing software projects from a single interface.

Built with Go, ForgePath scans configured workspaces, identifies projects and their main technologies, and provides shortcuts for common development actions such as opening a project, starting its development environment, checking Git information, and launching it in an editor.

> ForgePath is currently under development.

## Overview

Developers often keep multiple projects across different folders, technologies, and environments. Navigating between them usually requires remembering paths, repeating commands, and manually checking how each project should be started.

ForgePath centralizes this workflow in an interactive terminal interface.

```text
┌──────────────────────────────────────────────────────────────┐
│ ForgePath                                D:\Development       │
├──────────────────────────────────────────────────────────────┤
│ Search: story_                                             │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│ > 󰛦 Story Pilot       TypeScript · Next.js      main          │
│    Operis            PHP · Laravel · Vue       develop       │
│    Mastermind        Java · Spring Boot        main          │
│    Residuum          Python · FastAPI          feature/api   │
│    ForgePath         Go                        main          │
│                                                              │
├──────────────────────────────────────────────────────────────┤
│ enter select   / search   r run   g git   q quit             │
└──────────────────────────────────────────────────────────────┘
```

## Goals

ForgePath is being developed as both a practical developer tool and a study project focused on:

* Go fundamentals
* Terminal user interfaces
* Filesystem traversal
* Process execution
* Shell integration
* Cross-platform development
* Project and framework detection
* Software architecture
* Configuration management
* Testing and release automation

## Planned Features

### Project discovery

* Scan multiple configured workspaces
* Detect projects through manifest and configuration files
* Ignore generated and dependency directories
* Support configurable scan depth
* Cache detected projects for faster startup

### Technology detection

ForgePath will initially detect projects from the following ecosystems:

* JavaScript and TypeScript
* PHP
* Java
* Python
* Go

Supported frameworks and tools will include:

* Next.js
* React
* Vue.js
* Nuxt
* NestJS
* Express
* Laravel
* Spring Boot
* FastAPI
* Docker
* Docker Compose

### Interactive terminal interface

* Search projects using fuzzy filtering
* Navigate using the keyboard
* Display language and framework icons
* Show the current Git branch
* Indicate uncommitted changes
* Display recent and favorite projects
* Provide ASCII fallbacks when Nerd Fonts are unavailable

### Project actions

* Open a project in the configured editor
* Open the project directory
* Start the development environment
* Execute custom project commands
* Start Docker Compose services
* Copy the project path
* Open the remote Git repository
* Launch a terminal in the selected directory

### Shell integration

ForgePath will integrate with shells such as PowerShell and Bash, allowing the selected project to become the current shell directory.

```powershell
fp
```

After selecting a project, the shell will navigate directly to its directory.

## Tech Stack

* [Go](https://go.dev/)
* [Bubble Tea](https://github.com/charmbracelet/bubbletea)
* [Bubbles](https://github.com/charmbracelet/bubbles)
* [Lip Gloss](https://github.com/charmbracelet/lipgloss)
* [Cobra](https://github.com/spf13/cobra)
* [Huh](https://github.com/charmbracelet/huh)
* Nerd Fonts

## Architecture

The project is organized into isolated modules so that terminal rendering, project detection, configuration, filesystem access, and process execution remain independent.

```text
forgepath/
├── cmd/
│   └── forgepath/
│       └── main.go
│
├── internal/
│   ├── cli/
│   ├── config/
│   ├── detector/
│   ├── project/
│   ├── action/
│   ├── git/
│   ├── icon/
│   ├── platform/
│   └── tui/
│
├── go.mod
├── go.sum
├── LICENSE
└── README.md
```

### Main modules

| Module     | Responsibility                                            |
| ---------- | --------------------------------------------------------- |
| `cli`      | Commands, flags, aliases, and shell-facing output         |
| `config`   | Configuration loading, validation, and default paths      |
| `detector` | Language, framework, package manager, and tool detection  |
| `project`  | Project scanning, indexing, and domain models             |
| `action`   | Safe execution of project commands                        |
| `git`      | Branch, repository, remote, and working tree information  |
| `icon`     | Nerd Font icons and ASCII fallbacks                       |
| `platform` | Operating system-specific behavior                        |
| `tui`      | Terminal state, events, keyboard shortcuts, and rendering |

## Project Detection

ForgePath detects projects through marker files and dependency manifests.

| Ecosystem               | Marker files                                       |
| ----------------------- | -------------------------------------------------- |
| JavaScript / TypeScript | `package.json`, `tsconfig.json`                    |
| PHP                     | `composer.json`, `artisan`                         |
| Java                    | `pom.xml`, `build.gradle`, `build.gradle.kts`      |
| Python                  | `pyproject.toml`, `requirements.txt`, `Pipfile`    |
| Go                      | `go.mod`, `go.work`                                |
| Docker                  | `Dockerfile`, `compose.yaml`, `docker-compose.yml` |

Frameworks are identified by inspecting project dependencies and configuration files.

For example:

```text
next                 → Next.js
@nestjs/core         → NestJS
vue                  → Vue.js
laravel/framework    → Laravel
spring-boot          → Spring Boot
fastapi              → FastAPI
```

Directories such as the following will be ignored during project analysis:

```text
.git
.idea
.vscode
node_modules
vendor
.next
dist
build
target
.venv
venv
__pycache__
coverage
```

## Configuration

ForgePath will use a local configuration file to define workspaces, editor preferences, project commands, and interface settings.

Example:

```json
{
  "workspaces": [
    "D:\\Development",
    "D:\\SyncForge",
    "D:\\College"
  ],
  "editor": {
    "name": "phpstorm",
    "executable": "phpstorm64.exe"
  },
  "scan": {
    "maxDepth": 2,
    "ignoreHidden": true
  },
  "icons": "nerd-font",
  "projects": {
    "story-pilot": {
      "command": "pnpm dev"
    },
    "operis": {
      "command": "composer dev"
    },
    "mastermind": {
      "command": "docker compose up"
    }
  }
}
```

Expected configuration paths:

```text
Windows: %APPDATA%\forgepath\config.json
Linux:   ~/.config/forgepath/config.json
macOS:   ~/Library/Application Support/forgepath/config.json
```

## Planned Commands

```bash
forgepath
```

Open the interactive terminal interface.

```bash
forgepath list
```

List all detected projects.

```bash
forgepath scan
```

Scan configured workspaces and rebuild the project index.

```bash
forgepath pick --print-path
```

Select a project and print only its directory path.

```bash
forgepath open <project>
```

Open a project in the configured editor.

```bash
forgepath run <project>
```

Run the configured development command.

```bash
forgepath config init
```

Create an initial configuration file.

```bash
forgepath completion powershell
```

Generate shell completion scripts.

## PowerShell Integration

A shell function is required because a child process cannot directly change the working directory of its parent shell.

The intended PowerShell integration is:

```powershell
function fp {
    $previousOutputEncoding = [Console]::OutputEncoding

    try {
        [Console]::OutputEncoding = [System.Text.UTF8Encoding]::new($false)
        $target = & forgepath pick --print-path @args
        $exitCode = $LASTEXITCODE
    }
    finally {
        [Console]::OutputEncoding = $previousOutputEncoding
    }

    if ($exitCode -ne 0) {
        Write-Error "forgepath pick failed with exit code $exitCode"
        return
    }

    if ($target) {
        Set-Location -LiteralPath $target
    }
}
```

After adding the function to the PowerShell profile:

```powershell
fp
```

ForgePath opens the project selector and navigates the current terminal to the selected project.

## Bash Integration

```bash
fp() {
    local target status
    target="$(forgepath pick --print-path "$@")"
    status=$?

    if [ "$status" -ne 0 ]; then
        return "$status"
    fi

    if [ -n "$target" ]; then
        cd -- "$target"
    fi
}
```

## Development

### Requirements

* Go 1.25 or newer
* Git
* A terminal with ANSI color support
* A Nerd Font for language and tool icons

### Clone the repository

```bash
git clone https://github.com/daviPeter07/forgepath.git
cd forgepath
```

### Install dependencies

```bash
go mod download
```

### Run the application

```bash
go run ./cmd/forgepath
```

### Build

```bash
go build -o forgepath ./cmd/forgepath
```

On Windows:

```powershell
go build -o forgepath.exe ./cmd/forgepath
```

### Run tests

```bash
go test ./...
```

### Format the code

```bash
go fmt ./...
```

### Analyze the code

```bash
go vet ./...
```

## Security

Project commands will be executed using explicit command names and argument lists.

ForgePath will avoid concatenating user-controlled values into commands executed through `sh -c`, `cmd /c`, or `powershell -Command`.

Detected commands will be presented as suggestions. Custom commands must be explicitly configured or confirmed before execution.

## Contributing

ForgePath is currently a personal study and portfolio project, but suggestions, bug reports, and contributions are welcome.

Before submitting a pull request:

1. Open an issue describing the change.
2. Keep the change focused.
3. Add or update tests when applicable.
4. Run formatting, tests, and static analysis.
5. Explain the motivation and technical decisions in the pull request.

## License

This project is licensed under the MIT License.

## Author

Developed by [Davi Peterson](https://github.com/daviPeter07).
