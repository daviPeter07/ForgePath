package git

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// RecentCommits is how many commits Details lists.
const RecentCommits = 3

// Details is the repository information shown for one project at a time,
// loaded on demand because it costs several git calls.
type Details struct {
	// Upstream is the tracking branch ("origin/main"), empty when there is none.
	Upstream string
	Ahead    int
	Behind   int
	Commits  []Commit
}

type Commit struct {
	Hash    string
	Subject string
	When    time.Time
}

// LoadDetails reads the recent commits of path and, when the branch tracks a
// remote, how many commits it is ahead of and behind it. It uses the last
// fetched state of the remote and never contacts the network.
func LoadDetails(ctx context.Context, path string) (Details, error) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	if inside, err := run(ctx, path, "rev-parse", "--is-inside-work-tree"); err != nil || inside != "true" {
		return Details{}, errors.New("not a Git repository")
	}

	var details Details
	// An empty repository has no HEAD yet: no commits and no upstream.
	if _, err := run(ctx, path, "rev-parse", "--verify", "--quiet", "HEAD"); err != nil {
		return details, nil
	}
	log, err := run(ctx, path, "log", "-"+strconv.Itoa(RecentCommits), "--format=%h%x1f%s%x1f%ct")
	if err != nil {
		return Details{}, fmt.Errorf("read commits: %w", err)
	}
	for _, line := range strings.Split(log, "\n") {
		fields := strings.Split(line, "\x1f")
		if len(fields) != 3 {
			continue
		}
		seconds, _ := strconv.ParseInt(fields[2], 10, 64)
		details.Commits = append(details.Commits, Commit{Hash: fields[0], Subject: fields[1], When: time.Unix(seconds, 0)})
	}

	upstream, err := run(ctx, path, "rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{upstream}")
	if err != nil || upstream == "" {
		return details, nil
	}
	counts, err := run(ctx, path, "rev-list", "--left-right", "--count", "@{upstream}...HEAD")
	if err != nil {
		return details, nil
	}
	behind, ahead, found := strings.Cut(counts, "\t")
	if !found {
		return details, nil
	}
	details.Upstream = upstream
	details.Behind, _ = strconv.Atoi(strings.TrimSpace(behind))
	details.Ahead, _ = strconv.Atoi(strings.TrimSpace(ahead))
	return details, nil
}
