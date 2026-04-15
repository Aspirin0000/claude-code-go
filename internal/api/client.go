// Package api provides a Claude API client.
package api

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client is an API client.
type Client struct {
	apiKey     string
	baseURL    string
	model      string
	httpClient *http.Client
}

// Message represents a chat message.
type Message struct {
	Role    string         `json:"role"`
	Content string         `json:"content,omitempty"`
	Blocks  []ContentBlock `json:"-"`
}

// Tool defines an available tool.
type Tool struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"input_schema"`
}

// ContentBlock represents a block in the API response or request.
type ContentBlock struct {
	Type      string          `json:"type"`
	Text      string          `json:"text,omitempty"`
	Content   string          `json:"content,omitempty"`
	ID        string          `json:"id,omitempty"`
	Name      string          `json:"name,omitempty"`
	Input     json.RawMessage `json:"input,omitempty"`
	ToolUseID string          `json:"tool_use_id,omitempty"`
}

// Response is an API response.
type Response struct {
	Role    string         `json:"role"`
	Content []ContentBlock `json:"content"`
}

// StreamEvent is a streaming event.
type StreamEvent struct {
	Type    string `json:"type"`
	Delta   Delta  `json:"delta,omitempty"`
	Content string `json:"content,omitempty"`
	Index   int    `json:"index,omitempty"`
}

// Delta is streamed delta content.
type Delta struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// NewClient creates a new API client.
func NewClient(apiKey, model string) *Client {
	return &Client{
		apiKey:  apiKey,
		baseURL: "https://api.anthropic.com/v1",
		model:   model,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// ChatWithBlocks sends a chat request and returns content blocks.
func (c *Client) ChatWithBlocks(ctx context.Context, messages []Message, tools []Tool) (*Response, error) {
	// Build request body.
	type toolDef struct {
		Name        string          `json:"name"`
		Description string          `json:"description"`
		InputSchema json.RawMessage `json:"input_schema"`
	}

	toolDefs := make([]toolDef, len(tools))
	for i, t := range tools {
		toolDefs[i] = toolDef{
			Name:        t.Name,
			Description: t.Description,
			InputSchema: t.InputSchema,
		}
	}

	type reqMsg struct {
		Role    string      `json:"role"`
		Content interface{} `json:"content"`
	}

	reqMessages := make([]reqMsg, len(messages))
	for i, msg := range messages {
		if len(msg.Blocks) > 0 {
			reqMessages[i] = reqMsg{Role: msg.Role, Content: msg.Blocks}
		} else {
			reqMessages[i] = reqMsg{Role: msg.Role, Content: msg.Content}
		}
	}

	reqBody := map[string]interface{}{
		"model":      c.model,
		"max_tokens": 4096,
		"messages":   reqMessages,
	}
	if len(toolDefs) > 0 {
		reqBody["tools"] = toolDefs
	}

	if c.apiKey == "" {
		return &Response{
			Role: "assistant",
			Content: []ContentBlock{
				{
					Type: "text",
					Text: "⚠️ API key is not configured. Set api_key in ~/.config/claude/config.json.",
				},
			},
		}, nil
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/messages", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var result Response
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result, nil
}

// Chat sends a chat request using the legacy message interface.
func (c *Client) Chat(ctx context.Context, messages []Message, tools []Tool) (*Message, error) {
	resp, err := c.ChatWithBlocks(ctx, messages, tools)
	if err != nil {
		return nil, err
	}

	// Build response message text.
	var content strings.Builder
	for _, block := range resp.Content {
		switch block.Type {
		case "text":
			content.WriteString(block.Text)
		case "tool_use":
			content.WriteString(fmt.Sprintf("\n🔧 Tool call: %s\n", block.Name))
			content.WriteString(fmt.Sprintf("Arguments: %s\n", string(block.Input)))
		}
	}

	return &Message{
		Role:    "assistant",
		Content: content.String(),
	}, nil
}

// ChatStream sends a streaming chat request.
func (c *Client) ChatStream(ctx context.Context, messages []Message, tools []Tool) (<-chan StreamEvent, error) {
	if c.apiKey == "" {
		ch := make(chan StreamEvent, 1)
		ch <- StreamEvent{
			Type:    "error",
			Content: "API key is not configured",
		}
		close(ch)
		return ch, nil
	}

	// Build request body.
	type toolDef struct {
		Name        string          `json:"name"`
		Description string          `json:"description"`
		InputSchema json.RawMessage `json:"input_schema"`
	}

	toolDefs := make([]toolDef, len(tools))
	for i, t := range tools {
		toolDefs[i] = toolDef{
			Name:        t.Name,
			Description: t.Description,
			InputSchema: t.InputSchema,
		}
	}

	type reqMsg struct {
		Role    string      `json:"role"`
		Content interface{} `json:"content"`
	}

	reqMessages := make([]reqMsg, len(messages))
	for i, msg := range messages {
		if len(msg.Blocks) > 0 {
			reqMessages[i] = reqMsg{Role: msg.Role, Content: msg.Blocks}
		} else {
			reqMessages[i] = reqMsg{Role: msg.Role, Content: msg.Content}
		}
	}

	reqBody := map[string]interface{}{
		"model":      c.model,
		"max_tokens": 4096,
		"messages":   reqMessages,
	}
	if len(toolDefs) > 0 {
		reqBody["tools"] = toolDefs
	}
	reqBody["stream"] = true

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/messages", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("API error %d", resp.StatusCode)
	}

	ch := make(chan StreamEvent)

	go func() {
		defer resp.Body.Close()
		defer close(ch)

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

			var event StreamEvent
			if err := json.Unmarshal([]byte(data), &event); err != nil {
				continue
			}

			select {
			case ch <- event:
			case <-ctx.Done():
				return
			}
		}
	}()

	return ch, nil
}

// SetProvider sets the API provider.
func (c *Client) SetProvider(provider string) {
	switch provider {
	case "anthropic":
		c.baseURL = "https://api.anthropic.com/v1"
	case "bedrock":
		c.baseURL = "https://bedrock-runtime.us-east-1.amazonaws.com"
	case "vertex":
		c.baseURL = "https://us-central1-aiplatform.googleapis.com"
	}
}
