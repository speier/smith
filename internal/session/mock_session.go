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

		// Generate a contextual response based on the message
		response := m.generateResponse(message)

		// Stream word by word for realistic typing effect
		for _, word := range splitWords(response) {
			ch <- word
			time.Sleep(50 * time.Millisecond) // Slower for more natural typing feel
		}

		// Add complete response to history after streaming
		m.history = append(m.history, Message{
			Role:    "assistant",
			Content: response,
		})
	}()

	return ch, nil
}

// generateResponse creates a contextual response based on user input
func (m *MockSession) generateResponse(message string) string {
	// Simple responses based on keywords
	msgLower := message

	if len(message) < 10 {
		// Short message - give a brief response
		return "I understand. Could you elaborate on that?"
	}

	if contains(msgLower, "hello") || contains(msgLower, "hi") {
		return `Hello! I'm Smith, your AI coding assistant.

I can help you with:
• Writing code
• Debugging issues
• Explaining concepts
• Planning architecture

What would you like to work on?`
	}

	if contains(msgLower, "code") || contains(msgLower, "function") || contains(msgLower, "implement") {
		return `Sure! Here's an example implementation:

` + "```go\nfunc Example() string {\n    return \"This is working!\"\n}\n```" + `

This demonstrates the basic structure. Would you like me to:
• Add error handling?
• Include tests?
• Explain how it works?`
	}

	if contains(msgLower, "test") {
		return `Great idea! Here's a test example:

` + "```go\nfunc TestExample(t *testing.T) {\n    result := Example()\n    if result == \"\" {\n        t.Error(\"expected non-empty result\")\n    }\n}\n```" + `

All tests should pass with this approach.`
	}

	// Default response with user's message echoed
	return "You said: \"" + message + "\"\n\nI'm processing that request. In a real system, the Planning Agent would analyze this, the Implementation Agent would write code, the Testing Agent would verify it, and the Review Agent would ensure quality.\n\nWhat specific task would you like help with?"
}

// Helper function to check if string contains substring (case-insensitive)
func contains(s, substr string) bool {
	sLower := ""
	substrLower := ""
	for _, r := range s {
		if r >= 'A' && r <= 'Z' {
			sLower += string(r + 32)
		} else {
			sLower += string(r)
		}
	}
	for _, r := range substr {
		if r >= 'A' && r <= 'Z' {
			substrLower += string(r + 32)
		} else {
			substrLower += string(r)
		}
	}

	for i := 0; i <= len(sLower)-len(substrLower); i++ {
		if sLower[i:i+len(substrLower)] == substrLower {
			return true
		}
	}
	return false
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
