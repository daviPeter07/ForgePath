//go:build windows

package action

import (
	"os/exec"
	"strconv"
	"syscall"

	"golang.org/x/sys/windows"
)

func prepareCommand(command *exec.Cmd, _ bool) {
	command.SysProcAttr = &syscall.SysProcAttr{CreationFlags: windows.CREATE_NEW_PROCESS_GROUP}
}

func terminateProcessTree(command *exec.Cmd, force bool) error {
	if command.Process == nil {
		return nil
	}
	if !force {
		return windows.GenerateConsoleCtrlEvent(windows.CTRL_BREAK_EVENT, uint32(command.Process.Pid))
	}
	arguments := []string{"/T", "/PID", strconv.Itoa(command.Process.Pid)}
	arguments = append([]string{"/F"}, arguments...)
	killer := exec.Command("taskkill.exe", arguments...)
	if err := killer.Run(); err != nil {
		return command.Process.Kill()
	}
	return nil
}
