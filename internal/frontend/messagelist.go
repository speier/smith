package frontend

import (
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/speier/smith/pkg/lotus/vdom"
)

// Message represents a single message in the list
type Message struct {
	Role    string // "user", "assistant", "system"
	Content string
}

// MessageList is a scrollable list of chat messages with markdown rendering
type MessageList struct {
	ID         string        // Component ID
	Messages   []Message     // Chat messages
	Streaming  bool          // Currently streaming a response
	StreamBuf  string        // Partial streaming content
	Header     *vdom.Element // Optional header (e.g., logo, banner)
	mdRenderer *glamour.TermRenderer
}

// NewMessageList creates a new message list with markdown support
func NewMessageList() *MessageList {
	// Create markdown renderer for assistant messages with no left padding
	mdRenderer, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),         // Auto dark/light mode
		glamour.WithWordWrap(80),        // Wrap at 80 chars
		glamour.WithPreservedNewLines(), // Preserve hard line breaks
		glamour.WithStylesFromJSONBytes([]byte(`{
			"document": {
				"margin": 0
			},
			"paragraph": {
				"margin": 0
			}
		}`)),
	)

	return &MessageList{
		ID:         "messages",
		Messages:   []Message{},
		Streaming:  false,
		mdRenderer: mdRenderer,
	}
}

// AddMessage adds a message to the list
func (m *MessageList) AddMessage(role, content string) {
	m.Messages = append(m.Messages, Message{Role: role, Content: content})
}

// SetHeader sets the optional header element (e.g., logo and banner)
func (m *MessageList) SetHeader(elem *vdom.Element) {
	m.Header = elem
}

// Clear removes all messages
func (m *MessageList) Clear() {
	m.Messages = []Message{}
}

// SetStreaming sets streaming state and partial content
func (m *MessageList) SetStreaming(streaming bool, partial string) {
	m.Streaming = streaming
	m.StreamBuf = partial
}

// formatMessage formats a message with role prefix and markdown rendering
func (m *MessageList) formatMessage(role, content string) string {
	var prefix string
	switch role {
	case "user":
		prefix = "\x1b[36m>\x1b[0m " // Cyan arrow for user
	case "assistant":
		prefix = "\x1b[32m●\x1b[0m " // Green bullet for assistant
	case "system":
		// No prefix for system messages (welcome banner, errors, etc.)
		prefix = ""
	default:
		prefix = "  "
	}

	// Render markdown for assistant messages only (not system/user)
	rendered := content
	if role == "assistant" {
		if r, err := m.mdRenderer.Render(content); err == nil {
			// Trim all whitespace and split into lines
			r = strings.TrimSpace(r)
			// Also trim leading spaces from each line that glamour adds
			lines := strings.Split(r, "\n")
			for i, line := range lines {
				lines[i] = strings.TrimLeft(line, " ")
			}
			rendered = strings.Join(lines, "\n")
		}
	}

	// Add prefix to first line (if prefix exists)
	if prefix != "" {
		lines := strings.Split(rendered, "\n")
		if len(lines) > 0 {
			lines[0] = prefix + lines[0]
			// Indent subsequent lines to align with first line content
			indent := "  " // Match the "● " width
			for i := 1; i < len(lines); i++ {
				lines[i] = indent + lines[i]
			}
			rendered = strings.Join(lines, "\n")
		}
	}

	return rendered
}

// Render generates the Element for the message list
func (m *MessageList) Render() *vdom.Element {
	// Build message elements from formatted strings
	messageElements := make([]any, 0, len(m.Messages)+2)

	// Add header if set (logo, banner, etc.)
	if m.Header != nil {
		messageElements = append(messageElements, m.Header)
	}

	for i, msg := range m.Messages {
		formatted := m.formatMessage(msg.Role, msg.Content)
		messageElements = append(messageElements, vdom.Text(formatted))

		// Add spacing after assistant messages (before next user message)
		if msg.Role == "assistant" && i < len(m.Messages)-1 {
			messageElements = append(messageElements, vdom.Text(""))
		}
	}

	// If streaming, show partial response with indicator
	if m.Streaming {
		if m.StreamBuf != "" {
			streaming := m.formatMessage("assistant", m.StreamBuf+" ▌") // Blinking cursor indicator
			messageElements = append(messageElements, vdom.Text(streaming))
		} else {
			messageElements = append(messageElements, vdom.Text("\x1b[32m●\x1b[0m Thinking... ▌"))
		}
	}

	// Return VStack of messages
	return vdom.VStack(messageElements...).WithID(m.ID)
}

// IsNode implements vdom.Node interface
func (m *MessageList) IsNode() {}
