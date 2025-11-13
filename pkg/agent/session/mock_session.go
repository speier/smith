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

	// Special test command for scrolling
	if contains(msgLower, "scroll") || contains(msgLower, "long") {
		return `This is a very long response to test scrolling functionality.

Line 1: Lorem ipsum dolor sit amet, consectetur adipiscing elit.
Line 2: Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.
Line 3: Ut enim ad minim veniam, quis nostrud exercitation ullamco.
Line 4: Duis aute irure dolor in reprehenderit in voluptate velit.
Line 5: Esse cillum dolore eu fugiat nulla pariatur.
Line 6: Excepteur sint occaecat cupidatat non proident.
Line 7: Sunt in culpa qui officia deserunt mollit anim id est laborum.
Line 8: Sed ut perspiciatis unde omnis iste natus error sit.
Line 9: Voluptatem accusantium doloremque laudantium totam rem.
Line 10: Aperiam eaque ipsa quae ab illo inventore veritatis.
Line 11: Et quasi architecto beatae vitae dicta sunt explicabo.
Line 12: Nemo enim ipsam voluptatem quia voluptas sit aspernatur.
Line 13: Aut odit aut fugit sed quia consequuntur magni dolores.
Line 14: Eos qui ratione voluptatem sequi nesciunt neque porro.
Line 15: Quisquam est qui dolorem ipsum quia dolor sit amet.
Line 16: Consectetur adipisci velit sed quia non numquam eius.
Line 17: Modi tempora incidunt ut labore et dolore magnam aliquam.
Line 18: Quaerat voluptatem ut enim ad minima veniam quis nostrum.
Line 19: Exercitationem ullam corporis suscipit laboriosam nisi.
Line 20: Ut aliquid ex ea commodi consequatur quis autem vel eum.

This should be enough lines to trigger scrolling in a typical terminal window.
The ScrollView component should automatically scroll to show the latest content.`
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
