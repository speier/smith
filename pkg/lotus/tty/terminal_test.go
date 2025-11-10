package tty

import (
	"testing"
)

func TestKeyEvent(t *testing.T) {
	tests := []struct {
		name        string
		key         byte
		isCtrlC     bool
		isCtrlD     bool
		isEnter     bool
		isBackspace bool
		isPrintable bool
	}{
		{"Ctrl+C", KeyCtrlC, true, false, false, false, false},
		{"Ctrl+D", KeyCtrlD, false, true, false, false, false},
		{"Enter (CR)", KeyEnter, false, false, true, false, false},
		{"Enter (LF)", KeyEnter2, false, false, true, false, false},
		{"Backspace", KeyBackspace, false, false, false, true, false},
		{"Backspace2", KeyBackspace2, false, false, false, true, false},
		{"Letter A", 65, false, false, false, false, true},
		{"Space", 32, false, false, false, false, true},
		{"Tilde", 126, false, false, false, false, true},
		{"Escape", KeyEscape, false, false, false, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := KeyEvent{Key: tt.key}
			if tt.isPrintable {
				event.Char = string(tt.key)
			}

			if event.IsCtrlC() != tt.isCtrlC {
				t.Errorf("IsCtrlC() = %v, want %v", event.IsCtrlC(), tt.isCtrlC)
			}
			if event.IsCtrlD() != tt.isCtrlD {
				t.Errorf("IsCtrlD() = %v, want %v", event.IsCtrlD(), tt.isCtrlD)
			}
			if event.IsEnter() != tt.isEnter {
				t.Errorf("IsEnter() = %v, want %v", event.IsEnter(), tt.isEnter)
			}
			if event.IsBackspace() != tt.isBackspace {
				t.Errorf("IsBackspace() = %v, want %v", event.IsBackspace(), tt.isBackspace)
			}
			if event.IsPrintable() != tt.isPrintable {
				t.Errorf("IsPrintable() = %v, want %v", event.IsPrintable(), tt.isPrintable)
			}
		})
	}
}

func TestInputReader(t *testing.T) {
	reader, err := NewInputReader()
	if err != nil {
		t.Fatalf("NewInputReader() failed: %v", err)
	}
	if reader == nil {
		t.Fatal("NewInputReader() returned nil")
	}
	defer func() { _ = reader.Close() }()

	// Test default state
	if reader.filterMouse != false {
		t.Error("Expected filterMouse to be false by default")
	}
	if reader.inEscapeSeq != false {
		t.Error("Expected inEscapeSeq to be false by default")
	}

	// Test SetFilterMouse
	reader.SetFilterMouse(true)
	if reader.filterMouse != true {
		t.Error("SetFilterMouse(true) did not set filterMouse")
	}

	reader.SetFilterMouse(false)
	if reader.filterMouse != false {
		t.Error("SetFilterMouse(false) did not unset filterMouse")
	}
}

func TestScreen(t *testing.T) {
	// Note: NewScreen() will try to detect terminal size
	// This test may fail in non-terminal environments
	screen, err := NewScreen()
	if err != nil {
		t.Fatalf("NewScreen() error = %v", err)
	}

	width, height := screen.Size()
	if width <= 0 || height <= 0 {
		t.Errorf("Invalid screen size: %dx%d", width, height)
	}

	// Test state tracking
	if screen.inAltMode {
		t.Error("Expected inAltMode to be false initially")
	}
	if screen.cursorHidden {
		t.Error("Expected cursorHidden to be false initially")
	}
}
