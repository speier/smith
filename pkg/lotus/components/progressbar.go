package components

import "fmt"

// ProgressBar displays a progress indicator
type ProgressBar struct {
	Value       float64 // 0.0 to 1.0
	Width       int
	Height      int
	Char        string
	EmptyChar   string
	Color       string
	ShowPercent bool
}

// NewProgressBar creates a new progress bar
func NewProgressBar(width int) *ProgressBar {
	return &ProgressBar{
		Value:       0,
		Width:       width,
		Height:      1,
		Char:        "█",
		EmptyChar:   "░",
		Color:       "#5af",
		ShowPercent: true,
	}
}

// SetValue sets the progress value (0.0 to 1.0)
func (p *ProgressBar) SetValue(value float64) {
	if value < 0 {
		value = 0
	}
	if value > 1 {
		value = 1
	}
	p.Value = value
}

// Render generates the visual representation of the progress bar
func (p *ProgressBar) Render() string {
	barWidth := p.Width
	if p.ShowPercent {
		barWidth -= 5 // Reserve space for "100%"
	}

	if barWidth < 1 {
		barWidth = 1
	}

	filled := int(float64(barWidth) * p.Value)
	empty := barWidth - filled

	bar := ""
	for i := 0; i < filled; i++ {
		bar += p.Char
	}
	for i := 0; i < empty; i++ {
		bar += p.EmptyChar
	}

	if p.ShowPercent {
		bar += fmt.Sprintf(" %3.0f%%", p.Value*100)
	}

	return fmt.Sprintf(`<box class="progress">%s</box>`, bar)
}

// GetCSS returns the CSS for progress bar styling
func (p *ProgressBar) GetCSS() string {
	return fmt.Sprintf(`
		.progress {
			height: %d;
			color: %s;
		}
	`, p.Height, p.Color)
}
