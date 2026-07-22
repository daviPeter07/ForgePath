package icons

import "embed"

// Files contains the pinned Devicon SVG assets used by the terminal renderer.
//
//go:embed *.svg
var Files embed.FS
