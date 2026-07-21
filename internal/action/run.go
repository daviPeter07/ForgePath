package action

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

func RunCommand(ctx context.Context, path string, arguments []string, stdin io.Reader, stdout, stderr io.Writer) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if len(arguments) == 0 || arguments[0] == "" {
		return fmt.Errorf("development command cannot be empty")
	}

	executable, err := safeExecutable(runtime.GOOS, arguments[0])
	if err != nil {
		return err
	}
	command := exec.Command(executable, arguments[1:]...)
	command.Dir = path
	command.Stdin = stdin
	command.Stdout = stdout
	command.Stderr = stderr
	prepareCommand(command, isInteractive(stdin))
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := command.Start(); err != nil {
		return err
	}

	finished := make(chan error, 1)
	go func() {
		finished <- command.Wait()
	}()

	select {
	case err := <-finished:
		return err
	case <-ctx.Done():
		_ = terminateProcessTree(command, false)
	}

	timer := time.NewTimer(3 * time.Second)
	defer timer.Stop()
	select {
	case <-finished:
		return ctx.Err()
	case <-timer.C:
		_ = terminateProcessTree(command, true)
		<-finished
		return ctx.Err()
	}
}

func isInteractive(input io.Reader) bool {
	file, ok := input.(*os.File)
	if !ok {
		return false
	}
	if filepath.Clean(file.Name()) == filepath.Clean(os.DevNull) {
		return false
	}
	info, err := file.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}
