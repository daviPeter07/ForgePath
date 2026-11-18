package tui

import (
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/daviPeter07/forgepath/internal/icon"
	"github.com/daviPeter07/forgepath/internal/ide"
	"github.com/daviPeter07/forgepath/internal/project"
)

// The inspector is the right-hand panel of wide terminals: it describes the
// highlighted item and is the only place logos are drawn, at a size where
// they are legible.
const (
	splitMinWidth = 100
	inspectorGap  = 2
	// inspectorTop aligns the panel with the list's status line.
	inspectorTop = 2
	maxLogoSize  = 20
	minLogoSize  = 8
	logoGap      = 3
)

func (m Model) splitLayout() bool {
	return m.width >= splitMinWidth
}

func listWidthFor(total int) int {
	return min(max(total*42/100, 38), 60)
}

// resize fits the list into the window, leaving room for the footer and, on
// wide terminals, the inspector.
func (m *Model) resize() {
	width := m.width
	if m.splitLayout() {
		width = listWidthFor(m.width)
	}
	m.list.SetSize(max(width, 1), max(m.height-footerHeight, 1))
	m.list.SetDelegate(projectDelegate{compact: m.splitLayout()})
}

func (m Model) bodyView() string {
	body := m.list.View()
	if !m.splitLayout() {
		return body
	}
	// An empty list renders narrower than its width; pad it so the panel
	// never jumps sideways.
	body = lipgloss.NewStyle().Width(m.list.Width()).Render(body)
	width := m.width - m.list.Width() - inspectorGap
	height := m.list.Height() - inspectorTop
	if width < 20 || height < 6 {
		return body
	}
	panel := strings.Repeat("\n", inspectorTop) + m.inspectorView(width, height)
	return lipgloss.JoinHorizontal(lipgloss.Top, body, strings.Repeat(" ", inspectorGap), panel)
}

func (m Model) inspectorView(width, height int) string {
	innerWidth, innerHeight := width-6, height-4
	var content string
	switch item := m.list.SelectedItem().(type) {
	case projectItem:
		content = m.projectInspector(item.project, innerWidth, innerHeight)
	case directoryItem:
		content = m.folderInspector(item.directory, innerHeight)
	case editorItem:
		content = m.editorInspector(item)
	case dockerItem:
		content = m.dockerInspector(item, innerHeight)
	default:
		content = m.emptyInspector(innerHeight)
	}
	lines := strings.Split(content, "\n")
	if len(lines) > innerHeight {
		lines = lines[:innerHeight]
	}
	for index, line := range lines {
		lines[index] = ansi.Truncate(line, innerWidth, "…")
	}
	// The panel is as tall as its content: a short card reads better than a
	// tall, mostly empty frame.
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(palette.border).
		Padding(1, 2).
		Width(width).
		Render(strings.Join(lines, "\n"))
}

// emptyInspector explains an empty list: the current folder's files when
// browsing, otherwise how to add projects.
func (m Model) emptyInspector(height int) string {
	if m.mode == directoryScreen && m.currentPath != "" {
		return m.folderInspector(Directory{Name: filepath.Base(m.currentPath), Path: m.currentPath}, height)
	}
	return heading("No projects yet", mutedStyle().Render("ForgePath lists the project folders inside your workspaces.")) +
		"\n\n" + renderFacts([]fact{{"add one", "fg workspace add <folder>"}, {"or browse", "run fg inside a folder"}})
}

func mutedStyle() lipgloss.Style { return lipgloss.NewStyle().Foreground(palette.muted) }

// heading is the name line and the description line next to a logo.
func heading(name, description string) string {
	return lipgloss.NewStyle().Bold(true).Foreground(palette.text).Render(safeTerminalText(name)) + "\n" + description
}

func (m Model) projectInspector(found project.Project, width, height int) string {
	textWidth := width
	logo := ""
	if m.options.Icons == icon.ModeGraphics {
		size := min(maxLogoSize, width/3, height*2)
		size -= size % 2
		if size >= minLogoSize {
			if graphic, err := icon.GraphicSized(found.Technology, size); err == nil {
				logo = graphic
				textWidth = width - size - logoGap
			}
		}
	}

	name := lipgloss.NewStyle().Bold(true).Foreground(palette.text).Render(safeTerminalText(found.Name))
	if logo == "" {
		label := icon.Label(found.Technology, m.options.Icons)
		if label != "" {
			name = badge(label, technologyColor(found.Technology)) + "  " + name
		}
	}
	if found.Favorite {
		name = lipgloss.NewStyle().Foreground(palette.favorite).Render("★") + " " + name
	}
	techLine := []string{lipgloss.NewStyle().Bold(true).Foreground(technologyColor(found.Technology)).Render(string(found.Technology))}
	for _, framework := range found.Frameworks {
		techLine = append(techLine, mutedStyle().Render(safeTerminalText(string(framework))))
	}
	block := name + "\n" + strings.Join(techLine, mutedStyle().Render("  ·  ")) +
		"\n\n" + renderFacts(m.projectFacts(found, textWidth-factKeyWidth))
	if logo == "" {
		return block
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, logo, strings.Repeat(" ", logoGap), block)
}

type fact struct{ key, value string }

const factKeyWidth = 10

func (m Model) projectFacts(found project.Project, valueWidth int) []fact {
	var facts []fact
	facts = append(facts, fact{"branch", gitSummary(found)})
	var tooling []string
	for _, manager := range found.PackageManagers {
		tooling = append(tooling, string(manager))
	}
	if found.HasDocker && found.Technology != project.TechnologyDocker {
		tooling = append(tooling, "Docker")
	}
	if len(tooling) > 0 {
		facts = append(facts, fact{"tooling", strings.Join(tooling, " · ")})
	}
	facts = append(facts, fact{"folder", safeTerminalText(truncateStart(shortenHome(filepath.Dir(found.Path), m.home), max(valueWidth, 8)))})
	opened := relativeTime(found.LastOpened, m.options.Now())
	if opened == "" {
		opened = mutedStyle().Render("not opened yet")
	}
	facts = append(facts, fact{"opened", opened})
	if ranked := ide.Rank(m.options.IDEs, found.Technology); len(ranked) > 0 && ranked[0].Supports(found.Technology) {
		facts = append(facts, fact{"opens in", safeTerminalText(ranked[0].Name) + mutedStyle().Render("  o to choose")})
	}
	if len(found.Markers) > 0 {
		facts = append(facts, fact{"found by", safeTerminalText(strings.Join(found.Markers, ", "))})
	}
	return facts
}

func gitSummary(found project.Project) string {
	branch := safeTerminalText(found.GitBranch)
	switch {
	case found.GitBranch == "":
		return mutedStyle().Render("not a Git repository")
	case !found.GitStatusKnown:
		return branch + mutedStyle().Render(" · status unavailable")
	case found.GitDirty:
		return branch + lipgloss.NewStyle().Foreground(palette.changed).Render(" ● uncommitted changes")
	default:
		return branch + lipgloss.NewStyle().Foreground(palette.success).Render(" ✓ clean")
	}
}

// renderFacts aligns key/value pairs. Values may already be styled, so
// callers sanitize raw text (paths, names) before building them.
func renderFacts(facts []fact) string {
	keyStyle := lipgloss.NewStyle().Foreground(palette.muted).Width(factKeyWidth)
	valueStyle := lipgloss.NewStyle().Foreground(palette.text)
	lines := make([]string, len(facts))
	for index, fact := range facts {
		lines[index] = keyStyle.Render(fact.key) + valueStyle.Render(fact.value)
	}
	return strings.Join(lines, "\n")
}

// shortenHome replaces the home directory prefix with ~.
func shortenHome(path, home string) string {
	if home == "" {
		return path
	}
	if path == home {
		return "~"
	}
	if strings.HasPrefix(path, home+string(filepath.Separator)) {
		return "~" + path[len(home):]
	}
	return path
}

type previewEntry struct {
	name string
	dir  bool
}

func (m Model) folderInspector(directory Directory, height int) string {
	entries, ok := m.previews[directory.Path]
	if !ok {
		entries = readPreview(directory.Path)
		m.previews[directory.Path] = entries
	}
	summary := "Empty folder"
	switch len(entries) {
	case 0:
	case 1:
		summary = "1 item"
	default:
		summary = strconv.Itoa(len(entries)) + " items"
	}
	lines := []string{heading(directory.Name, mutedStyle().Render(summary)), ""}
	room := height - len(lines) - 1
	for index, entry := range entries {
		if index == room-1 && len(entries) > room {
			lines = append(lines, mutedStyle().Render("+"+strconv.Itoa(len(entries)-index)+" more"))
			break
		}
		if entry.dir {
			lines = append(lines, lipgloss.NewStyle().Foreground(palette.bright).Render(safeTerminalText(entry.name)+"/"))
		} else {
			lines = append(lines, mutedStyle().Render(safeTerminalText(entry.name)))
		}
	}
	return strings.Join(lines, "\n")
}

// readPreview lists a folder for the inspector: folders first, generated
// folders skipped.
func readPreview(path string) []previewEntry {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil
	}
	preview := make([]previewEntry, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() && ignoredBrowserDirectory(entry.Name()) {
			continue
		}
		preview = append(preview, previewEntry{name: entry.Name(), dir: entry.IsDir()})
	}
	sort.SliceStable(preview, func(i, j int) bool {
		if preview[i].dir != preview[j].dir {
			return preview[i].dir
		}
		return strings.ToLower(preview[i].name) < strings.ToLower(preview[j].name)
	})
	return preview
}

func (m Model) editorInspector(item editorItem) string {
	label := "Installed"
	if item.recommended {
		label = "Suggested for " + string(item.technology)
	} else if item.editor.Supports(item.technology) {
		label = "Compatible with " + string(item.technology)
	}
	command := append([]string{item.editor.Executable}, item.editor.Arguments...)
	supports := make([]string, 0, len(item.editor.Technologies))
	for _, technology := range item.editor.Technologies {
		supports = append(supports, string(technology))
	}
	if len(supports) > 6 {
		supports = append(supports[:6], "…")
	}
	facts := []fact{
		{"command", safeTerminalText(strings.Join(command, " "))},
		{"opens", safeTerminalText(shortenHome(m.editorPath, m.home))},
		{"supports", strings.Join(supports, ", ")},
	}
	return heading(item.editor.Name, mutedStyle().Render(label)) + "\n\n" + renderFacts(facts)
}

func (m Model) dockerInspector(item dockerItem, height int) string {
	target := mutedStyle().Render("Creates docker-compose.yml in " + safeTerminalText(filepath.Base(item.projectPath)))
	lines := []string{heading(item.label, target), ""}
	keyStyle := lipgloss.NewStyle().Foreground(palette.bright)
	valueStyle := lipgloss.NewStyle().Foreground(palette.text)
	commentStyle := lipgloss.NewStyle().Foreground(palette.muted).Italic(true)
	for _, line := range strings.Split(strings.TrimRight(item.compose, "\n"), "\n") {
		if len(lines) >= height {
			break
		}
		trimmed := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(trimmed, "#"):
			lines = append(lines, commentStyle.Render(line))
		case strings.Contains(line, ":") && !strings.HasPrefix(trimmed, "- "):
			key, value, _ := strings.Cut(line, ":")
			lines = append(lines, keyStyle.Render(key+":")+valueStyle.Render(value))
		default:
			lines = append(lines, valueStyle.Render(line))
		}
	}
	return strings.Join(lines, "\n")
}
