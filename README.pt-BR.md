[![en](https://img.shields.io/badge/lang-en-red.svg)](README.md) [![pt-br](https://img.shields.io/badge/lang-pt--br-green.svg)](README.pt-BR.md)

# ForgePath — Português

ForgePath é uma aplicação interativa de terminal para descobrir, navegar e gerenciar projetos de software a partir de uma única interface.

Desenvolvido em Go, o ForgePath analisa workspaces configurados, identifica projetos e suas principais tecnologias e oferece atalhos para ações comuns de desenvolvimento, como abrir um projeto, iniciar seu ambiente de desenvolvimento, consultar informações do Git e iniciá-lo em um editor.

> O ForgePath está atualmente em desenvolvimento.

## Visão geral

Desenvolvedores frequentemente mantêm vários projetos distribuídos entre diferentes pastas, tecnologias e ambientes. Navegar entre eles normalmente exige lembrar caminhos, repetir comandos e verificar manualmente como cada projeto deve ser iniciado.

O ForgePath centraliza esse fluxo de trabalho em uma interface interativa de terminal.

![Demonstração do ForgePath CLI](docs/demo.png)

*Uma rápida demonstração do ForgePath em ação, incluindo navegação de projetos e geração de Docker Compose.*

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
* Armazenar projetos detectados em cache para uma inicialização mais rápida

### Detecção de tecnologias

O ForgePath detecta projetos dos seguintes ecossistemas:

* JavaScript e TypeScript
* PHP
* Java
* Python
* Go
* Rust
* Ruby
* Swift
* Elixir

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
* Navegar com as setas renderizando apenas os projetos visíveis
* Navegar pelas pastas dos projetos sem sair da interface do terminal
* Renderizar os logos Devicon embutidos com pixels ANSI truecolor
* Mostrar a branch atual do Git
* Indicar alterações não commitadas
* Exibir projetos recentes e favoritos
* Oferecer badges de texto coloridos e glifos Nerd Font como alternativas
* Detectar IDEs instaladas e ordenar sugestões para cada tecnologia

### Ações de projeto

* Escolher entre IDEs compatíveis instaladas sem sair do seletor
* Abrir o diretório do projeto
* Iniciar o ambiente de desenvolvimento
* Executar comandos personalizados do projeto

### Integração com o shell

O ForgePath se integra a shells como PowerShell e Bash. A navegação e a abertura de IDEs permanecem dentro da aplicação; o diretório do shell só muda quando o usuário pressiona `c`.

```powershell
fg
```

Após pressionar `c`, o shell navega para o diretório exibido no ForgePath.

## Tecnologias utilizadas

* [Go](https://go.dev/)
* [Bubble Tea](https://github.com/charmbracelet/bubbletea)
* [Bubbles](https://github.com/charmbracelet/bubbles)
* [Lip Gloss](https://github.com/charmbracelet/lipgloss)
* [Cobra](https://github.com/spf13/cobra)
* Nerd Fonts

## Arquitetura

O projeto é organizado em módulos isolados para que a renderização do terminal, a detecção de projetos, a configuração, o acesso ao sistema de arquivos e a execução de processos permaneçam independentes.

```text
forgepath/
├── cmd/
│   └── fg/
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
│   ├── ide/
│   ├── platform/
│   └── tui/
│
├── public/
│   └── icons/
│       └── *.svg
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
| `ide` | Descoberta de IDEs instaladas e sugestões por tecnologia |
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
| Rust | `Cargo.toml` |
| Ruby | `Gemfile` |
| Swift | `Package.swift` |
| Elixir | `mix.exs` |
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

O ForgePath utiliza um arquivo de configuração do usuário para workspaces globais, preferências do editor e comandos dos projetos.

Exemplo:

```json
{
  "workspaces": ["D:\\Development", "D:\\Clientes"],
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

Primeiro, adicione uma ou mais pastas que contenham seus projetos. Esse catálogo global permite usar o ForgePath a partir de qualquer diretório atual:

```bash
fg workspace add D:\Development
fg workspace add D:\Clientes
fg workspace list
```

```bash
fg
```

Abre o navegador persistente do terminal. Pesquisar, entrar nos projetos e subdiretórios, voltar e iniciar uma IDE acontecem sem interromper o ForgePath.

O seletor verifica os executáveis instalados antes de exibi-los e prioriza ferramentas específicas da tecnologia. Por exemplo, um projeto PHP sugere PhpStorm antes do Visual Studio Code quando ambos estão instalados; TypeScript prioriza WebStorm, Python prioriza PyCharm, Go prioriza GoLand, Java prioriza IntelliJ IDEA e Rust prioriza RustRover.

### Descoberta de IDEs

| Tecnologia | IDE preferencial | Alternativas compatíveis |
| --- | --- | --- |
| PHP | PhpStorm | Visual Studio Code, Cursor, Zed, Sublime Text |
| TypeScript / JavaScript | WebStorm | Visual Studio Code, Cursor, Zed, Sublime Text |
| Python | PyCharm | Visual Studio Code, Cursor, Zed, Sublime Text |
| Go | GoLand | Visual Studio Code, Cursor, Zed, Sublime Text |
| Java | IntelliJ IDEA | Visual Studio Code, Cursor, Zed, Sublime Text |
| Rust | RustRover | Visual Studio Code, Cursor, Zed, Sublime Text |
| Ruby | RubyMine | Visual Studio Code, Cursor, Zed, Sublime Text |
| Swift | Xcode | Visual Studio Code, Cursor, Zed, Sublime Text |
| Elixir / Docker | Visual Studio Code | Cursor, Zed, Sublime Text |

O ForgePath verifica o `PATH` e locais comuns de instalação. A descoberta cobre instalações independentes e pelo JetBrains Toolbox no Windows; `/Applications`, aplicações do usuário e Toolbox no macOS; e `PATH`, Toolbox, Snap e Flatpak no Linux. Editores ausentes ou incompatíveis com a tecnologia detectada não são exibidos.

### Controles do teclado

| Tecla | Ação |
| --- | --- |
| `↑` / `↓` | Percorre os projetos, diretórios ou IDEs visíveis |
| `/` | Pesquisa a lista atual por nome ou tecnologia |
| `Enter` | Entra no projeto/diretório selecionado ou inicia a IDE selecionada |
| `Backspace` / `←` | Volta ao diretório pai ou à lista de projetos sem fechar o ForgePath |
| `o` | Exibe IDEs compatíveis cuja instalação foi verificada |
| `c` | Fecha o ForgePath e muda o shell para o diretório exibido |
| `Esc` | Fecha a pesquisa ou retorna à tela interna anterior |
| `q` / `Ctrl+C` | Encerra sem alterar o diretório do shell |

Depois de iniciar uma IDE, o ForgePath retorna ao navegador de diretórios e continua aberto. Diretórios gerados como `.git`, `node_modules`, `vendor`, `target`, `dist` e `build` são omitidos do navegador interno.

```bash
fg list
```

Lista todos os projetos detectados.

```bash
fg scan
```

Analisa um workspace e reconstrói seu cache de projetos. `fg`, `list` e `pick` reutilizam entradas do cache por até 30 segundos; use `--refresh` para ignorá-las.

```bash
fg pick --print-path
```

Navega pelos projetos e imprime o diretório atual somente depois que `c` é pressionado. Essa saída legível por máquina é usada pela integração com o shell.

O seletor incorpora os SVGs locais do Devicon no binário, rasteriza os arquivos e renderiza suas cores reais com pixels ANSI de meio bloco por padrão. Isso não exige Nerd Font nem um protocolo de imagens específico do terminal. Use badges de texto ou glifos Nerd Font explicitamente quando preferir:

```bash
fg --icons ascii
fg pick --icons nerd-font
fg --icons nerd-font
```

Os assets fixados do Devicon v2.17.0 e sua licença ficam em `public/icons`. O `go:embed` inclui esses arquivos no `fg`, enquanto `oksvg` e `rasterx` realizam a rasterização em memória. Nenhum acesso à internet é necessário durante a execução.

```bash
fg open <projeto> [workspace] --editor <executável>
```

Abre um projeto em um editor. Defina um executável com `--editor` ou `FORGEPATH_EDITOR`.

No Windows, informe o caminho do arquivo `.exe` do editor em vez de um launcher `.cmd` ou `.bat`.

```bash
fg reveal <projeto> [workspace]
```

Revela um projeto no Explorer, Finder ou gerenciador de arquivos do Linux.

```bash
fg run <projeto> [workspace]
```

Executa o comando de desenvolvimento configurado para um projeto. Os comandos são arrays de argumentos e nunca são interpretados por um shell.

No Windows, launchers `.cmd` e `.bat` são rejeitados. Configure um `.exe` real ou execute um script por meio de seu interpretador, como `node.exe` ou `php.exe`.

```bash
fg config init
```

Cria um arquivo de configuração inicial.

Use `--config <caminho>` ou `FORGEPATH_CONFIG` para substituir o caminho padrão da configuração.

```bash
fg favorite add <projeto> [workspace]
fg favorite remove <projeto> [workspace]
fg favorite list
fg recent
```

Os favoritos são exibidos primeiro no seletor, seguidos pelos projetos usados recentemente. Use `--state <caminho>` ou `FORGEPATH_STATE` para substituir o local do arquivo de estado.

Use `--cache <diretório>` ou `FORGEPATH_CACHE` para substituir o local do cache de projetos.

```bash
fg completion powershell
```

Gera scripts de autocompletar para o shell.

## Integração com PowerShell

Uma função de shell é necessária porque um processo filho não pode alterar diretamente o diretório de trabalho de seu processo pai.

A integração prevista com PowerShell é:

```powershell
function fg {
    if ($args.Count -eq 1 -and $args[0] -in @("back", "-")) {
        if ((Get-Location -Stack -ErrorAction SilentlyContinue).Count -eq 0) {
            Write-Error "ForgePath não possui um diretório anterior"
            return
        }
        try {
            Pop-Location -ErrorAction Stop
        }
        catch {
            Write-Error "ForgePath não possui um diretório anterior"
        }
        return
    }

    $commands = @("list", "pick", "scan", "open", "reveal", "run", "config", "workspace", "favorite", "recent", "completion", "help")
    $dispatch = $false
    $expectValue = $false
    foreach ($argument in $args) {
        if ($expectValue) {
            $expectValue = $false
            continue
        }
        if ($argument -in @("--config", "--state", "--cache", "--icons")) {
            $expectValue = $true
            continue
        }
        if ($argument -match '^--(config|state|cache|icons)=') {
            continue
        }
        if ($argument -in @("-h", "--help", "--version")) {
            $dispatch = $true
            break
        }
        if ($argument.StartsWith("-")) {
            continue
        }
        if ($commands -contains $argument) {
            $dispatch = $true
        }
        break
    }
    $applicationName = if ($env:OS -eq "Windows_NT") { "fg.exe" } else { "fg" }
    $executable = @(Get-Command $applicationName -CommandType Application -ErrorAction Stop)[0].Source
    if ($dispatch) {
        & $executable @args
        return
    }

    $previousOutputEncoding = [Console]::OutputEncoding

    try {
        [Console]::OutputEncoding = [System.Text.UTF8Encoding]::new($false)
        $target = & $executable pick --print-path @args
        $exitCode = $LASTEXITCODE
    }
    finally {
        [Console]::OutputEncoding = $previousOutputEncoding
    }

    if ($exitCode -ne 0) {
        Write-Error "ForgePath falhou com o código de saída $exitCode"
        return
    }

    if ($target) {
        Push-Location -LiteralPath $target
    }
}
```

Após adicionar a função ao perfil do PowerShell:

```powershell
fg
```

O ForgePath permanece aberto enquanto você navega pelas pastas ou inicia IDEs. Pressione `c` quando quiser fechar o seletor e navegar o terminal atual para o diretório exibido.

Cada diretório confirmado com `c` é adicionado à pilha de diretórios do shell. A navegação interna usa `Backspace`/`←`; depois de sair do ForgePath, desfaça a mudança do shell com:

```powershell
fg back
# ou
fg -
```

## Integração com Bash

```bash
fg() {
    local argument dispatch expect_value target status

    if [ "$#" -eq 1 ] && { [ "$1" = "back" ] || [ "$1" = "-" ]; }; then
        if ! popd >/dev/null 2>&1; then
            printf '%s\n' "ForgePath não possui um diretório anterior" >&2
            return 1
        fi
        return 0
    fi

    dispatch=0
    expect_value=0
    for argument in "$@"; do
        if [ "$expect_value" -eq 1 ]; then
            expect_value=0
            continue
        fi
        case "$argument" in
            --config|--state|--cache|--icons)
                expect_value=1
                continue
                ;;
            --config=*|--state=*|--cache=*|--icons=*)
                continue
                ;;
            -h|--help|--version)
                dispatch=1
                break
                ;;
            -*)
                continue
                ;;
            list|pick|scan|open|reveal|run|config|workspace|favorite|recent|completion|help)
                dispatch=1
                ;;
        esac
        break
    done
    if [ "$dispatch" -eq 1 ]; then
        command fg "$@"
        return $?
    fi

    target="$(command fg pick --print-path "$@")"
    status=$?

    if [ "$status" -ne 0 ]; then
        return "$status"
    fi

    if [ -n "$target" ]; then
        pushd "$target" >/dev/null
    fi
}
```

## Desenvolvimento

### Requisitos

* Go 1.25 ou mais recente
* Git
* Um terminal com suporte a ANSI truecolor para os logos gráficos
* Opcional: uma Nerd Font ao usar `--icons nerd-font`

### Clonar o repositório

```bash
git clone https://github.com/daviPeter07/forgepath.git
cd forgepath
```

### Instalar dependências

```bash
go mod download
```

### Instalar o comando

```bash
go install ./cmd/fg
```

O executável é instalado como `fg` em `GOBIN` ou `GOPATH/bin`. Adicione esse diretório ao `PATH` para executar `fg` de qualquer lugar.

### Executar a aplicação

```bash
go run ./cmd/fg
```

### Compilar

```bash
go build -o fg ./cmd/fg
```

No Windows:

```powershell
go build -o fg.exe ./cmd/fg
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
