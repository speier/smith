package lotusui

import (
	"github.com/speier/smith/pkg/lotus/tty"
	"github.com/speier/smith/pkg/lotus/vdom"
)

// ModalButton represents a button in the modal
type ModalButton struct {
	Label    string
	Variant  string // "primary", "danger", "secondary"
	OnClick  func()
	Disabled bool
}

// Modal is a dialog overlay component
type Modal struct {
	// Component metadata
	ID string

	// State
	Open bool

	// Content
	Title   string
	Content *vdom.Element
	Buttons []ModalButton

	// Visual style
	Width  int // Modal width (default: 60)
	Height int // Modal height (default: auto)
	Border vdom.BorderStyle

	// Behavior
	CloseOnEscape  bool // Close on Escape key (default: true)
	CloseOnOutside bool // Close on click outside (not yet implemented)

	// Callbacks
	OnClose func()

	// Internal
	focusedButton int
}

// NewModal creates a new modal component
func NewModal() *Modal {
	return &Modal{
		Width:         60,
		Border:        vdom.BorderStyleRounded,
		CloseOnEscape: true,
		focusedButton: 0,
	}
}

// WithID sets the component ID
func (m *Modal) WithID(id string) *Modal {
	m.ID = id
	return m
}

// WithTitle sets the modal title
func (m *Modal) WithTitle(title string) *Modal {
	m.Title = title
	return m
}

// WithContent sets the modal content
func (m *Modal) WithContent(content *vdom.Element) *Modal {
	m.Content = content
	return m
}

// WithButtons sets the modal buttons
func (m *Modal) WithButtons(buttons []ModalButton) *Modal {
	m.Buttons = buttons
	return m
}

// WithWidth sets the modal width
func (m *Modal) WithWidth(width int) *Modal {
	m.Width = width
	return m
}

// WithCloseOnEscape sets whether Escape closes the modal
func (m *Modal) WithCloseOnEscape(enable bool) *Modal {
	m.CloseOnEscape = enable
	return m
}

// WithOnClose sets the close callback
func (m *Modal) WithOnClose(callback func()) *Modal {
	m.OnClose = callback
	return m
}

// Render returns the modal element
func (m *Modal) Render() *vdom.Element {
	if !m.Open {
		return nil // Not visible
	}

	// Build modal structure:
	// - Backdrop (dimmed overlay)
	// - Modal box (centered)
	//   - Title
	//   - Content
	//   - Buttons

	// Title section
	var titleElem *vdom.Element
	if m.Title != "" {
		titleElem = vdom.Box(vdom.Text(" "+m.Title+" ")).
			WithStyle("color", "#ffffff").
			WithStyle("font-weight", "bold")
	}

	// Content section
	contentElem := vdom.Box(m.Content).
		WithFlexGrow(1)

	// Buttons section
	var buttonsElem *vdom.Element
	if len(m.Buttons) > 0 {
		buttonElements := make([]any, len(m.Buttons))
		for i, btn := range m.Buttons {
			buttonElements[i] = m.renderButton(i, btn)
		}
		buttonsElem = vdom.HStack(buttonElements...)
	}

	// Assemble modal content
	modalParts := make([]any, 0, 3)
	if titleElem != nil {
		modalParts = append(modalParts, titleElem)
	}
	if contentElem != nil {
		modalParts = append(modalParts, contentElem)
	}
	if buttonsElem != nil {
		modalParts = append(modalParts, buttonsElem)
	}

	modalContent := vdom.VStack(modalParts...)

	// Modal box with border, background, and fixed width
	modalBox := vdom.Box(modalContent).
		WithBorderStyle(m.Border).
		WithStyle("background-color", "#1a1a1a").
		WithStyle("box-shadow", "0 4px 6px rgba(0, 0, 0, 0.3)") // Subtle shadow

	// Set width if specified
	if m.Width > 0 {
		modalBox = modalBox.WithWidth(m.Width)
	}

	// Backdrop: full-screen overlay that centers the modal
	// Uses flexbox to center both horizontally and vertically
	backdrop := vdom.Box(modalBox).
		WithStyle("background-color", "rgba(0, 0, 0, 0.5)"). // Semi-transparent dark overlay
		WithAlignItems(vdom.AlignItemsCenter).               // Center horizontally
		WithJustifyContent(vdom.JustifyContentCenter).       // Center vertically
		WithStyle("position", "absolute").                   // Overlay on top
		WithStyle("top", "0").
		WithStyle("left", "0").
		WithStyle("right", "0").
		WithStyle("bottom", "0").
		WithStyle("z-index", "1000") // On top of everything

	return backdrop
}

// renderButton renders a single button
func (m *Modal) renderButton(index int, btn ModalButton) *vdom.Element {
	// Button text with padding
	text := " " + btn.Label + " "

	elem := vdom.Box(vdom.Text(text)).
		WithBorderStyle(vdom.BorderStyleRounded)

	// Variant colors
	color := "#ffffff"
	bgColor := "#404040"
	switch btn.Variant {
	case "primary":
		color = "#ffffff"
		bgColor = "#007acc"
	case "danger":
		color = "#ffffff"
		bgColor = "#dc3545"
	case "secondary":
		color = "#d4d4d4"
		bgColor = "#6c757d"
	}

	// Apply styling
	if btn.Disabled {
		elem = elem.WithStyle("color", "#606060")
	} else if index == m.focusedButton {
		// Focused button (highlighted)
		elem = elem.WithStyle("color", "#000000").
			WithStyle("background-color", "#00ff00")
	} else {
		elem = elem.WithStyle("color", color).
			WithStyle("background-color", bgColor)
	}

	return elem
}

// Show opens the modal
func (m *Modal) Show() {
	m.Open = true
	m.focusedButton = 0
}

// Close closes the modal
func (m *Modal) Close() {
	m.Open = false
	if m.OnClose != nil {
		m.OnClose()
	}
}

// IsOpen returns whether the modal is currently open
func (m *Modal) IsOpen() bool {
	return m.Open
}

// ShouldCloseOnEscape returns whether ESC key should close the modal
func (m *Modal) ShouldCloseOnEscape() bool {
	return m.CloseOnEscape
}

// ClickButton triggers the focused button
func (m *Modal) ClickButton() {
	if m.focusedButton >= 0 && m.focusedButton < len(m.Buttons) {
		btn := m.Buttons[m.focusedButton]
		if !btn.Disabled && btn.OnClick != nil {
			btn.OnClick()
		}
	}
}

// FocusNextButton moves focus to next button
func (m *Modal) FocusNextButton() {
	if len(m.Buttons) == 0 {
		return
	}
	m.focusedButton = (m.focusedButton + 1) % len(m.Buttons)
	// Skip disabled buttons
	start := m.focusedButton
	for m.Buttons[m.focusedButton].Disabled {
		m.focusedButton = (m.focusedButton + 1) % len(m.Buttons)
		if m.focusedButton == start {
			break // All disabled
		}
	}
}

// FocusPreviousButton moves focus to previous button
func (m *Modal) FocusPreviousButton() {
	if len(m.Buttons) == 0 {
		return
	}
	m.focusedButton--
	if m.focusedButton < 0 {
		m.focusedButton = len(m.Buttons) - 1
	}
	// Skip disabled buttons
	start := m.focusedButton
	for m.Buttons[m.focusedButton].Disabled {
		m.focusedButton--
		if m.focusedButton < 0 {
			m.focusedButton = len(m.Buttons) - 1
		}
		if m.focusedButton == start {
			break // All disabled
		}
	}
}

// HandleKey processes keyboard events
func (m *Modal) HandleKey(event tty.KeyEvent) bool {
	if !m.Open {
		return false
	}

	// Escape to close
	if event.Key == tty.KeyEscape && m.CloseOnEscape {
		m.Close()
		return true
	}

	// Tab to move between buttons
	if event.Key == '\t' {
		m.FocusNextButton()
		return true
	}

	// Arrow keys to move between buttons
	switch event.Code {
	case tty.SeqLeft:
		m.FocusPreviousButton()
		return true
	case tty.SeqRight:
		m.FocusNextButton()
		return true
	}

	// Enter or Space to click focused button
	if event.Key == ' ' || event.IsEnter() {
		m.ClickButton()
		return true
	}

	return false
}

// Focusable interface implementation

// HandleKeyEvent implements Focusable interface
func (m *Modal) HandleKeyEvent(event tty.KeyEvent) bool {
	return m.HandleKey(event)
}

// IsFocusable implements Focusable interface
func (m *Modal) IsFocusable() bool {
	return m.Open
}

// IsNode implements vdom.Node interface
func (m *Modal) IsNode() {}

// GetID returns the component ID
func (m *Modal) GetID() string {
	return m.ID
}
