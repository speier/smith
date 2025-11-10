package layout

import "github.com/charmbracelet/lipgloss"

// Style wraps lipgloss.Style for text styling
type Style struct {
	style lipgloss.Style
}

// NewStyle creates a new Style
func NewStyle() Style {
	return Style{style: lipgloss.NewStyle()}
}

// Foreground sets the foreground color (e.g., "#0f0", "#5af")
func (s Style) Foreground(color string) Style {
	s.style = s.style.Foreground(lipgloss.Color(color))
	return s
}

// Background sets the background color
func (s Style) Background(color string) Style {
	s.style = s.style.Background(lipgloss.Color(color))
	return s
}

// Bold makes the text bold
func (s Style) Bold(v bool) Style {
	s.style = s.style.Bold(v)
	return s
}

// Italic makes the text italic
func (s Style) Italic(v bool) Style {
	s.style = s.style.Italic(v)
	return s
}

// Underline underlines the text
func (s Style) Underline(v bool) Style {
	s.style = s.style.Underline(v)
	return s
}

// Render applies the style to the given text
func (s Style) Render(text string) string {
	return s.style.Render(text)
}

// Convenience functions for common colors

// Green returns green text
func Green(text string) string {
	return NewStyle().Foreground("#0f0").Render(text)
}

// Blue returns blue text
func Blue(text string) string {
	return NewStyle().Foreground("#5af").Render(text)
}

// Gray returns gray text
func Gray(text string) string {
	return NewStyle().Foreground("#888").Render(text)
}

// GrayItalic returns gray italic text
func GrayItalic(text string) string {
	return NewStyle().Foreground("#888").Italic(true).Render(text)
}
