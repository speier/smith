package lotus

import (
	"github.com/charmbracelet/glamour"
)

// MarkdownRenderer provides markdown rendering functionality
type MarkdownRenderer struct {
	renderer *glamour.TermRenderer
}

// NewMarkdownRenderer creates a new markdown renderer
func NewMarkdownRenderer(width int) (*MarkdownRenderer, error) {
	r, err := glamour.NewTermRenderer(
		glamour.WithStylePath("dark"),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return nil, err
	}

	return &MarkdownRenderer{
		renderer: r,
	}, nil
}

// Render renders markdown to ANSI-styled text
func (m *MarkdownRenderer) Render(markdown string) (string, error) {
	return m.renderer.Render(markdown)
}

// RenderMarkdown is a convenience function to render markdown with default settings
func RenderMarkdown(markdown string, width int) (string, error) {
	r, err := NewMarkdownRenderer(width)
	if err != nil {
		return markdown, err // Return original on error
	}
	
	rendered, err := r.Render(markdown)
	if err != nil {
		return markdown, err // Return original on error
	}
	
	return rendered, nil
}

// Markdown creates a text element with rendered markdown content
func Markdown(content string, width int) string {
	rendered, err := RenderMarkdown(content, width)
	if err != nil {
		// Fallback to plain text if markdown rendering fails
		return Text(content)
	}
	return Text(rendered)
}
