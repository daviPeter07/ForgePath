package tui

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/daviPeter07/forgepath/internal/project"
)

// RunPlan is the command the r key starts for a project.
type RunPlan struct {
	Command []string
	// Configured commands come from config.json and start immediately;
	// detected ones are suggestions the user confirms with a second r.
	Configured bool
}

// RunFunc runs a command in a folder with the given standard streams.
type RunFunc func(ctx context.Context, path string, command []string, stdin io.Reader, stdout, stderr io.Writer) error

type runFinishedMsg struct {
	command string
	err     error
}

type runPlanEntry struct {
	plan RunPlan
	ok   bool
}

// planFor returns the run plan of a project, cached because the inspector
// asks for it on every render.
func (m Model) planFor(found project.Project) (RunPlan, bool) {
	if m.options.PlanRun == nil || m.options.RunCommand == nil {
		return RunPlan{}, false
	}
	if entry, cached := m.runPlans[found.Path]; cached {
		return entry.plan, entry.ok
	}
	plan, ok := m.options.PlanRun(found)
	ok = ok && len(plan.Command) > 0
	m.runPlans[found.Path] = runPlanEntry{plan: plan, ok: ok}
	return plan, ok
}

// runSelected starts the highlighted project's command, handing the terminal
// to it until it exits. Detected commands need a second r to confirm.
func (m *Model) runSelected() tea.Cmd {
	item, ok := m.list.SelectedItem().(projectItem)
	if !ok || m.options.RunCommand == nil || m.options.PlanRun == nil {
		return nil
	}
	plan, ok := m.planFor(item.project)
	if !ok {
		return m.newErrorMessage(fmt.Errorf("nothing to run for %s; set projects.%q.command in config.json", item.project.Name, item.project.Name))
	}
	text := strings.Join(plan.Command, " ")
	key := item.project.Path + "\x00" + text
	if !plan.Configured && m.pendingRun != key {
		m.pendingRun = key
		return m.newHintMessage("Run " + text + "? Press r again")
	}
	m.pendingRun = ""
	process := &processCommand{
		ctx:     m.options.Context,
		path:    item.project.Path,
		home:    m.home,
		command: append([]string(nil), plan.Command...),
		run:     m.options.RunCommand,
	}
	return m.exec(process, func(err error) tea.Msg { return runFinishedMsg{command: text, err: err} })
}

func (m *Model) runFinished(message runFinishedMsg) tea.Cmd {
	switch {
	case message.err == nil:
		return m.newSuccessMessage(message.command + " finished")
	case interrupted(message.err):
		return m.newSuccessMessage(message.command + " stopped")
	default:
		return m.newErrorMessage(fmt.Errorf("%s failed: %w", message.command, message.err))
	}
}

// interrupted reports whether a command ended because the user pressed
// Ctrl+C: exit code 130 by convention, or killed by the signal itself.
func interrupted(err error) bool {
	if errors.Is(err, context.Canceled) {
		return true
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		code := exitErr.ExitCode()
		return code == 130 || code == -1
	}
	return false
}

// processCommand adapts a RunFunc to tea.ExecCommand, announcing what runs
// before the program takes over the terminal.
type processCommand struct {
	ctx     context.Context
	path    string
	home    string
	command []string
	run     RunFunc
	stdin   io.Reader
	stdout  io.Writer
	stderr  io.Writer
}

func (p *processCommand) SetStdin(reader io.Reader)  { p.stdin = reader }
func (p *processCommand) SetStdout(writer io.Writer) { p.stdout = writer }
func (p *processCommand) SetStderr(writer io.Writer) { p.stderr = writer }

func (p *processCommand) Run() error {
	_, _ = fmt.Fprintf(p.stdout, "\n▶ %s  %s\n  Ctrl+C stops it and returns to ForgePath\n\n",
		safeTerminalText(strings.Join(p.command, " ")), safeTerminalText(shortenHome(p.path, p.home)))
	return p.run(p.ctx, p.path, p.command, p.stdin, p.stdout, p.stderr)
}
