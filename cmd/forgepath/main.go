package main

import (
	"fmt"
	"io"
	"os"

	"github.com/daviPeter07/forgepath/internal/cli"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	command := cli.NewRootCommand(stdout, stderr)
	command.SetArgs(args)
	if err := command.Execute(); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}
