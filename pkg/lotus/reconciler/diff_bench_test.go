package reconciler

import (
	"fmt"
	"testing"

	"github.com/speier/smith/pkg/lotus/core"
)

// BenchmarkDiff_TextChange benchmarks diffing a simple text change
func BenchmarkDiff_TextChange(b *testing.B) {
	old := core.NewMarkupElement("Hello World")
	old.SetID("text-1")
	
	new := core.NewMarkupElement("Hello Universe")
	new.SetID("text-1")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Diff(old, new)
	}
}

// BenchmarkDiff_ComplexTree benchmarks diffing a complex element tree
func BenchmarkDiff_ComplexTree(b *testing.B) {
	// Create old tree with 10 children
	oldChildren := make([]*core.Element, 10)
	for i := 0; i < 10; i++ {
		child := core.NewMarkupElement(fmt.Sprintf("Item %d", i))
		child.SetID(fmt.Sprintf("item-%d", i))
		oldChildren[i] = child
	}
	old := core.NewContainerElement(oldChildren...)
	old.SetID("list")
	
	// Create new tree with one text change
	newChildren := make([]*core.Element, 10)
	for i := 0; i < 10; i++ {
		text := fmt.Sprintf("Item %d", i)
		if i == 5 {
			text = "Item 5 CHANGED"
		}
		child := core.NewMarkupElement(text)
		child.SetID(fmt.Sprintf("item-%d", i))
		newChildren[i] = child
	}
	new := core.NewContainerElement(newChildren...)
	new.SetID("list")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Diff(old, new)
	}
}

// BenchmarkFullRender_TextChange simulates full re-render cost
func BenchmarkFullRender_TextChange(b *testing.B) {
	markup := `<text>Hello World</text>`
	css := `text { color: #fff; }`
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ui := NewUI(markup, css, 80, 24)
		_ = ui.RenderToTerminal(false)
	}
}

// BenchmarkDiffRender_TextChange simulates diff-based update
func BenchmarkDiffRender_TextChange(b *testing.B) {
	// Initial render
	old := core.NewMarkupElement("Hello World")
	old.SetID("text-1")
	
	markup := old.ToMarkup()
	css := old.ToCSS()
	ui := NewUI(markup, css, 80, 24)
	ui.previousTree = old
	
	// Update with text change
	new := core.NewMarkupElement("Hello Universe")
	new.SetID("text-1")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		patches := Diff(ui.previousTree, new)
		_ = ui.ApplyPatches(patches)
		_ = ui.RenderToTerminal(false)
	}
}
