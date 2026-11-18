package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/daviPeter07/forgepath/internal/detector"
	"github.com/spf13/cobra"
)

type cloneFunc func(ctx context.Context, url, destination string, stdout, stderr io.Writer) error

var (
	repositoryPart = regexp.MustCompile(`^[A-Za-z0-9_.-]+$`)
	scpRepository  = regexp.MustCompile(`^[A-Za-z0-9_.-]+@[A-Za-z0-9_.-]+:[^/-].*$`)
	cloneSchemes   = []string{"https://", "http://", "ssh://", "git://", "file://"}
)

func newCloneCommand(configPath configPathFunc, clone cloneFunc) *cobra.Command {
	var workspace, name string
	command := &cobra.Command{
		Use:   "clone <owner/repo | url>",
		Short: "Clone a repository into a workspace",
		Long: "Clone a repository into the first configured workspace (or --workspace).\n" +
			"owner/repo is cloned from GitHub; any https, ssh or git@host: URL also works.",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			url, repositoryName, err := parseRepository(args[0])
			if err != nil {
				return err
			}
			if name != "" {
				if !validFolderName(name) {
					return fmt.Errorf("invalid folder name %q", name)
				}
				repositoryName = name
			}
			target, err := cloneWorkspace(workspace, configPath)
			if err != nil {
				return err
			}
			destination := filepath.Join(target, repositoryName)
			if _, err := os.Lstat(destination); err == nil {
				return fmt.Errorf("%q already exists; pass --name to clone under another folder name", destination)
			}
			if err := clone(cmd.Context(), url, destination, cmd.ErrOrStderr(), cmd.ErrOrStderr()); err != nil {
				return fmt.Errorf("clone %s: %w", url, err)
			}
			if _, err := fmt.Fprintln(cmd.OutOrStdout(), destination); err != nil {
				return err
			}
			if result, found, err := detector.Detect(destination); err == nil && found {
				_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "%s project ready; open it with fg\n", result.Technology)
			}
			return nil
		},
	}
	command.Flags().StringVar(&workspace, "workspace", "", "folder to clone into (default: first configured workspace)")
	command.Flags().StringVar(&name, "name", "", "folder name for the clone (default: the repository name)")
	return command
}

// parseRepository turns owner/repo (GitHub) or a clone URL into the URL to
// clone and the folder name. Values starting with "-" are rejected so they
// can never be read as git options.
func parseRepository(spec string) (string, string, error) {
	spec = strings.TrimSpace(spec)
	invalid := fmt.Errorf("invalid repository %q: use owner/repo or a clone URL", spec)
	if spec == "" || strings.HasPrefix(spec, "-") {
		return "", "", invalid
	}
	url := ""
	for _, scheme := range cloneSchemes {
		if strings.HasPrefix(spec, scheme) {
			url = spec
		}
	}
	switch {
	case url != "":
	case strings.Contains(spec, "://"):
		return "", "", invalid
	case scpRepository.MatchString(spec):
		url = spec
	default:
		parts := strings.Split(spec, "/")
		if len(parts) != 2 || !repositoryPart.MatchString(parts[0]) || !repositoryPart.MatchString(parts[1]) {
			return "", "", invalid
		}
		url = "https://github.com/" + parts[0] + "/" + strings.TrimSuffix(parts[1], ".git") + ".git"
	}
	trimmed := strings.TrimRight(url, "/")
	name := trimmed[strings.LastIndexAny(trimmed, "/:")+1:]
	name = strings.TrimSuffix(name, ".git")
	if !validFolderName(name) {
		return "", "", invalid
	}
	return url, name, nil
}

func validFolderName(name string) bool {
	return name != "" && name != "." && name != ".." && repositoryPart.MatchString(name)
}

// cloneWorkspace is --workspace, else the first configured workspace, else
// the current directory.
func cloneWorkspace(flag string, configPath configPathFunc) (string, error) {
	workspace := flag
	if workspace == "" {
		workspaces, err := configuredWorkspaces(nil, configPath)
		if err != nil {
			return "", err
		}
		workspace = workspaces[0]
	}
	workspace, err := filepath.Abs(workspace)
	if err != nil {
		return "", err
	}
	info, err := os.Stat(workspace)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", fmt.Errorf("workspace %q is not a directory", workspace)
	}
	return workspace, nil
}

// gitClone runs git clone with explicit arguments; "--" keeps the URL from
// ever being parsed as an option.
func gitClone(ctx context.Context, url, destination string, stdout, stderr io.Writer) error {
	command := exec.CommandContext(ctx, "git", "clone", "--", url, destination)
	command.Stdout = stdout
	command.Stderr = stderr
	return command.Run()
}
