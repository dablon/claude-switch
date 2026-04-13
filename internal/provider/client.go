package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Client struct {
	APIKey   string
	Provider ProviderType
	Model    string
	Endpoint string
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
}

type ChatResponse struct {
	Choices []Choice  `json:"choices"`
	Error   *APIError `json:"error,omitempty"`
}

type Choice struct {
	Message Message `json:"message"`
}

type APIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}

func NewClient(apiKey string, provider ProviderType, model, endpoint string) *Client {
	return &Client{
		APIKey:   apiKey,
		Provider: provider,
		Model:    model,
		Endpoint: endpoint,
	}
}

func (c *Client) Test() error {
	testMsg := "Hello"

	switch c.Provider {
	case ProviderAnthropic:
		return c.testAnthropic(testMsg)
	case ProviderOpenAI, ProviderGitHubCopilot:
		return c.testOpenAI(testMsg)
	case ProviderMinimax:
		return c.testMinimax(testMsg)
	case ProviderAzure:
		return c.testAzure(testMsg)
	case ProviderOllama, ProviderCustom:
		return c.testOllama(testMsg)
	default:
		return fmt.Errorf("unsupported provider: %s", c.Provider)
	}
}

func (c *Client) testAnthropic(msg string) error {
	url := "https://api.anthropic.com/v1/messages"

	body := map[string]interface{}{
		"model":      c.Model,
		"max_tokens": 10,
		"messages": []map[string]string{
			{"role": "user", "content": msg},
		},
	}

	bodyBytes, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
	req.Header.Set("x-api-key", c.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return c.parseError(resp.Body, "anthropic")
	}

	return nil
}

func (c *Client) testOpenAI(msg string) error {
	url := "https://api.openai.com/v1/chat/completions"

	reqBody := ChatRequest{
		Model: c.Model,
		Messages: []Message{
			{Role: "user", Content: msg},
		},
		MaxTokens: 10,
	}

	bodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return c.parseError(resp.Body, "openai")
	}

	return nil
}

func (c *Client) testMinimax(msg string) error {
	url := "https://api.minimax.chat/v1/text/chatcompletion_v2"

	reqBody := map[string]interface{}{
		"model": c.Model,
		"messages": []map[string]string{
			{"role": "user", "content": msg},
		},
		"max_tokens": 10,
	}

	bodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return c.parseError(resp.Body, "minimax")
	}

	return nil
}

func (c *Client) testAzure(msg string) error {
	if c.Endpoint == "" {
		return fmt.Errorf("Azure endpoint required")
	}

	url := c.Endpoint + "/openai/deployments/" + c.Model + "/chat/completions?api-version=2024-02-01"

	reqBody := ChatRequest{
		Model: c.Model,
		Messages: []Message{
			{Role: "user", Content: msg},
		},
		MaxTokens: 10,
	}

	bodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
	req.Header.Set("api-key", c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return c.parseError(resp.Body, "azure")
	}

	return nil
}

func (c *Client) testOllama(msg string) error {
	baseURL := c.Endpoint
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	// Normalize: strip trailing /api or /v1
	baseURL = strings.TrimSuffix(baseURL, "/api")
	baseURL = strings.TrimSuffix(baseURL, "/v1")

	// Try both native /chat and OpenAI-compatible /v1/chat/completions
	url := baseURL + "/api/chat"

	reqBody := map[string]interface{}{
		"model": c.Model,
		"messages": []map[string]string{
			{"role": "user", "content": msg},
		},
		"stream": false,
	}

	bodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	// Only add auth for local Ollama; remote ollama.com uses Bearer token
	isLocal := strings.HasPrefix(baseURL, "http://localhost") || strings.HasPrefix(baseURL, "http://127.0.0.1")
	if isLocal && c.APIKey != "" && c.APIKey != "ollama" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}
	if !isLocal && c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return c.parseError(resp.Body, "ollama")
	}

	return nil
}

func (c *Client) parseError(body io.Reader, provider string) error {
	data, _ := io.ReadAll(body)

	var errResp map[string]interface{}
	json.Unmarshal(data, &errResp)

	if errMsg, ok := errResp["error"].(map[string]interface{}); ok {
		if msg, ok := errMsg["message"].(string); ok {
			return fmt.Errorf("API error: %s", msg)
		}
	}

	return fmt.Errorf("API error (status %s): %s", provider, string(data))
}

func (c *Client) Chat(messages []Message) (string, error) {
	switch c.Provider {
	case ProviderAnthropic:
		return c.chatAnthropic(messages)
	case ProviderOpenAI, ProviderGitHubCopilot:
		return c.chatOpenAI(messages)
	case ProviderMinimax:
		return c.chatMinimax(messages)
	case ProviderAzure:
		return c.chatAzure(messages)
	case ProviderOllama, ProviderCustom:
		return c.chatOllama(messages)
	default:
		return "", fmt.Errorf("unsupported provider: %s", c.Provider)
	}
}

func (c *Client) chatAnthropic(msgs []Message) (string, error) {
	url := "https://api.anthropic.com/v1/messages"

	anthropicMsgs := make([]map[string]string, len(msgs))
	for i, m := range msgs {
		anthropicMsgs[i] = map[string]string{"role": m.Role, "content": m.Content}
	}

	body := map[string]interface{}{
		"model":      c.Model,
		"max_tokens": 4096,
		"messages":   anthropicMsgs,
	}

	bodyBytes, _ := json.Marshal(body)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
	req.Header.Set("x-api-key", c.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		var errResp map[string]interface{}
		json.Unmarshal(data, &errResp)
		if errMsg, ok := errResp["error"].(map[string]interface{}); ok {
			if msg, ok := errMsg["message"].(string); ok {
				return "", fmt.Errorf("API error: %s", msg)
			}
		}
		return "", fmt.Errorf("API error: %s", string(data))
	}

	var result struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}

	json.Unmarshal(data, &result)

	if len(result.Content) > 0 {
		return result.Content[0].Text, nil
	}

	return "", nil
}

func (c *Client) chatOpenAI(msgs []Message) (string, error) {
	url := "https://api.openai.com/v1/chat/completions"

	reqBody := ChatRequest{
		Model:     c.Model,
		Messages:  msgs,
		MaxTokens: 4096,
	}

	bodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)

	var chatResp ChatResponse
	json.Unmarshal(data, &chatResp)

	if chatResp.Error != nil {
		return "", fmt.Errorf("API error: %s", chatResp.Error.Message)
	}

	if len(chatResp.Choices) > 0 {
		return chatResp.Choices[0].Message.Content, nil
	}

	return "", nil
}

func (c *Client) chatMinimax(msgs []Message) (string, error) {
	url := "https://api.minimax.chat/v1/text/chatcompletion_v2"

	minimaxMsgs := make([]map[string]string, len(msgs))
	for i, m := range msgs {
		minimaxMsgs[i] = map[string]string{"role": m.Role, "content": m.Content}
	}

	reqBody := map[string]interface{}{
		"model":      c.Model,
		"messages":   minimaxMsgs,
		"max_tokens": 4096,
	}

	bodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error *APIError `json:"error,omitempty"`
	}

	json.Unmarshal(data, &result)

	if result.Error != nil {
		return "", fmt.Errorf("API error: %s", result.Error.Message)
	}

	if len(result.Choices) > 0 {
		return result.Choices[0].Message.Content, nil
	}

	return "", nil
}

func (c *Client) chatAzure(msgs []Message) (string, error) {
	if c.Endpoint == "" {
		return "", fmt.Errorf("Azure endpoint required")
	}

	url := c.Endpoint + "/openai/deployments/" + c.Model + "/chat/completions?api-version=2024-02-01"

	reqBody := ChatRequest{
		Model:     c.Model,
		Messages:  msgs,
		MaxTokens: 4096,
	}

	bodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
	req.Header.Set("api-key", c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)

	var chatResp ChatResponse
	json.Unmarshal(data, &chatResp)

	if chatResp.Error != nil {
		return "", fmt.Errorf("API error: %s", chatResp.Error.Message)
	}

	if len(chatResp.Choices) > 0 {
		return chatResp.Choices[0].Message.Content, nil
	}

	return "", nil
}

func (c *Client) chatOllama(msgs []Message) (string, error) {
	baseURL := c.Endpoint
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	// Normalize: strip trailing /api or /v1
	baseURL = strings.TrimSuffix(baseURL, "/api")
	baseURL = strings.TrimSuffix(baseURL, "/v1")

	// Try both native /chat and OpenAI-compatible /v1/chat/completions
	url := baseURL + "/api/chat"

	ollamaMsgs := make([]map[string]string, len(msgs))
	for i, m := range msgs {
		ollamaMsgs[i] = map[string]string{"role": m.Role, "content": m.Content}
	}

	reqBody := map[string]interface{}{
		"model":    c.Model,
		"messages":  ollamaMsgs,
		"stream":    false,
	}

	bodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	// Only add auth for local Ollama; remote ollama.com uses Bearer token
	isLocal := strings.HasPrefix(baseURL, "http://localhost") || strings.HasPrefix(baseURL, "http://127.0.0.1")
	if isLocal && c.APIKey != "" && c.APIKey != "ollama" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}
	if !isLocal && c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)

	var result struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
		Error   *APIError `json:"error,omitempty"`
	}

	json.Unmarshal(data, &result)

	if result.Error != nil {
		return "", fmt.Errorf("API error: %s", result.Error.Message)
	}

	return result.Message.Content, nil
}

func FormatMessage(role, content string) string {
	switch role {
	case "user":
		return fmt.Sprintf("\033[32mYou:\033[0m %s", content)
	case "assistant":
		lines := strings.Split(content, "\n")
		formatted := "\033[36mAssistant:\033[0m"
		for _, line := range lines {
			formatted += "\n" + line
		}
		return formatted
	default:
		return content
	}
}
