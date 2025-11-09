package frontend

import (
	"strings"
	"testing"
)

func TestGetWelcomeBanner(t *testing.T) {
	banner := GetWelcomeBanner()

	// Should contain the logo (ASCII art uses block characters)
	if !strings.Contains(banner, "███") {
		t.Error("Welcome banner should contain SMITH ASCII logo")
	}

	// Should contain version
	if !strings.Contains(banner, "v") {
		t.Error("Welcome banner should contain version")
	}

	// Should contain welcome message
	if !strings.Contains(banner, "Mr. Anderson") {
		t.Error("Welcome banner should contain 'Mr. Anderson' message")
	}

	if !strings.Contains(banner, "Welcome back") {
		t.Error("Welcome banner should contain 'Welcome back' message")
	}
}

func TestGetGoodbyeBanner(t *testing.T) {
	goodbye := GetGoodbyeBanner()

	// Should contain the goodbye message with wave emoji
	if !strings.Contains(goodbye, "Goodbye") {
		t.Error("Goodbye banner should contain farewell message")
	}

	if !strings.Contains(goodbye, "👋") {
		t.Error("Goodbye banner should contain wave emoji")
	}
}

func TestGetLogoLines(t *testing.T) {
	lines := GetLogoLines()

	// Should have 6 lines
	if len(lines) != 6 {
		t.Errorf("Logo should have 6 lines, got: %d", len(lines))
	}

	// Each line should contain block characters (█ or box drawing chars)
	for _, line := range lines {
		if len(line) == 0 {
			t.Error("Logo lines should not be empty")
		}
		// All logo lines contain "S", "M", "I", "T", or "H" made from box chars
		if !strings.Contains(line, "█") && !strings.Contains(line, "╗") && !strings.Contains(line, "╚") && !strings.Contains(line, "═") {
			t.Errorf("Logo line should contain ASCII art characters, got: %q", line)
		}
	}
}
