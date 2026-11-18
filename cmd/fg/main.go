package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/daviPeter07/forgepath/internal/action"
	"github.com/daviPeter07/forgepath/internal/cli"
)

func main() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	ctx, stop := watchSignals(context.Background(), signals, action.InteractiveChildRunning)
	defer stop()
	code := run(ctx, os.Args[1:], os.Stdout, os.Stderr)
	signal.Stop(signals)
	stop()
	os.Exit(code)
}

// watchSignals returns a context canceled by SIGTERM, and by Ctrl+C unless an
// interactive child process (a dev server started from the selector or fg
// run) owns the terminal: that child receives Ctrl+C itself, and ForgePath
// keeps running so it can return to the selector.
func watchSignals(parent context.Context, signals <-chan os.Signal, childOwnsInterrupt func() bool) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(parent)
	go func() {
		for {
			select {
			case received := <-signals:
				if received == os.Interrupt && childOwnsInterrupt() {
					continue
				}
				cancel()
				return
			case <-ctx.Done():
				return
			}
		}
	}()
	return ctx, cancel
}

func run(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	command := cli.NewRootCommand(stdout, stderr)
	command.SetArgs(args)
	if err := command.ExecuteContext(ctx); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}
