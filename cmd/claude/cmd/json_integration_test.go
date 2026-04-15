package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Aspirin0000/claude-code-go/internal/api"
	"github.com/Aspirin0000/claude-code-go/internal/config"
	"github.com/Aspirin0000/claude-code-go/internal/state"
	"github.com/Aspirin0000/claude-code-go/internal/tools"
)

func TestJSONMode_SimpleResponse(t *testing.T) {
	mockAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(api.Response{
			Role: "assistant",
			Content: []api.ContentBlock{
				{Type: "text", Text: "Hello from JSON mode"},
			},
		})
	}))
	defer mockAPI.Close()

	client := api.NewClient("test-key", "claude-test")
	client.SetBaseURL(mockAPI.URL)

	app := &App{
		apiClient:    client,
		toolRegistry: tools.NewDefaultRegistry(),
		config:       config.DefaultConfig(),
	}

	origState := state.GlobalState
	state.GlobalState = state.NewAppState()
	defer func() { state.GlobalState = origState }()

	stdin := bytes.NewReader([]byte(`{"prompt":"hello","system":"be helpful"}`))
	stdout := &bytes.Buffer{}

	if err := runJSONModeWithApp(app, stdin, stdout); err != nil {
		t.Fatalf("json mode failed: %v", err)
	}

	var resp JSONResponse
	if err := json.Unmarshal(stdout.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal stdout: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got error: %s", resp.Error)
	}
	if resp.Response != "Hello from JSON mode" {
		t.Errorf("unexpected response: %s", resp.Response)
	}
	if len(resp.Messages) != 2 {
		t.Errorf("expected 2 messages, got %d", len(resp.Messages))
	}
}

func TestJSONMode_MissingPrompt(t *testing.T) {
	app := &App{
		apiClient:    api.NewClient("test-key", "claude-test"),
		toolRegistry: tools.NewDefaultRegistry(),
		config:       config.DefaultConfig(),
	}

	origState := state.GlobalState
	state.GlobalState = state.NewAppState()
	defer func() { state.GlobalState = origState }()

	stdin := bytes.NewReader([]byte(`{}`))
	stdout := &bytes.Buffer{}

	if err := runJSONModeWithApp(app, stdin, stdout); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	var resp JSONResponse
	if err := json.Unmarshal(stdout.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure for missing prompt")
	}
}

func TestJSONMode_InvalidJSON(t *testing.T) {
	app := &App{
		apiClient:    api.NewClient("test-key", "claude-test"),
		toolRegistry: tools.NewDefaultRegistry(),
		config:       config.DefaultConfig(),
	}

	origState := state.GlobalState
	state.GlobalState = state.NewAppState()
	defer func() { state.GlobalState = origState }()

	stdin := bytes.NewReader([]byte(`not json`))
	stdout := &bytes.Buffer{}

	if err := runJSONModeWithApp(app, stdin, stdout); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	var resp JSONResponse
	if err := json.Unmarshal(stdout.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure for invalid JSON")
	}
}

func TestJSONMode_NoClient(t *testing.T) {
	app := &App{
		apiClient:    nil,
		toolRegistry: tools.NewDefaultRegistry(),
		config:       config.DefaultConfig(),
	}

	origState := state.GlobalState
	state.GlobalState = state.NewAppState()
	defer func() { state.GlobalState = origState }()

	stdout := &bytes.Buffer{}
	if err := runJSONModeWithApp(app, nil, stdout); err != nil {
		t.Fatalf("expected nil error (written to stdout), got %v", err)
	}

	var resp JSONResponse
	if err := json.Unmarshal(stdout.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure for nil client")
	}
}

func TestJSONMode_WithToolUse(t *testing.T) {
	callCount := 0
	mockAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if callCount == 1 {
			json.NewEncoder(w).Encode(api.Response{
				Role: "assistant",
				Content: []api.ContentBlock{
					{Type: "tool_use", ID: "tu_1", Name: "bash", Input: json.RawMessage(`{"command":"echo test"}`)},
				},
			})
		} else {
			json.NewEncoder(w).Encode(api.Response{
				Role: "assistant",
				Content: []api.ContentBlock{
					{Type: "text", Text: "Result received"},
				},
			})
		}
	}))
	defer mockAPI.Close()

	client := api.NewClient("test-key", "claude-test")
	client.SetBaseURL(mockAPI.URL)

	app := &App{
		apiClient:    client,
		toolRegistry: tools.NewDefaultRegistry(),
		config:       config.DefaultConfig(),
	}

	origState := state.GlobalState
	state.GlobalState = state.NewAppState()
	defer func() { state.GlobalState = origState }()

	stdin := bytes.NewReader([]byte(`{"prompt":"run echo test"}`))
	stdout := &bytes.Buffer{}

	if err := runJSONModeWithApp(app, stdin, stdout); err != nil {
		t.Fatalf("json mode with tool failed: %v", err)
	}

	var resp JSONResponse
	if err := json.Unmarshal(stdout.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got error: %s", resp.Error)
	}
	if resp.Response != "Result received" {
		t.Errorf("unexpected response: %s", resp.Response)
	}
	if len(resp.ToolCalls) != 1 {
		t.Errorf("expected 1 tool call, got %d", len(resp.ToolCalls))
	}
	if callCount != 2 {
		t.Errorf("expected 2 API calls, got %d", callCount)
	}
}

func TestJSONMode_PromptFlagFallback(t *testing.T) {
	mockAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(api.Response{
			Role: "assistant",
			Content: []api.ContentBlock{
				{Type: "text", Text: "Got it"},
			},
		})
	}))
	defer mockAPI.Close()

	client := api.NewClient("test-key", "claude-test")
	client.SetBaseURL(mockAPI.URL)

	app := &App{
		apiClient:    client,
		toolRegistry: tools.NewDefaultRegistry(),
		config:       config.DefaultConfig(),
	}

	origState := state.GlobalState
	state.GlobalState = state.NewAppState()
	defer func() { state.GlobalState = origState }()

	origPrompt := promptFlag
	promptFlag = "flag prompt"
	defer func() { promptFlag = origPrompt }()

	stdin := bytes.NewReader([]byte{})
	stdout := &bytes.Buffer{}

	if err := runJSONModeWithApp(app, stdin, stdout); err != nil {
		t.Fatalf("json mode with prompt flag failed: %v", err)
	}

	var resp JSONResponse
	if err := json.Unmarshal(stdout.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success: %s", resp.Error)
	}
	if resp.Response != "Got it" {
		t.Errorf("unexpected response: %s", resp.Response)
	}
}
