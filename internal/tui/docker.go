package tui

import (
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
	badge := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#17111F")).
		Background(lipgloss.Color("#3B82F6")). // Docker blue
		Padding(0, 1).
		Render("DOCKER")
	title := badge + "  " + lipgloss.NewStyle().Bold(true).Foreground(palette.text).Render(item.label)
	description := lipgloss.NewStyle().Foreground(palette.muted).Render(item.description)
	renderItemBlock(writer, model, index, title, description)
}

func (m *Model) showDocker() tea.Cmd {
	var path string
	var selectedProject project.Project
	switch item := m.list.SelectedItem().(type) {
	case projectItem:
		path = item.project.Path
		selectedProject = item.project
	case directoryItem:
		path = m.currentPath
		selectedProject = m.currentProject
	default:
		if m.mode == directoryScreen {
			path = m.currentPath
			selectedProject = m.currentProject
		}
	}
	if path == "" {
		return nil
	}

	m.returnMode = m.mode
	m.mode = dockerScreen
	m.list.SetDelegate(projectDelegate{})
	m.list.ResetFilter()
	m.list.Title = "  FORGEPATH  /  DOCKER COMPOSE  "
	m.list.SetStatusBarItemName("option", "options")

	options := generateDockerOptions(path, selectedProject.Technology)
	items := make([]list.Item, len(options))
	for i, opt := range options {
		items[i] = opt
	}
	return m.list.SetItems(items)
}

func (m *Model) generateDockerCompose(item dockerItem) tea.Cmd {
	composePath := filepath.Join(item.projectPath, "docker-compose.yml")
	if _, err := os.Stat(composePath); err == nil {
		return m.list.NewStatusMessage(safeTerminalText("docker-compose.yml already exists"))
	}
	err := os.WriteFile(composePath, []byte(item.compose), 0o644)

	// Switch back to previous screen
	var command tea.Cmd
	if m.returnMode == projectScreen {
		command = m.showProjects()
	} else {
		command, _ = m.showDirectory(m.currentPath)
	}

	if err != nil {
		return tea.Batch(command, m.list.NewStatusMessage(safeTerminalText("Failed to generate: "+err.Error())))
	}
	return tea.Batch(command, m.list.NewStatusMessage(safeTerminalText("Generated docker-compose.yml for "+item.label)))
}

func generateDockerOptions(path string, tech project.Technology) []dockerItem {
	var options []dockerItem

	switch tech {
	case project.TechnologyGo:
		options = append(options, dockerItem{
			label:       "Go App + PostgreSQL",
			description: "Generate docker-compose.yml for a Go application and PostgreSQL database",
			projectPath: path,
			compose: `version: '3.8'
services:
  app:
    build: .
    ports:
      - "8080:8080"
    depends_on:
      - db
    environment:
      - DATABASE_URL=postgres://user:password@db:5432/mydb?sslmode=disable
  db:
    image: postgres:15-alpine
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
	case project.TechnologyJavaScript, project.TechnologyTypeScript:
		options = append(options, dockerItem{
			label:       "Node App + MongoDB",
			description: "Generate docker-compose.yml for a Node application and MongoDB database",
			projectPath: path,
			compose: `version: '3.8'
services:
  app:
    build: .
    ports:
      - "3000:3000"
    depends_on:
      - db
    environment:
      - MONGO_URI=mongodb://db:27017/mydb
  db:
    image: mongo:6
    ports:
      - "27017:27017"
    volumes:
      - mongodata:/data/db
volumes:
  mongodata:
`,
		})
	case project.TechnologyPHP:
		options = append(options, dockerItem{
			label:       "PHP/Laravel + MySQL + Redis",
			description: "Generate docker-compose.yml for PHP, MySQL and Redis",
			projectPath: path,
			compose: `version: '3.8'
services:
  app:
    build: .
    ports:
      - "8000:8000"
    depends_on:
      - db
      - redis
    environment:
      - DB_CONNECTION=mysql
      - DB_HOST=db
      - DB_PORT=3306
      - DB_DATABASE=mydb
      - DB_USERNAME=user
      - DB_PASSWORD=password
      - REDIS_HOST=redis
  db:
    image: mysql:8
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
	case project.TechnologyPython:
		options = append(options, dockerItem{
			label:       "Python App + PostgreSQL + Redis",
			description: "Generate docker-compose.yml for Python, PostgreSQL and Redis",
			projectPath: path,
			compose: `version: '3.8'
services:
  app:
    build: .
    ports:
      - "8000:8000"
    depends_on:
      - db
      - redis
    environment:
      - DATABASE_URL=postgresql://user:password@db:5432/mydb
      - REDIS_URL=redis://redis:6379/0
  db:
    image: postgres:15-alpine
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
	case project.TechnologyJava:
		options = append(options, dockerItem{
			label:       "Java Spring Boot + PostgreSQL",
			description: "Generate docker-compose.yml for Java and PostgreSQL",
			projectPath: path,
			compose: `version: '3.8'
services:
  app:
    build: .
    ports:
      - "8080:8080"
    depends_on:
      - db
    environment:
      - SPRING_DATASOURCE_URL=jdbc:postgresql://db:5432/mydb
      - SPRING_DATASOURCE_USERNAME=user
      - SPRING_DATASOURCE_PASSWORD=password
  db:
    image: postgres:15-alpine
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
	}

	options = append(options, dockerItem{
		label:       "PostgreSQL Database Only",
		description: "Generate docker-compose.yml for a standalone PostgreSQL database",
		projectPath: path,
		compose: `version: '3.8'
services:
  db:
    image: postgres:15-alpine
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
		description: "Generate docker-compose.yml for a standalone MySQL database",
		projectPath: path,
		compose: `version: '3.8'
services:
  db:
    image: mysql:8
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
		description: "Generate docker-compose.yml for a standalone Redis server",
		projectPath: path,
		compose: `version: '3.8'
services:
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
	options = append(options, dockerItem{
		label:       "MongoDB Database Only",
		description: "Generate docker-compose.yml for a standalone MongoDB database",
		projectPath: path,
		compose: `version: '3.8'
services:
  db:
    image: mongo:6
    ports:
      - "27017:27017"
    volumes:
      - mongodata:/data/db
volumes:
  mongodata:
`,
	})

	return options
}
