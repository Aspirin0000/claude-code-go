package cmd

import (
	"testing"

	"github.com/Aspirin0000/claude-code-go/internal/api"
	"github.com/Aspirin0000/claude-code-go/internal/state"
	"github.com/Aspirin0000/claude-code-go/internal/tools"
)

func TestAppHandleAPIResponse_TextOnly(t *testing.T) {
	origState := state.GlobalState
	state.GlobalState = state.NewAppState()
	defer func() { state.GlobalState = origState }()

	app := &App{
		apiClient:    api.NewClient("test-key", "claude-test"),
		toolRegistry: tools.NewDefaultRegistry(),
	}

	resp := &api.Response{
		Role: "assistant",
		Content: []api.ContentBlock{
			{Type: "text", Text: "Hello there"},
		},
	}

	app.handleAPIResponse(resp)

	msgs := state.GlobalState.GetMessages()
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}
	if msgs[0].Role != "assistant" {
		t.Errorf("expected assistant role, got %s", msgs[0].Role)
	}
	if msgs[0].Content != "Hello there" {
		t.Errorf("unexpected content: %s", msgs[0].Content)
	}
	if app.loading {
		t.Error("expected loading to be false after text-only response")
	}
}

func TestAppHandleAPIResponse_WithToolUse(t *testing.T) {
	origState := state.GlobalState
	state.GlobalState = state.NewAppState()
	defer func() { state.GlobalState = origState }()

	app := &App{
		apiClient:    api.NewClient("test-key", "claude-test"),
		toolRegistry: tools.NewDefaultRegistry(),
	}

	resp := &api.Response{
		Role: "assistant",
		Content: []api.ContentBlock{
			{Type: "tool_use", ID: "tu_1", Name: "bash", Input: []byte(`{"command":"echo hi"}`)},
		},
	}

	_, cmd := app.handleAPIResponse(resp)
	if cmd == nil {
		t.Fatal("expected non-nil cmd for tool use")
	}

	msgs := state.GlobalState.GetMessages()
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}
	if len(msgs[0].Blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(msgs[0].Blocks))
	}
	if msgs[0].Blocks[0].Type != "tool_use" {
		t.Errorf("expected tool_use block, got %s", msgs[0].Blocks[0].Type)
	}
}

func TestAppProcessStreamEvent(t *testing.T) {
	app := &App{}

	app.processStreamEvent(api.StreamEvent{
		Type: "content_block_start",
		ContentBlock: &api.ContentBlock{
			Type: "text",
			Text: "",
		},
	})

	if len(app.streamBlocks) != 1 {
		t.Fatalf("expected 1 stream block, got %d", len(app.streamBlocks))
	}

	app.processStreamEvent(api.StreamEvent{
		Type:  "content_block_delta",
		Index: 0,
		Delta: api.Delta{
			Type: "text_delta",
			Text: "hello",
		},
	})

	if app.streamBlocks[0].Text != "hello" {
		t.Errorf("expected 'hello', got %s", app.streamBlocks[0].Text)
	}
	if app.streamingText != "hello" {
		t.Errorf("expected streamingText 'hello', got %s", app.streamingText)
	}
}

func TestAppFinishStream(t *testing.T) {
	app := &App{
		streamBlocks: []api.ContentBlock{
			{Type: "text", Text: "streamed text"},
		},
		streamingText: "streamed text",
	}

	resp, err := app.finishStream()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Role != "assistant" {
		t.Errorf("expected assistant role, got %s", resp.Role)
	}
	if len(resp.Content) != 1 || resp.Content[0].Text != "streamed text" {
		t.Errorf("unexpected response content: %+v", resp.Content)
	}
	if app.streamingText != "" {
		t.Error("expected streamingText to be cleared")
	}
	if app.streamBlocks != nil {
		t.Error("expected streamBlocks to be cleared")
	}
}
