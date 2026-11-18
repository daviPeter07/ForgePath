package main

import (
	"context"
	"os"
	"syscall"
	"testing"
	"time"
)

func TestInterruptIsLeftToAnInteractiveChild(t *testing.T) {
	signals := make(chan os.Signal, 1)
	childRunning := true
	ctx, stop := watchSignals(context.Background(), signals, func() bool { return childRunning })
	defer stop()

	signals <- os.Interrupt
	select {
	case <-ctx.Done():
		t.Fatal("Ctrl+C canceled ForgePath while a child owned the terminal")
	case <-time.After(50 * time.Millisecond):
	}

	childRunning = false
	signals <- os.Interrupt
	select {
	case <-ctx.Done():
	case <-time.After(time.Second):
		t.Fatal("Ctrl+C did not cancel ForgePath without a child")
	}
}

func TestTerminateAlwaysCancels(t *testing.T) {
	signals := make(chan os.Signal, 1)
	ctx, stop := watchSignals(context.Background(), signals, func() bool { return true })
	defer stop()
	signals <- syscall.SIGTERM
	select {
	case <-ctx.Done():
	case <-time.After(time.Second):
		t.Fatal("SIGTERM did not cancel ForgePath")
	}
}
