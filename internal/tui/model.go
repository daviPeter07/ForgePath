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
	"github.com/daviPeter07/forgepath/internal/icon"
	"github.com/daviPeter07/forgepath/internal/project"
)

// Each screen advertises only the keys that do something on it, and Enter is
// described by what it actually does there.
var (
	browseKey   = key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "browse"))
	openKey     = key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "open"))
	generateKey = key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "generate"))
	chooseKey   = key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "cd here"))
	editorKey   = key.NewBinding(key.WithKeys("o"), key.WithHelp("o", "open IDE"))
	dockerKey   = key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "compose"))
	backKey     = key.NewBinding(key.WithKeys("backspace", "left"), key.WithHelp("←", "back"))
)

var screenKeys = map[screenMode][]key.Binding{
	projectScreen:   {browseKey, chooseKey, editorKey, dockerKey},
	directoryScreen: {browseKey, chooseKey, editorKey, dockerKey, backKey},
	editorScreen:    {openKey, chooseKey, backKey},
	dockerScreen:    {generateKey, backKey},
}

var palette = struct {
	primary  color.Color
	bright   color.Color
	surface  color.Color
	text     color.Color
	muted    color.Color
	favorite color.Color
	ink      color.Color
	danger   color.Color
	success  color.Color
}{
	primary:  lipgloss.Color("#A855F7"),
	bright:   lipgloss.Color("#D8B4FE"),
	surface:  lipgloss.Color("#241532"),
	text:     lipgloss.Color("#F7F2FF"),
	muted:    lipgloss.Color("#A99DB8"),
	favorite: lipgloss.Color("#FACC15"),
	ink:      lipgloss.Color("#17111F"),
	danger:   lipgloss.Color("#EF4444"),
	success:  lipgloss.Color("#22C55E"),
}

const breadcrumbSeparator = " › "

// renderTitle draws the header as a brand chip followed by a breadcrumb.
func renderTitle(crumb string) string {
	brand := lipgloss.NewStyle().Bold(true).Foreground(palette.ink).Background(palette.primary).Padding(0, 1).Render("◆ FORGEPATH")
	trail := lipgloss.NewStyle().Bold(true).Foreground(palette.bright).Background(palette.surface).Padding(0, 1).Render(safeTerminalText(crumb))
	return brand + trail
}

// breadcrumb names path relative to base ("app › src › api"). Deep paths keep
// the base and the last two segments so the header stays on one line.
func breadcrumb(base, path string) string {
	if base != "" {
		relative, err := filepath.Rel(base, path)
		if err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			parts := []string{filepath.Base(base)}
			if relative != "." {
				parts = append(parts, strings.Split(relative, string(filepath.Separator))...)
			}
			if len(parts) > 4 {
				parts = append([]string{parts[0], "…"}, parts[len(parts)-2:]...)
			}
			return strings.Join(parts, breadcrumbSeparator)
		}
	}
	return truncateStart(path, maxCrumbWidth)
}

const maxCrumbWidth = 60

// truncateStart keeps the end of value, the most specific part of a path.
func truncateStart(value string, width int) string {
	runes := []rune(value)
	if len(runes) <= width {
		return value
	}
	return "…" + string(runes[len(runes)-width+1:])
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

	selected := index == model.Index()
	styles := newRowStyles(selected)
	iconLabel := icon.Label(item.project.Technology, item.icons)
	if iconLabel == "" {
		iconLabel = "◆"
	}
	iconStyle := styles.base.Bold(true).Foreground(technologyColor(item.project.Technology))
	if item.icons != icon.ModeNerdFont {
		iconStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(technologyBadgeTextColor(item.project.Technology)).
			Background(technologyColor(item.project.Technology)).
			Padding(0, 1)
	}

	width := max(model.Width()-4, 1)
	lineWidth := max(width-2, 1)
	title := styles.name.Render(safeTerminalText(item.project.Name))
	if model.Width() >= 20 {
		title = iconStyle.Render(iconLabel) + styles.base.Render("  ") + title
	}
	if item.project.Favorite {
		title = styles.star.Render("★") + styles.base.Render("  ") + title
	}
	description := styles.meta.Render(safeTerminalText(projectSubtitle(item)))
	content := ansi.Truncate(title, lineWidth, "…") + "\n" + ansi.Truncate(description, lineWidth, "…")
	_, _ = fmt.Fprint(writer, rowContainer(selected, width, 2).Render(content))
}

// rowStyles carries the selection background into every segment, because each
// inner style ends with a reset that would otherwise clear the row background.
type rowStyles struct {
	base lipgloss.Style
	name lipgloss.Style
	meta lipgloss.Style
	star lipgloss.Style
}

func newRowStyles(selected bool) rowStyles {
	base := lipgloss.NewStyle()
	if selected {
		base = base.Background(palette.surface)
	}
	return rowStyles{
		base: base,
		name: base.Bold(true).Foreground(palette.text),
		meta: base.Foreground(palette.muted),
		star: base.Bold(true).Foreground(palette.favorite),
	}
}

// rowContainer gives selected and unselected rows the same geometry: a left
// border that is only visible on the selected row.
func rowContainer(selected bool, width, height int) lipgloss.Style {
	style := lipgloss.NewStyle().
		Width(width).
		MaxWidth(width).
		MaxHeight(height).
		BorderLeft(true).
		PaddingLeft(1)
	if selected {
		return style.
			Background(palette.surface).
			BorderStyle(lipgloss.ThickBorder()).
			BorderForeground(palette.primary)
	}
	return style.BorderStyle(lipgloss.HiddenBorder())
}

// projectSubtitle is the metadata line under a project name, followed by the
// folder that contains the project.
func projectSubtitle(item projectItem) string {
	description := item.Description()
	if parent := filepath.Base(filepath.Dir(item.project.Path)); parent != "." && parent != "" && parent != string(filepath.Separator) {
		description += "  ·  " + parent
	}
	return description
}

func renderGraphicProjectItem(writer io.Writer, model list.Model, index int, item projectItem, graphic string) {
	selected := index == model.Index()
	styles := newRowStyles(selected)
	const gap = 2
	width := max(model.Width()-4, 1)
	graphicWidth := lipgloss.Width(graphic)
	textWidth := max(width-2-graphicWidth-gap, 1)

	prefix, prefixWidth := "", 0
	if item.project.Favorite {
		prefix, prefixWidth = styles.star.Render("★")+styles.base.Render(" "), 2
	}
	name := ansi.Truncate(safeTerminalText(item.project.Name), max(textWidth-prefixWidth, 1), "…")
	description := ansi.Truncate(safeTerminalText(projectSubtitle(item)), textWidth, "…")

	line := styles.base.Width(textWidth)
	lines := []string{line.Render(prefix + styles.name.Render(name)), line.Render(styles.meta.Render(description))}
	for len(lines) < lipgloss.Height(graphic) {
		lines = append(lines, line.Render("")) // fill the icon height so the selection band is solid
	}
	text := strings.Join(lines, "\n")
	content := lipgloss.JoinHorizontal(lipgloss.Top, graphic, styles.base.Render(strings.Repeat(" ", gap)), text)
	_, _ = fmt.Fprint(writer, rowContainer(selected, width, 4).Render(content))
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
	// parentProjects holds the enclosing projects while browsing into a nested
	// project, so going back restores the right root and IDE suggestions.
	parentProjects []project.Project
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
	projectList := list.New(nil, projectDelegate{graphics: options.Icons == icon.ModeGraphics}, 80, 24)
	projectList.Styles.Title = lipgloss.NewStyle()
	projectList.Styles.DefaultFilterCharacterMatch = lipgloss.NewStyle().
		Bold(true).
		Underline(true).
		Foreground(palette.favorite)
	projectList.StatusMessageLifetime = 5 * time.Second
	// Short cursor labels leave room in the help bar for the screen actions.
	projectList.KeyMap.CursorUp.SetHelp("↑", "up")
	projectList.KeyMap.CursorDown.SetHelp("↓", "down")

	model := Model{list: projectList, projects: append([]project.Project(nil), projects...), options: options}
	if len(projects) == 0 && options.StartPath != "" {
		if _, err := model.showDirectory(options.StartPath); err == nil {
			return model
		}
	}
	model.showProjects()
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
		return m, tea.Batch(command, m.newSuccessMessage("Opened "+filepath.Base(message.path)+" in "+message.editor))
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

// newErrorMessage must use a pointer receiver: the status message is stored in
// the list, and a value receiver would drop it with the copy.
func (m *Model) newErrorMessage(err error) tea.Cmd {
	msg := lipgloss.NewStyle().
		Bold(true).
		Foreground(palette.ink).
		Background(palette.danger).
		Padding(0, 1).
		Render("ERROR") + " " +
		lipgloss.NewStyle().Foreground(palette.danger).Render(safeTerminalText(err.Error()))
	return m.list.NewStatusMessage(msg)
}

func (m *Model) newSuccessMessage(text string) tea.Cmd {
	msg := lipgloss.NewStyle().
		Bold(true).
		Foreground(palette.ink).
		Background(palette.success).
		Padding(0, 1).
		Render("OK") + " " +
		lipgloss.NewStyle().Foreground(palette.success).Render(safeTerminalText(text))
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
