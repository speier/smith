package ui

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
