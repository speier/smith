package lotusui

import (
	"testing"

	"github.com/speier/smith/pkg/lotus/vdom"
)

func TestModalBasic(t *testing.T) {
	modal := NewModal().
		WithTitle("Test Modal").
		WithContent(vdom.Text("Test content"))

	if modal.Open {
		t.Error("Modal should be closed initially")
	}

	if modal.Title != "Test Modal" {
		t.Errorf("Title = %q, want 'Test Modal'", modal.Title)
	}
}

func TestModalShowClose(t *testing.T) {
	closed := false
	modal := NewModal().
		WithContent(vdom.Text("Content")).
		WithOnClose(func() {
			closed = true
		})

	// Show
	modal.Show()
	if !modal.Open {
		t.Error("Modal should be open after Show()")
	}

	// Close
	modal.Close()
	if modal.Open {
		t.Error("Modal should be closed after Close()")
	}
	if !closed {
		t.Error("OnClose callback not called")
	}
}

func TestModalButtons(t *testing.T) {
	clicked := ""
	modal := NewModal().
		WithContent(vdom.Text("Content")).
		WithButtons([]ModalButton{
			{
				Label:   "Cancel",
				Variant: "secondary",
				OnClick: func() {
					clicked = "cancel"
				},
			},
			{
				Label:   "Confirm",
				Variant: "primary",
				OnClick: func() {
					clicked = "confirm"
				},
			},
		})

	modal.Show()

	// Initially focused on first button
	if modal.focusedButton != 0 {
		t.Errorf("focusedButton = %d, want 0", modal.focusedButton)
	}

	// Click first button
	modal.ClickButton()
	if clicked != "cancel" {
		t.Errorf("clicked = %q, want 'cancel'", clicked)
	}

	// Move to next button
	modal.FocusNextButton()
	if modal.focusedButton != 1 {
		t.Errorf("focusedButton = %d, want 1", modal.focusedButton)
	}

	// Click second button
	modal.ClickButton()
	if clicked != "confirm" {
		t.Errorf("clicked = %q, want 'confirm'", clicked)
	}
}

func TestModalButtonNavigation(t *testing.T) {
	modal := NewModal().
		WithContent(vdom.Text("Content")).
		WithButtons([]ModalButton{
			{Label: "Button 1"},
			{Label: "Button 2"},
			{Label: "Button 3"},
		})

	modal.Show()

	// Start at 0
	if modal.focusedButton != 0 {
		t.Fatalf("focusedButton = %d, want 0", modal.focusedButton)
	}

	// Next -> 1
	modal.FocusNextButton()
	if modal.focusedButton != 1 {
		t.Errorf("After FocusNextButton, got %d, want 1", modal.focusedButton)
	}

	// Next -> 2
	modal.FocusNextButton()
	if modal.focusedButton != 2 {
		t.Errorf("After FocusNextButton, got %d, want 2", modal.focusedButton)
	}

	// Next -> wrap to 0
	modal.FocusNextButton()
	if modal.focusedButton != 0 {
		t.Errorf("After FocusNextButton (wrap), got %d, want 0", modal.focusedButton)
	}

	// Previous -> wrap to 2
	modal.FocusPreviousButton()
	if modal.focusedButton != 2 {
		t.Errorf("After FocusPreviousButton (wrap), got %d, want 2", modal.focusedButton)
	}
}

func TestModalDisabledButton(t *testing.T) {
	clicked := false
	modal := NewModal().
		WithContent(vdom.Text("Content")).
		WithButtons([]ModalButton{
			{Label: "Button 1", Disabled: true, OnClick: func() { clicked = true }},
			{Label: "Button 2"},
		})

	modal.Show()

	// Try to click disabled button
	modal.ClickButton()
	if clicked {
		t.Error("Disabled button should not trigger OnClick")
	}

	// Navigation should skip disabled
	modal.focusedButton = 0
	modal.FocusNextButton()
	if modal.focusedButton == 0 {
		t.Error("FocusNextButton should skip disabled button")
	}
}

func TestModalRender(t *testing.T) {
	modal := NewModal().
		WithTitle("Test").
		WithContent(vdom.Text("Content"))

	// Not visible when closed
	elem := modal.Render()
	if elem != nil {
		t.Error("Closed modal should render nil")
	}

	// Visible when open
	modal.Show()
	elem = modal.Render()
	if elem == nil {
		t.Error("Open modal should render non-nil")
	}
}

func TestModalCentering(t *testing.T) {
	modal := NewModal().
		WithTitle("Centered Modal").
		WithContent(vdom.Text("This should be centered")).
		WithButtons([]ModalButton{
			{Label: "OK", Variant: "primary"},
		})

	modal.Show()
	elem := modal.Render()

	if elem == nil {
		t.Fatal("Open modal should render non-nil element")
	}

	// Verify backdrop (root element) has centering properties
	backdrop := elem
	if backdrop.Props.Styles == nil {
		t.Fatal("Backdrop should have styles")
	}

	// Check background overlay
	bgColor := backdrop.Props.Styles["background-color"]
	if bgColor != "rgba(0, 0, 0, 0.5)" {
		t.Errorf("Backdrop background = %q, want 'rgba(0, 0, 0, 0.5)'", bgColor)
	}

	// Check positioning
	if backdrop.Props.Styles["position"] != "absolute" {
		t.Error("Backdrop should have position: absolute")
	}
	if backdrop.Props.Styles["top"] != "0" {
		t.Error("Backdrop should have top: 0")
	}
	if backdrop.Props.Styles["left"] != "0" {
		t.Error("Backdrop should have left: 0")
	}
	if backdrop.Props.Styles["right"] != "0" {
		t.Error("Backdrop should have right: 0")
	}
	if backdrop.Props.Styles["bottom"] != "0" {
		t.Error("Backdrop should have bottom: 0")
	}
	if backdrop.Props.Styles["z-index"] != "1000" {
		t.Error("Backdrop should have z-index: 1000")
	}

	// Check centering using flexbox
	if backdrop.Props.Styles["align-items"] != string(vdom.AlignItemsCenter) {
		t.Errorf("Backdrop align-items = %q, want %q", backdrop.Props.Styles["align-items"], string(vdom.AlignItemsCenter))
	}
	if backdrop.Props.Styles["justify-content"] != string(vdom.JustifyContentCenter) {
		t.Errorf("Backdrop justify-content = %q, want %q", backdrop.Props.Styles["justify-content"], string(vdom.JustifyContentCenter))
	}

	// Verify modal box is a child of backdrop
	if len(backdrop.Children) != 1 {
		t.Fatalf("Backdrop should have exactly 1 child (modal box), got %d", len(backdrop.Children))
	}

	modalBox := backdrop.Children[0]
	if modalBox.Props.Styles == nil {
		t.Fatal("Modal box should have styles")
	}

	// Check shadow
	shadow := modalBox.Props.Styles["box-shadow"]
	if shadow == "" {
		t.Error("Modal box should have box-shadow for visibility")
	}

	// Check background
	modalBg := modalBox.Props.Styles["background-color"]
	if modalBg != "#1a1a1a" {
		t.Errorf("Modal background = %q, want '#1a1a1a'", modalBg)
	}

	// Check border style
	borderStyle := modalBox.Props.Styles["border-style"]
	if borderStyle != string(vdom.BorderStyleRounded) {
		t.Errorf("Modal border-style = %q, want %q", borderStyle, string(vdom.BorderStyleRounded))
	}
}

func TestModalWidth(t *testing.T) {
	// Default width
	modal1 := NewModal().WithContent(vdom.Text("Content"))
	if modal1.Width != 60 {
		t.Errorf("Default width = %d, want 60", modal1.Width)
	}

	// Custom width
	modal2 := NewModal().
		WithContent(vdom.Text("Content")).
		WithWidth(80)

	modal2.Show()
	elem := modal2.Render()

	// Get modal box (child of backdrop)
	modalBox := elem.Children[0]

	// Width should be set as a style on the modal box
	width := modalBox.Props.Styles["width"]
	if width != "80" {
		t.Errorf("Modal width style = %q, want '80'", width)
	}
}

func TestModalBackdropOverlay(t *testing.T) {
	modal := NewModal().
		WithContent(vdom.Text("Content")).
		WithTitle("Test")

	modal.Show()
	backdrop := modal.Render()

	// Backdrop should be a Box that fills the screen
	if backdrop.Type != vdom.BoxElement {
		t.Errorf("Backdrop type = %v, want BoxElement", backdrop.Type)
	}

	// Should have semi-transparent background
	bg := backdrop.Props.Styles["background-color"]
	if bg != "rgba(0, 0, 0, 0.5)" {
		t.Errorf("Backdrop overlay color = %q, want semi-transparent black", bg)
	}

	// Should be positioned to cover entire screen
	styles := backdrop.Props.Styles
	requiredPositions := map[string]string{
		"position": "absolute",
		"top":      "0",
		"left":     "0",
		"right":    "0",
		"bottom":   "0",
	}

	for prop, want := range requiredPositions {
		if got := styles[prop]; got != want {
			t.Errorf("Backdrop %s = %q, want %q", prop, got, want)
		}
	}
}
