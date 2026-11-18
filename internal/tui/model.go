package tui

import (
	"context"
	"fmt"
	"image/color"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

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
	favoriteKey = key.NewBinding(key.WithKeys("f"), key.WithHelp("f", "favorite"))
	backKey     = key.NewBinding(key.WithKeys("backspace", "left"), key.WithHelp("←", "back"))
	moveKey     = key.NewBinding(key.WithKeys("up", "down"), key.WithHelp("↑↓", "move"))
	filterKey   = key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "filter"))
	clearKey    = key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "clear filter"))
	quitKey     = key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit"))
	applyKey    = key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "apply"))
	cancelKey   = key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel"))
)

// screenKeys are the actions of each screen, shown on the footer's second line.
var screenKeys = map[screenMode][]key.Binding{
	projectScreen:   {browseKey, chooseKey, editorKey, dockerKey, favoriteKey},
	directoryScreen: {browseKey, chooseKey, editorKey, dockerKey},
	editorScreen:    {openKey, chooseKey},
	dockerScreen:    {generateKey},
}

// footerHeight is the blank line plus the navigation and action lines.
const footerHeight = 3

// footerView replaces the list's one-line help with two lines, navigation and
// actions, so every key of the screen fits without truncation.
func (m Model) footerView() string {
	if m.list.SettingFilter() {
		return renderKeyLine("filter", []key.Binding{applyKey, cancelKey}) + "\n"
	}
	navigation := []key.Binding{moveKey, filterKey}
	if m.list.IsFiltered() {
		navigation = append(navigation, clearKey)
	}
	if m.mode != projectScreen {
		navigation = append(navigation, backKey)
	}
	navigation = append(navigation, quitKey)

	var actions []key.Binding
	for _, binding := range screenKeys[m.mode] {
		if binding.Help().Desc == "favorite" && m.options.ToggleFavorite == nil {
			continue
		}
		actions = append(actions, binding)
	}
	if m.mode == projectScreen {
		description := "changes only"
		if m.changesOnly {
			description = "show all"
		}
		actions = append(actions, key.NewBinding(key.WithKeys("!"), key.WithHelp("!", description)))
	}
	return renderKeyLine("navigate", navigation) + "\n" + renderKeyLine("actions", actions)
}

func renderKeyLine(label string, bindings []key.Binding) string {
	keyStyle := lipgloss.NewStyle().Bold(true).Foreground(palette.bright)
	descriptionStyle := lipgloss.NewStyle().Foreground(palette.muted)
	parts := make([]string, len(bindings))
	for index, binding := range bindings {
		parts[index] = keyStyle.Render(binding.Help().Key) + " " + descriptionStyle.Render(binding.Help().Desc)
	}
	labelStyle := lipgloss.NewStyle().Foreground(palette.primary).Width(10)
	return "  " + labelStyle.Render(label) + strings.Join(parts, descriptionStyle.Render(" · "))
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
	changed  color.Color
	border   color.Color
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
	changed:  lipgloss.Color("#FBBF24"),
	border:   lipgloss.Color("#4C3A63"),
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
	now     time.Time
}

// projectDelegate renders every list item. compact rows (wide terminals,
// where the inspector shows the full details) keep only the essentials.
type projectDelegate struct {
	compact bool
}

func (projectDelegate) Height() int  { return 2 }
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
	title := highlightMatches(safeTerminalText(item.project.Name), model.MatchesForItem(index), 0, styles.name)
	if model.Width() >= 20 {
		title = iconStyle.Render(iconLabel) + styles.base.Render("  ") + title
	}
	// The star column is always reserved so badges line up.
	if item.project.Favorite {
		title = styles.star.Render("★") + styles.base.Render(" ") + title
	} else {
		title = styles.base.Render("  ") + title
	}
	description := renderSubtitle(item, styles, delegate.compact)
	content := ansi.Truncate(title, lineWidth, "…") + "\n" + ansi.Truncate(description, lineWidth, "…")
	_, _ = fmt.Fprint(writer, rowContainer(selected, width, 2).Render(content))
}

// highlightMatches styles the runes of text that the fuzzy filter matched.
// matches index runes of the item's FilterValue, where text starts at offset.
func highlightMatches(text string, matches []int, offset int, base lipgloss.Style) string {
	length := utf8.RuneCountInString(text)
	var indices []int
	for _, match := range matches {
		if match >= offset && match < offset+length {
			indices = append(indices, match-offset)
		}
	}
	if len(indices) == 0 {
		return base.Render(text)
	}
	return lipgloss.StyleRunes(text, indices, base.Underline(true).Foreground(palette.favorite), base)
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

// renderSubtitle is the metadata line under a project name: technology in its
// color, tooling, Git branch (● when it has uncommitted changes), the folder
// that contains the project and when it was last opened.
func renderSubtitle(item projectItem, styles rowStyles, compact bool) string {
	found := item.project
	parts := []string{styles.base.Foreground(technologyColor(found.Technology)).Render(safeTerminalText(string(found.Technology)))}
	meta := func(value string) {
		parts = append(parts, styles.meta.Render(safeTerminalText(value)))
	}
	frameworks := found.Frameworks
	if compact && len(frameworks) > 1 {
		frameworks = frameworks[:1]
	}
	for _, framework := range frameworks {
		meta(string(framework))
	}
	if !compact {
		for _, manager := range found.PackageManagers {
			meta(string(manager))
		}
		if found.HasDocker && found.Technology != project.TechnologyDocker {
			meta("Docker")
		}
	}
	if found.GitBranch != "" {
		branch := styles.meta.Render(safeTerminalText(found.GitBranch))
		switch {
		case !found.GitStatusKnown:
			branch += styles.meta.Render(" ?")
		case found.GitDirty:
			branch += styles.base.Bold(true).Foreground(palette.changed).Render(" ●")
		}
		parts = append(parts, branch)
	}
	if parent := filepath.Base(filepath.Dir(found.Path)); !compact && parent != "." && parent != "" && parent != string(filepath.Separator) {
		meta(parent)
	}
	if opened := relativeTime(found.LastOpened, item.now); opened != "" {
		meta(opened)
	}
	return strings.Join(parts, styles.meta.Render("  ·  "))
}

// relativeTime formats when a project was last opened, compactly.
func relativeTime(when, now time.Time) string {
	if when.IsZero() || now.IsZero() {
		return ""
	}
	elapsed := now.Sub(when)
	switch {
	case elapsed < time.Minute:
		return "just now"
	case elapsed < time.Hour:
		return fmt.Sprintf("%dm ago", int(elapsed.Minutes()))
	case elapsed < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(elapsed.Hours()))
	case elapsed < 7*24*time.Hour:
		return fmt.Sprintf("%dd ago", int(elapsed.Hours()/24))
	case elapsed < 30*24*time.Hour:
		return fmt.Sprintf("%dw ago", int(elapsed.Hours()/(24*7)))
	case when.Year() == now.Year():
		return when.Format("Jan 2")
	default:
		return when.Format("Jan 2, 2006")
	}
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
	initCommand    tea.Cmd
	// changesOnly limits the project list to projects with uncommitted changes.
	changesOnly   bool
	width, height int
	home          string
	previews      map[string][]previewEntry
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
	if options.Now == nil {
		options.Now = time.Now
	}
	projectList := list.New(nil, projectDelegate{}, 80, 24)
	projectList.Styles.Title = lipgloss.NewStyle()
	projectList.Styles.DefaultFilterCharacterMatch = lipgloss.NewStyle().
		Bold(true).
		Underline(true).
		Foreground(palette.favorite)
	projectList.StatusMessageLifetime = 6 * time.Second
	// The status bar already says "No items"; the inspector explains why.
	projectList.Styles.NoItems = lipgloss.NewStyle().Transform(func(string) string { return "" })
	projectList.SetShowHelp(false) // replaced by footerView
	projectList.SetSize(80, 24-footerHeight)
	// f and d are page keys in the list; they are ForgePath actions here.
	projectList.KeyMap.NextPage.SetKeys("right", "l", "pgdown")

	home, _ := os.UserHomeDir()
	model := Model{
		list:     projectList,
		projects: append([]project.Project(nil), projects...),
		options:  options,
		home:     home,
		previews: map[string][]previewEntry{},
	}
	if len(projects) == 0 && options.StartPath != "" {
		if _, err := model.showDirectory(options.StartPath); err != nil {
			model.showProjects()
		}
	} else {
		model.showProjects()
	}
	if options.Hint != "" {
		model.initCommand = model.newHintMessage(options.Hint)
	}
	return model
}

func (m Model) Init() tea.Cmd {
	return m.initCommand
}

func (m Model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch message := message.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = message.Width, message.Height
		m.resize()
	case editorOpenedMsg:
		if message.request != m.editorRequest {
			return m, nil
		}
		m.editorOpening = false
		command := m.returnFromAction()
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
			if m.mode == dockerScreen {
				return m, nil
			}
			return m, m.showDocker()
		case "o":
			if m.mode == editorScreen {
				return m, nil
			}
			return m, m.showEditors()
		case "f":
			if m.mode != projectScreen {
				return m, nil
			}
			return m, m.toggleFavorite()
		case "!":
			if m.mode != projectScreen {
				return m, nil
			}
			m.changesOnly = !m.changesOnly
			return m, m.showProjects()
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
	footer := m.footerView()
	if m.width > 0 {
		lines := strings.Split(footer, "\n")
		for index, line := range lines {
			lines[index] = ansi.Truncate(line, m.width, "…")
		}
		footer = strings.Join(lines, "\n")
	}
	view := tea.NewView(m.bodyView() + "\n\n" + footer)
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

func (m *Model) newHintMessage(text string) tea.Cmd {
	msg := lipgloss.NewStyle().
		Bold(true).
		Foreground(palette.ink).
		Background(palette.bright).
		Padding(0, 1).
		Render("TIP") + " " +
		lipgloss.NewStyle().Foreground(palette.bright).Render(safeTerminalText(text))
	return m.list.NewStatusMessage(msg)
}

// toggleFavorite flips the highlighted project's favorite flag, keeps the
// list in favorites → recent → name order and moves the cursor with it.
func (m *Model) toggleFavorite() tea.Cmd {
	item, ok := m.list.SelectedItem().(projectItem)
	if !ok {
		return nil
	}
	if m.options.ToggleFavorite == nil {
		return m.newErrorMessage(fmt.Errorf("favorites are unavailable"))
	}
	favorite, err := m.options.ToggleFavorite(item.project)
	if err != nil {
		return m.newErrorMessage(fmt.Errorf("could not update favorites: %w", err))
	}
	for index := range m.projects {
		if samePath(m.projects[index].Path, item.project.Path) {
			m.projects[index].Favorite = favorite
		}
	}
	sortProjects(m.projects)
	command := m.list.SetItems(m.projectItems())
	if !m.list.IsFiltered() {
		for index, candidate := range m.list.Items() {
			if found, ok := candidate.(projectItem); ok && samePath(found.project.Path, item.project.Path) {
				m.list.Select(index)
				break
			}
		}
	}
	text := item.project.Name + " removed from favorites"
	if favorite {
		text = "★ " + item.project.Name + " added to favorites"
	}
	return tea.Batch(command, m.newSuccessMessage(text))
}

// sortProjects uses the same order as state.Decorate: favorites, then the most
// recently opened, then by name.
func sortProjects(projects []project.Project) {
	sort.SliceStable(projects, func(i, j int) bool {
		if projects[i].Favorite != projects[j].Favorite {
			return projects[i].Favorite
		}
		if !projects[i].LastOpened.Equal(projects[j].LastOpened) {
			return projects[i].LastOpened.After(projects[j].LastOpened)
		}
		return projects[i].Name < projects[j].Name
	})
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
