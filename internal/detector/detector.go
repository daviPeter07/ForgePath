package detector

import "github.com/daviPeter07/forgepath/internal/project"

type Result struct {
	Technology      project.Technology
	Markers         []string
	Frameworks      []project.Framework
	PackageManagers []project.PackageManager
	HasDocker       bool
}
