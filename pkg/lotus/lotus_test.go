package lotus

import (
	"testing"
)

// Integration test for the full UI API
func TestUI(t *testing.T) {
	markup := `
		<box id="root">
			<box id="header">Terminal UI</box>
			<box id="content">Main content area</box>
		</box>
	`
	css := `
		#root {
			display: flex;
			flex-direction: column;
		}
		#header {
			height: 3;
			border: 1px solid;
		}
		#content {
			flex: 1;
		}
	`

	ui := New(markup, css, 100, 40)

	if ui == nil {
		t.Fatal("New() returned nil")
	}

	if ui.Width != 100 {
		t.Errorf("expected width 100, got %d", ui.Width)
	}

	if ui.Height != 40 {
		t.Errorf("expected height 40, got %d", ui.Height)
	}

	header := ui.FindByID("header")
	if header == nil {
		t.Fatal("FindByID('header') returned nil")
	}

	if header.Height != 3 {
		t.Errorf("expected header height 3, got %d", header.Height)
	}

	output := ui.RenderToTerminal()
	if output == "" {
		t.Error("RenderToTerminal() produced empty output")
	}
}

func TestNewFullscreen(t *testing.T) {
	markup := `<box>Test</box>`
	css := ``

	ui, err := NewFullscreen(markup, css)
	if err != nil {
		t.Fatalf("NewFullscreen() error: %v", err)
	}

	if ui.Width <= 0 || ui.Height <= 0 {
		t.Errorf("invalid dimensions: %dx%d", ui.Width, ui.Height)
	}
}

func TestReflow(t *testing.T) {
	markup := `<box id="test">Content</box>`
	css := `#test { width: 100%; height: 100%; }`

	ui := New(markup, css, 100, 40)

	// Reflow to new size
	ui.Reflow(80, 30)

	if ui.Width != 80 {
		t.Errorf("expected width 80 after reflow, got %d", ui.Width)
	}

	if ui.Height != 30 {
		t.Errorf("expected height 30 after reflow, got %d", ui.Height)
	}

	test := ui.FindByID("test")
	if test.Width != 80 || test.Height != 30 {
		t.Errorf("expected test node 80x30, got %dx%d", test.Width, test.Height)
	}
}
