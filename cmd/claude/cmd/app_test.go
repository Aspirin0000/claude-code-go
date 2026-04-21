package cmd

import (
	"strings"
	"testing"
	"time"

	"github.com/Aspirin0000/claude-code-go/internal/api"
	"github.com/Aspirin0000/claude-code-go/internal/state"
	"github.com/Aspirin0000/claude-code-go/internal/tools"
	tea "github.com/charmbracelet/bubbletea"
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

func TestMouseScroll(t *testing.T) {
	app := &App{scrollOffset: 0}

	// Wheel up should increase scroll offset
	_, _ = app.Update(tea.MouseMsg{Button: tea.MouseButtonWheelUp})
	if app.scrollOffset != 3 {
		t.Errorf("expected scrollOffset 3 after wheel up, got %d", app.scrollOffset)
	}

	// Wheel down should decrease scroll offset
	_, _ = app.Update(tea.MouseMsg{Button: tea.MouseButtonWheelDown})
	if app.scrollOffset != 0 {
		t.Errorf("expected scrollOffset 0 after wheel down, got %d", app.scrollOffset)
	}

	// Wheel down should not go below zero
	_, _ = app.Update(tea.MouseMsg{Button: tea.MouseButtonWheelDown})
	if app.scrollOffset != 0 {
		t.Errorf("expected scrollOffset to stay at 0, got %d", app.scrollOffset)
	}
}

func TestMessageLines(t *testing.T) {
	app := &App{width: 40, styles: newStyles("dark")}

	msg := state.Message{
		Role:      "user",
		Content:   "hello world",
		Timestamp: time.Now(),
	}
	lines := app.messageLines(msg)
	if lines < 3 {
		t.Errorf("expected at least 3 lines for short message, got %d", lines)
	}

	// Long message should wrap
	msg.Content = strings.Repeat("a ", 30)
	linesLong := app.messageLines(msg)
	if linesLong <= lines {
		t.Errorf("expected more lines for long message, got %d vs %d", linesLong, lines)
	}
}

func TestCalculateStartIdx(t *testing.T) {
	app := &App{width: 40, height: 20, styles: newStyles("dark")}

	messages := []state.Message{
		{Role: "user", Content: "msg 1"},
		{Role: "assistant", Content: "msg 2"},
		{Role: "user", Content: "msg 3"},
	}

	idx := app.calculateStartIdx(messages)
	if idx != 0 {
		t.Errorf("expected startIdx 0 for few messages, got %d", idx)
	}

	// Add many long messages
	for i := 0; i < 20; i++ {
		messages = append(messages, state.Message{
			Role:    "assistant",
			Content: strings.Repeat("word ", 20),
		})
	}
	idx = app.calculateStartIdx(messages)
	if idx == 0 {
		t.Error("expected startIdx > 0 when messages exceed screen height")
	}
}

func TestMultiLineInput(t *testing.T) {
	app := &App{width: 40, styles: newStyles("dark")}

	// Alt+Enter should add a newline
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyEnter, Alt: true})
	if app.input != "\n" {
		t.Errorf("expected newline in input, got %q", app.input)
	}

	// Regular Enter should not add newline when loading or empty
	app.input = "hello"
	app.loading = true
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if app.input != "hello" {
		t.Errorf("expected input unchanged when loading, got %q", app.input)
	}
}

func TestRenderInputText(t *testing.T) {
	app := &App{width: 20, input: "hello world this is long", styles: newStyles("dark")}
	rendered := app.renderInputText()
	if !strings.Contains(rendered, "> ") {
		t.Error("expected input prefix")
	}
	if !strings.Contains(rendered, "█") {
		t.Error("expected cursor")
	}
	// Should wrap and indent continuation lines
	lines := strings.Split(rendered, "\n")
	if len(lines) <= 1 {
		t.Error("expected wrapped input to have multiple lines")
	}
	for i := 1; i < len(lines); i++ {
		if !strings.HasPrefix(lines[i], "  ") {
			t.Errorf("expected continuation line %d to be indented, got %q", i, lines[i])
		}
	}
}

func TestTabCompletion(t *testing.T) {
	app := &App{width: 40, styles: newStyles("dark")}

	// Test completing a partial command
	app.input = "/he"
	result := app.completeCommand(app.input)
	if !strings.HasPrefix(result, "/help") {
		t.Errorf("expected /help completion, got %q", result)
	}

	// Test no match returns input unchanged
	app.input = "/xyz"
	result = app.completeCommand(app.input)
	if result != "/xyz" {
		t.Errorf("expected unchanged input for no match, got %q", result)
	}

	// Test non-slash input
	app.input = "hello"
	result = app.completeCommand(app.input)
	if result != "hello" {
		t.Errorf("expected unchanged non-slash input, got %q", result)
	}
}

func TestLongestCommonPrefix(t *testing.T) {
	tests := []struct {
		input []string
		want  string
	}{
		{[]string{"/help", "/history"}, "/h"},
		{[]string{"/exit", "/exit"}, "/exit"},
		{[]string{"/a", "/b"}, "/"},
		{[]string{"test"}, "test"},
		{[]string{}, ""},
	}

	for _, tt := range tests {
		got := longestCommonPrefix(tt.input)
		if got != tt.want {
			t.Errorf("longestCommonPrefix(%v) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
