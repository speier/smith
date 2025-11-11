package components

import (
	"testing"

	"github.com/speier/smith/pkg/lotus/tty"
	"github.com/speier/smith/pkg/lotus/vdom"
)

func TestTabsBasic(t *testing.T) {
	tabs := NewTabs().WithTabs([]Tab{
		{Label: "Tab 1", Content: vdom.Text("Content 1")},
		{Label: "Tab 2", Content: vdom.Text("Content 2")},
		{Label: "Tab 3", Content: vdom.Text("Content 3")},
	})

	if len(tabs.Tabs) != 3 {
		t.Errorf("Expected 3 tabs, got %d", len(tabs.Tabs))
	}

	if tabs.Active != 0 {
		t.Errorf("Expected active=0, got %d", tabs.Active)
	}
}

func TestTabsSetActive(t *testing.T) {
	called := false
	var receivedIndex int

	tabs := NewTabs().
		WithTabs([]Tab{
			{Label: "Tab 1", Content: vdom.Text("Content 1")},
			{Label: "Tab 2", Content: vdom.Text("Content 2")},
		}).
		WithOnChange(func(index int) {
			called = true
			receivedIndex = index
		})

	tabs.SetActive(1)

	if !called {
		t.Error("OnChange callback not called")
	}
	if receivedIndex != 1 {
		t.Errorf("OnChange received %d, want 1", receivedIndex)
	}
	if tabs.Active != 1 {
		t.Errorf("Active tab not updated, got %d", tabs.Active)
	}
}

func TestTabsNavigation(t *testing.T) {
	tabs := NewTabs().WithTabs([]Tab{
		{Label: "Tab 1", Content: vdom.Text("Content 1")},
		{Label: "Tab 2", Content: vdom.Text("Content 2")},
		{Label: "Tab 3", Content: vdom.Text("Content 3")},
	})

	// Start at 0
	if tabs.Active != 0 {
		t.Fatalf("Expected active=0, got %d", tabs.Active)
	}

	// Next -> 1
	tabs.Next()
	if tabs.Active != 1 {
		t.Errorf("After Next, expected active=1, got %d", tabs.Active)
	}

	// Next -> 2
	tabs.Next()
	if tabs.Active != 2 {
		t.Errorf("After Next, expected active=2, got %d", tabs.Active)
	}

	// Next -> wrap to 0
	tabs.Next()
	if tabs.Active != 0 {
		t.Errorf("After Next (wrap), expected active=0, got %d", tabs.Active)
	}

	// Previous -> wrap to 2
	tabs.Previous()
	if tabs.Active != 2 {
		t.Errorf("After Previous (wrap), expected active=2, got %d", tabs.Active)
	}
}

func TestTabsDisabled(t *testing.T) {
	tabs := NewTabs().WithTabs([]Tab{
		{Label: "Tab 1", Content: vdom.Text("Content 1")},
		{Label: "Tab 2", Content: vdom.Text("Content 2"), Disabled: true},
		{Label: "Tab 3", Content: vdom.Text("Content 3")},
	})

	// Try to set active to disabled tab
	tabs.SetActive(1)
	if tabs.Active == 1 {
		t.Error("Should not allow selecting disabled tab")
	}

	// Navigation should skip disabled tabs
	tabs.SetActive(0)
	tabs.Next()
	if tabs.Active == 1 {
		t.Error("Next should skip disabled tab")
	}
	if tabs.Active != 2 {
		t.Errorf("Next should go to tab 2, got %d", tabs.Active)
	}
}

func TestTabsStatePersistence(t *testing.T) {
	tabs := NewTabs().
		WithID("test-tabs").
		WithTabs([]Tab{
			{Label: "Tab 1", Content: vdom.Text("Content 1")},
			{Label: "Tab 2", Content: vdom.Text("Content 2")},
		}).
		WithActive(1)

	// Save state
	state := tabs.SaveState()
	if active, ok := state["active"].(float64); !ok || int(active) != 1 {
		t.Errorf("SaveState returned wrong value: %v", state)
	}

	// Load state
	newTabs := NewTabs().
		WithID("test-tabs").
		WithTabs([]Tab{
			{Label: "Tab 1", Content: vdom.Text("Content 1")},
			{Label: "Tab 2", Content: vdom.Text("Content 2")},
		})

	if err := newTabs.LoadState(state); err != nil {
		t.Fatalf("LoadState failed: %v", err)
	}

	if newTabs.Active != 1 {
		t.Errorf("State not restored, got active=%d", newTabs.Active)
	}
}

func TestTabsRender(t *testing.T) {
	tabs := NewTabs().WithTabs([]Tab{
		{Label: "Tab 1", Content: vdom.Text("Content 1")},
		{Label: "Tab 2", Content: vdom.Text("Content 2")},
	})

	elem := tabs.Render()
	if elem == nil {
		t.Fatal("Render returned nil")
	}

	// Should be VStack with tab bar + content
	if len(elem.Children) != 2 {
		t.Errorf("Expected 2 children (tab bar + content), got %d", len(elem.Children))
	}
}

func TestTabsLazyRender(t *testing.T) {
	tabs := NewTabs().
		WithTabs([]Tab{
			{Label: "Tab 1", Content: vdom.Text("Content 1")},
			{Label: "Tab 2", Content: vdom.Text("Content 2")},
		}).
		WithLazyRender(true).
		WithActive(1)

	elem := tabs.Render()

	// With lazy render, only active tab content should be rendered
	// This is harder to test without inspecting the full tree
	// But we can verify it doesn't crash
	if elem == nil {
		t.Error("Lazy render returned nil")
	}
}

func TestTabsKeyboardNavigation(t *testing.T) {
	tabs := NewTabs().WithTabs([]Tab{
		{Label: "Tab 1", Content: vdom.Text("Content 1")},
		{Label: "Tab 2", Content: vdom.Text("Content 2")},
		{Label: "Tab 3", Content: vdom.Text("Content 3")},
	})

	t.Run("CtrlNumberSwitchesTabs", func(t *testing.T) {
		// Ctrl+1 (byte value 1)
		event := tty.KeyEvent{Key: 1}
		handled := tabs.HandleKey(event)
		if !handled {
			t.Error("Tabs should handle Ctrl+1")
		}
		if tabs.Active != 0 {
			t.Errorf("Expected active tab 0 after Ctrl+1, got %d", tabs.Active)
		}

		// Ctrl+2 (byte value 2)
		event = tty.KeyEvent{Key: 2}
		handled = tabs.HandleKey(event)
		if !handled {
			t.Error("Tabs should handle Ctrl+2")
		}
		if tabs.Active != 1 {
			t.Errorf("Expected active tab 1 after Ctrl+2, got %d", tabs.Active)
		}

		// Ctrl+3 (byte value 3)
		event = tty.KeyEvent{Key: 3}
		handled = tabs.HandleKey(event)
		if !handled {
			t.Error("Tabs should handle Ctrl+3")
		}
		if tabs.Active != 2 {
			t.Errorf("Expected active tab 2 after Ctrl+3, got %d", tabs.Active)
		}
	})

	t.Run("ArrowKeysNavigate", func(t *testing.T) {
		tabs := NewTabs().WithTabs([]Tab{
			{Label: "A", Content: vdom.Text("A")},
			{Label: "B", Content: vdom.Text("B")},
		})

		// Right arrow
		rightEvent := tty.KeyEvent{Key: 27, Code: tty.SeqRight}
		handled := tabs.HandleKey(rightEvent)
		if !handled {
			t.Error("Tabs should handle Right arrow")
		}
		if tabs.Active != 1 {
			t.Errorf("Expected active tab 1 after Right, got %d", tabs.Active)
		}

		// Left arrow
		leftEvent := tty.KeyEvent{Key: 27, Code: tty.SeqLeft}
		handled = tabs.HandleKey(leftEvent)
		if !handled {
			t.Error("Tabs should handle Left arrow")
		}
		if tabs.Active != 0 {
			t.Errorf("Expected active tab 0 after Left, got %d", tabs.Active)
		}
	})

	t.Run("IsFocusable", func(t *testing.T) {
		tabs := NewTabs().WithTabs([]Tab{
			{Label: "Tab 1", Content: vdom.Text("Content 1")},
		})

		// Tabs should NOT be focusable (acts as global event handler)
		if tabs.IsFocusable() {
			t.Error("Tabs should not be focusable (should return false)")
		}
	})
}
