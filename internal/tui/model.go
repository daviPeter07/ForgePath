package tui

import (
	"context"
	"fmt"
	"io"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/daviPeter07/forgepath/internal/project"
)

var selectKey = key.NewBinding(
	key.WithKeys("enter"),
	key.WithHelp("enter", "select"),
)

type projectItem struct {
	project project.Project
}

func (item projectItem) Title() string {
	return item.project.Name
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
	return item.project.Name + " " + string(item.project.Technology)
}

type Model struct {
	list         list.Model
	selected     project.Project
	hasSelection bool
	cancelled    bool
}

func NewModel(projects []project.Project) Model {
	items := make([]list.Item, len(projects))
	for index, found := range projects {
		items[index] = projectItem{project: found}
	}

	accent := lipgloss.Color("#F59E0B")
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Bold(true).
		Foreground(accent).
		BorderLeftForeground(accent)
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(lipgloss.Color("#A8A29E")).
		BorderLeftForeground(accent)

	projectList := list.New(items, delegate, 80, 24)
	projectList.Title = "ForgePath"
	projectList.SetStatusBarItemName("project", "projects")
	projectList.Styles.Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#1C1917")).
		Background(accent).
		Padding(0, 1)
	projectList.Styles.DefaultFilterCharacterMatch = lipgloss.NewStyle().
		Bold(true).
		Foreground(accent)
	projectList.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{selectKey}
	}
	projectList.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{selectKey}
	}

	return Model{list: projectList}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch message := message.(type) {
	case tea.WindowSizeMsg:
		m.list.SetSize(message.Width, message.Height)
	case tea.KeyPressMsg:
		switch message.String() {
		case "ctrl+c":
			m.cancelled = true
			return m, tea.Quit
		case "enter":
			if !m.list.SettingFilter() {
				if item, ok := m.list.SelectedItem().(projectItem); ok {
					m.selected = item.project
					m.hasSelection = true
					return m, tea.Quit
				}
			}
		case "q":
			if !m.list.SettingFilter() {
				m.cancelled = true
				return m, tea.Quit
			}
		case "esc":
			if !m.list.SettingFilter() && !m.list.IsFiltered() {
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
	view.WindowTitle = "ForgePath project selector"
	return view
}

func (m Model) Selection() (project.Project, bool) {
	return m.selected, m.hasSelection
}

func (m Model) Cancelled() bool {
	return m.cancelled
}

func Select(ctx context.Context, projects []project.Project, input io.Reader, output io.Writer) (project.Project, bool, error) {
	program := tea.NewProgram(
		NewModel(projects),
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
