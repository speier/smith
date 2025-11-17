package lotusui

import (
	"github.com/speier/smith/pkg/lotus/context"
	"github.com/speier/smith/pkg/lotus/tty"
	"github.com/speier/smith/pkg/lotus/vdom"
)

// Checkbox is a toggleable checkbox component
type Checkbox struct {
	// Component metadata
	ID string

	// State
	Label   string
	Checked bool

	// Visual style
	Icon      CheckboxIcon
	FocusChar string // Character to show when focused (default: ">")

	// Behavior
	Disabled bool

	// Callbacks
	OnChange func(context.Context, bool)

	// Internal
	focused bool
}

// CheckboxIcon defines the visual style of the checkbox
type CheckboxIcon struct {
	Unchecked string // Default: "[ ]"
	Checked   string // Default: "[✓]"
}

// Default checkbox icons
var (
	CheckboxIconDefault = CheckboxIcon{
		Unchecked: "[ ]",
		Checked:   "[✓]",
	}
	CheckboxIconX = CheckboxIcon{
		Unchecked: "[ ]",
		Checked:   "[x]",
	}
	CheckboxIconDot = CheckboxIcon{
		Unchecked: "( )",
		Checked:   "(•)",
	}
	CheckboxIconSquare = CheckboxIcon{
		Unchecked: "☐",
		Checked:   "☑",
	}
)

// NewCheckbox creates a new checkbox
func NewCheckbox() *Checkbox {
	return &Checkbox{
		Icon:      CheckboxIconDefault,
		FocusChar: ">",
	}
}

// WithID sets the component ID
func (c *Checkbox) WithID(id string) *Checkbox {
	c.ID = id
	return c
}

// WithLabel sets the checkbox label
func (c *Checkbox) WithLabel(label string) *Checkbox {
	c.Label = label
	return c
}

// WithChecked sets the initial checked state
func (c *Checkbox) WithChecked(checked bool) *Checkbox {
	c.Checked = checked
	return c
}

// WithIcon sets the checkbox icon style
func (c *Checkbox) WithIcon(icon CheckboxIcon) *Checkbox {
	c.Icon = icon
	return c
}

// WithDisabled sets the disabled state
func (c *Checkbox) WithDisabled(disabled bool) *Checkbox {
	c.Disabled = disabled
	return c
}

// WithOnChange sets the change callback
func (c *Checkbox) WithOnChange(callback func(context.Context, bool)) *Checkbox {
	c.OnChange = callback
	return c
}

// Render returns the checkbox element
func (c *Checkbox) Render() *vdom.Element {
	// Choose icon based on state
	icon := c.Icon.Unchecked
	if c.Checked {
		icon = c.Icon.Checked
	}

	// Focus indicator
	focusIndicator := " "
	if c.focused {
		focusIndicator = c.FocusChar
	}

	// Build layout: [focus] [icon] label
	text := focusIndicator + " " + icon
	if c.Label != "" {
		text += " " + c.Label
	}

	// Apply styling
	elem := vdom.Text(text)

	if c.Disabled {
		elem = elem.WithStyle("color", "#808080") // Gray for disabled
	} else if c.focused {
		elem = elem.WithStyle("color", "#00ff00") // Green for focused
	}

	return elem
}

// Toggle toggles the checkbox state
func (c *Checkbox) Toggle() {
	if c.Disabled {
		return
	}
	c.Checked = !c.Checked
}

// SetChecked sets the checked state
func (c *Checkbox) SetChecked(checked bool) {
	if c.Disabled {
		return
	}
	if c.Checked != checked {
		c.Checked = checked
	}
}

// emitChange triggers the OnChange callback
func (c *Checkbox) emitChange(ctx context.Context) {
	if c.OnChange != nil {
		c.OnChange(ctx, c.Checked)
	}
}

// HandleKey processes keyboard events with context
func (c *Checkbox) HandleKey(ctx context.Context, event tty.KeyEvent) bool {
	if c.Disabled {
		return false
	}

	// Space key (ASCII 32) or Enter
	if event.Key == ' ' || event.Key == 13 {
		// Space or Enter: toggle
		c.Checked = !c.Checked
		c.emitChange(ctx)
		return true
	}

	return false
}

// Focusable interface implementation

// IsFocusable implements Focusable interface
func (c *Checkbox) IsFocusable() bool {
	return !c.Disabled
}

// SetFocus sets the focus state
func (c *Checkbox) SetFocus(focused bool) {
	c.focused = focused
}

// IsNode implements vdom.Node interface
func (c *Checkbox) IsNode() {}

// State persistence

// GetID returns the component ID
func (c *Checkbox) GetID() string {
	return c.ID
}

// SaveState saves the checkbox state
func (c *Checkbox) SaveState() map[string]interface{} {
	return map[string]interface{}{
		"checked": c.Checked,
	}
}

// LoadState restores the checkbox state
func (c *Checkbox) LoadState(state map[string]interface{}) error {
	if checked, ok := state["checked"].(bool); ok {
		c.Checked = checked
	}
	return nil
}
