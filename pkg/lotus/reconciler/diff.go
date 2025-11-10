package reconciler

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/speier/smith/pkg/lotus/core"
)

// Patch represents a change to apply to the UI tree
type Patch interface {
	Apply(ui *UI) error
}

// UpdateTextPatch updates text content of a node without re-parsing or re-layout
type UpdateTextPatch struct {
	NodeID  string
	NewText string
}

// Apply updates the text content of a node
func (p UpdateTextPatch) Apply(ui *UI) error {
	node := ui.findNodeByID(p.NodeID)
	if node == nil {
		return fmt.Errorf("node not found: %s", p.NodeID)
	}
	node.Content = p.NewText
	return nil
}

// UpdateStylePatch updates styles and re-layouts only the affected subtree
type UpdateStylePatch struct {
	NodeID    string
	NewStyles map[string]string
}

// Apply updates styles and re-layouts the subtree
func (p UpdateStylePatch) Apply(ui *UI) error {
	node := ui.findNodeByID(p.NodeID)
	if node == nil {
		return fmt.Errorf("node not found: %s", p.NodeID)
	}

	// Update computed styles
	if node.Styles == nil {
		node.Styles = &core.ComputedStyle{}
	}

	for key, value := range p.NewStyles {
		switch key {
		case "color":
			node.Styles.Color = value
		case "background":
			node.Styles.BgColor = value
		case "width":
			node.Styles.Width = value
		case "height":
			node.Styles.Height = value
		case "border":
			node.Styles.Border = value != ""
		case "flex":
			node.Styles.Flex = value
			// TODO: Add padding, margin, etc. - need to parse values
		}
	}

	// Re-layout only this subtree
	// TODO: Implement incremental layout
	// For now, we'll need to re-layout the entire tree
	// This is still faster than re-parsing!

	return nil
}

// ReplaceNodePatch replaces an entire node subtree
type ReplaceNodePatch struct {
	NodeID     string
	NewElement *core.Element
}

// Apply replaces a node with a new subtree
func (p ReplaceNodePatch) Apply(ui *UI) error {
	// For full replacement, we need to re-render the subtree
	// This is the fallback when diffing can't help
	// TODO: Implement node replacement
	return nil
}

// InsertNodePatch inserts a new node
type InsertNodePatch struct {
	ParentID   string
	NewElement *core.Element
	Index      int
}

// Apply inserts a new node at the specified position
func (p InsertNodePatch) Apply(ui *UI) error {
	// TODO: Implement node insertion
	return nil
}

// DeleteNodePatch removes a node
type DeleteNodePatch struct {
	NodeID string
}

// Apply removes a node from the tree
func (p DeleteNodePatch) Apply(ui *UI) error {
	// TODO: Implement node deletion
	return nil
}

// Diff compares two element trees and generates minimal patches
func Diff(old, new *core.Element) []Patch {
	// Fast path: If pointers are equal, no changes needed
	if old == new {
		return nil
	}

	if old == nil && new == nil {
		return nil
	}

	if old == nil {
		// New tree - no patches needed, will do full render
		return nil
	}

	if new == nil {
		// Deleted - remove everything
		if old.ID != "" {
			return []Patch{DeleteNodePatch{NodeID: old.ID}}
		}
		return nil
	}

	patches := []Patch{}

	// Different types = full replacement
	if old.Type != new.Type {
		if old.ID != "" {
			patches = append(patches, ReplaceNodePatch{
				NodeID:     old.ID,
				NewElement: new,
			})
		}
		return patches
	}

	// Same type - check what changed
	switch old.Type {
	case "markup":
		// Text content changed?
		if old.Markup != new.Markup && old.ID != "" {
			patches = append(patches, UpdateTextPatch{
				NodeID:  old.ID,
				NewText: new.Markup,
			})
		}
	case "component":
		// Component changed - need to re-render
		// For now, treat as replacement
		if old.ID != "" {
			patches = append(patches, ReplaceNodePatch{
				NodeID:     old.ID,
				NewElement: new,
			})
		}
	case "container", "container-with-markup":
		// Check if pre-rendered markup changed
		if old.Markup != new.Markup && old.ID != "" {
			// Container markup changed - check children
			patches = append(patches, diffChildren(old, new)...)
		}

		// Check if inline styles changed
		if !stylesEqual(old.Styles, new.Styles) && old.ID != "" {
			patches = append(patches, UpdateStylePatch{
				NodeID:    old.ID,
				NewStyles: new.Styles,
			})
		}

		// Diff children recursively
		patches = append(patches, diffChildren(old, new)...)
	}

	return patches
}

// diffChildren compares children arrays and generates patches
func diffChildren(old, new *core.Element) []Patch {
	patches := []Patch{}

	oldChildren := old.Children
	newChildren := new.Children

	// Simple diffing strategy: compare by index
	// TODO: Implement keyed diffing for better list performance

	maxLen := len(oldChildren)
	if len(newChildren) > maxLen {
		maxLen = len(newChildren)
	}

	for i := 0; i < maxLen; i++ {
		if i >= len(oldChildren) {
			// New child added
			if new.ID != "" {
				patches = append(patches, InsertNodePatch{
					ParentID:   new.ID,
					NewElement: newChildren[i],
					Index:      i,
				})
			}
		} else if i >= len(newChildren) {
			// Child removed
			if oldChildren[i].ID != "" {
				patches = append(patches, DeleteNodePatch{
					NodeID: oldChildren[i].ID,
				})
			}
		} else {
			// Child exists in both - recursively diff
			childPatches := Diff(oldChildren[i], newChildren[i])
			patches = append(patches, childPatches...)
		}
	}

	return patches
}

// stylesEqual compares two style maps
func stylesEqual(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}

	for key, valA := range a {
		if valB, ok := b[key]; !ok || valA != valB {
			return false
		}
	}

	return true
}

// findNodeByID recursively searches for a node by ID
func (ui *UI) findNodeByID(id string) *core.Node {
	if ui.Root == nil {
		return nil
	}
	return findNodeByIDRecursive(ui.Root, id)
}

func findNodeByIDRecursive(node *core.Node, id string) *core.Node {
	if node.ID == id {
		return node
	}

	for _, child := range node.Children {
		if found := findNodeByIDRecursive(child, id); found != nil {
			return found
		}
	}

	return nil
}

// GenerateID generates a unique ID for an element
func GenerateID() string {
	bytes := make([]byte, 8)
	_, _ = rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
