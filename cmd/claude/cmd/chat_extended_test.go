package cmd

import (
	"strings"
	"testing"
	"time"

	"github.com/Aspirin0000/claude-code-go/internal/state"
)

// ============================================================================
// App messageLines Tests
// ============================================================================

func TestAppMessageLines(t *testing.T) {
	app := &App{width: 80}

	tests := []struct {
		name     string
		msg      state.Message
		expected int
	}{
		{
			name:     "simple user message",
			msg:      state.Message{Role: "user", Content: "Hello"},
			expected: 3, // 1 line + 2 trailing
		},
		{
			name:     "assistant with tool use",
			msg:      state.Message{Role: "assistant", Content: "Let me check", Blocks: []state.ContentBlock{{Type: "tool_use"}}},
			expected: 3, // 1 line + "[using tools...]" + 2 trailing
		},
		{
			name:     "system message",
			msg:      state.Message{Role: "system", Content: "Error"},
			expected: 3,
		},
		{
			name:     "empty message",
			msg:      state.Message{Role: "user", Content: ""},
			expected: 3,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := app.messageLines(test.msg)
			if result < test.expected {
				t.Errorf("expected at least %d lines, got %d", test.expected, result)
			}
		})
	}
}

func TestAppMessageLinesWithTimestamp(t *testing.T) {
	app := &App{width: 80}
	msg := state.Message{
		Role:      "user",
		Content:   "Hello",
		Timestamp: time.Now(),
	}
	result := app.messageLines(msg)
	if result < 3 {
		t.Errorf("expected at least 3 lines, got %d", result)
	}
}

// ============================================================================
// App calculateStartIdx Tests
// ============================================================================

func TestAppCalculateStartIdx(t *testing.T) {
	app := &App{width: 80, height: 24}

	messages := []state.Message{
		{Role: "user", Content: "Hello"},
		{Role: "assistant", Content: "Hi there"},
	}

	idx := app.calculateStartIdx(messages)
	if idx != 0 {
		t.Errorf("expected 0, got %d", idx)
	}
}

func TestAppCalculateStartIdxEmpty(t *testing.T) {
	app := &App{width: 80, height: 24}
	idx := app.calculateStartIdx([]state.Message{})
	if idx != 0 {
		t.Errorf("expected 0 for empty messages, got %d", idx)
	}
}

func TestAppCalculateStartIdxWithScroll(t *testing.T) {
	app := &App{width: 80, height: 24, scrollOffset: 5}

	messages := []state.Message{
		{Role: "user", Content: "Hello"},
		{Role: "assistant", Content: "Hi there"},
	}

	idx := app.calculateStartIdx(messages)
	// With scroll offset, should still be reasonable
	if idx < 0 || idx > len(messages) {
		t.Errorf("unexpected index: %d", idx)
	}
}

// ============================================================================
// App renderInputText Tests
// ============================================================================

func TestAppRenderInputText(t *testing.T) {
	app := &App{width: 80, input: "hello"}
	result := app.renderInputText()
	if !strings.Contains(result, "hello") {
		t.Errorf("expected 'hello' in result, got %q", result)
	}
	if !strings.HasPrefix(result, "> ") {
		t.Errorf("expected prefix '> ', got %q", result)
	}
}

func TestAppRenderInputTextEmpty(t *testing.T) {
	app := &App{width: 80, input: ""}
	result := app.renderInputText()
	if !strings.HasPrefix(result, "> ") {
		t.Errorf("expected prefix '> ', got %q", result)
	}
}

func TestAppRenderInputTextMultiline(t *testing.T) {
	app := &App{width: 20, input: "hello world this is a long input"}
	result := app.renderInputText()
	lines := strings.Split(result, "\n")
	if len(lines) < 2 {
		t.Errorf("expected multiple lines, got %d", len(lines))
	}
}

// ============================================================================
// newStyles Tests
// ============================================================================

func TestNewStyles(t *testing.T) {
	dark := newStyles("dark")
	if dark == nil {
		t.Fatal("expected non-nil dark styles")
	}

	light := newStyles("light")
	if light == nil {
		t.Fatal("expected non-nil light styles")
	}
}

func TestNewStylesDefault(t *testing.T) {
	// Unknown theme should default to dark
	styles := newStyles("unknown")
	if styles == nil {
		t.Fatal("expected non-nil styles for unknown theme")
	}
}

// ============================================================================
// Helper function
// ============================================================================

func splitLines(s string) []string {
	return strings.Split(s, "\n")
}
