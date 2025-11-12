package runtime

import (
	"testing"

	"github.com/speier/smith/pkg/lotus/primitives"
	"github.com/speier/smith/pkg/lotus/vdom"
)

// TestComponentReferencePreservation tests that components maintain their reference
// when converted to elements, enabling event routing to work
func TestComponentReferencePreservation(t *testing.T) {
	// Create a component
	input := primitives.NewInput()

	// Convert to element (this is what Box() does internally)
	element := vdom.ToElement(input)

	if element == nil {
		t.Fatal("ToElement returned nil")
	}

	// Verify component reference is preserved
	if element.Component == nil {
		t.Error("Component reference was lost during ToElement conversion")
	}

	if element.Component != input {
		t.Error("Component reference doesn't match original component")
	}

	// Verify type is set correctly
	if element.Type != vdom.ComponentElement {
		t.Errorf("Expected ComponentElement type, got %v", element.Type)
	}

	// Verify we can find focusable component in tree
	if focusable, ok := element.Component.(Focusable); !ok {
		t.Error("Component should be Focusable")
	} else if !focusable.IsFocusable() {
		t.Error("TextInput should be focusable")
	}
}

// TestEventRoutingInTree tests that handleEventInTree can find components
func TestEventRoutingInTree(t *testing.T) {
	// Create a simple app with a text input
	type TestApp struct {
		input *primitives.Input
	}

	app := &TestApp{
		input: primitives.NewInput(),
	}

	// Render method that boxes the input
	renderFunc := func() *vdom.Element {
		return vdom.VStack(
			vdom.Box(vdom.Text("Header")),
			vdom.Box(app.input), // Component should be preserved here
		)
	}

	element := renderFunc()

	// Traverse tree looking for the component
	var foundComponent vdom.Component
	var traverse func(*vdom.Element)
	traverse = func(e *vdom.Element) {
		if e == nil {
			return
		}
		if e.Component != nil {
			foundComponent = e.Component
		}
		for _, child := range e.Children {
			traverse(child)
		}
	}

	traverse(element)

	if foundComponent == nil {
		t.Error("Component not found in element tree")
	}

	if foundComponent != app.input {
		t.Error("Found component doesn't match original input")
	}
}

// TestFocusManagerAndGlobalEventRouting is commented out because it requires
// ui components (Tabs) which are now in pkg/lotus-ui
// TODO: Move this test to pkg/lotus-ui or create mock components
/*
func TestFocusManagerAndGlobalEventRouting(t *testing.T) {
	// Create components
	nameInput := primitives.NewInput().WithPlaceholder("Name")
	emailInput := primitives.NewInput().WithPlaceholder("Email")

	// Track which tab is active
	activeTab := 0
	tabs := ui.NewTabs().WithTabs([]ui.Tab{
		{Label: "Form", Content: vdom.VStack(
			vdom.Box(nameInput),
			vdom.Box(emailInput),
		)},
		{Label: "Other", Content: vdom.Text("Other content")},
	}).WithActive(0)

	tabs.OnChange = func(index int) {
		activeTab = index
	}

	// Create app
	app := &struct {
		tabs *ui.Tabs
	}{
		tabs: tabs,
	}

	renderFunc := func() *vdom.Element {
		return vdom.Box(app.tabs)
	}

	// Test 1: Focus manager should find both inputs
	t.Run("CollectsFocusables", func(t *testing.T) {
		fm := newFocusManager()
		element := renderFunc()
		fm.rebuild(element)

		if len(fm.focusables) != 2 {
			t.Errorf("Expected 2 focusable components, got %d", len(fm.focusables))
		}

		// First focusable should be focused
		if !nameInput.Focused {
			t.Error("First input should be focused after rebuild")
		}
		if emailInput.Focused {
			t.Error("Second input should not be focused after rebuild")
		}
	})

	// Test 2: Tab key should cycle focus
	t.Run("TabCyclesFocus", func(t *testing.T) {
		fm := newFocusManager()
		element := renderFunc()
		fm.rebuild(element)

		// Initially first input is focused
		if fm.focusIndex != 0 {
			t.Errorf("Expected focusIndex 0, got %d", fm.focusIndex)
		}

		// Press Tab
		fm.next()

		if fm.focusIndex != 1 {
			t.Errorf("Expected focusIndex 1 after Tab, got %d", fm.focusIndex)
		}

		// Should have updated focus states
		if nameInput.Focused {
			t.Error("First input should not be focused after Tab")
		}
		if !emailInput.Focused {
			t.Error("Second input should be focused after Tab")
		}

		// Press Tab again - should wrap to first
		fm.next()
		if fm.focusIndex != 0 {
			t.Errorf("Expected focusIndex 0 after second Tab, got %d", fm.focusIndex)
		}
	})

	// Test 3: Global handler (Tabs) should receive Ctrl+Number even when input is focused
	t.Run("GlobalHandlerReceivesCtrlNumber", func(t *testing.T) {
		fm := newFocusManager()
		element := renderFunc()
		fm.rebuild(element)

		// Ensure first input is focused
		if fm.getFocused() != nameInput {
			t.Fatal("First input should be focused")
		}

		// Simulate Ctrl+2 event (should switch to second tab)
		event := tty.KeyEvent{Key: 2} // Ctrl+2

		// First, focused component gets a chance
		handled := nameInput.HandleKeyEvent(event)
		// TextInput doesn't handle Ctrl+2, so it returns false
		if handled {
			t.Error("TextInput should not handle Ctrl+2")
		}

		// Verify event is not printable
		if event.IsPrintable() {
			t.Error("Ctrl+2 should not be printable")
		}

		// Then, global handlers get a chance
		handled = handleEventInTreeGlobal(element, event, fm)
		if !handled {
			t.Error("Tabs should handle Ctrl+2 as global handler")
		}

		// Verify tab switched
		if activeTab != 1 {
			t.Errorf("Expected active tab 1 after Ctrl+2, got %d", activeTab)
		}
	})

	// Test 4: Tabs.IsFocusable should return false
	t.Run("TabsIsNotFocusable", func(t *testing.T) {
		if tabs.IsFocusable() {
			t.Error("Tabs should not be focusable (IsFocusable should return false)")
		}
	})

	// Test 5: Regular typing should go to focused input
	t.Run("FocusedInputReceivesTyping", func(t *testing.T) {
		fm := newFocusManager()
		element := renderFunc()
		fm.rebuild(element)

		// Type 'a' in first input
		event := tty.KeyEvent{Key: 'a', Char: "a"}
		handled := nameInput.HandleKeyEvent(event)
		if !handled {
			t.Error("Focused input should handle regular key")
		}

		if nameInput.Value != "a" {
			t.Errorf("Expected input value 'a', got '%s'", nameInput.Value)
		}
	})
}
*/

// TestFocusWithMixedComponents is commented out because it requires
// ui components (RadioGroup) which are now in pkg/lotus-ui
// TODO: Move this test to pkg/lotus-ui or create mock components
/*
func TestFocusWithMixedComponents(t *testing.T) {
	// Mimic the kitchensink forms tab structure
	nameInput := primitives.NewInput().WithPlaceholder("Name")
	emailInput := primitives.NewInput().WithPlaceholder("Email")
	radioGroup := ui.NewRadioGroup().WithOptions([]ui.RadioOption{
		{Label: "Light", Value: "light"},
		{Label: "Dark", Value: "dark"},
	})

	renderFunc := func() *vdom.Element {
		return vdom.VStack(
			vdom.Box(nameInput),
			vdom.Box(emailInput),
			vdom.Box(radioGroup),
		)
	}

	t.Run("CollectsAllFocusables", func(t *testing.T) {
		fm := newFocusManager()
		element := renderFunc()
		fm.rebuild(element)

		// Should collect: nameInput, emailInput, radioGroup
		if len(fm.focusables) != 3 {
			t.Errorf("Expected 3 focusables (2 inputs + 1 radio group), got %d", len(fm.focusables))
			for i, f := range fm.focusables {
				t.Logf("  [%d] %T", i, f)
			}
		}

		// First input should be focused
		if !nameInput.Focused {
			t.Error("First input should be focused initially")
		}
	})

	t.Run("TabCyclesToRadioGroup", func(t *testing.T) {
		fm := newFocusManager()
		element := renderFunc()
		fm.rebuild(element)

		// Tab to second input
		fm.next()
		if fm.focusIndex != 1 {
			t.Errorf("Expected focusIndex 1, got %d", fm.focusIndex)
		}
		if !emailInput.Focused {
			t.Error("Second input should be focused after Tab")
		}

		// Tab to radio group
		fm.next()
		if fm.focusIndex != 2 {
			t.Errorf("Expected focusIndex 2, got %d", fm.focusIndex)
		}

		// RadioGroup should be the focused component
		if fm.getFocused() == nil {
			t.Fatal("Something should be focused")
		}
		if _, ok := fm.getFocused().(*ui.RadioGroup); !ok {
			t.Errorf("RadioGroup should be focused, got %T", fm.getFocused())
		}
	})

	t.Run("RadioGroupReceivesArrowKeys", func(t *testing.T) {
		fm := newFocusManager()
		element := renderFunc()
		fm.rebuild(element)

		// Tab to radio group
		fm.next() // email
		fm.next() // radio

		// Verify radioGroup is focused (check type)
		_, isRadioGroup := fm.getFocused().(*ui.RadioGroup)
		if !isRadioGroup {
			t.Fatalf("RadioGroup should be focused, got %T", fm.getFocused())
		}

		// Try to press Down arrow on radio group
		downEvent := tty.KeyEvent{Key: 27, Code: tty.SeqDown}
		handled := radioGroup.HandleKeyEvent(downEvent)
		if !handled {
			t.Error("RadioGroup should handle Down arrow when focused")
		}

		// Check that selection changed (second option should be selected after Down)
		// Initial selection is empty, first option gets focus
		// After Down, second option gets focus and can be selected
		spaceEvent := tty.KeyEvent{Key: ' ', Char: " "}
		radioGroup.HandleKeyEvent(spaceEvent)
		if radioGroup.Selected != "dark" {
			t.Errorf("Expected 'dark' to be selected after Down+Space, got %q", radioGroup.Selected)
		}
	})
}
*/

// TestKitchensinkScenario is commented out because it requires
// ui components (Tabs) which are now in pkg/lotus-ui
// TODO: Move this test to pkg/lotus-ui
/*
func TestKitchensinkScenario(t *testing.T) {
	// Mimic kitchensink structure
	nameInput := primitives.NewInput().WithPlaceholder("Name")
	emailInput := primitives.NewInput().WithPlaceholder("Email")

	formsTab := vdom.VStack(
		vdom.Box(nameInput),
		vdom.Box(emailInput),
	)

	otherTab := vdom.Text("Other content")

	tabs := ui.NewTabs().WithTabs([]ui.Tab{
		{Label: "Forms", Content: formsTab},
		{Label: "Other", Content: otherTab},
	}).WithActive(0)

	// Critical: Pass component, not pre-rendered element
	app := &struct{}{}
	renderFunc := func() *vdom.Element {
		return vdom.VStack(
			vdom.Box(vdom.Text("Header")),
			vdom.Box(tabs), // Component, not tabs.Render()
		)
	}

	t.Run("TabsComponentPreservedInTree", func(t *testing.T) {
		element := renderFunc()

		// Find the tabs component in the tree
		var foundTabs *ui.Tabs
		var traverse func(*vdom.Element)
		traverse = func(e *vdom.Element) {
			if e == nil {
				return
			}
			if e.Component != nil {
				if t, ok := e.Component.(*ui.Tabs); ok {
					foundTabs = t
				}
			}
			for _, child := range e.Children {
				traverse(child)
			}
		}
		traverse(element)

		if foundTabs == nil {
			t.Fatal("Tabs component not found in tree - component reference lost!")
		}
		if foundTabs != tabs {
			t.Error("Found Tabs doesn't match original - wrong reference")
		}
	})

	t.Run("CtrlNumberSwitchesTabsWhileInputFocused", func(t *testing.T) {
		fm := newFocusManager()
		element := renderFunc()
		fm.rebuild(element)

		// First input should be focused
		if !nameInput.Focused {
			t.Fatal("Name input should be focused initially")
		}

		// Simulate user typing in input (to prove it has focus)
		typingEvent := tty.KeyEvent{Key: 'a', Char: "a"}
		handled := nameInput.HandleKeyEvent(typingEvent)
		if !handled {
			t.Error("Focused input should handle typing")
		}
		if nameInput.Value != "a" {
			t.Errorf("Expected input value 'a', got '%s'", nameInput.Value)
		}

		// Now press Ctrl+2 to switch tabs
		ctrl2Event := tty.KeyEvent{Key: 2}

		// Input shouldn't handle it
		handled = nameInput.HandleKeyEvent(ctrl2Event)
		if handled {
			t.Error("TextInput should not handle Ctrl+2")
		}

		// Global handler should handle it
		handled = handleEventInTreeGlobal(element, ctrl2Event, fm)
		if !handled {
			t.Error("Global handler (Tabs) should handle Ctrl+2")
		}

		// Tab should have switched
		if tabs.Active != 1 {
			t.Errorf("Expected active tab 1 after Ctrl+2, got %d", tabs.Active)
		}
	})
}
*/
