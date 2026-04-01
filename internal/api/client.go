// Package api 提供 Claude API 客户端
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

// Client API 客户端
type Client struct {
	apiKey     string
	baseURL    string
	model      string
	httpClient *http.Client
}

// Message 消息类型
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Tool 工具定义
type Tool struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"input_schema"`
}

// ContentBlock 内容块
type ContentBlock struct {
	Type  string          `json:"type"`
	Text  string          `json:"text,omitempty"`
	ID    string          `json:"id,omitempty"`
	Name  string          `json:"name,omitempty"`
	Input json.RawMessage `json:"input,omitempty"`
}

// Response API 响应
type Response struct {
	Role    string         `json:"role"`
	Content []ContentBlock `json:"content"`
}

// StreamEvent 流式事件
type StreamEvent struct {
	Type    string `json:"type"`
	Delta   Delta  `json:"delta,omitempty"`
	Content string `json:"content,omitempty"`
	Index   int    `json:"index,omitempty"`
}

// Delta 增量内容
type Delta struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// NewClient 创建新的 API 客户端
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

// ChatWithBlocks 发送聊天请求并返回内容块
func (c *Client) ChatWithBlocks(ctx context.Context, messages []Message, tools []Tool) (*Response, error) {
	// 构建请求体
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

	reqBody := map[string]interface{}{
		"model":      c.model,
		"max_tokens": 4096,
		"messages":   messages,
		"tools":      toolDefs,
	}

	if c.apiKey == "" {
		return &Response{
			Role: "assistant",
			Content: []ContentBlock{
				{
					Type: "text",
					Text: "⚠️ 未配置 API Key。请在 ~/.config/claude/config.json 中设置 api_key。",
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

// Chat 发送聊天请求（兼容旧接口）
func (c *Client) Chat(ctx context.Context, messages []Message, tools []Tool) (*Message, error) {
	resp, err := c.ChatWithBlocks(ctx, messages, tools)
	if err != nil {
		return nil, err
	}

	// 构建响应消息文本
	var content strings.Builder
	for _, block := range resp.Content {
		switch block.Type {
		case "text":
			content.WriteString(block.Text)
		case "tool_use":
			content.WriteString(fmt.Sprintf("\n🔧 调用工具: %s\n", block.Name))
			content.WriteString(fmt.Sprintf("参数: %s\n", string(block.Input)))
		}
	}

	return &Message{
		Role:    "assistant",
		Content: content.String(),
	}, nil
}

// ChatStream 发送流式聊天请求
func (c *Client) ChatStream(ctx context.Context, messages []Message, tools []Tool) (<-chan StreamEvent, error) {
	if c.apiKey == "" {
		ch := make(chan StreamEvent, 1)
		ch <- StreamEvent{
			Type:    "error",
			Content: "未配置 API Key",
		}
		close(ch)
		return ch, nil
	}

	// 构建请求体
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

	reqBody := map[string]interface{}{
		"model":      c.model,
		"max_tokens": 4096,
		"messages":   messages,
		"tools":      toolDefs,
		"stream":     true,
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

// SetProvider 设置 API 提供商
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
