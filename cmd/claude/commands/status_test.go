package commands

import (
	"context"
	"testing"
	"time"

	"github.com/Aspirin0000/claude-code-go/internal/state"
)

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		duration time.Duration
		expected string
	}{
		{30 * time.Second, "30s"},
		{90 * time.Second, "1m 30s"},
		{5 * time.Minute, "5m 0s"},
		{3661 * time.Second, "1h 1m 1s"},
		{0, "0s"},
	}

	for _, tt := range tests {
		result := formatDuration(tt.duration)
		if result != tt.expected {
			t.Errorf("formatDuration(%v) = %q, want %q", tt.duration, result, tt.expected)
		}
	}
}

func TestEstimateTokensForText(t *testing.T) {
	tests := []struct {
		text      string
		minTokens int
		maxTokens int
	}{
		{"", 0, 0},
		{"hello", 5, 6},       // 5 ascii chars / 4 = 1-2 + overhead 4 = 5-6
		{"hello world", 6, 7}, // 11 ascii chars / 4 = 3 + overhead 4 = 7
		{"你好世界", 6, 6},        // 4 cjk chars / 2 = 2 + overhead 4 = 6
		{"hello你好", 6, 7},     // 5 ascii/4=1 + 2 cjk/2=1 + 4 overhead = 6
	}

	for _, tt := range tests {
		result := estimateTokensForText(tt.text)
		if result < tt.minTokens || result > tt.maxTokens {
			t.Errorf("estimateTokensForText(%q) = %d, want between %d and %d", tt.text, result, tt.minTokens, tt.maxTokens)
		}
	}
}

func TestCalculateTokenUsage(t *testing.T) {
	messages := []state.Message{
		{Role: "user", Content: "hello"},
		{Role: "assistant", Content: "world"},
		{Role: "system", Content: "system msg"},
		{Role: "unknown", Content: "unknown msg"},
	}

	input, output := calculateTokenUsage(messages)
	if input <= 0 {
		t.Errorf("expected positive input tokens, got %d", input)
	}
	if output <= 0 {
		t.Errorf("expected positive output tokens, got %d", output)
	}

	// user, system, and unknown should count as input
	// assistant should count as output
}

func TestStatusCommandExecution(t *testing.T) {
	cmd := NewStatusCommand()
	ctx := context.Background()

	if err := cmd.Execute(ctx, []string{}); err != nil {
		t.Fatalf("status command failed: %v", err)
	}
}

func TestStatusCommandAliases(t *testing.T) {
	cmd := NewStatusCommand()
	aliases := cmd.Aliases()

	found := false
	for _, alias := range aliases {
		if alias == "st" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'st' alias")
	}
}

func TestGetCurrentModelInfo(t *testing.T) {
	// Default should return a model
	info := getCurrentModelInfo()
	if info.Name == "" {
		t.Error("expected non-empty model name")
	}
}
