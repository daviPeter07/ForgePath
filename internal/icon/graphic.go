package icon

import (
	"fmt"
	"image"
	"image/color"
	"regexp"
	"strings"
	"sync"

	"github.com/daviPeter07/forgepath/internal/project"
	iconassets "github.com/daviPeter07/forgepath/public/icons"
	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
)

const (
	graphicWidth   = 8
	graphicHeight  = 8
	minGraphicSize = 4
)

// darkLogoTint is Rust's crab color, readable on dark and light backgrounds.
var darkLogoTint = color.NRGBA{R: 0xDE, G: 0xA5, B: 0x84, A: 0xFF}

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

// Graphic renders a technology logo at the small list size.
func Graphic(technology project.Technology) (string, error) {
	return GraphicSized(technology, graphicWidth)
}

// GraphicSized renders a technology logo as a size×size pixel image, which
// takes size cells horizontally and size/2 rows (two pixels per half block).
func GraphicSized(technology project.Technology, size int) (string, error) {
	if size < minGraphicSize {
		return "", fmt.Errorf("logo size %d is below the minimum of %d", size, minGraphicSize)
	}
	size -= size % 2
	key := graphicKey{technology: technology, size: size}
	if cached, ok := graphicCache.Load(key); ok {
		return cached.(string), nil
	}
	graphicMu.Lock()
	defer graphicMu.Unlock()
	if cached, ok := graphicCache.Load(key); ok {
		return cached.(string), nil
	}
	name := graphicFiles[technology]
	if name == "" {
		return "", fmt.Errorf("no graphic icon for %q", technology)
	}
	data, err := iconassets.Files.ReadFile(name)
	if err != nil {
		return "", err
	}
	parsed, err := oksvg.ReadIconStream(strings.NewReader(normalizeArcFlags(string(data))), oksvg.StrictErrorMode)
	if err != nil {
		return "", fmt.Errorf("parse icon %q: %w", name, err)
	}
	canvas := image.NewRGBA(image.Rect(0, 0, size, size))
	scanner := rasterx.NewScannerGV(size, size, canvas, canvas.Bounds())
	raster := rasterx.NewDasher(size, size, scanner)
	parsed.SetTarget(0, 0, float64(size), float64(size))
	parsed.Draw(raster, 1)
	lightenDarkLogo(canvas)

	rendered := renderANSI(canvas)
	graphicCache.Store(key, rendered)
	return rendered, nil
}

type graphicKey struct {
	technology project.Technology
	size       int
}

var (
	pathData   = regexp.MustCompile(`\bd="([^"]*)"`)
	pathNumber = regexp.MustCompile(`^[-+]?(?:\d+\.?\d*|\.\d+)(?:[eE][-+]?\d+)?`)
)

// normalizeArcFlags rewrites arc commands whose flags are packed against the
// next number ("a15 15 0 017.8 4.5" means flags 0 and 1, then x 7.8). That
// form is valid SVG, but oksvg misreads it and drops the shapes after it,
// which removed the letters from the TypeScript logo.
func normalizeArcFlags(svg string) string {
	return pathData.ReplaceAllStringFunc(svg, func(attribute string) string {
		data := attribute[len(`d="`) : len(attribute)-1]
		if !strings.ContainsAny(data, "aA") {
			return attribute
		}
		return `d="` + expandPathData(data) + `"`
	})
}

func expandPathData(data string) string {
	var out strings.Builder
	var command byte
	argument := 0
	first := true
	for index := 0; index < len(data); {
		character := data[index]
		switch {
		case strings.IndexByte(" ,\t\r\n", character) >= 0:
			index++
		case strings.IndexByte("MmZzLlHhVvCcSsQqTtAa", character) >= 0:
			command, argument, first = character, 0, true
			out.WriteByte(character)
			index++
		default:
			var token string
			if (command == 'a' || command == 'A') && (argument%7 == 3 || argument%7 == 4) {
				token = data[index : index+1]
				index++
			} else {
				token = pathNumber.FindString(data[index:])
				if token == "" {
					return data // not path data we understand; leave it untouched
				}
				index += len(token)
			}
			if !first {
				out.WriteByte(' ')
			}
			out.WriteString(token)
			first = false
			argument++
		}
	}
	return out.String()
}

// lightenDarkLogo recolors logos drawn almost entirely in black (Rust), which
// would be invisible on dark terminals, with a light tone that also reads on
// light backgrounds.
func lightenDarkLogo(canvas *image.RGBA) {
	visible, dark := 0, 0
	bounds := canvas.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			pixel := color.NRGBAModel.Convert(canvas.At(x, y)).(color.NRGBA)
			if pixel.A < 32 {
				continue
			}
			visible++
			if int(pixel.R)+int(pixel.G)+int(pixel.B) < 150 {
				dark++
			}
		}
	}
	if visible == 0 || dark*10 < visible*6 {
		return
	}
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			pixel := color.NRGBAModel.Convert(canvas.At(x, y)).(color.NRGBA)
			if pixel.A == 0 {
				continue
			}
			canvas.Set(x, y, color.NRGBA{R: darkLogoTint.R, G: darkLogoTint.G, B: darkLogoTint.B, A: pixel.A})
		}
	}
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
