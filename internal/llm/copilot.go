package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/speier/smith/internal/config"
)

// CopilotProvider implements GitHub Copilot LLM access
type CopilotProvider struct {
	clientID        string
	deviceCodeURL   string
	accessTokenURL  string
	copilotTokenURL string
	client          *http.Client
	auth            *CopilotAuth
}

// CopilotAuth stores authentication tokens
type CopilotAuth struct {
	RefreshToken string    // GitHub OAuth token
	AccessToken  string    // Copilot API token
	ExpiresAt    time.Time // When access token expires
}

// DeviceCodeResponse from GitHub device flow
type DeviceCodeResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

// AccessTokenResponse from GitHub OAuth
type AccessTokenResponse struct {
	AccessToken string `json:"access_token,omitempty"`
	Error       string `json:"error,omitempty"`
	ErrorDesc   string `json:"error_description,omitempty"`
}

// CopilotTokenResponse from Copilot API
type CopilotTokenResponse struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"`
	RefreshIn int64  `json:"refresh_in"`
	Endpoints struct {
		API string `json:"api"`
	} `json:"endpoints"`
}

// NewCopilotProvider creates a GitHub Copilot provider
func NewCopilotProvider() *CopilotProvider {
	return &CopilotProvider{
		clientID:        "Iv1.b507a08c87ecfe98",
		deviceCodeURL:   "https://github.com/login/device/code",
		accessTokenURL:  "https://github.com/login/oauth/access_token",
		copilotTokenURL: "https://api.github.com/copilot_internal/v2/token",
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Authorize starts the device flow and returns instructions for the user
func (c *CopilotProvider) Authorize() (*DeviceCodeResponse, error) {
	payload := map[string]string{
		"client_id": c.clientID,
		"scope":     "read:user",
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", c.deviceCodeURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "GitHubCopilotChat/0.26.7")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("device code request: %w", err)
	}
	defer resp.Body.Close()

	var result DeviceCodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &result, nil
}

// PollForToken polls GitHub for the OAuth token after user authorization
func (c *CopilotProvider) PollForToken(deviceCode string) (string, error) {
	payload := map[string]string{
		"client_id":   c.clientID,
		"device_code": deviceCode,
		"grant_type":  "urn:ietf:params:oauth:grant-type:device_code",
	}
	body, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", c.accessTokenURL, bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "GitHubCopilotChat/0.26.7")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("token request: %w", err)
	}
	defer resp.Body.Close()

	var result AccessTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decoding response: %w", err)
	}

	if result.Error != "" {
		if result.Error == "authorization_pending" {
			return "pending", nil
		}
		return "", fmt.Errorf("oauth error: %s - %s", result.Error, result.ErrorDesc)
	}

	if result.AccessToken == "" {
		return "", fmt.Errorf("no access token in response")
	}

	return result.AccessToken, nil
}

// GetCopilotToken exchanges GitHub OAuth token for Copilot API token
func (c *CopilotProvider) GetCopilotToken(githubToken string) (*CopilotTokenResponse, error) {
	req, err := http.NewRequest("GET", c.copilotTokenURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+githubToken)
	req.Header.Set("User-Agent", "GitHubCopilotChat/0.26.7")
	req.Header.Set("Editor-Version", "vscode/1.99.3")
	req.Header.Set("Editor-Plugin-Version", "copilot-chat/0.26.7")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("copilot token request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("copilot token failed (%d): %s", resp.StatusCode, string(body))
	}

	var result CopilotTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &result, nil
}

// EnsureAuth ensures we have a valid Copilot token, refreshing if needed
func (c *CopilotProvider) EnsureAuth() error {
	if c.auth == nil {
		return fmt.Errorf("not authenticated - run authentication flow first")
	}

	// If access token is still valid, we're done
	if c.auth.AccessToken != "" && time.Now().Before(c.auth.ExpiresAt) {
		return nil
	}

	// Get new Copilot API token using refresh token
	copilotToken, err := c.GetCopilotToken(c.auth.RefreshToken)
	if err != nil {
		return fmt.Errorf("refreshing copilot token: %w", err)
	}

	// Update auth
	c.auth.AccessToken = copilotToken.Token
	c.auth.ExpiresAt = time.Unix(copilotToken.ExpiresAt, 0)

	return nil
}

// SetAuth sets the authentication tokens (after successful login)
func (c *CopilotProvider) SetAuth(refreshToken string) error {
	// Get initial Copilot token
	copilotToken, err := c.GetCopilotToken(refreshToken)
	if err != nil {
		return fmt.Errorf("getting initial copilot token: %w", err)
	}

	c.auth = &CopilotAuth{
		RefreshToken: refreshToken,
		AccessToken:  copilotToken.Token,
		ExpiresAt:    time.Unix(copilotToken.ExpiresAt, 0),
	}

	// Save to disk
	authData := map[string]interface{}{
		"refresh_token": refreshToken,
		"access_token":  copilotToken.Token,
		"expires_at":    copilotToken.ExpiresAt,
	}
	if err := config.SaveAuth("copilot", authData); err != nil {
		return fmt.Errorf("saving auth: %w", err)
	}

	return nil
}

// LoadAuth loads authentication from disk
func (c *CopilotProvider) LoadAuth() error {
	auth, err := config.LoadAuth("copilot")
	if err != nil {
		return fmt.Errorf("loading auth: %w", err)
	}

	if auth == nil {
		return fmt.Errorf("no authentication found - please run 'smith auth login'")
	}

	refreshToken, ok := auth.Data["refresh_token"].(string)
	if !ok {
		return fmt.Errorf("invalid refresh token in auth data")
	}

	accessToken, ok := auth.Data["access_token"].(string)
	if !ok {
		accessToken = ""
	}

	expiresAt, ok := auth.Data["expires_at"].(float64)
	if !ok {
		expiresAt = 0
	}

	c.auth = &CopilotAuth{
		RefreshToken: refreshToken,
		AccessToken:  accessToken,
		ExpiresAt:    time.Unix(int64(expiresAt), 0),
	}

	return nil
}

// Chat implements the Provider interface
func (c *CopilotProvider) Chat(messages []Message, tools []Tool) (*Response, error) {
	if err := c.EnsureAuth(); err != nil {
		return nil, err
	}

	// Copilot uses OpenAI-compatible chat completions API
	apiURL := "https://api.githubcopilot.com/chat/completions"

	// Convert messages to OpenAI format
	apiMessages := make([]map[string]string, len(messages))
	for i, msg := range messages {
		apiMessages[i] = map[string]string{
			"role":    msg.Role,
			"content": msg.Content,
		}
	}

	payload := map[string]interface{}{
		"messages": apiMessages,
		"model":    "gpt-4o", // Copilot model
		"stream":   false,
	}

	// Add tools if provided (for future function calling support)
	if len(tools) > 0 {
		// TODO: Convert tools to OpenAI function calling format
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.auth.AccessToken)
	req.Header.Set("User-Agent", "GitHubCopilotChat/0.26.7")
	req.Header.Set("Editor-Version", "vscode/1.99.3")
	req.Header.Set("Editor-Plugin-Version", "copilot-chat/0.26.7")

	resp, err := c.client.Do(req)
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
