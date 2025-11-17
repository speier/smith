package lotusui

import (
	"testing"

	"github.com/speier/smith/pkg/lotus/context"
	"github.com/speier/smith/pkg/lotus/tty"
)

func TestRadioGroupSelection(t *testing.T) {
	options := []RadioOption{
		{Label: "Option 1", Value: "opt1"},
		{Label: "Option 2", Value: "opt2"},
		{Label: "Option 3", Value: "opt3"},
	}

	group := NewRadioGroup().
		WithOptions(options).
		WithSelected("opt2")

	if group.Selected != "opt2" {
		t.Errorf("Expected selected='opt2', got %q", group.Selected)
	}

	// Verify radios are built correctly
	if len(group.radios) != 3 {
		t.Fatalf("Expected 3 radios, got %d", len(group.radios))
	}

	if !group.radios[1].Selected {
		t.Error("Second radio should be selected")
	}
	if group.radios[0].Selected || group.radios[2].Selected {
		t.Error("Other radios should not be selected")
	}
}

func TestRadioGroupChangeSelection(t *testing.T) {
	options := []RadioOption{
		{Label: "Option 1", Value: "opt1"},
		{Label: "Option 2", Value: "opt2"},
	}

	called := false
	var receivedValue string

	group := NewRadioGroup().
		WithOptions(options).
		WithSelected("opt1").
		WithOnChange(func(ctx context.Context, value string) {
			called = true
			receivedValue = value
		})

	// Change selection
	group.selectValue(context.Context{}, "opt2")

	if !called {
		t.Error("OnChange callback not called")
	}
	if receivedValue != "opt2" {
		t.Errorf("OnChange callback received %q, want 'opt2'", receivedValue)
	}
	if group.Selected != "opt2" {
		t.Errorf("Selected not updated, got %q", group.Selected)
	}
}

func TestRadioGroupStatePersistence(t *testing.T) {
	options := []RadioOption{
		{Label: "Option 1", Value: "opt1"},
		{Label: "Option 2", Value: "opt2"},
	}

	group := NewRadioGroup().
		WithID("test-radio").
		WithOptions(options).
		WithSelected("opt2")

	// Save state
	state := group.SaveState()
	if selected, ok := state["selected"].(string); !ok || selected != "opt2" {
		t.Errorf("SaveState returned wrong value: %v", state)
	}

	// Create new group and load state
	newGroup := NewRadioGroup().
		WithID("test-radio").
		WithOptions(options)

	if err := newGroup.LoadState(state); err != nil {
		t.Fatalf("LoadState failed: %v", err)
	}

	if newGroup.Selected != "opt2" {
		t.Errorf("State not restored correctly, got %q", newGroup.Selected)
	}
	if !newGroup.radios[1].Selected {
		t.Error("Second radio should be selected after restore")
	}
}

func TestRadioGroupDirections(t *testing.T) {
	options := []RadioOption{
		{Label: "A", Value: "a"},
		{Label: "B", Value: "b"},
	}

	// Test vertical (default)
	vGroup := NewRadioGroup().WithOptions(options)
	if vGroup.Direction != "vertical" {
		t.Errorf("Default direction should be 'vertical', got %q", vGroup.Direction)
	}

	// Test horizontal
	hGroup := NewRadioGroup().
		WithOptions(options).
		WithDirection("horizontal")
	if hGroup.Direction != "horizontal" {
		t.Errorf("Expected 'horizontal', got %q", hGroup.Direction)
	}
}

func TestRadioGroupKeyboardNavigation(t *testing.T) {
	options := []RadioOption{
		{Label: "Option 1", Value: "opt1"},
		{Label: "Option 2", Value: "opt2"},
		{Label: "Option 3", Value: "opt3"},
	}

	t.Run("VerticalNavigation", func(t *testing.T) {
		group := NewRadioGroup().
			WithOptions(options).
			WithDirection("vertical")

		// Initially first option should be focused
		if group.focusedIndex != 0 {
			t.Errorf("Expected focusedIndex 0, got %d", group.focusedIndex)
		}

		// Press Down arrow
		downEvent := tty.KeyEvent{Key: 27, Code: tty.SeqDown}
		handled := group.HandleKey(downEvent)
		if !handled {
			t.Error("RadioGroup should handle Down arrow")
		}

		if group.focusedIndex != 1 {
			t.Errorf("Expected focusedIndex 1 after Down, got %d", group.focusedIndex)
		}

		// Press Down again
		group.HandleKey(downEvent)
		if group.focusedIndex != 2 {
			t.Errorf("Expected focusedIndex 2 after second Down, got %d", group.focusedIndex)
		}

		// Press Down again - should wrap to 0
		group.HandleKey(downEvent)
		if group.focusedIndex != 0 {
			t.Errorf("Expected focusedIndex 0 after wrap, got %d", group.focusedIndex)
		}

		// Press Up arrow
		upEvent := tty.KeyEvent{Key: 27, Code: tty.SeqUp}
		handled = group.HandleKey(upEvent)
		if !handled {
			t.Error("RadioGroup should handle Up arrow")
		}

		if group.focusedIndex != 2 {
			t.Errorf("Expected focusedIndex 2 after Up (wrap), got %d", group.focusedIndex)
		}
	})

	t.Run("SpaceSelectsOption", func(t *testing.T) {
		group := NewRadioGroup().
			WithOptions(options).
			WithDirection("vertical")

		selectedValue := ""
		group.OnChange = func(ctx context.Context, value string) {
			selectedValue = value
		}

		// Focus second option
		group.focusedIndex = 1

		// Press Space
		spaceEvent := tty.KeyEvent{Key: ' ', Char: " "}
		handled := group.HandleKey(spaceEvent)
		if !handled {
			t.Error("RadioGroup should handle Space key")
		}

		if selectedValue != "opt2" {
			t.Errorf("Expected selected value 'opt2', got %q", selectedValue)
		}

		if group.Selected != "opt2" {
			t.Errorf("Expected group.Selected='opt2', got %q", group.Selected)
		}
	})

	t.Run("IsFocusable", func(t *testing.T) {
		group := NewRadioGroup().WithOptions(options)
		if !group.IsFocusable() {
			t.Error("RadioGroup with options should be focusable")
		}

		emptyGroup := NewRadioGroup()
		if emptyGroup.IsFocusable() {
			t.Error("Empty RadioGroup should not be focusable")
		}
	})
}
