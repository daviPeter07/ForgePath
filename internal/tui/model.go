package tui

import (
	"context"
	"fmt"
	"image/color"
	"io"
	"path/filepath"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/daviPeter07/forgepath/internal/detector"
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

var dockerKey = key.NewBinding(
	key.WithKeys("d"),
	key.WithHelp("d", "docker compose"),
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

type projectDelegate struct {
	graphics bool
}

func (delegate projectDelegate) Height() int {
	if delegate.graphics {
		return 4
	}
	return 2
}
func (projectDelegate) Spacing() int { return 1 }
func (projectDelegate) Update(tea.Msg, *list.Model) tea.Cmd {
	return nil
}

func (delegate projectDelegate) Render(writer io.Writer, model list.Model, index int, raw list.Item) {
	switch item := raw.(type) {
	case directoryItem:
		renderDirectoryItem(writer, model, index, item)
		return
	case editorItem:
		renderEditorItem(writer, model, index, item)
		return
	case dockerItem:
		renderDockerItem(writer, model, index, item)
		return
	}
	item, ok := raw.(projectItem)
	if !ok {
		return
	}
	if delegate.graphics && model.Width() >= 20 {
		if graphic, err := icon.Graphic(item.project.Technology); err == nil {
			renderGraphicProjectItem(writer, model, index, item, graphic)
			return
		}
	}

	iconLabel := icon.Label(item.project.Technology, item.icons)
	if iconLabel == "" {
		iconLabel = "◆"
	}
	iconStyle := lipgloss.NewStyle().Bold(true).Foreground(technologyColor(item.project.Technology))
	if item.icons != icon.ModeNerdFont {
		iconStyle = iconStyle.
			Foreground(technologyBadgeTextColor(item.project.Technology)).
			Background(technologyColor(item.project.Technology)).
			Padding(0, 1)
	}
	nameStyle := lipgloss.NewStyle().Bold(true).Foreground(palette.text)
	metaStyle := lipgloss.NewStyle().Foreground(palette.muted)

	compact := model.Width() < 20
	title := nameStyle.Render(safeTerminalText(item.project.Name))
	if !compact {
		title = iconStyle.Render(iconLabel) + "  " + title
	}
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
			BorderStyle(lipgloss.NormalBorder()).
			BorderLeft(true).
			BorderForeground(palette.primary).
			PaddingLeft(1)
	}
	_, _ = fmt.Fprint(writer, container.Render(content))
}

func renderGraphicProjectItem(writer io.Writer, model list.Model, index int, item projectItem, graphic string) {
	name := safeTerminalText(item.project.Name)
	if item.project.Favorite {
		name = "★  " + name
	}
	description := item.Description()
	if parent := filepath.Base(filepath.Dir(item.project.Path)); parent != "." && parent != "" {
		description += "  ·  " + parent
	}
	description = safeTerminalText(description)

	outerWidth := model.Width() - 4
	if outerWidth < 1 {
		outerWidth = 1
	}
	innerWidth := outerWidth - 1
	if innerWidth < 1 {
		innerWidth = 1
	}
	textWidth := innerWidth - 10
	if textWidth < 1 {
		textWidth = 1
	}
	name = ansi.Truncate(name, textWidth, "…")
	description = ansi.Truncate(description, textWidth, "…")
	text := lipgloss.NewStyle().Bold(true).Foreground(palette.text).Render(name) + "\n" +
		lipgloss.NewStyle().Foreground(palette.muted).Render(description)
	if index == model.Index() {
		nameLine := lipgloss.NewStyle().Bold(true).Foreground(palette.text).Background(palette.surface).Width(textWidth).Render(name)
		descriptionLine := lipgloss.NewStyle().Foreground(palette.muted).Background(palette.surface).Width(textWidth).Render(description)
		blankLine := lipgloss.NewStyle().Background(palette.surface).Width(textWidth).Render("")
		text = nameLine + "\n" + descriptionLine + "\n" + blankLine + "\n" + blankLine
	}
	content := lipgloss.JoinHorizontal(lipgloss.Top, graphic, "  ", text)
	container := lipgloss.NewStyle().Width(innerWidth).MaxWidth(innerWidth).MaxHeight(4)
	if index == model.Index() {
		container = lipgloss.NewStyle().
			Width(outerWidth).
			MaxWidth(outerWidth).
			MaxHeight(4).
			BorderStyle(lipgloss.NormalBorder()).
			BorderLeft(true).
			BorderForeground(palette.primary)
		_, _ = fmt.Fprint(writer, container.Render(content))
		return
	}
	_, _ = fmt.Fprint(writer, " "+container.Render(content))
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

	projectList := list.New(items, projectDelegate{graphics: options.Icons == icon.ModeGraphics}, 80, 24)
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
		return []key.Binding{selectKey, chooseKey, editorKey, dockerKey, backKey}
	}
	projectList.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{selectKey, chooseKey, editorKey, dockerKey, backKey}
	}
	projectList.StatusMessageLifetime = 5 * time.Second

	model := Model{list: projectList, projects: append([]project.Project(nil), projects...), options: options}
	if len(projects) == 0 && options.StartPath != "" {
		// Initialize directly
		directories, err := options.ReadDirectories(options.StartPath)
		if err == nil {
			model.mode = directoryScreen
			model.setProjectDelegate(model.list.Width())
			model.currentPath = options.StartPath
			model.list.Title = "  FORGEPATH  /  " + safeTerminalText(filepath.Base(options.StartPath)) + "  "
			model.list.SetStatusBarItemName("item", "items")
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
					items = append(items, projectItem{project: p, icons: options.Icons})
				} else {
					items = append(items, directoryItem{directory: directory})
				}
			}
			model.list.SetItems(items)
		}
	}
	return model
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch message := message.(type) {
	case tea.WindowSizeMsg:
		m.list.SetSize(message.Width, message.Height)
		if m.mode == projectScreen {
			m.setProjectDelegate(message.Width)
		}
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
			return m, tea.Batch(command, m.newErrorMessage(fmt.Errorf("could not open %s: %w", message.editor, message.err)))
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
				return m, m.newErrorMessage(fmt.Errorf("could not browse directory: %w", err))
			}
			return m, command
		case "c":
			return m, m.confirmCurrentDirectory()
		case "d":
			if m.mode != dockerScreen {
				return m, m.showDocker()
			}
		case "o":
			if m.mode != editorScreen {
				return m, m.showEditors()
			}
		case "backspace", "left":
			command, err := m.goBack()
			if err != nil {
				return m, m.newErrorMessage(fmt.Errorf("could not go back: %w", err))
			}
			return m, command
		case "q":
			m.cancelled = true
			return m, tea.Quit
		case "esc":
			if !m.list.IsFiltered() && m.mode != projectScreen {
				command, err := m.goBack()
				if err != nil {
					return m, m.newErrorMessage(fmt.Errorf("could not go back: %w", err))
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

func (m Model) newErrorMessage(err error) tea.Cmd {
	msg := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#17111F")).
		Background(lipgloss.Color("#EF4444")). // Red background
		Padding(0, 1).
		Render("ERROR") + " " +
		lipgloss.NewStyle().Foreground(lipgloss.Color("#EF4444")).Render(err.Error())
	return m.list.NewStatusMessage(msg)
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
