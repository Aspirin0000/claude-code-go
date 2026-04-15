package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Aspirin0000/claude-code-go/internal/api"
	"github.com/Aspirin0000/claude-code-go/internal/tools"
)

func TestServerChat_SimpleResponse(t *testing.T) {
	mockAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(api.Response{
			Role: "assistant",
			Content: []api.ContentBlock{
				{Type: "text", Text: "Hello from mock API"},
			},
		})
	}))
	defer mockAPI.Close()

	client := api.NewClient("test-key", "claude-test")
	client.SetBaseURL(mockAPI.URL)

	registry := tools.NewDefaultRegistry()
	server := NewServer(client, registry, "claude-test")

	mux := http.NewServeMux()
	server.RegisterRoutes(mux)

	body, _ := json.Marshal(ChatRequest{Prompt: "hello"})
	req := httptest.NewRequest(http.MethodPost, "/chat", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var resp ChatResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got error: %s", resp.Error)
	}
	if resp.Response != "Hello from mock API" {
		t.Errorf("unexpected response: %s", resp.Response)
	}
	if len(resp.Messages) != 2 {
		t.Errorf("expected 2 messages (user + assistant), got %d", len(resp.Messages))
	}
}

func TestServerChat_WithToolUse(t *testing.T) {
	callCount := 0
	mockAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if callCount == 1 {
			json.NewEncoder(w).Encode(api.Response{
				Role: "assistant",
				Content: []api.ContentBlock{
					{Type: "tool_use", ID: "tu_1", Name: "bash", Input: json.RawMessage(`{"command":"echo hello"}`)},
				},
			})
		} else {
			json.NewEncoder(w).Encode(api.Response{
				Role: "assistant",
				Content: []api.ContentBlock{
					{Type: "text", Text: "Done!"},
				},
			})
		}
	}))
	defer mockAPI.Close()

	client := api.NewClient("test-key", "claude-test")
	client.SetBaseURL(mockAPI.URL)

	registry := tools.NewDefaultRegistry()
	server := NewServer(client, registry, "claude-test")

	mux := http.NewServeMux()
	server.RegisterRoutes(mux)

	body, _ := json.Marshal(ChatRequest{Prompt: "run echo hello"})
	req := httptest.NewRequest(http.MethodPost, "/chat", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var resp ChatResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got error: %s", resp.Error)
	}
	if resp.Response != "Done!" {
		t.Errorf("unexpected response: %s", resp.Response)
	}
	if len(resp.ToolCalls) != 1 {
		t.Errorf("expected 1 tool call, got %d", len(resp.ToolCalls))
	}
	if callCount != 2 {
		t.Errorf("expected 2 API calls, got %d", callCount)
	}
}

func TestServerChat_APIError(t *testing.T) {
	mockAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "mock api error"})
	}))
	defer mockAPI.Close()

	client := api.NewClient("test-key", "claude-test")
	client.SetBaseURL(mockAPI.URL)

	registry := tools.NewDefaultRegistry()
	server := NewServer(client, registry, "claude-test")

	mux := http.NewServeMux()
	server.RegisterRoutes(mux)

	body, _ := json.Marshal(ChatRequest{Prompt: "hello"})
	req := httptest.NewRequest(http.MethodPost, "/chat", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var resp ChatResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure")
	}
	if resp.Error == "" {
		t.Error("expected error message")
	}
}

func TestServerChat_WithSystemPrompt(t *testing.T) {
	mockAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqBody map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&reqBody)
		hasSystem := false
		if msgs, ok := reqBody["messages"].([]interface{}); ok {
			for _, m := range msgs {
				if msg, ok := m.(map[string]interface{}); ok && msg["role"] == "system" {
					hasSystem = true
					break
				}
			}
		}
		responseText := "No system prompt"
		if hasSystem {
			responseText = "System prompt received"
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(api.Response{
			Role: "assistant",
			Content: []api.ContentBlock{
				{Type: "text", Text: responseText},
			},
		})
	}))
	defer mockAPI.Close()

	client := api.NewClient("test-key", "claude-test")
	client.SetBaseURL(mockAPI.URL)

	registry := tools.NewDefaultRegistry()
	server := NewServer(client, registry, "claude-test")

	mux := http.NewServeMux()
	server.RegisterRoutes(mux)

	body, _ := json.Marshal(ChatRequest{Prompt: "hello", System: "be concise"})
	req := httptest.NewRequest(http.MethodPost, "/chat", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	var resp ChatResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success: %s", resp.Error)
	}
}
