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
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("no response from API")
	}

	return &Response{
		Content:          result.Choices[0].Message.Content,
		PromptTokens:     result.Usage.PromptTokens,
		CompletionTokens: result.Usage.CompletionTokens,
		TotalTokens:      result.Usage.TotalTokens,
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
	var totalPromptTokens, totalCompletionTokens, totalTokens int

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
			Usage struct {
				PromptTokens     int `json:"prompt_tokens"`
				CompletionTokens int `json:"completion_tokens"`
				TotalTokens      int `json:"total_tokens"`
			} `json:"usage"`
		}

		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue // Skip malformed chunks
		}

		// Capture usage data from any chunk that includes it
		if chunk.Usage.TotalTokens > 0 {
			totalPromptTokens = chunk.Usage.PromptTokens
			totalCompletionTokens = chunk.Usage.CompletionTokens
			totalTokens = chunk.Usage.TotalTokens
		}

		if len(chunk.Choices) > 0 {
			content := chunk.Choices[0].Delta.Content
			done := chunk.Choices[0].FinishReason != nil

			if err := callback(&Response{
				Content:          content,
				Done:             done,
				PromptTokens:     totalPromptTokens,
				CompletionTokens: totalCompletionTokens,
				TotalTokens:      totalTokens,
			}); err != nil {
				return err
			}
		}
	}

	return scanner.Err()
}

func (o *OpenRouterProvider) GetModels() ([]Model, error) {
	// Check authentication first
	if o.apiKey == "" {
		return nil, fmt.Errorf("not authenticated - set OPENROUTER_API_KEY environment variable")
	}

	// Fetch models from OpenRouter API
	apiURL := "https://openrouter.ai/api/v1/models"

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+o.apiKey)
	req.Header.Set("HTTP-Referer", "https://github.com/speier/smith")
	req.Header.Set("X-Title", "Smith Agent System")

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching models: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Data []struct {
			ID          string `json:"id"`
			Name        string `json:"name"`
			Description string `json:"description"`
			Pricing     struct {
				Prompt     string `json:"prompt"`
				Completion string `json:"completion"`
			} `json:"pricing"`
			ContextLength int `json:"context_length"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	// Convert to our Model format with intelligent sorting
	// Priority: Claude 3.5, GPT-4, Claude 3, GPT-3.5, then others
	models := make([]Model, 0, len(result.Data))

	// Categorize models by family/quality
	var claude35Models []Model
	var gpt4Models []Model
	var claude3Models []Model
	var gpt35Models []Model
	var otherModels []Model

	for _, m := range result.Data {
		model := Model{
			ID:          m.ID,
			Name:        m.Name,
			Description: m.Description,
			ContextSize: m.ContextLength,
		}

		idLower := strings.ToLower(m.ID)
		// Categorize by model family (most capable first)
		if strings.Contains(idLower, "claude-3.5") || strings.Contains(idLower, "claude-3-5") {
			claude35Models = append(claude35Models, model)
		} else if strings.Contains(idLower, "gpt-4") {
			gpt4Models = append(gpt4Models, model)
		} else if strings.Contains(idLower, "claude-3") {
			claude3Models = append(claude3Models, model)
		} else if strings.Contains(idLower, "gpt-3.5") {
			gpt35Models = append(gpt35Models, model)
		} else {
			otherModels = append(otherModels, model)
		}
	}

	// Sort within each category alphabetically by ID
	sortModelsByID := func(models []Model) {
		for i := 0; i < len(models); i++ {
			for j := i + 1; j < len(models); j++ {
				if models[i].ID > models[j].ID {
					models[i], models[j] = models[j], models[i]
				}
			}
		}
	}

	sortModelsByID(claude35Models)
	sortModelsByID(gpt4Models)
	sortModelsByID(claude3Models)
	sortModelsByID(gpt35Models)
	sortModelsByID(otherModels)

	// Combine: Best models first
	models = append(models, claude35Models...)
	models = append(models, gpt4Models...)
	models = append(models, claude3Models...)
	models = append(models, gpt35Models...)
	models = append(models, otherModels...)

	if len(models) == 0 {
		return nil, fmt.Errorf("no models available")
	}

	return models, nil
}

func (o *OpenRouterProvider) GetName() string {
	return "OpenRouter"
}

func (o *OpenRouterProvider) RequiresAuth() bool {
	return true // Requires API key
}
