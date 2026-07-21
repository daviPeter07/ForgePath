# ForgePath — Fases de Desenvolvimento

Este documento organiza o desenvolvimento do ForgePath em fases pequenas e progressivas.

A proposta é aprender Go durante a construção do projeto, evitando criar abstrações, dependências ou estruturas antes de existir uma necessidade real.

---

## Fase 1 — Modelagem do domínio

### Objetivo

Representar os dados básicos utilizados pelo ForgePath.

### Arquivos

```text
internal/
└── project/
    └── project.go
```

### Tarefas

- [x] Criar o tipo `Technology`
- [x] Criar as constantes das tecnologias suportadas
- [x] Criar a struct `Project`
- [x] Definir os campos iniciais do projeto:
  - [x] Nome
  - [x] Caminho
  - [x] Tecnologia
  - [x] Arquivos marcadores

### Conceitos de Go estudados

- Pacotes
- Tipos definidos
- Tipo subjacente
- Constantes agrupadas
- Structs
- Campos exportados
- Slices

### Resultado esperado

O projeto consegue representar uma tecnologia e um projeto em memória, mas ainda não realiza nenhuma detecção.

---

## Fase 2 — Resultado da detecção

### Objetivo

Representar o resultado produzido quando uma pasta é analisada.

### Arquivos

```text
internal/
└── detector/
    └── detector.go
```

### Tarefas

- [x] Criar a struct `Result`
- [x] Reutilizar o tipo `project.Technology`
- [x] Registrar os arquivos marcadores encontrados
- [x] Remover abstrações que ainda não possuem uso real

### Decisão arquitetural

A interface `Detector` não será utilizada inicialmente.

O projeto começará com uma função concreta. Uma interface só será introduzida quando houver mais de uma implementação ou quando ela facilitar os testes de outro pacote.

### Conceitos de Go estudados

- Imports internos
- Acesso a tipos de outro pacote
- Identificadores exportados
- Composição entre pacotes
- Responsabilidade de cada tipo

### Resultado esperado

O pacote `detector` possui uma estrutura capaz de representar a tecnologia identificada e os marcadores que levaram à identificação.

---

## Fase 3 — Detecção inicial de projetos Go

### Objetivo

Receber o caminho de uma pasta e verificar se ela contém um arquivo `go.mod`.

### Arquivos

```text
internal/
└── detector/
    ├── detector.go
    └── manifest.go
```

### Tarefas

- [x] Criar a função `Detect`
- [x] Receber o caminho da pasta como parâmetro
- [x] Construir o caminho do marcador com `filepath.Join`
- [x] Consultar o arquivo com `os.Stat`
- [x] Tratar a ausência de `go.mod`
- [x] Tratar erros reais do sistema de arquivos
- [x] Confirmar que `go.mod` é um arquivo, não um diretório
- [x] Retornar `project.TechnologyGo`
- [x] Retornar o marcador encontrado

### Assinatura inicial

```go
func Detect(path string) (Result, bool, error)
```

### Estados possíveis

#### Projeto Go encontrado

```text
Result preenchido, true, nil
```

#### Marcador não encontrado

```text
Result vazio, false, nil
```

#### Falha ao analisar a pasta

```text
Result vazio, false, error
```

### Conceitos de Go estudados

- Funções
- Parâmetros
- Múltiplos retornos
- Declaração curta com `:=`
- Biblioteca padrão
- Tratamento explícito de erros
- `nil`
- Condicionais
- Construção de structs
- Valor zero dos tipos

### Resultado esperado

O ForgePath consegue descobrir se uma pasta específica representa um projeto Go.

---

## Fase 4 — Testes do detector Go

### Objetivo

Garantir que a detecção funcione sem depender de pastas reais da máquina.

### Arquivos

```text
internal/
└── detector/
    ├── manifest.go
    └── manifest_test.go
```

### Tarefas

- [x] Criar teste para uma pasta com `go.mod`
- [x] Criar teste para uma pasta sem `go.mod`
- [x] Criar teste para uma pasta chamada `go.mod`
- [x] Criar teste para erro de acesso ou caminho inválido
- [x] Utilizar `t.TempDir()`
- [x] Criar arquivos temporários durante os testes
- [x] Validar tecnologia detectada
- [x] Validar marcadores encontrados
- [x] Executar todos os testes

### Comandos

```bash
go test ./internal/detector
go test ./...
```

### Conceitos de Go estudados

- Pacote `testing`
- Arquivos terminados em `_test.go`
- Funções iniciadas com `Test`
- Diretórios temporários
- Testes isolados
- Falhas com `t.Fatal` e `t.Fatalf`
- Comparação de valores

### Resultado esperado

A detecção de projetos Go está coberta por testes automatizados e funciona de maneira independente do sistema operacional.

---

## Fase 5 — Generalização da detecção

### Objetivo

Adicionar suporte às demais tecnologias previstas no projeto.

### Ordem recomendada

1. Go
2. PHP
3. Java
4. Python
5. TypeScript
6. JavaScript

### Marcadores

| Tecnologia | Marcadores |
| --- | --- |
| Go | `go.mod` |
| PHP | `composer.json` |
| Java | `pom.xml`, `build.gradle`, `build.gradle.kts` |
| Python | `pyproject.toml`, `requirements.txt`, `Pipfile` |
| TypeScript | `package.json` e `tsconfig.json` |
| JavaScript | `package.json` sem `tsconfig.json` |

### Tarefas

- [x] Definir as regras em um único local
- [x] Detectar PHP
- [x] Detectar Java com Maven
- [x] Detectar Java com Gradle
- [x] Detectar Python
- [x] Detectar TypeScript antes de JavaScript
- [x] Detectar JavaScript
- [x] Retornar todos os marcadores relevantes
- [x] Criar testes para cada tecnologia
- [x] Criar testes para prioridades entre regras

### Decisão importante

TypeScript deve ser verificado antes de JavaScript.

Projetos TypeScript geralmente possuem `package.json`, então verificar JavaScript primeiro produziria uma classificação incorreta.

### Conceitos de Go estudados

- Slices de structs
- Iterações com `for`
- Regras ordenadas
- Funções auxiliares
- Testes orientados a tabela
- Separação de responsabilidades
- Refatoração segura

### Resultado esperado

A função de detecção identifica corretamente as tecnologias iniciais do ForgePath.

---

## Fase 6 — Scanner de workspaces

### Objetivo

Analisar as subpastas de um workspace e encontrar todos os projetos reconhecidos.

### Arquivos

```text
internal/
└── scanner/
    ├── scanner.go
    └── scanner_test.go
```

### Responsabilidade do scanner

O detector analisa uma única pasta.

O scanner percorre várias pastas e chama o detector para cada uma delas.

### Tarefas

- [x] Receber o caminho de um workspace
- [x] Validar se o caminho existe
- [x] Validar se o caminho é um diretório
- [x] Ler as entradas do diretório
- [x] Ignorar arquivos comuns
- [x] Ignorar diretórios ocultos
- [x] Ignorar diretórios de dependências e build
- [x] Chamar o detector para cada subdiretório
- [x] Criar valores de `project.Project`
- [x] Ordenar os projetos pelo nome
- [x] Retornar uma slice de projetos
- [x] Criar testes do scanner

### Diretórios ignorados inicialmente

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
```

### Conceitos de Go estudados

- Leitura de diretórios
- Slices de structs
- Laços
- `continue`
- Ordenação
- Propagação de erros
- Funções auxiliares
- Testes de integração entre pacotes

### Resultado esperado

O ForgePath recebe uma pasta como `D:\Development` e retorna os projetos encontrados dentro dela.

---

## Fase 7 — Interface por linha de comando

### Objetivo

Permitir que o usuário execute o ForgePath pelo terminal.

### Arquivos

```text
cmd/
└── forgepath/
    └── main.go

internal/
└── cli/
    ├── root.go
    └── list.go
```

### Comando inicial

```bash
forgepath list [workspace]
```

### Tarefas

- [x] Criar o ponto de entrada em `main.go`
- [x] Configurar o Cobra
- [x] Criar o comando raiz
- [x] Criar o comando `list`
- [x] Usar o diretório atual quando nenhum caminho for informado
- [x] Executar o scanner
- [x] Imprimir os projetos encontrados
- [x] Imprimir erros em `stderr`
- [x] Retornar código de saída adequado
- [x] Exibir ajuda da CLI

### Conceitos de Go estudados

- Pacote `main`
- Função `main`
- Dependências externas
- Organização de comandos
- Entrada e saída do terminal
- Códigos de saída
- Integração entre camadas

### Resultado esperado

O comando abaixo lista projetos detectados:

```bash
forgepath list D:\Development
```

---

## Fase 8 — Integração com o shell

### Objetivo

Permitir que o usuário selecione um projeto e navegue até a pasta no terminal atual.

### Tarefas

- [x] Criar um comando que imprima apenas o caminho selecionado
- [x] Separar a saída visual da saída utilizada pelo shell
- [x] Criar função para PowerShell
- [x] Criar função para Bash
- [x] Testar caminhos com espaços
- [x] Tratar cancelamento da seleção

### Exemplo de uso esperado

```powershell
fp
```

### Conceitos estudados

- Processos pai e filho
- `stdout`
- `stderr`
- Integração entre executável e shell
- Escapamento de caminhos
- Comportamento multiplataforma

### Resultado esperado

O ForgePath consegue devolver um caminho para que o shell altere o diretório atual.

---

## Fase 9 — Interface interativa no terminal

### Objetivo

Substituir a listagem simples por uma TUI navegável.

### Tecnologias previstas

- Bubble Tea
- Bubbles
- Lip Gloss

### Tarefas

- [x] Criar o modelo inicial da TUI
- [x] Carregar os projetos encontrados pelo scanner
- [x] Implementar navegação pelo teclado
- [x] Implementar seleção com Enter
- [x] Implementar pesquisa
- [x] Implementar ajuda de atalhos
- [x] Tratar redimensionamento do terminal
- [x] Separar apresentação e regras de negócio
- [x] Manter detector e scanner independentes da TUI

### Conceitos de Go estudados

- Métodos
- Receivers
- Interfaces reais de bibliotecas
- Arquitetura baseada em mensagens
- Estado imutável ou controlado
- Eventos de teclado
- Renderização textual

### Resultado esperado

O usuário visualiza e seleciona projetos em uma interface interativa dentro do terminal.

---

## Fase 10 — Metadados e recursos adicionais

### Objetivo

Enriquecer a apresentação dos projetos sem comprometer o núcleo já testado.

### Funcionalidades futuras

- [ ] Detectar frameworks
- [ ] Detectar gerenciadores de pacotes
- [ ] Detectar Docker
- [ ] Mostrar branch atual do Git
- [ ] Mostrar alterações não commitadas
- [ ] Adicionar ícones Nerd Font
- [ ] Criar fallback sem ícones
- [ ] Abrir projeto no editor
- [ ] Abrir pasta no gerenciador de arquivos
- [ ] Executar comandos de desenvolvimento
- [ ] Adicionar configuração persistente
- [ ] Adicionar favoritos
- [ ] Adicionar projetos recentes
- [ ] Criar cache de projetos

### Regra de evolução

Cada recurso deve entrar somente quando:

1. Sua responsabilidade estiver clara
2. O comportamento principal estiver testado
3. A nova abstração resolver um problema real
4. A funcionalidade não obrigar detector, scanner e interface a ficarem acoplados

---

# Critérios gerais de qualidade

Em todas as fases:

- [ ] Escrever código simples antes de abstrair
- [ ] Evitar interfaces sem consumidores reais
- [ ] Manter cada pacote com uma responsabilidade clara
- [ ] Utilizar a biblioteca padrão quando ela for suficiente
- [ ] Tratar erros explicitamente
- [ ] Não comparar mensagens de erro por texto
- [ ] Utilizar `filepath` para caminhos multiplataforma
- [ ] Não depender de caminhos locais nos testes
- [ ] Executar `go fmt ./...`
- [ ] Executar `go vet ./...`
- [ ] Executar `go test ./...`
- [ ] Não avançar sem entender o código da fase atual

---

# Estado atual

- [x] Fase 1 — Modelagem do domínio
- [x] Fase 2 — Resultado da detecção
- [x] Fase 3 — Detecção inicial de projetos Go
- [x] Fase 4 — Testes do detector Go
- [x] Fase 5 — Generalização da detecção
- [x] Fase 6 — Scanner de workspaces
- [x] Fase 7 — Interface por linha de comando
- [x] Fase 8 — Integração com o shell
- [x] Fase 9 — Interface interativa no terminal
- [ ] Fase 10 — Metadados e recursos adicionais
