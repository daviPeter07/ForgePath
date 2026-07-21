package detector

import (
	"encoding/json"
	"encoding/xml"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/daviPeter07/forgepath/internal/project"
)

type metadata struct {
	frameworks      []project.Framework
	packageManagers []project.PackageManager
	hasDocker       bool
}

type dependencyManifest struct {
	Dependencies    map[string]json.RawMessage `json:"dependencies"`
	DevDependencies map[string]json.RawMessage `json:"devDependencies"`
	Require         map[string]json.RawMessage `json:"require"`
	RequireDev      map[string]json.RawMessage `json:"require-dev"`
	PackageManager  string                     `json:"packageManager"`
}

type mavenCoordinate struct {
	GroupID    string `xml:"groupId"`
	ArtifactID string `xml:"artifactId"`
}

type mavenProject struct {
	Parent       mavenCoordinate   `xml:"parent"`
	Dependencies []mavenCoordinate `xml:"dependencies>dependency"`
	Plugins      []mavenCoordinate `xml:"build>plugins>plugin"`
}

var (
	blockCommentPattern = regexp.MustCompile(`(?s)/\*.*?\*/`)
	quotedValuePattern  = regexp.MustCompile(`["']([^"']+)["']`)
)

// Metadata enrichment is best-effort so optional files cannot hide a valid project.
func detectMetadata(path string) metadata {
	result := metadata{}

	if metadataFileExists(filepath.Join(path, "go.mod")) {
		result.packageManagers = appendUniqueManager(result.packageManagers, project.PackageManagerGoModules)
	}

	if data, exists := readMetadataFile(filepath.Join(path, "composer.json")); exists {
		var manifest dependencyManifest
		if json.Unmarshal(data, &manifest) == nil && hasDependency(manifest, "laravel/framework") {
			result.frameworks = appendUniqueFramework(result.frameworks, project.FrameworkLaravel)
		}
		result.packageManagers = appendUniqueManager(result.packageManagers, project.PackageManagerComposer)
	}

	result = detectJavaMetadata(path, result)
	result = detectPythonMetadata(path, result)
	result = detectNodeMetadata(path, result)

	result.hasDocker = hasAnyMetadataFile(path, []string{
		"Dockerfile", "compose.yaml", "compose.yml", "docker-compose.yml", "docker-compose.yaml",
	})

	return result
}

func detectNodeMetadata(path string, result metadata) metadata {
	data, exists := readMetadataFile(filepath.Join(path, "package.json"))
	if !exists {
		return result
	}

	var manifest dependencyManifest
	if json.Unmarshal(data, &manifest) == nil {
		for _, framework := range nodeFrameworks(manifest) {
			result.frameworks = appendUniqueFramework(result.frameworks, framework)
		}
		if manager := packageManagerFromField(manifest.PackageManager); manager != "" {
			result.packageManagers = appendUniqueManager(result.packageManagers, manager)
		}
	}

	markers := []struct {
		name    string
		manager project.PackageManager
	}{
		{name: "bun.lock", manager: project.PackageManagerBun},
		{name: "bun.lockb", manager: project.PackageManagerBun},
		{name: "pnpm-lock.yaml", manager: project.PackageManagerPNPM},
		{name: "yarn.lock", manager: project.PackageManagerYarn},
		{name: "package-lock.json", manager: project.PackageManagerNPM},
	}
	for _, marker := range markers {
		if metadataFileExists(filepath.Join(path, marker.name)) {
			result.packageManagers = appendUniqueManager(result.packageManagers, marker.manager)
		}
	}

	return result
}

func nodeFrameworks(manifest dependencyManifest) []project.Framework {
	frameworks := make([]project.Framework, 0, 4)
	hasNext := hasDependency(manifest, "next")
	hasNuxt := hasDependency(manifest, "nuxt")

	if hasNext {
		frameworks = append(frameworks, project.FrameworkNextJS)
	}
	if hasNuxt {
		frameworks = append(frameworks, project.FrameworkNuxt)
	}

	candidates := []struct {
		dependency string
		framework  project.Framework
		suppressed bool
	}{
		{dependency: "@nestjs/core", framework: project.FrameworkNestJS},
		{dependency: "react", framework: project.FrameworkReact, suppressed: hasNext},
		{dependency: "vue", framework: project.FrameworkVue, suppressed: hasNuxt},
		{dependency: "express", framework: project.FrameworkExpress},
	}
	for _, candidate := range candidates {
		if !candidate.suppressed && hasDependency(manifest, candidate.dependency) {
			frameworks = append(frameworks, candidate.framework)
		}
	}
	return frameworks
}

func detectJavaMetadata(path string, result metadata) metadata {
	if data, exists := readMetadataFile(filepath.Join(path, "pom.xml")); exists {
		result.packageManagers = appendUniqueManager(result.packageManagers, project.PackageManagerMaven)
		var manifest mavenProject
		if xml.Unmarshal(data, &manifest) == nil && mavenUsesSpringBoot(manifest) {
			result.frameworks = appendUniqueFramework(result.frameworks, project.FrameworkSpringBoot)
		}
	}

	for _, name := range []string{"build.gradle", "build.gradle.kts"} {
		if data, exists := readMetadataFile(filepath.Join(path, name)); exists {
			result.packageManagers = appendUniqueManager(result.packageManagers, project.PackageManagerGradle)
			if gradleUsesSpringBoot(string(data)) {
				result.frameworks = appendUniqueFramework(result.frameworks, project.FrameworkSpringBoot)
			}
		}
	}
	return result
}

func mavenUsesSpringBoot(manifest mavenProject) bool {
	coordinates := append([]mavenCoordinate{manifest.Parent}, manifest.Dependencies...)
	coordinates = append(coordinates, manifest.Plugins...)
	for _, coordinate := range coordinates {
		value := strings.ToLower(coordinate.GroupID + "/" + coordinate.ArtifactID)
		if strings.Contains(value, "org.springframework.boot") || strings.Contains(value, "spring-boot") {
			return true
		}
	}
	return false
}

func gradleUsesSpringBoot(content string) bool {
	content = blockCommentPattern.ReplaceAllString(content, "")
	for _, line := range strings.Split(content, "\n") {
		line, _, _ = strings.Cut(line, "//")
		line = strings.TrimSpace(strings.ToLower(line))
		if strings.Contains(line, "org.springframework.boot") || strings.Contains(line, "spring-boot-") {
			return true
		}
	}
	return false
}

func detectPythonMetadata(path string, result metadata) metadata {
	markers := []struct {
		name    string
		manager project.PackageManager
	}{
		{name: "uv.lock", manager: project.PackageManagerUV},
		{name: "poetry.lock", manager: project.PackageManagerPoetry},
		{name: "Pipfile", manager: project.PackageManagerPipenv},
		{name: "requirements.txt", manager: project.PackageManagerPip},
	}
	for _, marker := range markers {
		if metadataFileExists(filepath.Join(path, marker.name)) {
			result.packageManagers = appendUniqueManager(result.packageManagers, marker.manager)
		}
	}

	for _, name := range []string{"pyproject.toml", "requirements.txt", "Pipfile"} {
		if data, exists := readMetadataFile(filepath.Join(path, name)); exists && containsFastAPI(name, string(data)) {
			result.frameworks = appendUniqueFramework(result.frameworks, project.FrameworkFastAPI)
			break
		}
	}
	return result
}

func containsFastAPI(name, content string) bool {
	lines := strings.Split(content, "\n")
	section := ""
	inDependencies := false

	for _, line := range lines {
		line, _, _ = strings.Cut(line, "#")
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if name == "requirements.txt" && isFastAPIRequirement(line) {
			return true
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = strings.ToLower(line)
			inDependencies = false
			continue
		}

		if name == "Pipfile" && (section == "[packages]" || section == "[dev-packages]") {
			dependency, _, _ := strings.Cut(line, "=")
			if strings.EqualFold(strings.TrimSpace(dependency), "fastapi") {
				return true
			}
		}

		if name != "pyproject.toml" {
			continue
		}
		if section == "[tool.poetry.dependencies]" || section == "[tool.poetry.group.dev.dependencies]" {
			dependency, _, _ := strings.Cut(line, "=")
			if strings.EqualFold(strings.TrimSpace(dependency), "fastapi") {
				return true
			}
		}
		if strings.HasPrefix(strings.ToLower(line), "dependencies") {
			inDependencies = strings.Contains(line, "[")
		}
		if inDependencies || section == "[project.optional-dependencies]" {
			for _, match := range quotedValuePattern.FindAllStringSubmatch(line, -1) {
				if isFastAPIRequirement(match[1]) {
					return true
				}
			}
		}
		if inDependencies && strings.Contains(line, "]") {
			inDependencies = false
		}
	}
	return false
}

func isFastAPIRequirement(value string) bool {
	value = strings.TrimSpace(strings.ToLower(value))
	if index := strings.IndexAny(value, "[<>=!~ "); index >= 0 {
		value = value[:index]
	}
	return value == "fastapi"
}

func hasDependency(manifest dependencyManifest, name string) bool {
	_, dependency := manifest.Dependencies[name]
	_, devDependency := manifest.DevDependencies[name]
	_, requirement := manifest.Require[name]
	_, devRequirement := manifest.RequireDev[name]
	return dependency || devDependency || requirement || devRequirement
}

func packageManagerFromField(value string) project.PackageManager {
	name, _, _ := strings.Cut(strings.TrimSpace(value), "@")
	switch strings.ToLower(name) {
	case "npm":
		return project.PackageManagerNPM
	case "pnpm":
		return project.PackageManagerPNPM
	case "yarn":
		return project.PackageManagerYarn
	case "bun":
		return project.PackageManagerBun
	default:
		return ""
	}
}

func appendUniqueFramework(values []project.Framework, value project.Framework) []project.Framework {
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}

func appendUniqueManager(values []project.PackageManager, value project.PackageManager) []project.PackageManager {
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}

func hasAnyMetadataFile(path string, names []string) bool {
	for _, name := range names {
		if metadataFileExists(filepath.Join(path, name)) {
			return true
		}
	}
	return false
}

func metadataFileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.Mode().IsRegular()
}

func readMetadataFile(path string) ([]byte, bool) {
	if !metadataFileExists(path) {
		return nil, false
	}
	data, err := os.ReadFile(path)
	return data, err == nil
}
