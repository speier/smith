package components

import "fmt"

// MenuItem represents a single menu option
type MenuItem struct {
	Label    string
	Value    string
	Disabled bool
}

// Menu is a list of selectable options
type Menu struct {
	Items         []MenuItem
	Selected      int
	Width         int
	Height        int
	Cursor        string
	Color         string
	SelectedColor string
	DisabledColor string
}

// NewMenu creates a new menu
func NewMenu(items []MenuItem) *Menu {
	return &Menu{
		Items:         items,
		Selected:      0,
		Width:         20,
		Height:        10,
		Cursor:        "> ",
		Color:         "#ddd",
		SelectedColor: "#5af",
		DisabledColor: "#666",
	}
}

// SelectNext moves selection down
func (m *Menu) SelectNext() {
	m.Selected++
	if m.Selected >= len(m.Items) {
		m.Selected = len(m.Items) - 1
	}
	// Skip disabled items
	for m.Selected < len(m.Items)-1 && m.Items[m.Selected].Disabled {
		m.Selected++
	}
}

// SelectPrev moves selection up
func (m *Menu) SelectPrev() {
	m.Selected--
	if m.Selected < 0 {
		m.Selected = 0
	}
	// Skip disabled items
	for m.Selected > 0 && m.Items[m.Selected].Disabled {
		m.Selected--
	}
}

// GetSelectedValue returns the currently selected item's value
func (m *Menu) GetSelectedValue() string {
	if m.Selected >= 0 && m.Selected < len(m.Items) {
		return m.Items[m.Selected].Value
	}
	return ""
}

// Render generates the markup for the menu
func (m *Menu) Render() string {
	menuBoxes := ""

	for i, item := range m.Items {
		cursor := "  "
		class := "menu-item"

		if i == m.Selected {
			cursor = m.Cursor
			class += " selected"
		}

		if item.Disabled {
			class += " disabled"
		}

		menuBoxes += fmt.Sprintf(`<box class="%s">%s%s</box>`, class, cursor, item.Label)
	}

	return menuBoxes
}

// GetCSS returns the CSS for menu styling
func (m *Menu) GetCSS() string {
	return fmt.Sprintf(`
		.menu-item {
			height: 1;
			color: %s;
		}
		.menu-item.selected {
			color: %s;
		}
		.menu-item.disabled {
			color: %s;
		}
	`, m.Color, m.SelectedColor, m.DisabledColor)
}
