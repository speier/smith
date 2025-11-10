package components

import "fmt"

// Panel is a bordered container
type Panel struct {
	ID          string
	Content     string
	Title       string
	BorderStyle string // "single", "rounded", "double", "none"
	Color       string
	Flex        int
	Padding     string
}

// NewPanel creates a new panel
func NewPanel(id, content string) *Panel {
	return &Panel{
		ID:          id,
		Content:     content,
		Title:       "",
		BorderStyle: "single",
		Color:       "#ddd",
		Flex:        0,
		Padding:     "0",
	}
}

// Render generates the markup for the panel
func (p *Panel) Render() string {
	if p.ID != "" {
		return fmt.Sprintf(`<box id="%s">%s</box>`, p.ID, p.Content)
	}
	return fmt.Sprintf(`<box>%s</box>`, p.Content)
}

// GetCSS returns the CSS for panel styling
func (p *Panel) GetCSS() string {
	css := ""

	if p.BorderStyle != "none" {
		css += fmt.Sprintf(`
			#%s {
				border: 1px solid;
				border-style: %s;
				color: %s;
				padding: %s;
		`, p.ID, p.BorderStyle, p.Color, p.Padding)
	} else {
		css += fmt.Sprintf(`
			#%s {
				color: %s;
				padding: %s;
		`, p.ID, p.Color, p.Padding)
	}

	if p.Flex > 0 {
		css += fmt.Sprintf(`
				flex: %d;
		`, p.Flex)
	}

	css += `
			}
	`

	return css
}
