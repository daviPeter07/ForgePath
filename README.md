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

## Features

### Project discovery

* Scan multiple configured workspaces
* Detect projects through manifest and configuration files
* Ignore generated and dependency directories
* Support configurable scan depth
* Cache detected projects for faster startup

### Technology detection

ForgePath detects projects from the following ecosystems:

* JavaScript and TypeScript
* PHP
* Java
* Python
* Go

Supported frameworks and tools include:

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

ForgePath integrates with shells such as PowerShell and Bash, allowing the selected project to become the current shell directory.

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

ForgePath uses a local configuration file for editor preferences and project commands.

Example:

```json
{
  "editor": {
    "name": "phpstorm",
    "executable": "phpstorm64.exe"
  },
  "projects": {
    "story-pilot": {
      "command": ["node", "node_modules/vite/bin/vite.js"]
    },
    "operis": {
      "command": ["php", "composer.phar", "dev"]
    },
    "mastermind": {
      "command": ["docker", "compose", "up"]
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

## Commands

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

Scan a workspace and rebuild its project cache. `forgepath`, `list`, and `pick` reuse cache entries for up to 30 seconds; pass `--refresh` to bypass them.

```bash
forgepath pick --print-path
```

Select a project and print only its directory path.

The selector uses portable ASCII labels by default. Enable technology icons in a Nerd Font terminal with:

```bash
forgepath pick --icons nerd-font
forgepath --icons nerd-font
```

```bash
forgepath open <project> [workspace] --editor code
```

Open a project in an editor. Set an executable with `--editor` or `FORGEPATH_EDITOR`.

On Windows, provide the editor `.exe` path rather than a `.cmd` or `.bat` launcher.

```bash
forgepath reveal <project> [workspace]
```

Reveal a project in Explorer, Finder, or the Linux file manager.

```bash
forgepath run <project> [workspace]
```

Run the development command configured for a project. Commands are argument arrays and are never interpreted by a shell.

On Windows, `.cmd` and `.bat` launchers are rejected. Configure a real `.exe` or invoke a script through its interpreter, such as `node.exe` or `php.exe`.

```bash
forgepath config init
```

Create an initial configuration file.

Use `--config <path>` or `FORGEPATH_CONFIG` to override the default configuration path.

```bash
forgepath favorite add <project> [workspace]
forgepath favorite remove <project> [workspace]
forgepath favorite list
forgepath recent
```

Favorites are shown first in the selector, followed by recently used projects. Use `--state <path>` or `FORGEPATH_STATE` to override the state file location.

Use `--cache <directory>` or `FORGEPATH_CACHE` to override the project cache location.

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
* Optional: a Nerd Font for language and tool icons

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

---

# ForgePath — Português

ForgePath é uma aplicação interativa de terminal para descobrir, navegar e gerenciar projetos de software a partir de uma única interface.

Desenvolvido em Go, o ForgePath analisa workspaces configurados, identifica projetos e suas principais tecnologias e oferece atalhos para ações comuns de desenvolvimento, como abrir um projeto, iniciar seu ambiente de desenvolvimento, consultar informações do Git e iniciá-lo em um editor.

> O ForgePath está atualmente em desenvolvimento.

## Visão geral

Desenvolvedores frequentemente mantêm vários projetos distribuídos entre diferentes pastas, tecnologias e ambientes. Navegar entre eles normalmente exige lembrar caminhos, repetir comandos e verificar manualmente como cada projeto deve ser iniciado.

O ForgePath centraliza esse fluxo de trabalho em uma interface interativa de terminal.

```text
┌──────────────────────────────────────────────────────────────┐
│ ForgePath                                D:\Development       │
├──────────────────────────────────────────────────────────────┤
│ Busca: story_                                                │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│ > 󰛦 Story Pilot       TypeScript · Next.js      main          │
│    Operis            PHP · Laravel · Vue       develop       │
│    Mastermind        Java · Spring Boot        main          │
│    Residuum          Python · FastAPI          feature/api   │
│    ForgePath         Go                        main          │
│                                                              │
├──────────────────────────────────────────────────────────────┤
│ enter selecionar   / buscar   r executar   g git   q sair    │
└──────────────────────────────────────────────────────────────┘
```

## Objetivos

O ForgePath é desenvolvido tanto como uma ferramenta prática para desenvolvedores quanto como um projeto de estudo focado em:

* Fundamentos de Go
* Interfaces de usuário no terminal
* Percurso do sistema de arquivos
* Execução de processos
* Integração com shells
* Desenvolvimento multiplataforma
* Detecção de projetos e frameworks
* Arquitetura de software
* Gerenciamento de configuração
* Testes e automação de releases

## Recursos

### Descoberta de projetos

* Analisar múltiplos workspaces configurados
* Detectar projetos por arquivos de manifesto e configuração
* Ignorar diretórios gerados e de dependências
* Suportar profundidade de análise configurável
* Armazenar projetos detectados em cache para uma inicialização mais rápida

### Detecção de tecnologias

O ForgePath detecta projetos dos seguintes ecossistemas:

* JavaScript e TypeScript
* PHP
* Java
* Python
* Go

Os frameworks e ferramentas compatíveis incluem:

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

### Interface interativa de terminal

* Pesquisar projetos com filtro fuzzy
* Navegar usando o teclado
* Exibir ícones de linguagens e frameworks
* Mostrar a branch atual do Git
* Indicar alterações não commitadas
* Exibir projetos recentes e favoritos
* Oferecer alternativas ASCII quando Nerd Fonts não estiverem disponíveis

### Ações de projeto

* Abrir um projeto no editor configurado
* Abrir o diretório do projeto
* Iniciar o ambiente de desenvolvimento
* Executar comandos personalizados do projeto
* Iniciar serviços do Docker Compose
* Copiar o caminho do projeto
* Abrir o repositório Git remoto
* Abrir um terminal no diretório selecionado

### Integração com o shell

O ForgePath se integra a shells como PowerShell e Bash, permitindo que o projeto selecionado se torne o diretório atual do shell.

```powershell
fp
```

Após selecionar um projeto, o shell navegará diretamente para seu diretório.

## Tecnologias utilizadas

* [Go](https://go.dev/)
* [Bubble Tea](https://github.com/charmbracelet/bubbletea)
* [Bubbles](https://github.com/charmbracelet/bubbles)
* [Lip Gloss](https://github.com/charmbracelet/lipgloss)
* [Cobra](https://github.com/spf13/cobra)
* [Huh](https://github.com/charmbracelet/huh)
* Nerd Fonts

## Arquitetura

O projeto é organizado em módulos isolados para que a renderização do terminal, a detecção de projetos, a configuração, o acesso ao sistema de arquivos e a execução de processos permaneçam independentes.

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

### Principais módulos

| Módulo | Responsabilidade |
| --- | --- |
| `cli` | Comandos, flags, aliases e saída destinada ao shell |
| `config` | Carregamento, validação e caminhos padrão da configuração |
| `detector` | Detecção de linguagens, frameworks, gerenciadores de pacotes e ferramentas |
| `project` | Análise, indexação e modelos de domínio dos projetos |
| `action` | Execução segura de comandos dos projetos |
| `git` | Informações de branch, repositório, remote e working tree |
| `icon` | Ícones Nerd Font e alternativas ASCII |
| `platform` | Comportamentos específicos de cada sistema operacional |
| `tui` | Estado do terminal, eventos, atalhos de teclado e renderização |

## Detecção de projetos

O ForgePath detecta projetos por meio de arquivos marcadores e manifestos de dependências.

| Ecossistema | Arquivos marcadores |
| --- | --- |
| JavaScript / TypeScript | `package.json`, `tsconfig.json` |
| PHP | `composer.json`, `artisan` |
| Java | `pom.xml`, `build.gradle`, `build.gradle.kts` |
| Python | `pyproject.toml`, `requirements.txt`, `Pipfile` |
| Go | `go.mod`, `go.work` |
| Docker | `Dockerfile`, `compose.yaml`, `docker-compose.yml` |

Os frameworks são identificados pela inspeção das dependências e dos arquivos de configuração do projeto.

Por exemplo:

```text
next                 → Next.js
@nestjs/core         → NestJS
vue                  → Vue.js
laravel/framework    → Laravel
spring-boot          → Spring Boot
fastapi              → FastAPI
```

Diretórios como os seguintes serão ignorados durante a análise dos projetos:

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

## Configuração

O ForgePath utiliza um arquivo de configuração local para preferências do editor e comandos dos projetos.

Exemplo:

```json
{
  "editor": {
    "name": "phpstorm",
    "executable": "phpstorm64.exe"
  },
  "projects": {
    "story-pilot": {
      "command": ["node", "node_modules/vite/bin/vite.js"]
    },
    "operis": {
      "command": ["php", "composer.phar", "dev"]
    },
    "mastermind": {
      "command": ["docker", "compose", "up"]
    }
  }
}
```

Caminhos esperados para a configuração:

```text
Windows: %APPDATA%\forgepath\config.json
Linux:   ~/.config/forgepath/config.json
macOS:   ~/Library/Application Support/forgepath/config.json
```

## Comandos

```bash
forgepath
```

Abre a interface interativa de terminal.

```bash
forgepath list
```

Lista todos os projetos detectados.

```bash
forgepath scan
```

Analisa um workspace e reconstrói seu cache de projetos. `forgepath`, `list` e `pick` reutilizam entradas do cache por até 30 segundos; use `--refresh` para ignorá-las.

```bash
forgepath pick --print-path
```

Seleciona um projeto e imprime somente o caminho de seu diretório.

O seletor usa rótulos ASCII portáveis por padrão. Ative os ícones de tecnologia em um terminal com Nerd Font usando:

```bash
forgepath pick --icons nerd-font
forgepath --icons nerd-font
```

```bash
forgepath open <projeto> [workspace] --editor code
```

Abre um projeto em um editor. Defina um executável com `--editor` ou `FORGEPATH_EDITOR`.

No Windows, informe o caminho do arquivo `.exe` do editor em vez de um launcher `.cmd` ou `.bat`.

```bash
forgepath reveal <projeto> [workspace]
```

Revela um projeto no Explorer, Finder ou gerenciador de arquivos do Linux.

```bash
forgepath run <projeto> [workspace]
```

Executa o comando de desenvolvimento configurado para um projeto. Os comandos são arrays de argumentos e nunca são interpretados por um shell.

No Windows, launchers `.cmd` e `.bat` são rejeitados. Configure um `.exe` real ou execute um script por meio de seu interpretador, como `node.exe` ou `php.exe`.

```bash
forgepath config init
```

Cria um arquivo de configuração inicial.

Use `--config <caminho>` ou `FORGEPATH_CONFIG` para substituir o caminho padrão da configuração.

```bash
forgepath favorite add <projeto> [workspace]
forgepath favorite remove <projeto> [workspace]
forgepath favorite list
forgepath recent
```

Os favoritos são exibidos primeiro no seletor, seguidos pelos projetos usados recentemente. Use `--state <caminho>` ou `FORGEPATH_STATE` para substituir o local do arquivo de estado.

Use `--cache <diretório>` ou `FORGEPATH_CACHE` para substituir o local do cache de projetos.

```bash
forgepath completion powershell
```

Gera scripts de autocompletar para o shell.

## Integração com PowerShell

Uma função de shell é necessária porque um processo filho não pode alterar diretamente o diretório de trabalho de seu processo pai.

A integração prevista com PowerShell é:

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
        Write-Error "forgepath pick falhou com o código de saída $exitCode"
        return
    }

    if ($target) {
        Set-Location -LiteralPath $target
    }
}
```

Após adicionar a função ao perfil do PowerShell:

```powershell
fp
```

O ForgePath abre o seletor de projetos e navega o terminal atual para o projeto selecionado.

## Integração com Bash

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

## Desenvolvimento

### Requisitos

* Go 1.25 ou mais recente
* Git
* Um terminal com suporte a cores ANSI
* Opcional: uma Nerd Font para ícones de linguagens e ferramentas

### Clonar o repositório

```bash
git clone https://github.com/daviPeter07/forgepath.git
cd forgepath
```

### Instalar dependências

```bash
go mod download
```

### Executar a aplicação

```bash
go run ./cmd/forgepath
```

### Compilar

```bash
go build -o forgepath ./cmd/forgepath
```

No Windows:

```powershell
go build -o forgepath.exe ./cmd/forgepath
```

### Executar os testes

```bash
go test ./...
```

### Formatar o código

```bash
go fmt ./...
```

### Analisar o código

```bash
go vet ./...
```

## Segurança

Os comandos dos projetos serão executados usando nomes de executáveis e listas de argumentos explícitos.

O ForgePath evitará concatenar valores controlados pelo usuário em comandos executados por `sh -c`, `cmd /c` ou `powershell -Command`.

Os comandos detectados serão apresentados como sugestões. Comandos personalizados devem ser configurados ou confirmados explicitamente antes da execução.

## Contribuição

O ForgePath é atualmente um projeto pessoal de estudo e portfólio, mas sugestões, relatos de bugs e contribuições são bem-vindos.

Antes de enviar um pull request:

1. Abra uma issue descrevendo a alteração.
2. Mantenha a alteração focada.
3. Adicione ou atualize testes quando aplicável.
4. Execute formatação, testes e análise estática.
5. Explique a motivação e as decisões técnicas no pull request.

## Licença

Este projeto é licenciado sob a licença MIT.

## Autor

Desenvolvido por [Davi Peterson](https://github.com/daviPeter07).
