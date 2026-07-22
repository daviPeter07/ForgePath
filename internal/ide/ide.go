package ide

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/daviPeter07/forgepath/internal/project"
)

type IDE struct {
	ID           string
	Name         string
	Executable   string
	Arguments    []string
	Technologies []project.Technology
}

type candidate struct {
	IDE
	commands   []string
	paths      []string
	requires   []string
	requireAny []string
}

type finder struct {
	goos     string
	lookPath func(string) (string, error)
	glob     func(string) ([]string, error)
	stat     func(string) (os.FileInfo, error)
	getenv   func(string) string
}

func Discover() []IDE {
	return discover(finder{
		goos:     runtime.GOOS,
		lookPath: exec.LookPath,
		glob:     filepath.Glob,
		stat:     os.Stat,
		getenv:   os.Getenv,
	})
}

func Rank(installed []IDE, technology project.Technology) []IDE {
	ranked := make([]IDE, 0, len(installed))
	for _, editor := range installed {
		if technology == "" || editor.Supports(technology) {
			ranked = append(ranked, editor)
		}
	}
	priority := preferredEditors[technology]
	score := func(editor IDE) int {
		for index, id := range priority {
			if editor.ID == id {
				return index
			}
		}
		if editor.Supports(technology) {
			return len(priority) + 10
		}
		return len(priority) + 100
	}
	sort.SliceStable(ranked, func(left, right int) bool {
		leftScore := score(ranked[left])
		rightScore := score(ranked[right])
		if leftScore != rightScore {
			return leftScore < rightScore
		}
		return ranked[left].Name < ranked[right].Name
	})
	return ranked
}

func (editor IDE) Supports(technology project.Technology) bool {
	for _, supported := range editor.Technologies {
		if supported == technology {
			return true
		}
	}
	return false
}

func discover(system finder) []IDE {
	candidates := platformCandidates(system.goos, system.getenv)
	installed := make([]IDE, 0, len(candidates))
	seenIDs := make(map[string]struct{})
	for _, candidate := range candidates {
		if _, exists := seenIDs[candidate.ID]; exists {
			continue
		}
		if !requirementsMet(system, candidate.requires) || !anyRequirementMet(system, candidate.requireAny) {
			continue
		}
		executable := findExecutable(system, candidate)
		if executable == "" {
			continue
		}
		seenIDs[candidate.ID] = struct{}{}
		found := candidate.IDE
		found.Executable = executable
		installed = append(installed, found)
	}
	return installed
}

func findExecutable(system finder, candidate candidate) string {
	for _, command := range candidate.commands {
		path, err := system.lookPath(command)
		if err == nil && supportedExecutable(system.goos, path) {
			return path
		}
	}

	var matches []string
	for _, pattern := range candidate.paths {
		found, err := system.glob(pattern)
		if err != nil {
			continue
		}
		if len(found) == 0 && !strings.ContainsAny(pattern, "*?[") {
			found = []string{pattern}
		}
		for _, path := range found {
			if info, err := system.stat(path); err == nil && executableFile(system.goos, path, info) {
				matches = append(matches, path)
			}
		}
	}
	sort.Sort(sort.Reverse(sort.StringSlice(matches)))
	if len(matches) > 0 {
		return matches[0]
	}
	return ""
}

func requirementsMet(system finder, paths []string) bool {
	for _, path := range paths {
		if _, err := system.stat(path); err != nil {
			return false
		}
	}
	return true
}

func anyRequirementMet(system finder, patterns []string) bool {
	if len(patterns) == 0 {
		return true
	}
	for _, pattern := range patterns {
		matches, _ := system.glob(pattern)
		if len(matches) == 0 && !strings.ContainsAny(pattern, "*?[") {
			matches = []string{pattern}
		}
		for _, path := range matches {
			if _, err := system.stat(path); err == nil {
				return true
			}
		}
	}
	return false
}

func executableFile(goos, path string, info os.FileInfo) bool {
	if !info.Mode().IsRegular() || !supportedExecutable(goos, path) {
		return false
	}
	return goos == "windows" || info.Mode().Perm()&0o111 != 0
}

func supportedExecutable(goos, path string) bool {
	if goos != "windows" {
		return true
	}
	extension := strings.ToLower(filepath.Ext(path))
	return extension != ".cmd" && extension != ".bat"
}

var allTechnologies = []project.Technology{
	project.TechnologyTypeScript, project.TechnologyJavaScript, project.TechnologyPython,
	project.TechnologyGo, project.TechnologyJava, project.TechnologyPHP, project.TechnologyDocker,
	project.TechnologyRust, project.TechnologyRuby, project.TechnologySwift, project.TechnologyElixir,
}

var preferredEditors = map[project.Technology][]string{
	project.TechnologyPHP:        {"phpstorm", "vscode", "cursor", "zed", "sublime"},
	project.TechnologyTypeScript: {"webstorm", "vscode", "cursor", "zed", "sublime"},
	project.TechnologyJavaScript: {"webstorm", "vscode", "cursor", "zed", "sublime"},
	project.TechnologyPython:     {"pycharm", "vscode", "cursor", "zed", "sublime"},
	project.TechnologyGo:         {"goland", "vscode", "cursor", "zed", "sublime"},
	project.TechnologyJava:       {"intellij", "vscode", "cursor", "zed", "sublime"},
	project.TechnologyRust:       {"rustrover", "vscode", "cursor", "zed", "sublime"},
	project.TechnologyRuby:       {"rubymine", "vscode", "cursor", "zed", "sublime"},
	project.TechnologySwift:      {"xcode", "vscode", "zed", "sublime"},
	project.TechnologyElixir:     {"vscode", "cursor", "zed", "sublime"},
	project.TechnologyDocker:     {"vscode", "cursor", "zed", "sublime"},
}

func platformCandidates(goos string, getenv func(string) string) []candidate {
	supports := func(technologies ...project.Technology) []project.Technology { return technologies }
	candidates := []candidate{
		{IDE: IDE{ID: "phpstorm", Name: "PhpStorm", Technologies: supports(project.TechnologyPHP, project.TechnologyTypeScript, project.TechnologyJavaScript)}, commands: []string{"phpstorm", "phpstorm64.exe"}},
		{IDE: IDE{ID: "webstorm", Name: "WebStorm", Technologies: supports(project.TechnologyTypeScript, project.TechnologyJavaScript)}, commands: []string{"webstorm", "webstorm64.exe"}},
		{IDE: IDE{ID: "pycharm", Name: "PyCharm", Technologies: supports(project.TechnologyPython)}, commands: []string{"pycharm", "pycharm64.exe"}},
		{IDE: IDE{ID: "goland", Name: "GoLand", Technologies: supports(project.TechnologyGo)}, commands: []string{"goland", "goland64.exe"}},
		{IDE: IDE{ID: "intellij", Name: "IntelliJ IDEA", Technologies: supports(project.TechnologyJava)}, commands: []string{"idea", "idea64.exe"}},
		{IDE: IDE{ID: "rustrover", Name: "RustRover", Technologies: supports(project.TechnologyRust)}, commands: []string{"rustrover", "rustrover64.exe"}},
		{IDE: IDE{ID: "rubymine", Name: "RubyMine", Technologies: supports(project.TechnologyRuby)}, commands: []string{"rubymine", "rubymine64.exe"}},
		{IDE: IDE{ID: "vscode", Name: "Visual Studio Code", Technologies: allTechnologies}, commands: []string{"code", "code.exe"}},
		{IDE: IDE{ID: "cursor", Name: "Cursor", Technologies: allTechnologies}, commands: []string{"cursor", "Cursor.exe"}},
		{IDE: IDE{ID: "zed", Name: "Zed", Technologies: allTechnologies}, commands: []string{"zed", "zed.exe"}},
		{IDE: IDE{ID: "sublime", Name: "Sublime Text", Technologies: allTechnologies}, commands: []string{"subl", "sublime_text.exe"}},
	}

	switch goos {
	case "windows":
		local := getenv("LOCALAPPDATA")
		programFiles := getenv("ProgramFiles")
		toolbox := filepath.Join(local, "JetBrains", "Toolbox", "apps")
		jetbrains := filepath.Join(programFiles, "JetBrains")
		windowsPaths := map[string][]string{
			"phpstorm":  {filepath.Join(toolbox, "PhpStorm", "*", "*", "bin", "phpstorm64.exe"), filepath.Join(jetbrains, "PhpStorm *", "bin", "phpstorm64.exe")},
			"webstorm":  {filepath.Join(toolbox, "WebStorm", "*", "*", "bin", "webstorm64.exe"), filepath.Join(jetbrains, "WebStorm *", "bin", "webstorm64.exe")},
			"pycharm":   {filepath.Join(toolbox, "PyCharm-*", "*", "*", "bin", "pycharm64.exe"), filepath.Join(jetbrains, "PyCharm *", "bin", "pycharm64.exe")},
			"goland":    {filepath.Join(toolbox, "GoLand", "*", "*", "bin", "goland64.exe"), filepath.Join(jetbrains, "GoLand *", "bin", "goland64.exe")},
			"intellij":  {filepath.Join(toolbox, "IDEA-*", "*", "*", "bin", "idea64.exe"), filepath.Join(jetbrains, "IntelliJ IDEA *", "bin", "idea64.exe")},
			"rustrover": {filepath.Join(toolbox, "RustRover", "*", "*", "bin", "rustrover64.exe"), filepath.Join(jetbrains, "RustRover *", "bin", "rustrover64.exe")},
			"rubymine":  {filepath.Join(toolbox, "RubyMine", "*", "*", "bin", "rubymine64.exe"), filepath.Join(jetbrains, "RubyMine *", "bin", "rubymine64.exe")},
			"vscode":    {filepath.Join(local, "Programs", "Microsoft VS Code", "Code.exe"), filepath.Join(programFiles, "Microsoft VS Code", "Code.exe")},
			"cursor":    {filepath.Join(local, "Programs", "cursor", "Cursor.exe")},
			"sublime":   {filepath.Join(programFiles, "Sublime Text", "sublime_text.exe")},
		}
		for index := range candidates {
			candidates[index].paths = windowsPaths[candidates[index].ID]
		}
	case "darwin":
		home := getenv("HOME")
		applications := map[string]string{
			"phpstorm": "PhpStorm.app/Contents/MacOS/phpstorm", "webstorm": "WebStorm.app/Contents/MacOS/webstorm",
			"pycharm": "PyCharm.app/Contents/MacOS/pycharm", "goland": "GoLand.app/Contents/MacOS/goland",
			"intellij": "IntelliJ IDEA.app/Contents/MacOS/idea", "rustrover": "RustRover.app/Contents/MacOS/rustrover",
			"rubymine": "RubyMine.app/Contents/MacOS/rubymine", "vscode": "Visual Studio Code.app/Contents/Resources/app/bin/code",
			"cursor": "Cursor.app/Contents/Resources/app/bin/cursor", "zed": "Zed.app/Contents/MacOS/zed",
			"sublime": "Sublime Text.app/Contents/MacOS/sublime_text",
		}
		for index := range candidates {
			if relative := applications[candidates[index].ID]; relative != "" {
				candidates[index].paths = []string{
					filepath.Join("/Applications", filepath.FromSlash(relative)),
					filepath.Join(home, "Applications", filepath.FromSlash(relative)),
				}
				if !slicesContains([]string{"vscode", "cursor", "zed", "sublime"}, candidates[index].ID) {
					candidates[index].paths = append(candidates[index].paths,
						filepath.Join(home, "Library", "Application Support", "JetBrains", "Toolbox", "apps", "*", "*", "*", filepath.FromSlash(relative)),
					)
				}
			}
		}
		candidates = append(candidates, candidate{
			IDE:      IDE{ID: "xcode", Name: "Xcode", Arguments: []string{"-a", "Xcode"}, Technologies: supports(project.TechnologySwift)},
			commands: []string{"open"}, requires: []string{"/Applications/Xcode.app"},
		})
	case "linux":
		home := getenv("HOME")
		toolbox := filepath.Join(home, ".local", "share", "JetBrains", "Toolbox", "apps")
		for index := range candidates {
			if strings.HasPrefix(candidates[index].ID, "vscode") {
				continue
			}
			binary := candidates[index].commands[0]
			candidates[index].paths = []string{filepath.Join(toolbox, "*", "*", "*", "bin", binary), filepath.Join("/snap/bin", binary)}
		}
		flatpakRoot := filepath.Join(home, ".local", "share", "flatpak", "app")
		candidates = append(candidates,
			candidate{IDE: IDE{ID: "vscode", Name: "Visual Studio Code", Arguments: []string{"run", "com.visualstudio.code"}, Technologies: allTechnologies}, commands: []string{"flatpak"}, requireAny: []string{filepath.Join(flatpakRoot, "com.visualstudio.code"), "/var/lib/flatpak/app/com.visualstudio.code"}},
			candidate{IDE: IDE{ID: "phpstorm", Name: "PhpStorm", Arguments: []string{"run", "com.jetbrains.PhpStorm"}, Technologies: supports(project.TechnologyPHP, project.TechnologyTypeScript, project.TechnologyJavaScript)}, commands: []string{"flatpak"}, requireAny: []string{filepath.Join(flatpakRoot, "com.jetbrains.PhpStorm"), "/var/lib/flatpak/app/com.jetbrains.PhpStorm"}},
		)
	}
	return candidates
}

func slicesContains(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}
