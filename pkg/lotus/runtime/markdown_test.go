package runtime

import (
	"strings"
	"testing"
)

func TestMarkdownRenderer(t *testing.T) {
	renderer, err := NewMarkdownRenderer(80)
	if err != nil {
		t.Fatalf("Failed to create markdown renderer: %v", err)
	}

	markdown := "# Hello\n\nThis is **bold** and `code`."
	result, err := renderer.Render(markdown)
	if err != nil {
		t.Fatalf("Failed to render markdown: %v", err)
	}

	// Should contain ANSI codes for styling
	if result == markdown {
		t.Error("Rendered markdown should be different from input (should have ANSI codes)")
	}
}

func TestRenderMarkdown(t *testing.T) {
	markdown := "## Heading\n\n- List item 1\n- List item 2"
	result, err := RenderMarkdown(markdown, 80)
	if err != nil {
		t.Fatalf("Failed to render markdown: %v", err)
	}

	if result == "" {
		t.Error("Rendered markdown should not be empty")
	}
}

func TestMarkdownHelper(t *testing.T) {
	markdown := "**Bold text**"
	result := Markdown(markdown, 80)

	if !strings.Contains(result, "<text>") {
		t.Error("Markdown helper should create text element")
	}
}

func TestMarkdownCodeBlock(t *testing.T) {
	markdown := "```go\nfunc main() {\n    fmt.Println(\"Hello\")\n}\n```"
	result, err := RenderMarkdown(markdown, 80)
	if err != nil {
		t.Fatalf("Failed to render code block: %v", err)
	}

	// Should contain the code content
	if !strings.Contains(result, "main") {
		t.Error("Rendered markdown should contain code content")
	}
}

func TestMarkdownFallback(t *testing.T) {
	// Test that invalid markdown doesn't crash
	markdown := "# Valid markdown"
	result := Markdown(markdown, 80)

	if result == "" {
		t.Error("Markdown helper should return something even on error")
	}
}
