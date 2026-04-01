// Package api 提供 Claude API 客户端
package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
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

// StreamEvent 流式事件
type StreamEvent struct {
	Type    string          `json:"type"`
	Delta   json.RawMessage `json:"delta,omitempty"`
	Content json.RawMessage `json:"content,omitempty"`
}

// NewClient 创建新的 API 客户端
func NewClient(apiKey, model string) *Client {
	return &Client{
		apiKey:  apiKey,
		baseURL: "https://api.anthropic.com/v1",
		model:   model,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// Chat 发送聊天请求
func (c *Client) Chat(ctx context.Context, messages []Message, tools []Tool) (*Message, error) {
	// TODO: 实现 API 调用
	return nil, fmt.Errorf("未实现")
}

// ChatStream 发送流式聊天请求
func (c *Client) ChatStream(ctx context.Context, messages []Message, tools []Tool) (<-chan StreamEvent, error) {
	// TODO: 实现流式 API 调用
	return nil, fmt.Errorf("未实现")
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
