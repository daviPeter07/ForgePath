package tui

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os/exec"
	"runtime"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/daviPeter07/forgepath/internal/icon"
	"github.com/daviPeter07/forgepath/internal/project"
)

type capturedExec struct {
	command  tea.ExecCommand
	callback tea.ExecCallback
}

func runModel(t *testing.T, plan RunPlan, planned bool, width int) (Model, *[]capturedExec) {
	t.Helper()
	model := sizedModel(t, []project.Project{
		{Name: "api", Path: "/work/api", Technology: project.TechnologyGo},
		{Name: "web", Path: "/work/web", Technology: project.TechnologyTypeScript},
	}, Options{
		Icons:      icon.ModeASCII,
		PlanRun:    func(project.Project) (RunPlan, bool) { return plan, planned },
		RunCommand: func(context.Context, string, []string, io.Reader, io.Writer, io.Writer) error { return nil },
	}, width, 30)
	var captured []capturedExec
	model.exec = func(command tea.ExecCommand, callback tea.ExecCallback) tea.Cmd {
		captured = append(captured, capturedExec{command, callback})
		return func() tea.Msg { return nil }
	}
	return model, &captured
}

func TestRunStartsConfiguredCommandImmediately(t *testing.T) {
	model, captured := runModel(t, RunPlan{Command: []string{"make", "dev"}, Configured: true}, true, 96)
	model = press(t, model, letterKey('r'))
	if len(*captured) != 1 {
		t.Fatalf("exec calls = %d, want 1", len(*captured))
	}
	process := (*captured)[0].command.(*processCommand)
	if process.path != "/work/api" || strings.Join(process.command, " ") != "make dev" {
		t.Fatalf("process = %q in %q", process.command, process.path)
	}
}

func TestRunAsksBeforeStartingDetectedCommand(t *testing.T) {
	model, captured := runModel(t, RunPlan{Command: []string{"pnpm", "dev"}}, true, 96)

	model = press(t, model, letterKey('r'))
	if len(*captured) != 0 {
		t.Fatal("a detected command started without confirmation")
	}
	if plain := ansi.Strip(model.View().Content); !strings.Contains(plain, "Run pnpm dev? Press r again") {
		t.Fatalf("view does not ask for confirmation:\n%s", plain)
	}

	model = press(t, model, downKey, letterKey('r'))
	if len(*captured) != 0 {
		t.Fatal("moving to another project must reset the confirmation")
	}
	model = press(t, model, letterKey('r'))
	if len(*captured) != 1 {
		t.Fatalf("exec calls after confirming = %d, want 1", len(*captured))
	}
	if path := (*captured)[0].command.(*processCommand).path; path != "/work/web" {
		t.Fatalf("ran in %q, want the confirmed project", path)
	}
}

func TestRunExplainsProjectsWithoutCommand(t *testing.T) {
	model, captured := runModel(t, RunPlan{}, false, 96)
	model = press(t, model, letterKey('r'))
	if len(*captured) != 0 {
		t.Fatal("ran without a command")
	}
	if plain := ansi.Strip(model.View().Content); !strings.Contains(plain, "nothing to run for api") {
		t.Fatalf("view does not explain the missing command:\n%s", plain)
	}
}

func TestRunFinishedReportsTheOutcome(t *testing.T) {
	tests := []struct {
		err  error
		want string
	}{
		{nil, "pnpm dev finished"},
		{context.Canceled, "pnpm dev stopped"},
		{errors.New("exit status 1"), "pnpm dev failed: exit status 1"},
	}
	for _, tt := range tests {
		model, _ := runModel(t, RunPlan{}, false, 120)
		updated, _ := model.Update(runFinishedMsg{command: "pnpm dev", err: tt.err})
		if plain := ansi.Strip(updated.(Model).View().Content); !strings.Contains(plain, tt.want) {
			t.Fatalf("err %v: view is missing %q:\n%s", tt.err, tt.want, plain)
		}
	}
}

func TestInterruptedRecognizesCtrlC(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("exit code 130 is a Unix convention")
	}
	err := exec.Command("sh", "-c", "exit 130").Run()
	if !interrupted(err) {
		t.Fatalf("interrupted(%v) = false, want true", err)
	}
	if interrupted(exec.Command("sh", "-c", "exit 1").Run()) {
		t.Fatal("a plain failure is not an interruption")
	}
}

func TestProcessCommandAnnouncesAndRuns(t *testing.T) {
	var gotPath string
	var gotCommand []string
	process := &processCommand{
		ctx:     context.Background(),
		path:    "/home/dev/work/api",
		home:    "/home/dev",
		command: []string{"go", "run", "."},
		run: func(_ context.Context, path string, command []string, _ io.Reader, stdout, _ io.Writer) error {
			gotPath, gotCommand = path, command
			_, err := io.WriteString(stdout, "serving\n")
			return err
		},
	}
	var output bytes.Buffer
	process.SetStdin(strings.NewReader(""))
	process.SetStdout(&output)
	process.SetStderr(&output)
	if err := process.Run(); err != nil {
		t.Fatal(err)
	}
	if gotPath != "/home/dev/work/api" || strings.Join(gotCommand, " ") != "go run ." {
		t.Fatalf("ran %q in %q", gotCommand, gotPath)
	}
	for _, want := range []string{"go run .", "~/work/api", "Ctrl+C", "serving"} {
		if !strings.Contains(output.String(), want) {
			t.Fatalf("output %q is missing %q", output.String(), want)
		}
	}
}

func TestInspectorAndFooterAdvertiseRun(t *testing.T) {
	model, _ := runModel(t, RunPlan{Command: []string{"pnpm", "dev"}}, true, 120)
	plain := ansi.Strip(model.View().Content)
	if !strings.Contains(plain, "pnpm dev") || !strings.Contains(plain, "detected") {
		t.Fatalf("inspector does not show the run command:\n%s", plain)
	}
	if footer := ansi.Strip(model.footerView()); !strings.Contains(footer, "r run") {
		t.Fatalf("footer = %q, want r run", footer)
	}
	if footer := ansi.Strip(testModel().footerView()); strings.Contains(footer, "r run") {
		t.Fatalf("footer = %q, r needs a runner", footer)
	}
}
