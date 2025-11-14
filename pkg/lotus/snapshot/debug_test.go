package snapshot

import (
	"testing"

	"github.com/speier/smith/pkg/lotus/vdom"
)

func TestDebugChatLayout(t *testing.T) {
	// Reproduce FULL chat structure
	messages := []any{
		vdom.Text("Welcome"),
		vdom.Text("Type a message"),
	}

	element := vdom.VStack(
		// Header
		vdom.Box(
			vdom.Text("Lotus Chat"),
		).WithBorderStyle(vdom.BorderStyleRounded),

		// Messages - now using overflow:auto via flex-grow
		vdom.Box(
			vdom.VStack(messages...).WithGap(1),
		).
			WithFlexGrow(1).
			WithBorderStyle(vdom.BorderStyleRounded),

		// Input
		vdom.Box(
			vdom.Text("Type a message..."),
		).WithBorderStyle(vdom.BorderStyleRounded),
	)

	snapshot := Render(element, 160, 40)

	t.Logf("Layout tree:\n%s", snapshot.DumpLayout())

	// Find where "Welcome" appears
	welcomeLine := -1
	for i, line := range snapshot.Lines {
		if containsText(line, "Welcome") {
			welcomeLine = i
			t.Logf("'Welcome' found at line %d", i)
			break
		}
	}

	if welcomeLine == -1 {
		t.Fatal("Could not find 'Welcome'")
	}

	// After header box (3 lines: top border, content, bottom border),
	// messages box starts, then border (1 line), then content should be at line 4
	if welcomeLine > 5 {
		t.Errorf("Content too far down: 'Welcome' at line %d (expected ~line 4-5)", welcomeLine)
		t.Logf("\nFirst 10 lines:\n")
		for i := 0; i < 10 && i < len(snapshot.Lines); i++ {
			t.Logf("Line %d: %q", i, snapshot.Lines[i])
		}
	}
}
