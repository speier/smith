// Package render provides ANSI terminal rendering.
//
// This package is independent of vdom, style, and layout. It takes a LayoutBox
// tree (with computed positions/sizes) and produces ANSI escape sequences.
//
// This is where all terminal-specific output happens - colors, positioning, borders.
package render

import (
	"fmt"
	"strings"
)

// colorToANSI converts a color string to ANSI code

// RenderBufferFull renders an entire buffer to ANSI output with synchronized output
func RenderBufferFull(buf *Buffer) string {
	var out strings.Builder

	// Begin synchronized output (CSI 2026) for flicker-free rendering
	out.WriteString("\x1b[?2026h")

	// Clear screen and move to home
	out.WriteString("\x1b[2J\x1b[H")

	var currentStyle Style
	styleActive := false

	for y := 0; y < buf.Height; y++ {
		for x := 0; x < buf.Width; x++ {
			cell := buf.Get(x, y)

			// Check if style changed
			if !stylesEqual(cell.Style, currentStyle) {
				// Reset previous style
				if styleActive {
					out.WriteString("\x1b[0m")
				}

				// Apply new style
				styleSeq := buildStyleSequence(cell.Style)
				if styleSeq != "" {
					out.WriteString(styleSeq)
					styleActive = true
				} else {
					styleActive = false
				}
				currentStyle = cell.Style
			}

			// Write character
			out.WriteRune(cell.Char)
		}

		// Move to next line (except for last line)
		if y < buf.Height-1 {
			out.WriteString("\r\n")
		}
	}

	// Reset style at end
	if styleActive {
		out.WriteString("\x1b[0m")
	}

	// End synchronized output
	out.WriteString("\x1b[?2026l")

	return out.String()
}

// RenderBufferDiff renders only the changed regions between two buffers
func RenderBufferDiff(prev, curr *Buffer, diff *DiffResult) string {
	if diff.FullRedraw {
		return RenderBufferFull(curr)
	}

	if len(diff.Regions) == 0 {
		return "" // No changes
	}

	var out strings.Builder

	// Begin synchronized output
	out.WriteString("\x1b[?2026h")

	// Render each changed region
	for _, region := range diff.Regions {
		var currentStyle Style
		styleActive := false

		for dy := 0; dy < region.Height; dy++ {
			y := region.Y + dy

			// Move cursor to start of region
			out.WriteString(fmt.Sprintf("\x1b[%d;%dH", y+1, region.X+1))

			for dx := 0; dx < region.Width; dx++ {
				x := region.X + dx
				cell := curr.Get(x, y)

				// Check if style changed
				if !stylesEqual(cell.Style, currentStyle) {
					// Reset previous style
					if styleActive {
						out.WriteString("\x1b[0m")
					}

					// Apply new style
					styleSeq := buildStyleSequence(cell.Style)
					if styleSeq != "" {
						out.WriteString(styleSeq)
						styleActive = true
					} else {
						styleActive = false
					}
					currentStyle = cell.Style
				}

				// Write character
				out.WriteRune(cell.Char)
			}
		}

		// Reset style after each region
		if styleActive {
			out.WriteString("\x1b[0m")
		}
	}

	// End synchronized output
	out.WriteString("\x1b[?2026l")

	return out.String()
}

// stylesEqual checks if two styles are identical
func stylesEqual(a, b Style) bool {
	return a.FgColor == b.FgColor &&
		a.BgColor == b.BgColor &&
		a.Bold == b.Bold &&
		a.Italic == b.Italic &&
		a.Underline == b.Underline &&
		a.Strikethrough == b.Strikethrough &&
		a.Dim == b.Dim &&
		a.Reverse == b.Reverse
}

// buildStyleSequence creates an ANSI escape sequence for a style
func buildStyleSequence(st Style) string {
	var codes []string

	// Style attributes first (traditional ANSI order)
	// Bold
	if st.Bold {
		codes = append(codes, "1")
	}

	// Dim
	if st.Dim {
		codes = append(codes, "2")
	}

	// Italic
	if st.Italic {
		codes = append(codes, "3")
	}

	// Underline
	if st.Underline {
		codes = append(codes, "4")
	}

	// Reverse
	if st.Reverse {
		codes = append(codes, "7")
	}

	// Strikethrough
	if st.Strikethrough {
		codes = append(codes, "9")
	}

	// Foreground color
	if st.FgColor != "" {
		if code := colorToANSI(st.FgColor, false); code != "" {
			codes = append(codes, code)
		}
	}

	// Background color
	if st.BgColor != "" {
		if code := colorToANSI(st.BgColor, true); code != "" {
			codes = append(codes, code)
		}
	}

	if len(codes) == 0 {
		return ""
	}

	return "\x1b[" + strings.Join(codes, ";") + "m"
}

// colorToANSI converts a color string to ANSI code
func colorToANSI(color string, background bool) string {
	offset := 30
	if background {
		offset = 40
	}

	// Handle hex colors
	if strings.HasPrefix(color, "#") {
		hexToANSI := map[string]int{
			"#ffffff": 97, // bright white
			"#ffff00": 93, // bright yellow
			"#ff0000": 91, // bright red
			"#00ff00": 92, // bright green
			"#0000ff": 94, // bright blue
			"#ff00ff": 95, // bright magenta
			"#00ffff": 96, // bright cyan
			"#000000": 30, // black
			"#808080": 90, // gray
		}

		if code, ok := hexToANSI[strings.ToLower(color)]; ok {
			if background {
				// Convert foreground code to background (add 10)
				return fmt.Sprintf("%d", code+10)
			}
			return fmt.Sprintf("%d", code)
		}
		// Default to white
		return fmt.Sprintf("%d", offset+7)
	}

	// Named colors
	switch color {
	case "black":
		return fmt.Sprintf("%d", offset+0)
	case "red":
		return fmt.Sprintf("%d", offset+1)
	case "green":
		return fmt.Sprintf("%d", offset+2)
	case "yellow":
		return fmt.Sprintf("%d", offset+3)
	case "blue":
		return fmt.Sprintf("%d", offset+4)
	case "magenta":
		return fmt.Sprintf("%d", offset+5)
	case "cyan":
		return fmt.Sprintf("%d", offset+6)
	case "white":
		return fmt.Sprintf("%d", offset+7)
	}

	// Bright colors
	switch color {
	case "bright-black", "gray", "grey":
		return fmt.Sprintf("%d", offset+60)
	case "bright-red":
		return fmt.Sprintf("%d", offset+61)
	case "bright-green":
		return fmt.Sprintf("%d", offset+62)
	case "bright-yellow":
		return fmt.Sprintf("%d", offset+63)
	case "bright-blue":
		return fmt.Sprintf("%d", offset+64)
	case "bright-magenta":
		return fmt.Sprintf("%d", offset+65)
	case "bright-cyan":
		return fmt.Sprintf("%d", offset+66)
	case "bright-white":
		return fmt.Sprintf("%d", offset+67)
	}

	return ""
}
