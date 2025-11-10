package components

import (
	"fmt"
	"strings"
)

// Message represents a single message in the list
type Message struct {
	Role    string // "user" or "assistant"
	Content string
}

// MessageList is a scrollable list of messages
type MessageList struct {
	ID             string // Component ID
	Messages       []Message
	Width          int  // Set by layout system (0 = auto)
	Height         int  // Set by layout system (0 = auto)
	Scroll         int  // Current scroll position
	AutoScroll     bool // Auto-scroll to bottom on new messages
	UserPrefix     string
	ShowSpacing    bool // Add blank line between messages
	UserColor      string
	AssistantColor string
	StreamingColor string
	IsStreaming    bool
}

// NewMessageList creates a new message list (layout-aware)
func NewMessageList(id string) *MessageList {
	return &MessageList{
		ID:             id,
		Messages:       []Message{},
		Width:          0, // Will be set by layout
		Height:         0, // Will be set by layout
		Scroll:         0,
		AutoScroll:     true,
		UserPrefix:     "| ",
		ShowSpacing:    true,
		UserColor:      "#fff",
		AssistantColor: "#ddd",
		StreamingColor: "#888",
	}
}

// GetID returns the component ID (implements ComponentRenderer)
func (m *MessageList) GetID() string {
	return m.ID
}

// SetDimensions allows layout system to set dimensions before render
func (m *MessageList) SetDimensions(width, height int) {
	m.Width = width
	m.Height = height
}

// AddMessage adds a message to the list
func (m *MessageList) AddMessage(role, content string) {
	m.Messages = append(m.Messages, Message{Role: role, Content: content})
	if m.AutoScroll {
		m.ScrollToBottom()
	}
}

// Clear removes all messages
func (m *MessageList) Clear() {
	m.Messages = []Message{}
	m.Scroll = 0
}

// ScrollUp scrolls up one line
func (m *MessageList) ScrollUp() {
	if m.Scroll > 0 {
		m.Scroll--
	}
}

// ScrollDown scrolls down one line
func (m *MessageList) ScrollDown() {
	m.Scroll++
	// Will be clamped in Render
}

// ScrollToBottom scrolls to the bottom
func (m *MessageList) ScrollToBottom() {
	m.Scroll = 999999 // Will be clamped in Render
}

// Render generates the markup for the message list
func (m *MessageList) Render() string {
	if len(m.Messages) == 0 {
		return `<box class="message"></box>`
	}

	// Use sensible defaults if dimensions not set by layout
	width := m.Width
	if width == 0 {
		width = 80 // Default terminal width
	}
	height := m.Height
	if height == 0 {
		height = 20 // Default visible lines
	}

	// Wrap all messages
	type wrappedMsg struct {
		role  string
		lines []string
	}
	var wrapped []wrappedMsg

	wrapWidth := width - 4 // Account for padding
	if wrapWidth < 10 {
		wrapWidth = 10
	}

	for _, msg := range m.Messages {
		lines := wrapText(msg.Content, wrapWidth)
		wrapped = append(wrapped, wrappedMsg{role: msg.Role, lines: lines})
	}

	// Calculate total lines
	totalLines := 0
	for _, msg := range wrapped {
		totalLines += len(msg.lines)
		if m.ShowSpacing {
			totalLines++ // Spacing line
		}
	}

	// Clamp scroll
	maxVisibleLines := height
	if maxVisibleLines < 1 {
		maxVisibleLines = 1
	}

	maxScroll := totalLines - maxVisibleLines
	if maxScroll < 0 {
		maxScroll = 0
	}
	if m.Scroll > maxScroll {
		m.Scroll = maxScroll
	}
	if m.Scroll < 0 {
		m.Scroll = 0
	}

	// Build visible messages
	messageBoxes := ""
	currentLine := 0
	visibleStart := m.Scroll
	visibleEnd := m.Scroll + maxVisibleLines

	for _, msg := range wrapped {
		for _, line := range msg.lines {
			if currentLine >= visibleStart && currentLine < visibleEnd {
				class := "message " + msg.role
				if m.IsStreaming && msg.role == "assistant" {
					class += " streaming"
				}
				// Add prefix for user messages
				prefix := ""
				if msg.role == "user" && m.UserPrefix != "" {
					prefix = m.UserPrefix
				}
				messageBoxes += fmt.Sprintf(`<box class="%s">%s%s</box>`, class, prefix, line)
			}
			currentLine++
		}
		// Add spacing line
		if m.ShowSpacing {
			if currentLine >= visibleStart && currentLine < visibleEnd {
				messageBoxes += `<box class="message"></box>`
			}
			currentLine++
		}
	}

	if messageBoxes == "" {
		messageBoxes = `<box class="message"></box>`
	}

	return messageBoxes
}

// GetCSS returns the CSS for message list styling
func (m *MessageList) GetCSS() string {
	return fmt.Sprintf(`
		.message {
			height: 1;
			margin: 0;
			padding: 0 1 0 1;
		}
		.message.user {
			color: %s;
		}
		.message.assistant {
			color: %s;
		}
		.message.streaming {
			color: %s;
		}
	`, m.UserColor, m.AssistantColor, m.StreamingColor)
}

// wrapText wraps text to fit within the given width
func wrapText(text string, width int) []string {
	if width < 10 {
		width = 10
	}

	var lines []string
	words := strings.Fields(text)

	if len(words) == 0 {
		return []string{text} // Empty or whitespace-only
	}

	currentLine := ""
	for _, word := range words {
		testLine := currentLine
		if testLine != "" {
			testLine += " "
		}
		testLine += word

		if len(testLine) > width {
			if currentLine != "" {
				lines = append(lines, currentLine)
				currentLine = word
			} else {
				// Word itself is too long - split it
				for len(word) > width {
					lines = append(lines, word[:width])
					word = word[width:]
				}
				currentLine = word
			}
		} else {
			currentLine = testLine
		}
	}

	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return lines
}
