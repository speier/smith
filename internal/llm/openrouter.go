package llm

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// OpenRouterProvider implements OpenRouter API access
type OpenRouterProvider struct {
	apiKey   string
	endpoint string
	client   *http.Client
}

func NewOpenRouterProvider() *OpenRouterProvider {
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	return &OpenRouterProvider{
		apiKey:   apiKey,
		endpoint: "https://openrouter.ai/api/v1/chat/completions",
		client:   &http.Client{Timeout: 60 * time.Second},
	}
}

func (o *OpenRouterProvider) Chat(messages []Message, tools []Tool) (*Response, error) {
	if o.apiKey == "" {
		return nil, fmt.Errorf("OPENROUTER_API_KEY environment variable not set")
	}

	// Convert to OpenAI format
	reqMessages := make([]map[string]string, len(messages))
	for i, msg := range messages {
		reqMessages[i] = map[string]string{
			"role":    msg.Role,
			"content": msg.Content,
		}
	}

	reqBody := map[string]interface{}{
		"model":    "anthropic/claude-3.5-sonnet", // Default model
		"messages": reqMessages,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	req, err := http.NewRequest("POST", o.endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.apiKey)
	req.Header.Set("HTTP-Referer", "https://github.com/speier/smith")
	req.Header.Set("X-Title", "Smith Agent System")

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("api request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("api error (%d): %s", resp.StatusCode, string(bodyBytes))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("no response from API")
	}

	return &Response{
		Content: result.Choices[0].Message.Content,
	}, nil
}

func (o *OpenRouterProvider) ChatStream(messages []Message, tools []Tool, callback func(*Response) error) error {
	if o.apiKey == "" {
		return fmt.Errorf("OPENROUTER_API_KEY environment variable not set")
	}

	// Convert to OpenAI format
	reqMessages := make([]map[string]string, len(messages))
	for i, msg := range messages {
		reqMessages[i] = map[string]string{
			"role":    msg.Role,
			"content": msg.Content,
		}
	}

	reqBody := map[string]interface{}{
		"model":    "anthropic/claude-3.5-sonnet", // Default model
		"messages": reqMessages,
		"stream":   true, // Enable streaming
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshaling request: %w", err)
	}

	req, err := http.NewRequest("POST", o.endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.apiKey)
	req.Header.Set("HTTP-Referer", "https://github.com/speier/smith")
	req.Header.Set("X-Title", "Smith Agent System")

	resp, err := o.client.Do(req)
	if err != nil {
		return fmt.Errorf("api request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("api error (%d): %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse streaming response
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		var chunk struct {
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
				FinishReason *string `json:"finish_reason"`
			} `json:"choices"`
		}

		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue // Skip malformed chunks
		}

		if len(chunk.Choices) > 0 {
			content := chunk.Choices[0].Delta.Content
			done := chunk.Choices[0].FinishReason != nil

			if err := callback(&Response{
				Content: content,
				Done:    done,
			}); err != nil {
				return err
			}
		}
	}

	return scanner.Err()
}

func (o *OpenRouterProvider) GetModels() ([]Model, error) {
	// Popular models on OpenRouter
	// TODO: Could fetch this dynamically from https://openrouter.ai/api/v1/models
	return []Model{
		{
			ID:          "anthropic/claude-3.5-sonnet",
			Name:        "Claude 3.5 Sonnet",
			Description: "Anthropic's most capable model",
			ContextSize: 200000,
		},
		{
			ID:          "anthropic/claude-3-opus",
			Name:        "Claude 3 Opus",
			Description: "Powerful model for complex tasks",
			ContextSize: 200000,
		},
		{
			ID:          "anthropic/claude-3-haiku",
			Name:        "Claude 3 Haiku",
			Description: "Fast and efficient",
			ContextSize: 200000,
		},
		{
			ID:          "openai/gpt-4-turbo",
			Name:        "GPT-4 Turbo",
			Description: "OpenAI's latest GPT-4",
			ContextSize: 128000,
		},
		{
			ID:          "google/gemini-pro-1.5",
			Name:        "Gemini Pro 1.5",
			Description: "Google's advanced model",
			ContextSize: 1000000,
		},
		{
			ID:          "meta-llama/llama-3.1-405b-instruct",
			Name:        "Llama 3.1 405B",
			Description: "Meta's largest open model",
			ContextSize: 128000,
		},
	}, nil
}

func (o *OpenRouterProvider) GetName() string {
	return "OpenRouter"
}

func (o *OpenRouterProvider) RequiresAuth() bool {
	return true // Requires API key
}
