package project

type Technology string

const (
	TechnologyTypeScript Technology = "TypeScript"
	TechnologyJavaScript Technology = "JavaScript"
	TechnologyPython     Technology = "Python"
	TechnologyGo         Technology = "Go"
	TechnologyJava       Technology = "Java"
	TechnologyPHP        Technology = "PHP"
)

type Project struct {
	Name       string
	Path       string
	Technology Technology
	Markers    []string
}
