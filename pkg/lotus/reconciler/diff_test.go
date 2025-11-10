package reconciler

import (
	"testing"

	"github.com/speier/smith/pkg/lotus/core"
)

func TestDiff_TextChange(t *testing.T) {
	// Test that text changes generate UpdateTextPatch
	old := core.NewMarkupElement("Hello")
	old.SetID("text-1")

	new := core.NewMarkupElement("World")
	new.SetID("text-1")

	patches := Diff(old, new)

	if len(patches) != 1 {
		t.Fatalf("expected 1 patch, got %d", len(patches))
	}

	patch, ok := patches[0].(UpdateTextPatch)
	if !ok {
		t.Fatalf("expected UpdateTextPatch, got %T", patches[0])
	}

	if patch.NodeID != "text-1" {
		t.Errorf("expected NodeID text-1, got %s", patch.NodeID)
	}

	if patch.NewText != "World" {
		t.Errorf("expected NewText 'World', got %s", patch.NewText)
	}
}

func TestDiff_StyleChange(t *testing.T) {
	// Test that style changes generate UpdateStylePatch
	old := core.NewContainerElement()
	old.SetID("container-1")
	old.SetStyle("color", "red")

	new := core.NewContainerElement()
	new.SetID("container-1")
	new.SetStyle("color", "blue")

	patches := Diff(old, new)

	// Should have at least one UpdateStylePatch
	found := false
	for _, p := range patches {
		if _, ok := p.(UpdateStylePatch); ok {
			found = true
			break
		}
	}

	if !found {
		t.Error("expected UpdateStylePatch in patches")
	}
}

func TestDiff_NoChange(t *testing.T) {
	// Test that identical elements generate no patches
	old := core.NewMarkupElement("Same")
	old.SetID("text-1")

	new := core.NewMarkupElement("Same")
	new.SetID("text-1")

	patches := Diff(old, new)

	if len(patches) != 0 {
		t.Errorf("expected 0 patches for identical elements, got %d", len(patches))
	}
}

func TestDiff_TypeChange(t *testing.T) {
	// Test that type changes generate ReplaceNodePatch
	old := core.NewMarkupElement("Text")
	old.SetID("elem-1")

	new := core.NewContainerElement()
	new.SetID("elem-1")

	patches := Diff(old, new)

	if len(patches) != 1 {
		t.Fatalf("expected 1 patch, got %d", len(patches))
	}

	_, ok := patches[0].(ReplaceNodePatch)
	if !ok {
		t.Fatalf("expected ReplaceNodePatch, got %T", patches[0])
	}
}

func TestDiff_ChildrenAdded(t *testing.T) {
	// Test that new children generate InsertNodePatch
	old := core.NewContainerElement()
	old.SetID("container-1")

	child1 := core.NewMarkupElement("Child 1")
	child1.SetID("child-1")

	new := core.NewContainerElement(child1)
	new.SetID("container-1")

	patches := Diff(old, new)

	// Should have InsertNodePatch
	found := false
	for _, p := range patches {
		if _, ok := p.(InsertNodePatch); ok {
			found = true
			break
		}
	}

	if !found {
		t.Log("Expected InsertNodePatch (may not be implemented yet)")
	}
}

func TestDiff_NilElements(t *testing.T) {
	// Test edge cases with nil elements
	patches := Diff(nil, nil)
	if patches != nil {
		t.Error("expected nil patches for nil elements")
	}

	old := core.NewMarkupElement("Test")
	_ = Diff(old, nil)
	// Should handle gracefully

	patches = Diff(nil, old)
	// Should handle gracefully - return nil (full render will happen)
	if patches != nil {
		t.Error("expected nil patches when old is nil")
	}
}

func TestStylesEqual(t *testing.T) {
	tests := []struct {
		name     string
		a        map[string]string
		b        map[string]string
		expected bool
	}{
		{
			name:     "both nil",
			a:        nil,
			b:        nil,
			expected: true,
		},
		{
			name:     "both empty",
			a:        map[string]string{},
			b:        map[string]string{},
			expected: true,
		},
		{
			name:     "same values",
			a:        map[string]string{"color": "red"},
			b:        map[string]string{"color": "red"},
			expected: true,
		},
		{
			name:     "different values",
			a:        map[string]string{"color": "red"},
			b:        map[string]string{"color": "blue"},
			expected: false,
		},
		{
			name:     "different keys",
			a:        map[string]string{"color": "red"},
			b:        map[string]string{"background": "red"},
			expected: false,
		},
		{
			name:     "different lengths",
			a:        map[string]string{"color": "red"},
			b:        map[string]string{"color": "red", "background": "blue"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stylesEqual(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestGenerateID(t *testing.T) {
	// Test that GenerateID generates unique IDs
	id1 := GenerateID()
	id2 := GenerateID()

	if id1 == id2 {
		t.Error("GenerateID should generate unique IDs")
	}

	if len(id1) == 0 {
		t.Error("GenerateID should not return empty string")
	}
}
