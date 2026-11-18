package tui

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/daviPeter07/forgepath/internal/project"
)

type dockerItem struct {
	label       string
	description string
	technology  project.Technology
	compose     string // compose file content
	projectPath string
}

func (item dockerItem) FilterValue() string { return item.label }

func renderDockerItem(writer io.Writer, model list.Model, index int, item dockerItem) {
	renderItemBlock(writer, model, index, badge("DOCKER", lipgloss.Color("#3B82F6")), item.label, item.description)
}

// composeFileNames are the files Docker Compose loads by default; generating a
// second one next to any of them would change which file Compose picks.
var composeFileNames = []string{"compose.yaml", "compose.yml", "docker-compose.yml", "docker-compose.yaml"}

func (m *Model) showDocker() tea.Cmd {
	path, selectedProject := m.actionTarget()
	if path == "" {
		return nil
	}
	m.returnMode = m.mode
	m.setScreen(dockerScreen, targetName(path, selectedProject)+breadcrumbSeparator+"Docker Compose", "template", "templates")
	options := generateDockerOptions(path, selectedProject.Technology)
	items := make([]list.Item, len(options))
	for index, option := range options {
		items[index] = option
	}
	return m.list.SetItems(items)
}

func (m *Model) generateDockerCompose(item dockerItem) tea.Cmd {
	for _, name := range composeFileNames {
		if _, err := os.Lstat(filepath.Join(item.projectPath, name)); err == nil {
			return m.newErrorMessage(fmt.Errorf("%s already exists", name))
		}
	}
	if err := writeNewFile(filepath.Join(item.projectPath, "docker-compose.yml"), item.compose); err != nil {
		return m.newErrorMessage(fmt.Errorf("could not generate docker-compose.yml: %w", err))
	}
	for index := range m.projects {
		if samePath(m.projects[index].Path, item.projectPath) {
			m.projects[index].HasDocker = true
		}
	}
	return tea.Batch(m.returnFromAction(), m.newSuccessMessage("Generated docker-compose.yml ("+item.label+")"))
}

// writeNewFile creates path exclusively, so a file created after the existence
// check is never overwritten.
func writeNewFile(path, content string) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	if _, err := file.WriteString(content); err != nil {
		_ = file.Close()
		_ = os.Remove(path)
		return err
	}
	return file.Close()
}

func generateDockerOptions(path string, tech project.Technology) []dockerItem {
	var options []dockerItem

	options = append(options, dockerItem{
		label:       "App + PostgreSQL + Redis",
		description: "App, PostgreSQL 15 and Redis 7",
		projectPath: path,
		compose: `services:
  app:
    build: .
    ports:
      - "8000:8000"
    depends_on:
      - db
      - redis
    # TODO: change to .env variables in production (e.g. ${DB_USER})
    environment:
      - DB_HOST=db
      - DB_PORT=5432
      - DB_NAME=mydb
      - DB_USER=user
      - DB_PASS=password
      - REDIS_HOST=redis
  db:
    image: postgres:15-alpine
    # TODO: change to .env variables in production (e.g. ${POSTGRES_USER})
    environment:
      POSTGRES_USER: user
      POSTGRES_PASSWORD: password
      POSTGRES_DB: mydb
    ports:
      - "5432:5432"
    volumes:
      - pgdata:/var/lib/postgresql/data
  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redisdata:/data
volumes:
  pgdata:
  redisdata:
`,
	})

	options = append(options, dockerItem{
		label:       "App + MySQL + Redis",
		description: "App, MySQL 8 and Redis 7",
		projectPath: path,
		compose: `services:
  app:
    build: .
    ports:
      - "8000:8000"
    depends_on:
      - db
      - redis
    # TODO: change to .env variables in production (e.g. ${DB_USER})
    environment:
      - DB_HOST=db
      - DB_PORT=3306
      - DB_NAME=mydb
      - DB_USER=user
      - DB_PASS=password
      - REDIS_HOST=redis
  db:
    image: mysql:8
    # TODO: change to .env variables in production (e.g. ${MYSQL_USER})
    environment:
      MYSQL_ROOT_PASSWORD: password
      MYSQL_DATABASE: mydb
      MYSQL_USER: user
      MYSQL_PASSWORD: password
    ports:
      - "3306:3306"
    volumes:
      - mysqldata:/var/lib/mysql
  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redisdata:/data
volumes:
  mysqldata:
  redisdata:
`,
	})

	options = append(options, dockerItem{
		label:       "App + PostgreSQL",
		description: "App and PostgreSQL 15",
		projectPath: path,
		compose: `services:
  app:
    build: .
    ports:
      - "8000:8000"
    depends_on:
      - db
    # TODO: change to .env variables in production (e.g. ${DB_USER})
    environment:
      - DB_HOST=db
      - DB_PORT=5432
      - DB_NAME=mydb
      - DB_USER=user
      - DB_PASS=password
  db:
    image: postgres:15-alpine
    # TODO: change to .env variables in production (e.g. ${POSTGRES_USER})
    environment:
      POSTGRES_USER: user
      POSTGRES_PASSWORD: password
      POSTGRES_DB: mydb
    ports:
      - "5432:5432"
    volumes:
      - pgdata:/var/lib/postgresql/data
volumes:
  pgdata:
`,
	})

	options = append(options, dockerItem{
		label:       "App + MySQL",
		description: "App and MySQL 8",
		projectPath: path,
		compose: `services:
  app:
    build: .
    ports:
      - "8000:8000"
    depends_on:
      - db
    # TODO: change to .env variables in production (e.g. ${DB_USER})
    environment:
      - DB_HOST=db
      - DB_PORT=3306
      - DB_NAME=mydb
      - DB_USER=user
      - DB_PASS=password
  db:
    image: mysql:8
    # TODO: change to .env variables in production (e.g. ${MYSQL_USER})
    environment:
      MYSQL_ROOT_PASSWORD: password
      MYSQL_DATABASE: mydb
      MYSQL_USER: user
      MYSQL_PASSWORD: password
    ports:
      - "3306:3306"
    volumes:
      - mysqldata:/var/lib/mysql
volumes:
  mysqldata:
`,
	})

	options = append(options, dockerItem{
		label:       "PostgreSQL Database Only",
		description: "PostgreSQL 15 on port 5432",
		projectPath: path,
		compose: `services:
  db:
    image: postgres:15-alpine
    # TODO: change to .env variables in production (e.g. ${POSTGRES_USER})
    environment:
      POSTGRES_USER: user
      POSTGRES_PASSWORD: password
      POSTGRES_DB: mydb
    ports:
      - "5432:5432"
    volumes:
      - pgdata:/var/lib/postgresql/data
volumes:
  pgdata:
`,
	})

	options = append(options, dockerItem{
		label:       "MySQL Database Only",
		description: "MySQL 8 on port 3306",
		projectPath: path,
		compose: `services:
  db:
    image: mysql:8
    # TODO: change to .env variables in production (e.g. ${MYSQL_USER})
    environment:
      MYSQL_ROOT_PASSWORD: password
      MYSQL_DATABASE: mydb
      MYSQL_USER: user
      MYSQL_PASSWORD: password
    ports:
      - "3306:3306"
    volumes:
      - mysqldata:/var/lib/mysql
volumes:
  mysqldata:
`,
	})

	options = append(options, dockerItem{
		label:       "Redis Server Only",
		description: "Redis 7 on port 6379",
		projectPath: path,
		compose: `services:
  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redisdata:/data
volumes:
  redisdata:
`,
	})

	return options
}
