package components

import (
	"github.com/speier/smith/pkg/lotus/vdom"
)

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

// Render generates the Element for the menu
func (m *Menu) Render() *vdom.Element {
	menuItems := []*vdom.Element{}

	for i, item := range m.Items {
		cursor := "  "
		color := m.Color

		if i == m.Selected {
			cursor = m.Cursor
			color = m.SelectedColor
		}

		if item.Disabled {
			color = m.DisabledColor
		}

		menuItems = append(menuItems,
			vdom.Box(
				vdom.Text(cursor+item.Label),
			).WithStyle("height", "1").
				WithStyle("color", color),
		)
	}

	// Convert []*Element to []any for VStack
	children := make([]any, len(menuItems))
	for i, item := range menuItems {
		children[i] = item
	}
	return vdom.VStack(children...)
}

// IsNode implements vdom.Node interface
func (m *Menu) IsNode() {}
