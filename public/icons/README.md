# Technology icons

These colored SVG assets come from [Devicon](https://devicon.dev/) and are
stored locally so ForgePath interfaces and documentation do not depend on a
remote CDN.

- Source repository: https://github.com/devicons/devicon
- Release: https://github.com/devicons/devicon/releases/tag/v2.17.0
- Download source: https://cdn.jsdelivr.net/gh/devicons/devicon@v2.17.0/icons
- License: MIT; see `LICENSE.devicon`

Product names and logos remain the property of their respective owners. The
assets are intended only to identify the detected technology.

ForgePath embeds these files in the Go binary, rasterizes each SVG to a small
RGBA canvas, and renders the pixels with ANSI truecolor half-block characters.
This keeps the original logo colors without requiring a terminal-specific
image protocol. Text badges and Nerd Font glyphs remain available as explicit
fallback modes.
