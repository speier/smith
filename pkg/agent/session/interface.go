package session

// Session interface - represents an interactive coding session
// Backed by the agent system (Planning, Implementation, Testing, Review)
type Session interface {
	// SendMessage sends a message and streams the response
	SendMessage(message string) (<-chan string, error)

	// GetHistory returns all messages in the conversation
	GetHistory() []Message

	// Reset clears the conversation history
	Reset()
}

// Message represents a chat message
type Message struct {
	Role    string // "user" or "assistant"
	Content string
}
