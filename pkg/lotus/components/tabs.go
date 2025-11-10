package components

import "fmt"

// Tab represents a single tab
type Tab struct {
	Label   string
	Content string
	ID      string
}

// Tabs is a tabbed view component
type Tabs struct {
	Tabs          []Tab
	Active        int
	Width         int
	Height        int
	ActiveColor   string
	InactiveColor string
	BorderStyle   string
}

// NewTabs creates a new tabs component
func NewTabs(tabs []Tab) *Tabs {
	return &Tabs{
		Tabs:          tabs,
		Active:        0,
		Width:         80,
		Height:        20,
		ActiveColor:   "#5af",
		InactiveColor: "#666",
		BorderStyle:   "single",
	}
}

// SelectNext moves to next tab
func (t *Tabs) SelectNext() {
	t.Active++
	if t.Active >= len(t.Tabs) {
		t.Active = len(t.Tabs) - 1
	}
}

// SelectPrev moves to previous tab
func (t *Tabs) SelectPrev() {
	t.Active--
	if t.Active < 0 {
		t.Active = 0
	}
}

// SelectTab selects a tab by index
func (t *Tabs) SelectTab(index int) {
	if index >= 0 && index < len(t.Tabs) {
		t.Active = index
	}
}

// GetActiveTab returns the currently active tab
func (t *Tabs) GetActiveTab() Tab {
	if t.Active >= 0 && t.Active < len(t.Tabs) {
		return t.Tabs[t.Active]
	}
	return Tab{}
}

// Render generates the markup for the tabs
func (t *Tabs) Render() string {
	// Build tab headers
	tabHeaders := ""
	for i, tab := range t.Tabs {
		class := "tab-header"
		if i == t.Active {
			class += " active"
		}

		tabHeaders += fmt.Sprintf(`<box class="%s">%s</box>`, class, tab.Label)
	}

	// Get active content
	activeContent := ""
	if t.Active >= 0 && t.Active < len(t.Tabs) {
		activeContent = t.Tabs[t.Active].Content
	}

	markup := fmt.Sprintf(`
		<box id="tabs">
			<box id="tab-headers">%s</box>
			<box id="tab-content">%s</box>
		</box>
	`, tabHeaders, activeContent)

	return markup
}

// GetCSS returns the CSS for tabs styling
func (t *Tabs) GetCSS() string {
	return fmt.Sprintf(`
		#tabs {
			width: %d;
			height: %d;
			display: flex;
			flex-direction: column;
		}
		#tab-headers {
			height: 3;
			display: flex;
			flex-direction: row;
			border-bottom: 1px solid;
		}
		.tab-header {
			padding: 0 2;
			color: %s;
		}
		.tab-header.active {
			color: %s;
			border-bottom: 2px solid;
		}
		#tab-content {
			flex: 1;
			border: 1px solid;
			border-style: %s;
			padding: 1;
		}
	`, t.Width, t.Height, t.InactiveColor, t.ActiveColor, t.BorderStyle)
}
