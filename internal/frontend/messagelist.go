package frontend

import (
	"fmt"

	"github.com/speier/smith/pkg/lotus/vdom"
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
func NewMessageList(id ...string) *MessageList {
	listID := ""
	if len(id) > 0 {
		listID = id[0]
	}
	return &MessageList{
		ID:             listID,
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

// Render generates the Element for the message list
func (m *MessageList) Render() *vdom.Element {
	if len(m.Messages) == 0 {
		return vdom.Box(vdom.Text(""))
	}

	height := m.Height
	if height == 0 {
		height = 20 // Default visible lines
	}

	// Calculate total lines
	totalLines := len(m.Messages)
	if m.ShowSpacing {
		totalLines += len(m.Messages)
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

	// Build visible message elements
	messageElements := []*vdom.Element{}
	currentLine := 0
	visibleStart := m.Scroll
	visibleEnd := m.Scroll + maxVisibleLines

	for _, msg := range m.Messages {
		// Render message line
		if currentLine >= visibleStart && currentLine < visibleEnd {
			color := m.AssistantColor
			prefix := ""

			if msg.Role == "user" {
				color = m.UserColor
				if m.UserPrefix != "" {
					prefix = m.UserPrefix
				}
			} else if m.IsStreaming {
				color = "#5af" // Streaming indicator
			}

			messageElements = append(messageElements,
				vdom.Box(
					vdom.Text(prefix+msg.Content),
				).WithStyle("color", color).
					WithStyle("padding", "0 1"),
			)
		}
		currentLine++

		// Add spacing line
		if m.ShowSpacing {
			if currentLine >= visibleStart && currentLine < visibleEnd {
				messageElements = append(messageElements,
					vdom.Box(vdom.Text("")),
				)
			}
			currentLine++
		}
	}

	if len(messageElements) == 0 {
		messageElements = append(messageElements, vdom.Box(vdom.Text("")))
	}

	// Convert []*Element to []any for VStack
	children := make([]any, len(messageElements))
	for i, elem := range messageElements {
		children[i] = elem
	}

	return vdom.VStack(children...).
		WithID(m.ID).
		WithStyle("height", fmt.Sprintf("%d", height))
}

// IsNode implements vdom.Node interface
func (m *MessageList) IsNode() {}
