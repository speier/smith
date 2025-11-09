package session

import (
	"time"
)

// MockSession is a simple session for testing/demo
// In production, this would be your actual agent system
type MockSession struct {
	history []Message
}

func NewMockSession() *MockSession {
	return &MockSession{
		history: []Message{},
	}
}

func (m *MockSession) SendMessage(message string) (<-chan string, error) {
	// Add user message to history
	m.history = append(m.history, Message{
		Role:    "user",
		Content: message,
	})

	// Create response channel
	ch := make(chan string)

	// Simulate streaming response
	go func() {
		defer close(ch)

		response := `Here's a response with **markdown**!

## Code Example

` + "```go\nfunc main() {\n    fmt.Println(\"Hello, Smith!\")\n}\n```" + `

- Point 1
- Point 2
- Point 3

What else would you like to know?`

		// Stream word by word
		for _, word := range splitWords(response) {
			ch <- word
			time.Sleep(30 * time.Millisecond)
		}
	}()

	return ch, nil
}

func (m *MockSession) GetHistory() []Message {
	return m.history
}

func (m *MockSession) Reset() {
	m.history = []Message{}
}

func splitWords(text string) []string {
	var words []string
	var current []rune

	for _, r := range text {
		if r == ' ' || r == '\n' {
			if len(current) > 0 {
				words = append(words, string(current))
				current = nil
			}
			if r == '\n' {
				words = append(words, "\n")
			} else {
				words = append(words, " ")
			}
		} else {
			current = append(current, r)
		}
	}

	if len(current) > 0 {
		words = append(words, string(current))
	}

	return words
}
