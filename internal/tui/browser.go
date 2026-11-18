package tui

import (
	"context"
	"fmt"
	"image/color"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

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
	// ToggleFavorite persists the favorite flag of a project and returns the
	// new value. Without it the f key and its help are hidden.
	ToggleFavorite func(project.Project) (bool, error)
	// Hint is shown as a status message when the browser opens.
	Hint string
	// Now is the clock used for "opened 2h ago" labels.
	Now func() time.Time
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
	renderItemBlock(writer, model, index, badge("DIR", palette.muted), safeTerminalText(item.directory.Name), "Folder")
}

func renderEditorItem(writer io.Writer, model list.Model, index int, item editorItem) {
	badgeColor := palette.primary
	label := "Installed"
	if item.recommended {
		badgeColor = palette.success
		label = "Suggested for " + string(item.technology)
	} else if item.editor.Supports(item.technology) {
		label = "Compatible with " + string(item.technology)
	}
	renderItemBlock(writer, model, index, badge("IDE", badgeColor), safeTerminalText(item.editor.Name), safeTerminalText(label+"  ·  "+item.editor.Executable))
}

func badge(label string, background color.Color) string {
	return lipgloss.NewStyle().
		Bold(true).
		Foreground(palette.ink).
		Background(background).
		Padding(0, 1).
		Render(label)
}

func renderItemBlock(writer io.Writer, model list.Model, index int, badge, title, description string) {
	selected := index == model.Index()
	styles := newRowStyles(selected)
	width := max(model.Width()-4, 1)
	lineWidth := max(width-2, 1)
	content := ansi.Truncate(badge+styles.base.Render("  ")+styles.name.Render(title), lineWidth, "…") + "\n" +
		ansi.Truncate(styles.meta.Render(description), lineWidth, "…")
	_, _ = fmt.Fprint(writer, rowContainer(selected, width, 2).Render(content))
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

// setScreen switches the list to a screen: header, help keys and item names.
func (m *Model) setScreen(mode screenMode, crumb, singular, plural string) {
	m.mode = mode
	m.list.ResetFilter()
	m.list.Title = renderTitle(crumb)
	m.list.SetStatusBarItemName(singular, plural)
	if mode == projectScreen || mode == directoryScreen {
		m.setProjectDelegate(m.list.Width())
	} else {
		m.list.SetDelegate(projectDelegate{})
	}
}

func (m *Model) showProjects() tea.Cmd {
	m.setScreen(projectScreen, "Projects", "project", "projects")
	m.currentPath = ""
	m.currentProject = project.Project{}
	m.parentProjects = nil
	return m.list.SetItems(m.projectItems())
}

func (m *Model) projectItems() []list.Item {
	now := m.options.Now()
	items := make([]list.Item, len(m.projects))
	for index, found := range m.projects {
		items[index] = projectItem{project: found, icons: m.options.Icons, now: now}
	}
	return items
}

func (m *Model) setProjectDelegate(width int) {
	m.list.SetDelegate(projectDelegate{graphics: m.options.Icons == icon.ModeGraphics && width >= 20})
}

// browseRoot is the folder the breadcrumb starts from: the outermost project
// being browsed, or the start folder when no project was found.
func (m *Model) browseRoot() string {
	if len(m.parentProjects) > 0 && m.parentProjects[0].Path != "" {
		return m.parentProjects[0].Path
	}
	if m.currentProject.Path != "" {
		return m.currentProject.Path
	}
	return m.options.StartPath
}

func (m *Model) showDirectory(path string) (tea.Cmd, error) {
	directories, err := m.options.ReadDirectories(path)
	if err != nil {
		return nil, err
	}
	m.currentPath = path
	m.setScreen(directoryScreen, breadcrumb(m.browseRoot(), path), "item", "items")
	items := make([]list.Item, 0, len(directories))
	for _, directory := range directories {
		result, found, err := detector.Detect(directory.Path)
		if err != nil || !found {
			items = append(items, directoryItem{directory: directory})
			continue
		}
		items = append(items, projectItem{icons: m.options.Icons, project: project.Project{
			Name:            directory.Name,
			Path:            directory.Path,
			Technology:      result.Technology,
			Markers:         result.Markers,
			Frameworks:      result.Frameworks,
			PackageManagers: result.PackageManagers,
			HasDocker:       result.HasDocker,
		}})
	}
	return m.list.SetItems(items), nil
}

func (m *Model) enterSelected() (tea.Cmd, error) {
	switch item := m.list.SelectedItem().(type) {
	case projectItem:
		previous, previousParents := m.currentProject, m.parentProjects
		if m.mode == directoryScreen {
			m.parentProjects = append(append([]project.Project(nil), m.parentProjects...), m.currentProject)
		}
		m.currentProject = item.project
		command, err := m.showDirectory(item.project.Path)
		if err != nil {
			m.currentProject, m.parentProjects = previous, previousParents
		}
		return command, err
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

func (m *Model) returnFromAction() tea.Cmd {
	if m.returnMode == projectScreen {
		return m.showProjects()
	}
	command, err := m.showDirectory(m.currentPath)
	if err != nil {
		return tea.Batch(m.showProjects(), m.newErrorMessage(fmt.Errorf("could not reopen directory: %w", err)))
	}
	return command
}

func (m *Model) goBack() (tea.Cmd, error) {
	switch m.mode {
	case editorScreen, dockerScreen:
		m.editorRequest++
		m.editorOpening = false
		return m.returnFromAction(), nil
	case directoryScreen:
		if m.currentProject.Path != "" && samePath(m.currentPath, m.currentProject.Path) {
			if len(m.parentProjects) == 0 {
				return m.showProjects(), nil
			}
			last := len(m.parentProjects) - 1
			m.currentProject = m.parentProjects[last]
			m.parentProjects = m.parentProjects[:last]
		}
		return m.showDirectory(filepath.Dir(m.currentPath))
	default:
		return nil, nil
	}
}

// actionTarget is the folder an action (IDE, compose) applies to: the
// highlighted project on the project list, otherwise the displayed folder.
func (m *Model) actionTarget() (string, project.Project) {
	switch m.mode {
	case projectScreen:
		if item, ok := m.list.SelectedItem().(projectItem); ok {
			return item.project.Path, item.project
		}
	case directoryScreen:
		if item, ok := m.list.SelectedItem().(projectItem); ok {
			return item.project.Path, item.project
		}
		return m.currentPath, m.currentProject
	}
	return "", project.Project{}
}

// targetName labels an action target by its project name when the target is
// the project itself, and by its folder name otherwise.
func targetName(path string, target project.Project) string {
	if target.Name != "" && samePath(path, target.Path) {
		return target.Name
	}
	return filepath.Base(path)
}

func (m *Model) showEditors() tea.Cmd {
	path, selectedProject := m.actionTarget()
	if path == "" {
		return nil
	}
	ranked := ide.Rank(m.options.IDEs, selectedProject.Technology)
	if len(ranked) == 0 {
		return m.newErrorMessage(fmt.Errorf("no supported IDE was found on this machine"))
	}
	m.returnMode = m.mode
	m.editorPath = path
	m.editorProject = selectedProject
	m.setScreen(editorScreen, targetName(path, selectedProject)+breadcrumbSeparator+"Open with…", "installed IDE", "installed IDEs")
	items := make([]list.Item, len(ranked))
	for index, editor := range ranked {
		items[index] = editorItem{editor: editor, technology: selectedProject.Technology, recommended: index == 0 && editor.Supports(selectedProject.Technology)}
	}
	return m.list.SetItems(items)
}

func (m *Model) openSelectedEditor(item editorItem) tea.Cmd {
	if m.options.OpenEditor == nil {
		return m.newErrorMessage(fmt.Errorf("editor launching is unavailable"))
	}
	if m.editorOpening {
		return m.list.NewStatusMessage("Waiting for the editor to start…")
	}
	m.editorRequest++
	request := m.editorRequest
	m.editorOpening = true
	ctx := m.options.Context
	openEditor := m.options.OpenEditor
	path := m.editorPath
	selectedProject := m.editorProject
	editor := item.editor
	return func() tea.Msg {
		err := openEditor(ctx, path, selectedProject, editor)
		return editorOpenedMsg{request: request, editor: editor.Name, path: path, err: err}
	}
}

// confirmCurrentDirectory selects the folder the user is looking at: the
// highlighted project on the project list, otherwise the displayed folder.
func (m *Model) confirmCurrentDirectory() tea.Cmd {
	var selected project.Project
	switch m.mode {
	case projectScreen:
		item, ok := m.list.SelectedItem().(projectItem)
		if !ok {
			return nil
		}
		selected = item.project
	case directoryScreen:
		selected = m.currentProject
		selected.Path = m.currentPath
		selected.Name = filepath.Base(m.currentPath)
	case editorScreen:
		selected = m.editorProject
		selected.Path = m.editorPath
		selected.Name = filepath.Base(m.editorPath)
	default:
		return nil
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
