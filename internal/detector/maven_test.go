package detector

import (
	"reflect"
	"testing"

	"github.com/daviPeter07/forgepath/internal/project"
)

func TestDetectSpringBootFromMaven(t *testing.T) {
	const header = `<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
  <modelVersion>4.0.0</modelVersion>
`
	tests := []struct {
		name       string
		pom        string
		frameworks []project.Framework
	}{
		{
			name: "starter parent",
			pom: header + `  <parent>
    <groupId>org.springframework.boot</groupId>
    <artifactId>spring-boot-starter-parent</artifactId>
    <version>3.5.0</version>
  </parent>
</project>`,
			frameworks: []project.Framework{project.FrameworkSpringBoot},
		},
		{
			name: "starter dependency without parent",
			pom: header + `  <dependencies>
    <dependency>
      <groupId>org.springframework.boot</groupId>
      <artifactId>spring-boot-starter-web</artifactId>
    </dependency>
  </dependencies>
</project>`,
			frameworks: []project.Framework{project.FrameworkSpringBoot},
		},
		{
			name: "maven plugin only",
			pom: header + `  <build>
    <plugins>
      <plugin>
        <groupId>org.springframework.boot</groupId>
        <artifactId>spring-boot-maven-plugin</artifactId>
      </plugin>
    </plugins>
  </build>
</project>`,
			frameworks: []project.Framework{project.FrameworkSpringBoot},
		},
		{
			name: "plain Java",
			pom: header + `  <dependencies>
    <dependency>
      <groupId>junit</groupId>
      <artifactId>junit</artifactId>
    </dependency>
  </dependencies>
</project>`,
		},
		{
			name: "malformed pom keeps Maven",
			pom:  `<project><parent><groupId>org.springframework.boot`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			writeProjectFiles(t, dir, map[string]string{"pom.xml": tt.pom})

			result, found, err := Detect(dir)
			if err != nil || !found {
				t.Fatalf("Detect() found = %t, error = %v", found, err)
			}
			if result.Technology != project.TechnologyJava {
				t.Fatalf("Technology = %q, want Java", result.Technology)
			}
			if !reflect.DeepEqual(result.Frameworks, tt.frameworks) {
				t.Fatalf("Frameworks = %q, want %q", result.Frameworks, tt.frameworks)
			}
			if !reflect.DeepEqual(result.PackageManagers, []project.PackageManager{project.PackageManagerMaven}) {
				t.Fatalf("PackageManagers = %q, want Maven", result.PackageManagers)
			}
		})
	}
}

func TestGradleSpringBootIgnoresComments(t *testing.T) {
	tests := map[string]bool{
		`plugins { id 'org.springframework.boot' version '3.5.0' }`:                          true,
		"dependencies {\n  implementation 'org.springframework.boot:spring-boot-starter'\n}": true,
		"// id 'org.springframework.boot'\nplugins { id 'java' }":                            false,
		"/* org.springframework.boot\n spring-boot-starter */\nplugins { id 'java' }":        false,
	}
	for content, want := range tests {
		if got := gradleUsesSpringBoot(content); got != want {
			t.Errorf("gradleUsesSpringBoot(%q) = %t, want %t", content, got, want)
		}
	}
}

func TestPackageManagerFromField(t *testing.T) {
	tests := map[string]project.PackageManager{
		"pnpm@10.0.0": project.PackageManagerPNPM,
		" yarn@4 ":    project.PackageManagerYarn,
		"npm":         project.PackageManagerNPM,
		"Bun@1.2":     project.PackageManagerBun,
		"deno@2":      "",
		"":            "",
	}
	for field, want := range tests {
		if got := packageManagerFromField(field); got != want {
			t.Errorf("packageManagerFromField(%q) = %q, want %q", field, got, want)
		}
	}
}

func TestDetectFallsBackToSourceFileExtension(t *testing.T) {
	dir := t.TempDir()
	writeProjectFiles(t, dir, map[string]string{"README.md": "", "main.rs": "fn main() {}"})

	result, found, err := Detect(dir)
	if err != nil || !found {
		t.Fatalf("Detect() found = %t, error = %v", found, err)
	}
	if result.Technology != project.TechnologyRust || !reflect.DeepEqual(result.Markers, []string{"main.rs"}) {
		t.Fatalf("Detect() = %+v, want Rust from main.rs", result)
	}
}
