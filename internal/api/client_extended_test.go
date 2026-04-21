package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// ============================================================================
// Client Configuration Tests
// ============================================================================

func TestNewClient(t *testing.T) {
	client := NewClient("test-key", "claude-test")
	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if client.apiKey != "test-key" {
		t.Errorf("expected api key 'test-key', got %q", client.apiKey)
	}
	if client.model != "claude-test" {
		t.Errorf("expected model 'claude-test', got %q", client.model)
	}
	if client.baseURL != "https://api.anthropic.com/v1" {
		t.Errorf("expected default base URL, got %q", client.baseURL)
	}
	if client.maxRetries != 3 {
		t.Errorf("expected 3 retries, got %d", client.maxRetries)
	}
	if client.httpClient == nil {
		t.Error("expected non-nil http client")
	}
}

func TestSetMaxRetries(t *testing.T) {
	client := NewClient("key", "model")
	client.SetMaxRetries(5)
	if client.maxRetries != 5 {
		t.Errorf("expected 5 retries, got %d", client.maxRetries)
	}
}

func TestSetBaseURL(t *testing.T) {
	client := NewClient("key", "model")
	client.SetBaseURL("https://custom.api.com")
	if client.baseURL != "https://custom.api.com" {
		t.Errorf("expected custom URL, got %q", client.baseURL)
	}
}

func TestGetModel(t *testing.T) {
	client := NewClient("key", "claude-test")
	if client.GetModel() != "claude-test" {
		t.Errorf("expected 'claude-test', got %q", client.GetModel())
	}
}

func TestGetProvider(t *testing.T) {
	client := NewClient("key", "model")
	if client.GetProvider() != "" {
		t.Errorf("expected empty provider, got %q", client.GetProvider())
	}

	client.SetProvider("anthropic")
	if client.GetProvider() != "anthropic" {
		t.Errorf("expected 'anthropic', got %q", client.GetProvider())
	}
}

func TestSetProviderAnthropic(t *testing.T) {
	client := NewClient("key", "model")
	client.SetProvider("anthropic")
	if client.baseURL != "https://api.anthropic.com/v1" {
		t.Errorf("expected anthropic URL, got %q", client.baseURL)
	}
}

func TestSetProviderUnknown(t *testing.T) {
	client := NewClient("key", "model")
	client.SetProvider("unknown")
	// Should not change baseURL
	if client.baseURL != "https://api.anthropic.com/v1" {
		t.Errorf("expected default URL to remain, got %q", client.baseURL)
	}
}

// ============================================================================
// Chat Method Tests
// ============================================================================

func TestChat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Response{
			Role: "assistant",
			Content: []ContentBlock{
				{Type: "text", Text: "Hello!"},
			},
		})
	}))
	defer server.Close()

	client := NewClient("test-key", "claude-test")
	client.baseURL = server.URL

	resp, err := client.Chat(context.Background(), []Message{{Role: "user", Content: "Hi"}}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Role != "assistant" {
		t.Errorf("expected role assistant, got %s", resp.Role)
	}
	if !strings.Contains(resp.Content, "Hello!") {
		t.Errorf("expected content to contain 'Hello!', got %q", resp.Content)
	}
}

func TestChatWithToolUse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Response{
			Role: "assistant",
			Content: []ContentBlock{
				{Type: "text", Text: "Let me check."},
				{Type: "tool_use", Name: "bash", Input: json.RawMessage(`{"command":"ls"}`)},
			},
		})
	}))
	defer server.Close()

	client := NewClient("test-key", "claude-test")
	client.baseURL = server.URL

	resp, err := client.Chat(context.Background(), []Message{{Role: "user", Content: "List files"}}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(resp.Content, "🔧 Tool call: bash") {
		t.Errorf("expected tool call info in content, got %q", resp.Content)
	}
}

// ============================================================================
// ChatStream Tests
// ============================================================================

func TestChatStream(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "data: {\"type\": \"content_block_start\", \"index\": 0, \"content_block\": {\"type\": \"text\"}}")
		fmt.Fprintln(w, "data: {\"type\": \"content_block_delta\", \"index\": 0, \"delta\": {\"type\": \"text_delta\", \"text\": \"Hello\"}}")
		fmt.Fprintln(w, "data: {\"type\": \"message_stop\"}")
		fmt.Fprintln(w, "data: [DONE]")
	}))
	defer server.Close()

	client := NewClient("test-key", "claude-test")
	client.baseURL = server.URL

	ch, err := client.ChatStream(context.Background(), []Message{{Role: "user", Content: "Hi"}}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	events := []StreamEvent{}
	for event := range ch {
		events = append(events, event)
	}

	if len(events) < 2 {
		t.Errorf("expected at least 2 events, got %d", len(events))
	}
}

func TestChatStreamNoAPIKey(t *testing.T) {
	client := NewClient("", "claude-test")
	ch, err := client.ChatStream(context.Background(), []Message{{Role: "user", Content: "Hi"}}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	events := []StreamEvent{}
	for event := range ch {
		events = append(events, event)
	}

	if len(events) != 1 {
		t.Errorf("expected 1 event, got %d", len(events))
	}
	if events[0].Type != "error" {
		t.Errorf("expected error event, got %s", events[0].Type)
	}
}

func TestChatStreamAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	client := NewClient("test-key", "claude-test")
	client.baseURL = server.URL

	_, err := client.ChatStream(context.Background(), []Message{{Role: "user", Content: "Hi"}}, nil)
	if err == nil {
		t.Fatal("expected error for unauthorized")
	}
}

// ============================================================================
// Retry Tests (Extended)
// ============================================================================

func TestChatWithBlocksRetryOnNetworkError(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 2 {
			// Close connection to simulate network error
			hj, ok := w.(http.Hijacker)
			if ok {
				conn, _, _ := hj.Hijack()
				conn.Close()
				return
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Response{
			Role:    "assistant",
			Content: []ContentBlock{{Type: "text", Text: "Success"}},
		})
	}))
	defer server.Close()

	client := NewClient("test-key", "claude-test")
	client.baseURL = server.URL
	client.SetMaxRetries(3)

	resp, err := client.ChatWithBlocks(context.Background(), []Message{{Role: "user", Content: "Hi"}}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Content[0].Text != "Success" {
		t.Errorf("unexpected response: %s", resp.Content[0].Text)
	}
}

func TestChatWithBlocksRetryExhausted(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"persistent failure"}`))
	}))
	defer server.Close()

	client := NewClient("test-key", "claude-test")
	client.baseURL = server.URL
	client.SetMaxRetries(2)

	_, err := client.ChatWithBlocks(context.Background(), []Message{{Role: "user", Content: "Hi"}}, nil)
	if err == nil {
		t.Fatal("expected error after retries exhausted")
	}
	if attempts != 3 { // initial + 2 retries
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestChatWithBlocksContextCancelled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient("test-key", "claude-test")
	client.baseURL = server.URL

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := client.ChatWithBlocks(ctx, []Message{{Role: "user", Content: "Hi"}}, nil)
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

// ============================================================================
// CollectStreamResponse Tests (Extended)
// ============================================================================

func TestCollectStreamResponseEmpty(t *testing.T) {
	ch := make(chan StreamEvent)
	close(ch)

	resp := CollectStreamResponse(ch)
	if len(resp.Content) != 0 {
		t.Errorf("expected 0 content blocks, got %d", len(resp.Content))
	}
	if resp.Role != "assistant" {
		t.Errorf("expected role assistant, got %s", resp.Role)
	}
}

func TestCollectStreamResponseMultipleBlocks(t *testing.T) {
	ch := make(chan StreamEvent, 6)
	ch <- StreamEvent{Type: "content_block_start", Index: 0, ContentBlock: &ContentBlock{Type: "text"}}
	ch <- StreamEvent{Type: "content_block_delta", Index: 0, Delta: Delta{Type: "text_delta", Text: "First"}}
	ch <- StreamEvent{Type: "content_block_start", Index: 1, ContentBlock: &ContentBlock{Type: "text"}}
	ch <- StreamEvent{Type: "content_block_delta", Index: 1, Delta: Delta{Type: "text_delta", Text: "Second"}}
	ch <- StreamEvent{Type: "content_block_stop", Index: 0}
	ch <- StreamEvent{Type: "content_block_stop", Index: 1}
	close(ch)

	resp := CollectStreamResponse(ch)
	if len(resp.Content) != 2 {
		t.Fatalf("expected 2 content blocks, got %d", len(resp.Content))
	}
	if resp.Content[0].Text != "First" {
		t.Errorf("expected 'First', got %q", resp.Content[0].Text)
	}
	if resp.Content[1].Text != "Second" {
		t.Errorf("expected 'Second', got %q", resp.Content[1].Text)
	}
}

func TestCollectStreamResponseInvalidIndex(t *testing.T) {
	ch := make(chan StreamEvent, 3)
	ch <- StreamEvent{Type: "content_block_start", Index: 0, ContentBlock: &ContentBlock{Type: "text"}}
	ch <- StreamEvent{Type: "content_block_delta", Index: -1, Delta: Delta{Type: "text_delta", Text: "Invalid"}}
	ch <- StreamEvent{Type: "content_block_delta", Index: 5, Delta: Delta{Type: "text_delta", Text: "Out of range"}}
	close(ch)

	resp := CollectStreamResponse(ch)
	if len(resp.Content) != 1 {
		t.Fatalf("expected 1 content block, got %d", len(resp.Content))
	}
	if resp.Content[0].Text != "" {
		t.Errorf("expected empty text, got %q", resp.Content[0].Text)
	}
}

func TestCollectStreamResponseToolUseWithoutJSON(t *testing.T) {
	ch := make(chan StreamEvent, 4)
	ch <- StreamEvent{Type: "content_block_start", Index: 0, ContentBlock: &ContentBlock{Type: "tool_use", ID: "tu_1", Name: "bash"}}
	ch <- StreamEvent{Type: "content_block_delta", Index: 0, Delta: Delta{Type: "input_json_delta", PartialJSON: `{"cmd":"ls"}`}}
	ch <- StreamEvent{Type: "content_block_stop", Index: 0}
	ch <- StreamEvent{Type: "message_stop"}
	close(ch)

	resp := CollectStreamResponse(ch)
	if len(resp.Content) != 1 {
		t.Fatalf("expected 1 content block, got %d", len(resp.Content))
	}
	if string(resp.Content[0].Input) != `{"cmd":"ls"}` {
		t.Errorf("unexpected input: %s", string(resp.Content[0].Input))
	}
}

// ============================================================================
// Data Types Tests
// ============================================================================

func TestMessage(t *testing.T) {
	msg := Message{
		Role:    "user",
		Content: "Hello",
		Blocks:  []ContentBlock{{Type: "text", Text: "Hello"}},
	}
	if msg.Role != "user" {
		t.Errorf("expected role 'user', got %q", msg.Role)
	}
}

func TestTool(t *testing.T) {
	tool := Tool{
		Name:        "bash",
		Description: "Run bash commands",
		InputSchema: json.RawMessage(`{"type":"object"}`),
	}
	if tool.Name != "bash" {
		t.Errorf("expected name 'bash', got %q", tool.Name)
	}
}

func TestContentBlock(t *testing.T) {
	block := ContentBlock{
		Type:      "tool_use",
		ID:        "tu_1",
		Name:      "bash",
		Input:     json.RawMessage(`{"cmd":"ls"}`),
		ToolUseID: "tu_1",
	}
	if block.Type != "tool_use" {
		t.Errorf("expected type 'tool_use', got %q", block.Type)
	}
}

func TestResponse(t *testing.T) {
	resp := Response{
		Role: "assistant",
		Content: []ContentBlock{
			{Type: "text", Text: "Hello"},
		},
	}
	if resp.Role != "assistant" {
		t.Errorf("expected role 'assistant', got %q", resp.Role)
	}
}

func TestStreamEvent(t *testing.T) {
	event := StreamEvent{
		Type:         "content_block_delta",
		Index:        0,
		Delta:        Delta{Type: "text_delta", Text: "Hello"},
		ContentBlock: &ContentBlock{Type: "text"},
	}
	if event.Type != "content_block_delta" {
		t.Errorf("expected type 'content_block_delta', got %q", event.Type)
	}
}

func TestDelta(t *testing.T) {
	delta := Delta{
		Type:        "text_delta",
		Text:        "Hello",
		PartialJSON: `{"key":"value"}`,
	}
	if delta.Type != "text_delta" {
		t.Errorf("expected type 'text_delta', got %q", delta.Type)
	}
}

// ============================================================================
// Backoff Tests (Extended)
// ============================================================================

func TestBackoffMaxDelay(t *testing.T) {
	// Test that backoff doesn't exceed max delay
	b10 := backoff(10)
	if b10 > 31*time.Second {
		t.Errorf("backoff(10) should not exceed 31s, got %v", b10)
	}
}

func TestBackoffIncreases(t *testing.T) {
	for i := 0; i < 5; i++ {
		b1 := backoff(i)
		b2 := backoff(i + 1)
		if b2 <= b1 {
			t.Errorf("expected backoff(%d) > backoff(%d), got %v vs %v", i+1, i, b2, b1)
		}
	}
}

// ============================================================================
// IsRetriable Tests (Extended)
// ============================================================================

func TestIsRetriableAll5xx(t *testing.T) {
	for code := 500; code <= 599; code++ {
		if !isRetriable(nil, code) {
			t.Errorf("expected %d to be retriable", code)
		}
	}
}

func TestIsRetriableNotRetriable(t *testing.T) {
	notRetriable := []int{200, 201, 400, 401, 403, 404}
	for _, code := range notRetriable {
		if isRetriable(nil, code) {
			t.Errorf("expected %d to not be retriable", code)
		}
	}
}
