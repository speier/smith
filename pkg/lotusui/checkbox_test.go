package lotusui

import (
	"testing"

	"github.com/speier/smith/pkg/lotus/context"
	"github.com/speier/smith/pkg/lotus/tty"
)

func TestCheckboxToggle(t *testing.T) {
	cb := NewCheckbox().WithLabel("Accept terms")

	if cb.Checked {
		t.Error("Expected unchecked initially")
	}

	cb.Toggle()
	if !cb.Checked {
		t.Error("Expected checked after toggle")
	}

	cb.Toggle()
	if cb.Checked {
		t.Error("Expected unchecked after second toggle")
	}
}

func TestCheckboxDisabled(t *testing.T) {
	cb := NewCheckbox().WithDisabled(true)

	// Toggle should not work when disabled
	cb.Toggle()
	if cb.Checked {
		t.Error("Disabled checkbox should not toggle")
	}

	// SetChecked should not work when disabled
	cb.SetChecked(true)
	if cb.Checked {
		t.Error("Disabled checkbox should not change state")
	}
}

func TestCheckboxCallback(t *testing.T) {
	called := false
	var receivedValue bool

	cb := NewCheckbox().WithOnChange(func(ctx context.Context, checked bool) {
		called = true
		receivedValue = checked
	})

	// Use keyboard event (space) to toggle - this triggers callback
	cb.HandleKey(context.Context{}, tty.KeyEvent{Key: ' '})

	if !called {
		t.Error("OnChange callback not called")
	}
	if !receivedValue {
		t.Error("OnChange callback received wrong value")
	}
}

func TestCheckboxStatePersistence(t *testing.T) {
	cb := NewCheckbox().WithID("test-checkbox").WithChecked(true)

	// Save state
	state := cb.SaveState()
	if checked, ok := state["checked"].(bool); !ok || !checked {
		t.Errorf("SaveState returned wrong value: %v", state)
	}

	// Create new checkbox and load state
	newCb := NewCheckbox().WithID("test-checkbox")
	if err := newCb.LoadState(state); err != nil {
		t.Fatalf("LoadState failed: %v", err)
	}

	if !newCb.Checked {
		t.Error("State not restored correctly")
	}
}

func TestCheckboxFocusable(t *testing.T) {
	cb := NewCheckbox()
	if !cb.IsFocusable() {
		t.Error("Checkbox should be focusable")
	}

	cb.Disabled = true
	if cb.IsFocusable() {
		t.Error("Disabled checkbox should not be focusable")
	}
}
