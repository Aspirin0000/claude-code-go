package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestChatWithBlocks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/messages" {
			t.Errorf("expected /messages, got %s", r.URL.Path)
		}
		if r.Header.Get("x-api-key") != "test-key" {
			t.Errorf("expected api key test-key, got %s", r.Header.Get("x-api-key"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Response{
			Role: "assistant",
			Content: []ContentBlock{
				{Type: "text", Text: "Hello from mock"},
			},
		})
	}))
	defer server.Close()

	client := NewClient("test-key", "claude-test")
	client.baseURL = server.URL

	messages := []Message{
		{Role: "user", Content: "Hi"},
	}
	resp, err := client.ChatWithBlocks(context.Background(), messages, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Role != "assistant" {
		t.Errorf("expected role assistant, got %s", resp.Role)
	}
	if len(resp.Content) != 1 || resp.Content[0].Type != "text" {
		t.Fatalf("unexpected content: %+v", resp.Content)
	}
	if resp.Content[0].Text != "Hello from mock" {
		t.Errorf("unexpected text: %s", resp.Content[0].Text)
	}
}

func TestChatWithBlocksToolUse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Response{
			Role: "assistant",
			Content: []ContentBlock{
				{Type: "text", Text: "Let me check that."},
				{Type: "tool_use", ID: "tu_1", Name: "bash", Input: json.RawMessage(`{"command":"ls"}`)},
			},
		})
	}))
	defer server.Close()

	client := NewClient("test-key", "claude-test")
	client.baseURL = server.URL

	messages := []Message{
		{Role: "user", Content: "List files"},
	}
	tools := []Tool{
		{
			Name:        "bash",
			Description: "Run bash",
			InputSchema: json.RawMessage(`{"type":"object","properties":{}}`),
		},
	}

	resp, err := client.ChatWithBlocks(context.Background(), messages, tools)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Content) != 2 {
		t.Fatalf("expected 2 content blocks, got %d", len(resp.Content))
	}
	if resp.Content[1].Type != "tool_use" || resp.Content[1].Name != "bash" {
		t.Errorf("unexpected tool_use block: %+v", resp.Content[1])
	}
}

func TestChatWithBlocksBlocksSupport(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request body contains blocks
		var reqBody map[string]interface{}
		json.NewDecoder(r.Body).Decode(&reqBody)
		messages, ok := reqBody["messages"].([]interface{})
		if !ok || len(messages) != 2 {
			t.Fatalf("expected 2 messages, got %+v", reqBody["messages"])
		}
		secondMsg, ok := messages[1].(map[string]interface{})
		if !ok {
			t.Fatalf("expected second message to be object")
		}
		content, ok := secondMsg["content"].([]interface{})
		if !ok || len(content) != 1 {
			t.Fatalf("expected content to be array with 1 block, got %+v", secondMsg["content"])
		}
		block, ok := content[0].(map[string]interface{})
		if !ok || block["type"] != "tool_result" {
			t.Fatalf("expected tool_result block, got %+v", content[0])
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Response{
			Role:    "assistant",
			Content: []ContentBlock{{Type: "text", Text: "Done"}},
		})
	}))
	defer server.Close()

	client := NewClient("test-key", "claude-test")
	client.baseURL = server.URL

	messages := []Message{
		{Role: "user", Content: "Run command"},
		{
			Role: "user",
			Blocks: []ContentBlock{
				{Type: "tool_result", ToolUseID: "tu_1", Content: "output"},
			},
		},
	}

	_, err := client.ChatWithBlocks(context.Background(), messages, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestChatWithBlocksAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"invalid request"}`))
	}))
	defer server.Close()

	client := NewClient("test-key", "claude-test")
	client.baseURL = server.URL

	_, err := client.ChatWithBlocks(context.Background(), []Message{{Role: "user", Content: "Hi"}}, nil)
	if err == nil {
		t.Fatal("expected error for bad request")
	}
}

func TestChatWithBlocksNoAPIKey(t *testing.T) {
	client := NewClient("", "claude-test")
	resp, err := client.ChatWithBlocks(context.Background(), []Message{{Role: "user", Content: "Hi"}}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Content) == 0 || resp.Content[0].Type != "text" {
		t.Fatalf("expected text block for missing key warning")
	}
}

func TestSetProvider(t *testing.T) {
	client := NewClient("key", "model")
	client.SetProvider("bedrock")
	if client.baseURL != "https://bedrock-runtime.us-east-1.amazonaws.com" {
		t.Errorf("unexpected baseURL for bedrock: %s", client.baseURL)
	}

	client.SetProvider("vertex")
	if client.baseURL != "https://us-central1-aiplatform.googleapis.com" {
		t.Errorf("unexpected baseURL for vertex: %s", client.baseURL)
	}
}
