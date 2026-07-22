package project

import "time"

type Technology string
type Framework string
type PackageManager string

const (
	TechnologyTypeScript Technology = "TypeScript"
	TechnologyJavaScript Technology = "JavaScript"
	TechnologyPython     Technology = "Python"
	TechnologyGo         Technology = "Go"
	TechnologyJava       Technology = "Java"
	TechnologyPHP        Technology = "PHP"
	TechnologyDocker     Technology = "Docker"
	TechnologyRust       Technology = "Rust"
	TechnologyRuby       Technology = "Ruby"
	TechnologySwift      Technology = "Swift"
	TechnologyElixir     Technology = "Elixir"
)

const (
	FrameworkNextJS     Framework = "Next.js"
	FrameworkReact      Framework = "React"
	FrameworkVue        Framework = "Vue.js"
	FrameworkNuxt       Framework = "Nuxt"
	FrameworkNestJS     Framework = "NestJS"
	FrameworkExpress    Framework = "Express"
	FrameworkLaravel    Framework = "Laravel"
	FrameworkSpringBoot Framework = "Spring Boot"
	FrameworkFastAPI    Framework = "FastAPI"
)

const (
	PackageManagerGoModules PackageManager = "Go Modules"
	PackageManagerComposer  PackageManager = "Composer"
	PackageManagerMaven     PackageManager = "Maven"
	PackageManagerGradle    PackageManager = "Gradle"
	PackageManagerPip       PackageManager = "pip"
	PackageManagerPipenv    PackageManager = "Pipenv"
	PackageManagerPoetry    PackageManager = "Poetry"
	PackageManagerUV        PackageManager = "uv"
	PackageManagerNPM       PackageManager = "npm"
	PackageManagerPNPM      PackageManager = "pnpm"
	PackageManagerYarn      PackageManager = "Yarn"
	PackageManagerBun       PackageManager = "Bun"
)

type Project struct {
	Name            string
	Path            string
	Technology      Technology
	Markers         []string
	Frameworks      []Framework
	PackageManagers []PackageManager
	HasDocker       bool
	GitBranch       string
	GitDirty        bool
	GitStatusKnown  bool
	Favorite        bool
	LastOpened      time.Time
}
