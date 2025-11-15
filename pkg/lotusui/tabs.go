package lotusui

import (
	"fmt"
	"strings"

	"github.com/speier/smith/pkg/lotus/tty"
	"github.com/speier/smith/pkg/lotus/vdom"
)

// Tab represents a single tab with label and content
type Tab struct {
	Label    string
	Content  any // Accepts *vdom.Element, vdom.Component, or vdom.Node
	Disabled bool
}

// Tabs is a tabbed navigation component
type Tabs struct {
	// Component metadata
	ID string

	// Tabs
	Tabs []Tab

	// State
	Active int // Index of active tab

	// Visual style
	TabBarStyle   TabBarStyle
	ActiveColor   string // Color for active tab (default: "#00ff00")
	InactiveColor string // Color for inactive tabs (default: "#808080")

	// Behavior
	LazyRender bool // Only render active tab content (performance optimization)

	// Callbacks
	OnChange func(int) // Called when active tab changes
}

// TabBarStyle defines how the tab bar is rendered
type TabBarStyle string

const (
	TabBarStyleLine     TabBarStyle = "line"     // ─────────
	TabBarStyleBrackets TabBarStyle = "brackets" // [ Tab1 ][ Tab2 ]
	TabBarStylePipes    TabBarStyle = "pipes"    // | Tab1 | Tab2 |
	TabBarStyleRounded  TabBarStyle = "rounded"  // ╭─Tab1─╮╭─Tab2─╮
)

// NewTabs creates a new tabs component
func NewTabs() *Tabs {
	return &Tabs{
		TabBarStyle:   TabBarStyleLine,
		ActiveColor:   "#00ff00",
		InactiveColor: "#808080",
		LazyRender:    true,
	}
}

// WithID sets the component ID
func (t *Tabs) WithID(id string) *Tabs {
	t.ID = id
	return t
}

// WithTabs sets the tabs
func (t *Tabs) WithTabs(tabs []Tab) *Tabs {
	t.Tabs = tabs
	return t
}

// WithActive sets the initially active tab
func (t *Tabs) WithActive(index int) *Tabs {
	if index >= 0 && index < len(t.Tabs) {
		t.Active = index
	}
	return t
}

// WithTabBarStyle sets the tab bar visual style
func (t *Tabs) WithTabBarStyle(style TabBarStyle) *Tabs {
	t.TabBarStyle = style
	return t
}

// WithActiveColor sets the active tab color
func (t *Tabs) WithActiveColor(color string) *Tabs {
	t.ActiveColor = color
	return t
}

// WithLazyRender enables/disables lazy rendering
func (t *Tabs) WithLazyRender(lazy bool) *Tabs {
	t.LazyRender = lazy
	return t
}

// WithOnChange sets the change callback
func (t *Tabs) WithOnChange(callback func(int)) *Tabs {
	t.OnChange = callback
	return t
}

// Render returns the tabs element
func (t *Tabs) Render() *vdom.Element {
	if len(t.Tabs) == 0 {
		return vdom.Text("No tabs")
	}

	// Ensure active index is valid
	if t.Active < 0 || t.Active >= len(t.Tabs) {
		t.Active = 0
	}

	// Build tab bar
	tabBar := t.renderTabBar()

	// Build content - convert tab content to elements
	var content *vdom.Element
	if t.LazyRender {
		// Only render active tab
		content = t.convertContent(t.Tabs[t.Active].Content)
	} else {
		// Render all tabs, show/hide via style
		children := make([]any, len(t.Tabs))
		for i, tab := range t.Tabs {
			children[i] = t.convertContent(tab.Content)
			if i != t.Active {
				// Hide inactive tabs via wrapper
				if elem, ok := children[i].(*vdom.Element); ok {
					children[i] = elem.WithStyle("display", "none")
				}
			}
		}
		content = vdom.VStack(children...)
	}

	// Layout: tab bar (auto height) + content (flex-grow: 1)
	return vdom.VStack(
		tabBar,
		vdom.Box(content).WithFlexGrow(1),
	)
}

// convertContent converts tab content (any type) to an Element
// Properly handles vdom.Node (Component or Element) to preserve component references
func (t *Tabs) convertContent(content any) *vdom.Element {
	switch v := content.(type) {
	case *vdom.Element:
		return v
	case vdom.Node:
		return vdom.ToElement(v)
	case string:
		return vdom.Text(v)
	default:
		return vdom.Text(fmt.Sprintf("%v", v))
	}
}

// renderTabBar creates the tab bar UI
func (t *Tabs) renderTabBar() *vdom.Element {
	var parts []string

	for i, tab := range t.Tabs {
		var label string
		isActive := i == t.Active

		switch t.TabBarStyle {
		case TabBarStyleBrackets:
			if isActive {
				label = fmt.Sprintf("[ %s ]", tab.Label)
			} else {
				label = fmt.Sprintf("  %s  ", tab.Label)
			}

		case TabBarStylePipes:
			label = fmt.Sprintf("| %s ", tab.Label)

		case TabBarStyleRounded:
			if isActive {
				label = fmt.Sprintf("╭─ %s ─╮", tab.Label)
			} else {
				label = fmt.Sprintf("   %s   ", tab.Label)
			}

		case TabBarStyleLine:
			fallthrough
		default:
			// Simple underline for active
			label = fmt.Sprintf(" %s ", tab.Label)
		}

		parts = append(parts, label)
	}

	tabBarText := strings.Join(parts, " ")

	// Add colored sections for active/inactive
	elem := vdom.Text(tabBarText)

	if t.Active >= 0 && t.Active < len(t.Tabs) {
		// Color active tab
		elem = elem.WithStyle("color", t.ActiveColor)
	}

	// Add underline for active tab (line style)
	if t.TabBarStyle == TabBarStyleLine {
		underline := t.renderUnderline()
		return vdom.VStack(
			elem,
			vdom.Text(underline),
		)
	}

	return elem
}

// renderUnderline creates the underline for active tab
func (t *Tabs) renderUnderline() string {
	if len(t.Tabs) == 0 {
		return ""
	}

	var parts []string
	for i, tab := range t.Tabs {
		labelLen := len(tab.Label) + 2 // +2 for spaces
		if i == t.Active {
			parts = append(parts, strings.Repeat("─", labelLen))
		} else {
			parts = append(parts, strings.Repeat(" ", labelLen))
		}
	}

	return " " + strings.Join(parts, " ")
}

// SetActive changes the active tab
func (t *Tabs) SetActive(index int) {
	if index < 0 || index >= len(t.Tabs) {
		return
	}
	if t.Tabs[index].Disabled {
		return
	}
	if t.Active != index {
		t.Active = index
		t.emitChange()
	}
}

// Next switches to the next tab
func (t *Tabs) Next() {
	next := t.Active + 1
	if next >= len(t.Tabs) {
		next = 0 // Wrap around
	}
	// Skip disabled tabs
	for next != t.Active && t.Tabs[next].Disabled {
		next++
		if next >= len(t.Tabs) {
			next = 0
		}
	}
	t.SetActive(next)
}

// Previous switches to the previous tab
func (t *Tabs) Previous() {
	prev := t.Active - 1
	if prev < 0 {
		prev = len(t.Tabs) - 1 // Wrap around
	}
	// Skip disabled tabs
	for prev != t.Active && t.Tabs[prev].Disabled {
		prev--
		if prev < 0 {
			prev = len(t.Tabs) - 1
		}
	}
	t.SetActive(prev)
}

// emitChange triggers the OnChange callback
func (t *Tabs) emitChange() {
	if t.OnChange != nil {
		t.OnChange(t.Active)
	}
}

// HandleKey processes keyboard events
func (t *Tabs) HandleKey(event tty.KeyEvent) bool {
	// Arrow keys for tab navigation (Left/Right at top level)
	switch event.Code {
	case tty.SeqLeft:
		t.Previous()
		return true
	case tty.SeqRight:
		t.Next()
		return true
	}

	// Ctrl+] for next tab
	if event.Key == 29 { // Ctrl+]
		t.Next()
		return true
	}

	// Ctrl+Number (1-9) for direct tab selection
	// Ctrl+1 through Ctrl+9 in terminals send byte values 1-9
	if event.Key >= 1 && event.Key <= 9 {
		index := int(event.Key - 1)
		if index < len(t.Tabs) {
			t.SetActive(index)
			return true
		}
	}

	return false
}

// Focusable interface implementation

// HandleKeyEvent implements Focusable interface
func (t *Tabs) HandleKeyEvent(event tty.KeyEvent) bool {
	return t.HandleKey(event)
}

// IsFocusable implements Focusable interface
// Tabs itself is not focusable - its children are focusable
// Tabs acts as a global event handler for tab switching shortcuts
func (t *Tabs) IsFocusable() bool {
	return false
}

// IsNode implements vdom.Node interface
func (t *Tabs) IsNode() {}

// State persistence

// GetID returns the component ID
func (t *Tabs) GetID() string {
	return t.ID
}

// SaveState saves the active tab index
func (t *Tabs) SaveState() map[string]interface{} {
	return map[string]interface{}{
		"active": float64(t.Active),
	}
}

// LoadState restores the active tab index
func (t *Tabs) LoadState(state map[string]interface{}) error {
	if active, ok := state["active"].(float64); ok {
		t.Active = int(active)
	}
	return nil
}
