package git

import (
	"context"
	"os/exec"
	"strings"
	"time"
)

type Info struct {
	Branch      string
	Dirty       bool
	StatusKnown bool
}

func Inspect(path string) (Info, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return InspectContext(ctx, path)
}

func InspectContext(ctx context.Context, path string) (Info, bool) {
	inside, err := run(ctx, path, "rev-parse", "--is-inside-work-tree")
	if err != nil || inside != "true" {
		return Info{}, false
	}

	branch, err := run(ctx, path, "symbolic-ref", "--quiet", "--short", "HEAD")
	if err != nil {
		commit, commitErr := run(ctx, path, "rev-parse", "--short", "HEAD")
		if commitErr != nil {
			return Info{}, false
		}
		branch = "detached@" + commit
	}

	status, err := run(ctx, path, "status", "--porcelain", "--untracked-files=normal")
	if err != nil {
		return Info{Branch: branch}, true
	}

	return Info{Branch: branch, Dirty: status != "", StatusKnown: true}, true
}

func run(ctx context.Context, path string, arguments ...string) (string, error) {
	args := append([]string{"--no-optional-locks", "-c", "core.fsmonitor=false", "-C", path}, arguments...)
	command := exec.CommandContext(ctx, "git", args...)
	command.WaitDelay = 100 * time.Millisecond
	output, err := command.Output()
	return strings.TrimSpace(string(output)), err
}
