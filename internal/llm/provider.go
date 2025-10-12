package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// Provider interface for different LLM providers
type Provider interface {
	Chat(messages []Message, tools []Tool) (*Response, error)
}

type Message struct {
	Role    string `json:"role"` // "user", "assistant", "system"
	Content string `json:"content"`
}

type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

type Response struct {
	Content   string
	ToolCalls []ToolCall
	Done      bool
}

type ToolCall struct {
	Name  string
	Input map[string]interface{}
}

// OpenAIProvider uses OpenAI-compatible APIs (OpenAI, Groq, OpenRouter, etc.)
type OpenAIProvider struct {
	apiKey   string
	endpoint string
	model    string
}

// NewOpenAI creates an OpenAI provider
// Reads API key from OPENAI_API_KEY env var
func NewOpenAI() *OpenAIProvider {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		fmt.Println("⚠️  OPENAI_API_KEY not set. LLM features will be limited.")
	}

	return &OpenAIProvider{
		apiKey:   apiKey,
		endpoint: "https://api.openai.com/v1/chat/completions",
		model:    "gpt-4o-mini", // Fast and cheap for testing
	}
}

// NewOpenAICustom creates a custom OpenAI-compatible provider
func NewOpenAICustom(apiKey, endpoint, model string) *OpenAIProvider {
	return &OpenAIProvider{
		apiKey:   apiKey,
		endpoint: endpoint,
		model:    model,
	}
}

func (p *OpenAIProvider) Chat(messages []Message, tools []Tool) (*Response, error) {
	if p.apiKey == "" {
		return &Response{
			Content: "⚠️  No API key configured. Set OPENAI_API_KEY environment variable.",
			Done:    true,
		}, nil
	}

	reqBody := map[string]interface{}{
		"model":    p.model,
		"messages": messages,
	}

	if len(tools) > 0 {
		reqBody["tools"] = convertTools(tools)
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", p.endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var apiResp struct {
		Choices []struct {
			Message struct {
				Role      string `json:"role"`
				Content   string `json:"content"`
				ToolCalls []struct {
					Function struct {
						Name      string          `json:"name"`
						Arguments json.RawMessage `json:"arguments"`
					} `json:"function"`
				} `json:"tool_calls,omitempty"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(apiResp.Choices) == 0 {
		return nil, fmt.Errorf("no response from API")
	}

	choice := apiResp.Choices[0]
	response := &Response{
		Content: choice.Message.Content,
		Done:    choice.FinishReason == "stop",
	}

	// Parse tool calls if present
	for _, tc := range choice.Message.ToolCalls {
		var input map[string]interface{}
		json.Unmarshal(tc.Function.Arguments, &input)
		response.ToolCalls = append(response.ToolCalls, ToolCall{
			Name:  tc.Function.Name,
			Input: input,
		})
	}

	return response, nil
}

// OllamaProvider for local LLMs
type OllamaProvider struct {
	endpoint string
	model    string
}

// NewOllama creates an Ollama provider for local LLMs
func NewOllama(model string) *OllamaProvider {
	if model == "" {
		model = "llama3.2" // Default model
	}

	return &OllamaProvider{
		endpoint: "http://localhost:11434/api/chat",
		model:    model,
	}
}

func (p *OllamaProvider) Chat(messages []Message, tools []Tool) (*Response, error) {
	reqBody := map[string]interface{}{
		"model":    p.model,
		"messages": messages,
		"stream":   false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(p.endpoint, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("Ollama request failed (is Ollama running?): %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Ollama error (status %d): %s", resp.StatusCode, string(body))
	}

	var apiResp struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		Done bool `json:"done"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &Response{
		Content: apiResp.Message.Content,
		Done:    apiResp.Done,
	}, nil
}

func convertTools(tools []Tool) []map[string]interface{} {
	result := make([]map[string]interface{}, len(tools))
	for i, tool := range tools {
		result[i] = map[string]interface{}{
			"type": "function",
			"function": map[string]interface{}{
				"name":        tool.Name,
				"description": tool.Description,
				"parameters":  tool.Parameters,
			},
		}
	}
	return result
}
