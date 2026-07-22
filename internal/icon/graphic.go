package icon

import (
	"fmt"
	"image"
	"image/color"
	"sync"

	"github.com/daviPeter07/forgepath/internal/project"
	iconassets "github.com/daviPeter07/forgepath/public/icons"
	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
)

const (
	graphicWidth  = 8
	graphicHeight = 8
)

var (
	graphicCache sync.Map
	graphicMu    sync.Mutex
)

var graphicFiles = map[project.Technology]string{
	project.TechnologyTypeScript: "typescript.svg",
	project.TechnologyJavaScript: "javascript.svg",
	project.TechnologyPython:     "python.svg",
	project.TechnologyGo:         "go.svg",
	project.TechnologyJava:       "java.svg",
	project.TechnologyPHP:        "php.svg",
	project.TechnologyDocker:     "docker.svg",
	project.TechnologyRust:       "rust.svg",
	project.TechnologyRuby:       "ruby.svg",
	project.TechnologySwift:      "swift.svg",
	project.TechnologyElixir:     "elixir.svg",
}

func Graphic(technology project.Technology) (string, error) {
	if cached, ok := graphicCache.Load(technology); ok {
		return cached.(string), nil
	}
	graphicMu.Lock()
	defer graphicMu.Unlock()
	if cached, ok := graphicCache.Load(technology); ok {
		return cached.(string), nil
	}
	name := graphicFiles[technology]
	if name == "" {
		return "", fmt.Errorf("no graphic icon for %q", technology)
	}
	file, err := iconassets.Files.Open(name)
	if err != nil {
		return "", err
	}
	defer file.Close()

	parsed, err := oksvg.ReadIconStream(file, oksvg.StrictErrorMode)
	if err != nil {
		return "", fmt.Errorf("parse icon %q: %w", name, err)
	}
	canvas := image.NewRGBA(image.Rect(0, 0, graphicWidth, graphicHeight))
	scanner := rasterx.NewScannerGV(graphicWidth, graphicHeight, canvas, canvas.Bounds())
	raster := rasterx.NewDasher(graphicWidth, graphicHeight, scanner)
	parsed.SetTarget(0, 0, graphicWidth, graphicHeight)
	parsed.Draw(raster, 1)

	rendered := renderANSI(canvas)
	graphicCache.Store(technology, rendered)
	return rendered, nil
}

func renderANSI(source image.Image) string {
	bounds := source.Bounds()
	result := make([]byte, 0, bounds.Dx()*bounds.Dy()*24)
	for y := bounds.Min.Y; y < bounds.Max.Y; y += 2 {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			top, topVisible := visiblePixel(source.At(x, y))
			bottom := color.NRGBA{}
			bottomVisible := false
			if y+1 < bounds.Max.Y {
				bottom, bottomVisible = visiblePixel(source.At(x, y+1))
			}
			switch {
			case topVisible && bottomVisible:
				result = fmt.Appendf(result, "\x1b[0;38;2;%d;%d;%d;48;2;%d;%d;%dm▀", top.R, top.G, top.B, bottom.R, bottom.G, bottom.B)
			case topVisible:
				result = fmt.Appendf(result, "\x1b[0;38;2;%d;%d;%dm▀", top.R, top.G, top.B)
			case bottomVisible:
				result = fmt.Appendf(result, "\x1b[0;38;2;%d;%d;%dm▄", bottom.R, bottom.G, bottom.B)
			default:
				result = append(result, "\x1b[0m "...)
			}
		}
		result = append(result, "\x1b[0m"...)
		if y+2 < bounds.Max.Y {
			result = append(result, '\n')
		}
	}
	return string(result)
}

func visiblePixel(source color.Color) (color.NRGBA, bool) {
	pixel := color.NRGBAModel.Convert(source).(color.NRGBA)
	return pixel, pixel.A >= 32
}
