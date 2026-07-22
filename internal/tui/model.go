package tui

import (
	"context"
	"fmt"
	"image/color"
	"io"
	"path/filepath"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/daviPeter07/forgepath/internal/icon"
	"github.com/daviPeter07/forgepath/internal/project"
)

var selectKey = key.NewBinding(
	key.WithKeys("enter"),
	key.WithHelp("enter", "browse"),
)

var chooseKey = key.NewBinding(
	key.WithKeys("c"),
	key.WithHelp("c", "cd here"),
)

var editorKey = key.NewBinding(
	key.WithKeys("o"),
	key.WithHelp("o", "open IDE"),
)

var backKey = key.NewBinding(
	key.WithKeys("backspace", "left"),
	key.WithHelp("←/backspace", "back"),
)

var palette = struct {
	primary  color.Color
	bright   color.Color
	surface  color.Color
	text     color.Color
	muted    color.Color
	favorite color.Color
}{
	primary:  lipgloss.Color("#A855F7"),
	bright:   lipgloss.Color("#D8B4FE"),
	surface:  lipgloss.Color("#241532"),
	text:     lipgloss.Color("#F7F2FF"),
	muted:    lipgloss.Color("#A99DB8"),
	favorite: lipgloss.Color("#FACC15"),
}

type projectItem struct {
	project project.Project
	icons   icon.Mode
}

type projectDelegate struct{}

func (projectDelegate) Height() int  { return 2 }
func (projectDelegate) Spacing() int { return 1 }
func (projectDelegate) Update(tea.Msg, *list.Model) tea.Cmd {
	return nil
}

func (projectDelegate) Render(writer io.Writer, model list.Model, index int, raw list.Item) {
	switch item := raw.(type) {
	case directoryItem:
		renderDirectoryItem(writer, model, index, item)
		return
	case editorItem:
		renderEditorItem(writer, model, index, item)
		return
	}
	item, ok := raw.(projectItem)
	if !ok {
		return
	}

	iconLabel := icon.Label(item.project.Technology, item.icons)
	if iconLabel == "" {
		iconLabel = "◆"
	}
	iconStyle := lipgloss.NewStyle().Bold(true).Foreground(technologyColor(item.project.Technology))
	if item.icons == icon.ModeASCII {
		iconStyle = iconStyle.
			Foreground(technologyBadgeTextColor(item.project.Technology)).
			Background(technologyColor(item.project.Technology)).
			Padding(0, 1)
	}
	nameStyle := lipgloss.NewStyle().Bold(true).Foreground(palette.text)
	metaStyle := lipgloss.NewStyle().Foreground(palette.muted)

	title := iconStyle.Render(iconLabel) + "  " + nameStyle.Render(safeTerminalText(item.project.Name))
	if item.project.Favorite {
		title = lipgloss.NewStyle().Foreground(palette.favorite).Render("★") + "  " + title
	}
	description := item.Description()
	if parent := filepath.Base(filepath.Dir(item.project.Path)); parent != "." && parent != "" {
		description += "  ·  " + parent
	}
	description = safeTerminalText(description)

	width := model.Width() - 4
	if width < 1 {
		width = 1
	}
	lineWidth := width - 2
	if lineWidth < 1 {
		lineWidth = 1
	}
	content := ansi.Truncate(title, lineWidth, "…") + "\n" + ansi.Truncate(metaStyle.Render(description), lineWidth, "…")
	container := lipgloss.NewStyle().Width(width).MaxWidth(width).MaxHeight(2).PaddingLeft(2)
	if index == model.Index() {
		container = container.
			Bold(true).
			Foreground(palette.bright).
			Background(palette.surface).
			BorderStyle(lipgloss.NormalBorder()).
			BorderLeft(true).
			BorderForeground(palette.primary).
			PaddingLeft(1)
	}
	_, _ = fmt.Fprint(writer, container.Render(content))
}

func technologyColor(technology project.Technology) color.Color {
	colors := map[project.Technology]color.Color{
		project.TechnologyTypeScript: lipgloss.Color("#38BDF8"),
		project.TechnologyJavaScript: lipgloss.Color("#FACC15"),
		project.TechnologyPython:     lipgloss.Color("#60A5FA"),
		project.TechnologyGo:         lipgloss.Color("#22D3EE"),
		project.TechnologyJava:       lipgloss.Color("#FB7185"),
		project.TechnologyPHP:        lipgloss.Color("#A78BFA"),
		project.TechnologyDocker:     lipgloss.Color("#3B82F6"),
		project.TechnologyRust:       lipgloss.Color("#F97316"),
		project.TechnologyRuby:       lipgloss.Color("#EF4444"),
		project.TechnologySwift:      lipgloss.Color("#FB7185"),
		project.TechnologyElixir:     lipgloss.Color("#C084FC"),
	}
	if color, ok := colors[technology]; ok {
		return color
	}
	return palette.bright
}

func technologyBadgeTextColor(_ project.Technology) color.Color {
	return lipgloss.Color("#17111F")
}

func (item projectItem) Title() string {
	prefix := icon.Label(item.project.Technology, item.icons)
	if item.project.Favorite {
		if item.icons == icon.ModeNerdFont {
			prefix = " " + prefix
		} else {
			prefix = "[F] " + prefix
		}
	}
	return prefix + " " + item.project.Name
}

func (item projectItem) Description() string {
	details := []string{string(item.project.Technology)}
	for _, framework := range item.project.Frameworks {
		details = append(details, string(framework))
	}
	for _, manager := range item.project.PackageManagers {
		details = append(details, string(manager))
	}
	if item.project.HasDocker && item.project.Technology != project.TechnologyDocker {
		details = append(details, "Docker")
	}
	if item.project.GitBranch != "" {
		branch := item.project.GitBranch
		if !item.project.GitStatusKnown {
			branch += "?"
		} else if item.project.GitDirty {
			branch += "*"
		}
		details = append(details, branch)
	}
	return strings.Join(details, " | ")
}

func (item projectItem) FilterValue() string {
	return item.project.Name + " " + item.Description()
}

type Model struct {
	list           list.Model
	projects       []project.Project
	options        Options
	mode           screenMode
	returnMode     screenMode
	currentPath    string
	currentProject project.Project
	editorPath     string
	editorProject  project.Project
	editorRequest  uint64
	editorOpening  bool
	selected       project.Project
	hasSelection   bool
	cancelled      bool
}

func NewModel(projects []project.Project, icons icon.Mode) Model {
	return NewModelWithOptions(projects, Options{Icons: icons})
}

func NewModelWithOptions(projects []project.Project, options Options) Model {
	if options.ReadDirectories == nil {
		options.ReadDirectories = readDirectories
	}
	if options.Context == nil {
		options.Context = context.Background()
	}
	items := make([]list.Item, len(projects))
	for index, found := range projects {
		items[index] = projectItem{project: found, icons: options.Icons}
	}

	projectList := list.New(items, projectDelegate{}, 80, 24)
	projectList.Title = "  FORGEPATH  /  PROJECT SWITCHER  "
	projectList.SetStatusBarItemName("project", "projects")
	projectList.Styles.Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(palette.text).
		Background(palette.primary).
		Padding(0, 1)
	projectList.Styles.DefaultFilterCharacterMatch = lipgloss.NewStyle().
		Bold(true).
		Foreground(palette.bright)
	projectList.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{selectKey, chooseKey, editorKey, backKey}
	}
	projectList.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{selectKey, chooseKey, editorKey, backKey}
	}

	return Model{list: projectList, projects: append([]project.Project(nil), projects...), options: options}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch message := message.(type) {
	case tea.WindowSizeMsg:
		m.list.SetSize(message.Width, message.Height)
	case editorOpenedMsg:
		if message.request != m.editorRequest {
			return m, nil
		}
		m.editorOpening = false
		var command tea.Cmd
		if m.returnMode == projectScreen {
			command = m.showProjects()
		} else {
			command, _ = m.showDirectory(m.currentPath)
		}
		if message.err != nil {
			return m, tea.Batch(command, m.list.NewStatusMessage(safeTerminalText("Could not open "+message.editor+": "+message.err.Error())))
		}
		return m, tea.Batch(command, m.list.NewStatusMessage(safeTerminalText("Opened "+filepath.Base(message.path)+" in "+message.editor)))
	case tea.KeyPressMsg:
		if message.String() == "ctrl+c" {
			m.cancelled = true
			return m, tea.Quit
		}
		if m.list.SettingFilter() {
			break
		}
		switch message.String() {
		case "enter":
			command, err := m.enterSelected()
			if err != nil {
				return m, m.list.NewStatusMessage(safeTerminalText("Could not browse directory: " + err.Error()))
			}
			return m, command
		case "c":
			return m, m.confirmCurrentDirectory()
		case "o":
			if m.mode != editorScreen {
				return m, m.showEditors()
			}
		case "backspace", "left":
			command, err := m.goBack()
			if err != nil {
				return m, m.list.NewStatusMessage(safeTerminalText("Could not go back: " + err.Error()))
			}
			return m, command
		case "q":
			m.cancelled = true
			return m, tea.Quit
		case "esc":
			if !m.list.IsFiltered() && m.mode != projectScreen {
				command, err := m.goBack()
				if err != nil {
					return m, m.list.NewStatusMessage(safeTerminalText("Could not go back: " + err.Error()))
				}
				return m, command
			}
			if !m.list.IsFiltered() {
				m.cancelled = true
				return m, tea.Quit
			}
		}
	}

	var command tea.Cmd
	m.list, command = m.list.Update(message)
	return m, command
}

func (m Model) View() tea.View {
	view := tea.NewView(m.list.View())
	view.AltScreen = true
	view.WindowTitle = "ForgePath project browser"
	return view
}

func (m Model) Selection() (project.Project, bool) {
	return m.selected, m.hasSelection
}

func (m Model) Cancelled() bool {
	return m.cancelled
}

func Select(ctx context.Context, projects []project.Project, icons icon.Mode, input io.Reader, output io.Writer) (project.Project, bool, error) {
	return SelectWithOptions(ctx, projects, Options{Icons: icons}, input, output)
}

func SelectWithOptions(ctx context.Context, projects []project.Project, options Options, input io.Reader, output io.Writer) (project.Project, bool, error) {
	options.Context = ctx
	program := tea.NewProgram(
		NewModelWithOptions(projects, options),
		tea.WithContext(ctx),
		tea.WithInput(input),
		tea.WithOutput(output),
	)

	finalModel, err := program.Run()
	if err != nil {
		return project.Project{}, false, err
	}

	model, ok := finalModel.(Model)
	if !ok {
		return project.Project{}, false, fmt.Errorf("unexpected TUI model %T", finalModel)
	}

	selected, found := model.Selection()
	return selected, found, nil
}
