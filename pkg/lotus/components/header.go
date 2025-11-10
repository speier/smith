package components

import "fmt"

// Header is a status bar / title bar component
type Header struct {
	Title       string
	Subtitle    string
	Height      int
	Color       string
	BorderStyle string
	Align       string // "left", "center", "right"
}

// NewHeader creates a new header
func NewHeader(title, subtitle string) *Header {
	return &Header{
		Title:       title,
		Subtitle:    subtitle,
		Height:      3,
		Color:       "#5af",
		BorderStyle: "rounded",
		Align:       "center",
	}
}

// Render generates the markup for the header
func (h *Header) Render() string {
	text := h.Title
	if h.Subtitle != "" {
		text += " - " + h.Subtitle
	}

	return fmt.Sprintf(`<box id="header">%s</box>`, text)
}

// GetCSS returns the CSS for header styling
func (h *Header) GetCSS() string {
	return fmt.Sprintf(`
		#header {
			height: %d;
			color: %s;
			border: 1px solid;
			border-style: %s;
			text-align: %s;
		}
	`, h.Height, h.Color, h.BorderStyle, h.Align)
}
