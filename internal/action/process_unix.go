//go:build !windows

package action

import (
	"errors"
	"os/exec"
	"syscall"
)

func prepareCommand(command *exec.Cmd, interactive bool) {
	if !interactive {
		command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	}
}

func terminateProcessTree(command *exec.Cmd, force bool) error {
	if command.Process == nil {
		return nil
	}
	pid := command.Process.Pid
	if command.SysProcAttr != nil && command.SysProcAttr.Setpgid {
		pid = -pid
	}
	signal := syscall.SIGINT
	if force {
		signal = syscall.SIGKILL
	}
	err := syscall.Kill(pid, signal)
	if errors.Is(err, syscall.ESRCH) {
		return nil
	}
	return err
}
