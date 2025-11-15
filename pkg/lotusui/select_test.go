package lotusui

import "testing"

func TestSelectBasic(t *testing.T) {
	sel := NewSelect().WithStringOptions([]string{"Option 1", "Option 2", "Option 3"})

	if len(sel.Options) != 3 {
		t.Errorf("Expected 3 options, got %d", len(sel.Options))
	}

	if sel.Selected != 0 {
		t.Errorf("Expected selected=0, got %d", sel.Selected)
	}

	if sel.Open {
		t.Error("Select should be closed initially")
	}
}

func TestSelectSetSelected(t *testing.T) {
	called := false
	var receivedIndex int
	var receivedValue string

	sel := NewSelect().
		WithStringOptions([]string{"Option 1", "Option 2"}).
		WithOnChange(func(index int, value string) {
			called = true
			receivedIndex = index
			receivedValue = value
		})

	sel.SetSelected(1)

	if !called {
		t.Error("OnChange callback not called")
	}
	if receivedIndex != 1 || receivedValue != "Option 2" {
		t.Errorf("OnChange received (%d, %q), want (1, 'Option 2')", receivedIndex, receivedValue)
	}
	if sel.Selected != 1 {
		t.Errorf("Selected not updated, got %d", sel.Selected)
	}
}

func TestSelectToggle(t *testing.T) {
	sel := NewSelect().WithStringOptions([]string{"Option 1", "Option 2"})

	// Initially closed
	if sel.Open {
		t.Fatal("Select should be closed initially")
	}

	// Toggle open
	sel.Toggle()
	if !sel.Open {
		t.Error("Select should be open after toggle")
	}

	// Toggle closed
	sel.Toggle()
	if sel.Open {
		t.Error("Select should be closed after second toggle")
	}
}

func TestSelectHighlighting(t *testing.T) {
	sel := NewSelect().WithStringOptions([]string{"A", "B", "C"}).WithSelected(0)
	sel.Open = true

	// Initially highlighted at selected (0)
	if sel.highlightedIndex != 0 {
		t.Fatalf("Expected highlighted=0, got %d", sel.highlightedIndex)
	}

	// Next
	sel.HighlightNext()
	if sel.highlightedIndex != 1 {
		t.Errorf("After HighlightNext, expected 1, got %d", sel.highlightedIndex)
	}

	// Next
	sel.HighlightNext()
	if sel.highlightedIndex != 2 {
		t.Errorf("After HighlightNext, expected 2, got %d", sel.highlightedIndex)
	}

	// Next (wrap)
	sel.HighlightNext()
	if sel.highlightedIndex != 0 {
		t.Errorf("After HighlightNext (wrap), expected 0, got %d", sel.highlightedIndex)
	}

	// Previous
	sel.HighlightPrevious()
	if sel.highlightedIndex != 2 {
		t.Errorf("After HighlightPrevious, expected 2, got %d", sel.highlightedIndex)
	}
}

func TestSelectDisabled(t *testing.T) {
	sel := NewSelect().WithStringOptions([]string{"A", "B"}).WithDisabled(true)

	// Can't toggle when disabled
	sel.Toggle()
	if sel.Open {
		t.Error("Disabled select should not open")
	}

	// Can't change selection when disabled
	sel.SetSelected(1)
	if sel.Selected == 1 {
		t.Error("Disabled select should not allow selection change")
	}
}

func TestSelectDisabledOptions(t *testing.T) {
	sel := NewSelect().WithOptions([]SelectOption{
		{Label: "A", Value: "a"},
		{Label: "B", Value: "b", Disabled: true},
		{Label: "C", Value: "c"},
	})
	sel.Open = true
	sel.highlightedIndex = 0

	// Should skip disabled option
	sel.HighlightNext()
	if sel.highlightedIndex == 1 {
		t.Error("HighlightNext should skip disabled option")
	}
	if sel.highlightedIndex != 2 {
		t.Errorf("Expected highlighted=2, got %d", sel.highlightedIndex)
	}
}

func TestSelectStatePersistence(t *testing.T) {
	sel := NewSelect().
		WithID("test-select").
		WithStringOptions([]string{"A", "B", "C"}).
		WithSelected(2)

	// Save state
	state := sel.SaveState()
	if selected, ok := state["selected"].(float64); !ok || int(selected) != 2 {
		t.Errorf("SaveState returned wrong value: %v", state)
	}

	// Load state
	newSel := NewSelect().
		WithID("test-select").
		WithStringOptions([]string{"A", "B", "C"})

	if err := newSel.LoadState(state); err != nil {
		t.Fatalf("LoadState failed: %v", err)
	}

	if newSel.Selected != 2 {
		t.Errorf("State not restored, got selected=%d", newSel.Selected)
	}
}

func TestSelectRender(t *testing.T) {
	sel := NewSelect().WithStringOptions([]string{"A", "B"})

	// Closed state
	elem := sel.Render()
	if elem == nil {
		t.Fatal("Render returned nil")
	}

	// Open state
	sel.Open = true
	elem = sel.Render()
	if elem == nil {
		t.Fatal("Render (open) returned nil")
	}
}
