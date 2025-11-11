package components

import (
	"github.com/speier/smith/pkg/lotus/tty"
	"github.com/speier/smith/pkg/lotus/vdom"
)

// Radio is a single radio button in a radio group
type Radio struct {
	// Component metadata
	ID string

	// State
	Label    string
	Value    string // Unique value for this option
	Selected bool   // Whether this radio is selected

	// Visual style
	Icon      RadioIcon
	FocusChar string // Character to show when focused (default: ">")

	// Behavior
	Disabled bool

	// Callbacks
	OnSelect func(string) // Called with this radio's value when selected

	// Internal
	focused bool
}

// RadioIcon defines the visual style of the radio button
type RadioIcon struct {
	Unselected string // Default: "( )"
	Selected   string // Default: "(•)"
}

// Default radio icons
var (
	RadioIconDefault = RadioIcon{
		Unselected: "( )",
		Selected:   "(•)",
	}
	RadioIconSquare = RadioIcon{
		Unselected: "[ ]",
		Selected:   "[•]",
	}
	RadioIconArrow = RadioIcon{
		Unselected: "  ",
		Selected:   "▶ ",
	}
)

// RadioGroup manages a group of radio buttons
type RadioGroup struct {
	// Component metadata
	ID string

	// Options
	Options []RadioOption // List of radio options

	// State
	Selected string // Currently selected value

	// Visual style
	Icon      RadioIcon
	Direction string // "vertical" or "horizontal"
	Spacing   int    // Space between radio buttons

	// Callbacks
	OnChange func(string)

	// Internal
	focusedIndex int
	radios       []*Radio // Generated radio components
}

// RadioOption defines a single radio option
type RadioOption struct {
	Label string
	Value string
}

// NewRadio creates a new radio button
func NewRadio() *Radio {
	return &Radio{
		Icon:      RadioIconDefault,
		FocusChar: ">",
	}
}

// WithID sets the component ID
func (r *Radio) WithID(id string) *Radio {
	r.ID = id
	return r
}

// WithLabel sets the radio label
func (r *Radio) WithLabel(label string) *Radio {
	r.Label = label
	return r
}

// WithValue sets the radio value
func (r *Radio) WithValue(value string) *Radio {
	r.Value = value
	return r
}

// WithSelected sets the selected state
func (r *Radio) WithSelected(selected bool) *Radio {
	r.Selected = selected
	return r
}

// WithIcon sets the radio icon style
func (r *Radio) WithIcon(icon RadioIcon) *Radio {
	r.Icon = icon
	return r
}

// WithDisabled sets the disabled state
func (r *Radio) WithDisabled(disabled bool) *Radio {
	r.Disabled = disabled
	return r
}

// WithOnSelect sets the select callback
func (r *Radio) WithOnSelect(callback func(string)) *Radio {
	r.OnSelect = callback
	return r
}

// Render returns the radio element
func (r *Radio) Render() *vdom.Element {
	// Choose icon based on state
	icon := r.Icon.Unselected
	if r.Selected {
		icon = r.Icon.Selected
	}

	// Focus indicator
	focusIndicator := " "
	if r.focused {
		focusIndicator = r.FocusChar
	}

	// Build layout: [focus] [icon] label
	text := focusIndicator + " " + icon
	if r.Label != "" {
		text += " " + r.Label
	}

	// Apply styling
	elem := vdom.Text(text)

	if r.Disabled {
		elem = elem.WithStyle("color", "#808080") // Gray for disabled
	} else if r.focused {
		elem = elem.WithStyle("color", "#00ff00") // Green for focused
	} else if r.Selected {
		elem = elem.WithStyle("color", "#ffffff") // White for selected
	}

	return elem
}

// Select marks this radio as selected
func (r *Radio) Select() {
	if r.Disabled {
		return
	}
	r.Selected = true
	r.emitSelect()
}

// emitSelect triggers the OnSelect callback
func (r *Radio) emitSelect() {
	if r.OnSelect != nil {
		r.OnSelect(r.Value)
	}
}

// HandleKey processes keyboard events
func (r *Radio) HandleKey(event tty.KeyEvent) bool {
	if r.Disabled {
		return false
	}

	// Space key (ASCII 32) or Enter
	if event.Key == ' ' || event.IsEnter() {
		r.Select()
		return true
	}

	return false
}

// Focusable interface implementation

// HandleKeyEvent implements Focusable interface
func (r *Radio) HandleKeyEvent(event tty.KeyEvent) bool {
	return r.HandleKey(event)
}

// IsFocusable implements Focusable interface
func (r *Radio) IsFocusable() bool {
	return !r.Disabled
}

// SetFocus sets the focus state
func (r *Radio) SetFocus(focused bool) {
	r.focused = focused
}

// IsNode implements vdom.Node interface
func (r *Radio) IsNode() {}

// --- RadioGroup ---

// NewRadioGroup creates a new radio group
func NewRadioGroup() *RadioGroup {
	return &RadioGroup{
		Icon:      RadioIconDefault,
		Direction: "vertical",
		Spacing:   0,
	}
}

// WithID sets the group ID
func (g *RadioGroup) WithID(id string) *RadioGroup {
	g.ID = id
	return g
}

// WithOptions sets the radio options
func (g *RadioGroup) WithOptions(options []RadioOption) *RadioGroup {
	g.Options = options
	g.buildRadios()
	return g
}

// WithSelected sets the initially selected value
func (g *RadioGroup) WithSelected(value string) *RadioGroup {
	g.Selected = value
	g.updateSelection()
	return g
}

// WithIcon sets the icon style for all radios
func (g *RadioGroup) WithIcon(icon RadioIcon) *RadioGroup {
	g.Icon = icon
	g.buildRadios()
	return g
}

// WithDirection sets the layout direction
func (g *RadioGroup) WithDirection(direction string) *RadioGroup {
	g.Direction = direction
	return g
}

// WithOnChange sets the change callback
func (g *RadioGroup) WithOnChange(callback func(string)) *RadioGroup {
	g.OnChange = callback
	return g
}

// buildRadios creates Radio components from options
func (g *RadioGroup) buildRadios() {
	g.radios = make([]*Radio, len(g.Options))
	for i, opt := range g.Options {
		radio := NewRadio().
			WithLabel(opt.Label).
			WithValue(opt.Value).
			WithIcon(g.Icon).
			WithSelected(opt.Value == g.Selected).
			WithOnSelect(func(value string) {
				g.selectValue(value)
			})
		g.radios[i] = radio
	}
}

// selectValue handles selection of a radio
func (g *RadioGroup) selectValue(value string) {
	if g.Selected == value {
		return
	}

	g.Selected = value
	g.updateSelection()

	if g.OnChange != nil {
		g.OnChange(value)
	}
}

// updateSelection updates which radio is selected
func (g *RadioGroup) updateSelection() {
	for _, radio := range g.radios {
		radio.Selected = (radio.Value == g.Selected)
	}
}

// Render returns the radio group element
func (g *RadioGroup) Render() *vdom.Element {
	if len(g.radios) == 0 {
		g.buildRadios()
	}

	// Build children elements
	children := make([]*vdom.Element, len(g.radios))
	for i, radio := range g.radios {
		children[i] = radio.Render()
	}

	// Convert to []any for variadic function
	childrenAny := make([]any, len(children))
	for i, child := range children {
		childrenAny[i] = child
	}

	// Layout based on direction
	if g.Direction == "horizontal" {
		return vdom.HStack(childrenAny...)
	}
	return vdom.VStack(childrenAny...)
}

// HandleKey processes keyboard events for navigation
func (g *RadioGroup) HandleKey(event tty.KeyEvent) bool {
	// Check arrow keys via escape sequences
	switch event.Code {
	case tty.SeqUp:
		if g.Direction == "vertical" {
			g.focusPrevious()
			return true
		}
	case tty.SeqDown:
		if g.Direction == "vertical" {
			g.focusNext()
			return true
		}
	case tty.SeqLeft:
		if g.Direction == "horizontal" {
			g.focusPrevious()
			return true
		}
	case tty.SeqRight:
		if g.Direction == "horizontal" {
			g.focusNext()
			return true
		}
	}

	// Space key (ASCII 32) or Enter
	if event.Key == ' ' || event.IsEnter() {
		if g.focusedIndex >= 0 && g.focusedIndex < len(g.radios) {
			g.radios[g.focusedIndex].Select()
			return true
		}
	}

	return false
}

// focusNext moves focus to next radio
func (g *RadioGroup) focusNext() {
	if len(g.radios) == 0 {
		return
	}
	g.radios[g.focusedIndex].SetFocus(false)
	g.focusedIndex = (g.focusedIndex + 1) % len(g.radios)
	g.radios[g.focusedIndex].SetFocus(true)
}

// focusPrevious moves focus to previous radio
func (g *RadioGroup) focusPrevious() {
	if len(g.radios) == 0 {
		return
	}
	g.radios[g.focusedIndex].SetFocus(false)
	g.focusedIndex--
	if g.focusedIndex < 0 {
		g.focusedIndex = len(g.radios) - 1
	}
	g.radios[g.focusedIndex].SetFocus(true)
}

// Focusable interface implementation

// HandleKeyEvent implements Focusable interface
func (g *RadioGroup) HandleKeyEvent(event tty.KeyEvent) bool {
	return g.HandleKey(event)
}

// IsFocusable implements Focusable interface
func (g *RadioGroup) IsFocusable() bool {
	return len(g.radios) > 0
}

// GetCursorOffset implements Focusable interface
// RadioGroup doesn't show a cursor, so always return 0
func (g *RadioGroup) GetCursorOffset() int {
	return 0
}

// IsNode implements vdom.Node interface
func (g *RadioGroup) IsNode() {}

// State persistence

// GetID returns the group ID
func (g *RadioGroup) GetID() string {
	return g.ID
}

// SaveState saves the selected value
func (g *RadioGroup) SaveState() map[string]interface{} {
	return map[string]interface{}{
		"selected": g.Selected,
	}
}

// LoadState restores the selected value
func (g *RadioGroup) LoadState(state map[string]interface{}) error {
	if selected, ok := state["selected"].(string); ok {
		g.Selected = selected
		g.updateSelection()
	}
	return nil
}
