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

func TestWrapText(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		width    int
		expected string
	}{
		{
			name:     "short text no wrap",
			text:     "hello world",
			width:    20,
			expected: "hello world",
		},
		{
			name:     "wrap long line",
			text:     "hello world this is a long sentence",
			width:    10,
			expected: "hello\nworld this\nis a long\nsentence",
		},
		{
			name:     "preserve newlines",
			text:     "line one\nline two",
			width:    20,
			expected: "line one\nline two",
		},
		{
			name:     "break very long word",
			text:     "supercalifragilisticexpialidocious",
			width:    10,
			expected: "supercalif\nragilistic\nexpialidoc\nious",
		},
		{
			name:     "zero width fallback",
			text:     "hello",
			width:    0,
			expected: "hello",
		},
		{
			name:     "multiple spaces collapsed",
			text:     "hello    world",
			width:    20,
			expected: "hello world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := wrapText(tt.text, tt.width)
			if got != tt.expected {
				t.Errorf("wrapText(%q, %d) = %q, want %q", tt.text, tt.width, got, tt.expected)
			}
		})
	}
}

func TestVisibleWidth(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"plain text", "hello", 5},
		{"with ansi", "\x1b[31mhello\x1b[0m", 5},
		{"empty", "", 0},
		{"only ansi", "\x1b[31m\x1b[0m", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := visibleWidth(tt.input)
			if got != tt.expected {
				t.Errorf("visibleWidth(%q) = %d, want %d", tt.input, got, tt.expected)
			}
		})
	}
}
