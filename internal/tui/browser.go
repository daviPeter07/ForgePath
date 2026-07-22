package tui

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/daviPeter07/forgepath/internal/detector"
	"github.com/daviPeter07/forgepath/internal/icon"
	"github.com/daviPeter07/forgepath/internal/ide"
	"github.com/daviPeter07/forgepath/internal/project"
)

type screenMode int

const (
	projectScreen screenMode = iota
	directoryScreen
	editorScreen
	dockerScreen
)

type Directory struct {
	Name string
	Path string
}

type ReadDirectoriesFunc func(string) ([]Directory, error)
type OpenEditorFunc func(context.Context, string, project.Project, ide.IDE) error

type Options struct {
	Icons           icon.Mode
	IDEs            []ide.IDE
	ReadDirectories ReadDirectoriesFunc
	OpenEditor      OpenEditorFunc
	Context         context.Context
	StartPath       string
}

type directoryItem struct {
	directory Directory
}

func (item directoryItem) FilterValue() string { return item.directory.Name }

type editorItem struct {
	editor      ide.IDE
	technology  project.Technology
	recommended bool
}

func (item editorItem) FilterValue() string { return item.editor.Name }

type editorOpenedMsg struct {
	request uint64
	editor  string
	path    string
	err     error
}

func renderDirectoryItem(writer io.Writer, model list.Model, index int, item directoryItem) {
	badge := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#17111F")).
		Background(palette.primary).
		Padding(0, 1).
		Render("DIR")
	title := badge + "  " + lipgloss.NewStyle().Bold(true).Foreground(palette.text).Render(safeTerminalText(item.directory.Name))
	description := lipgloss.NewStyle().Foreground(palette.muted).Render("Directory  ·  Enter to browse")
	renderItemBlock(writer, model, index, title, description)
}

func renderEditorItem(writer io.Writer, model list.Model, index int, item editorItem) {
	badgeColor := palette.primary
	label := "Installed"
	if item.recommended {
		badgeColor = lipgloss.Color("#22C55E")
		label = "Suggested for " + string(item.technology)
	} else if item.editor.Supports(item.technology) {
		label = "Compatible with " + string(item.technology)
	}
	badge := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#17111F")).
		Background(badgeColor).
		Padding(0, 1).
		Render("IDE")
	title := badge + "  " + lipgloss.NewStyle().Bold(true).Foreground(palette.text).Render(item.editor.Name)
	description := lipgloss.NewStyle().Foreground(palette.muted).Render(safeTerminalText(label + "  ·  " + item.editor.Executable))
	renderItemBlock(writer, model, index, title, description)
}

func renderItemBlock(writer io.Writer, model list.Model, index int, title, description string) {
	width := model.Width() - 4
	if width < 1 {
		width = 1
	}
	lineWidth := width - 2
	if lineWidth < 1 {
		lineWidth = 1
	}
	content := ansi.Truncate(title, lineWidth, "…") + "\n" + ansi.Truncate(description, lineWidth, "…")
	container := lipgloss.NewStyle().Width(width).MaxWidth(width).MaxHeight(2).PaddingLeft(2)
	if index == model.Index() {
		container = container.
			Background(palette.surface).
			BorderStyle(lipgloss.NormalBorder()).
			BorderLeft(true).
			BorderForeground(palette.primary).
			PaddingLeft(1)
	}
	_, _ = fmt.Fprint(writer, container.Render(content))
}

func readDirectories(path string) ([]Directory, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	directories := make([]Directory, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() || ignoredBrowserDirectory(entry.Name()) {
			continue
		}
		directories = append(directories, Directory{Name: entry.Name(), Path: filepath.Join(path, entry.Name())})
	}
	return directories, nil
}

func ignoredBrowserDirectory(name string) bool {
	ignored := map[string]struct{}{
		".git": {}, "node_modules": {}, "vendor": {}, "target": {}, ".venv": {}, "venv": {},
		"dist": {}, "build": {}, "__pycache__": {},
	}
	_, found := ignored[name]
	return found
}

func (m *Model) showProjects() tea.Cmd {
	m.mode = projectScreen
	m.setProjectDelegate(m.list.Width())
	m.currentPath = ""
	m.currentProject = project.Project{}
	m.list.ResetFilter()
	m.list.Title = "  FORGEPATH  /  PROJECTS  "
	m.list.SetStatusBarItemName("project", "projects")
	items := make([]list.Item, len(m.projects))
	for index, found := range m.projects {
		items[index] = projectItem{project: found, icons: m.options.Icons}
	}
	return m.list.SetItems(items)
}

func (m *Model) setProjectDelegate(width int) {
	m.list.SetDelegate(projectDelegate{graphics: m.options.Icons == icon.ModeGraphics && width >= 20})
}

func (m *Model) showDirectory(path string) (tea.Cmd, error) {
	directories, err := m.options.ReadDirectories(path)
	if err != nil {
		return nil, err
	}
	m.mode = directoryScreen
	m.list.SetDelegate(projectDelegate{graphics: m.options.Icons == icon.ModeGraphics && m.list.Width() >= 20})
	m.currentPath = path
	m.list.ResetFilter()
	m.list.Title = "  FORGEPATH  /  " + safeTerminalText(filepath.Base(path)) + "  "
	m.list.SetStatusBarItemName("item", "items")
	items := make([]list.Item, 0, len(directories))
	for _, directory := range directories {
		result, found, err := detector.Detect(directory.Path)
		if err == nil && found {
			p := project.Project{
				Name:            directory.Name,
				Path:            directory.Path,
				Technology:      result.Technology,
				Markers:         result.Markers,
				Frameworks:      result.Frameworks,
				PackageManagers: result.PackageManagers,
				HasDocker:       result.HasDocker,
			}
			items = append(items, projectItem{project: p, icons: m.options.Icons})
		} else {
			items = append(items, directoryItem{directory: directory})
		}
	}
	return m.list.SetItems(items), nil
}

func (m *Model) enterSelected() (tea.Cmd, error) {
	switch item := m.list.SelectedItem().(type) {
	case projectItem:
		m.currentProject = item.project
		return m.showDirectory(item.project.Path)
	case directoryItem:
		return m.showDirectory(item.directory.Path)
	case editorItem:
		return m.openSelectedEditor(item), nil
	case dockerItem:
		return m.generateDockerCompose(item), nil
	default:
		return nil, nil
	}
}

func (m *Model) goBack() (tea.Cmd, error) {
	switch m.mode {
	case editorScreen, dockerScreen:
		m.editorRequest++
		m.editorOpening = false
		if m.returnMode == projectScreen {
			return m.showProjects(), nil
		}
		return m.showDirectory(m.currentPath)
	case directoryScreen:
		if samePath(m.currentPath, m.currentProject.Path) {
			return m.showProjects(), nil
		}
		return m.showDirectory(filepath.Dir(m.currentPath))
	default:
		return nil, nil
	}
}

func (m *Model) showEditors() tea.Cmd {
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
	ranked := ide.Rank(m.options.IDEs, selectedProject.Technology)
	if len(ranked) == 0 {
		return m.list.NewStatusMessage("No supported IDE was found on this machine")
	}
	m.returnMode = m.mode
	m.editorPath = path
	m.editorProject = selectedProject
	m.mode = editorScreen
	m.list.SetDelegate(projectDelegate{})
	m.list.ResetFilter()
	m.list.Title = "  OPEN " + safeTerminalText(filepath.Base(path)) + " WITH…  "
	m.list.SetStatusBarItemName("installed IDE", "installed IDEs")
	items := make([]list.Item, len(ranked))
	for index, editor := range ranked {
		items[index] = editorItem{editor: editor, technology: selectedProject.Technology, recommended: index == 0}
	}
	return m.list.SetItems(items)
}

func (m *Model) openSelectedEditor(item editorItem) tea.Cmd {
	if m.options.OpenEditor == nil {
		return m.list.NewStatusMessage("Editor launching is unavailable")
	}
	if m.editorOpening {
		return m.list.NewStatusMessage("Waiting for the editor to start…")
	}
	m.editorRequest++
	request := m.editorRequest
	m.editorOpening = true
	ctx := m.options.Context
	path := m.editorPath
	selectedProject := m.editorProject
	editor := item.editor
	return func() tea.Msg {
		err := m.options.OpenEditor(ctx, path, selectedProject, editor)
		return editorOpenedMsg{request: request, editor: editor.Name, path: path, err: err}
	}
}

func (m *Model) confirmCurrentDirectory() tea.Cmd {
	selected := m.currentProject
	switch item := m.list.SelectedItem().(type) {
	case projectItem:
		selected = item.project
	case directoryItem:
		selected.Path = m.currentPath
		selected.Name = filepath.Base(m.currentPath)
	default:
		if m.mode == editorScreen {
			selected = m.editorProject
			selected.Path = m.editorPath
			selected.Name = filepath.Base(m.editorPath)
		} else if m.mode != directoryScreen {
			return nil
		} else {
			selected.Path = m.currentPath
			selected.Name = filepath.Base(m.currentPath)
		}
	}
	m.selected = selected
	m.hasSelection = true
	return tea.Quit
}

func safeTerminalText(value string) string {
	return strings.Map(func(character rune) rune {
		if character < 0x20 || character == 0x7f || (character >= 0x80 && character <= 0x9f) {
			return '�'
		}
		return character
	}, value)
}

func samePath(left, right string) bool {
	left = filepath.Clean(left)
	right = filepath.Clean(right)
	if runtime.GOOS == "windows" {
		return strings.EqualFold(left, right)
	}
	return left == right
}
