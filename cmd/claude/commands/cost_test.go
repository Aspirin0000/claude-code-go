package commands

import (
	"context"
	"testing"

	"github.com/Aspirin0000/claude-code-go/internal/state"
)

func TestCalculateCostTokens(t *testing.T) {
	tests := []struct {
		name       string
		messages   []state.Message
		wantInput  int
		wantOutput int
	}{
		{
			name:       "empty messages",
			messages:   []state.Message{},
			wantInput:  0,
			wantOutput: 0,
		},
		{
			name: "single user message",
			messages: []state.Message{
				{Role: "user", Content: "hello world"},
			},
			wantInput:  2, // len("hello world") / 4 = 11/4 = 2
			wantOutput: 0,
		},
		{
			name: "single assistant message",
			messages: []state.Message{
				{Role: "assistant", Content: "hi there"},
			},
			wantInput:  0,
			wantOutput: 2, // len("hi there") / 4 = 8/4 = 2
		},
		{
			name: "mixed messages",
			messages: []state.Message{
				{Role: "user", Content: "hello"},
				{Role: "assistant", Content: "world"},
				{Role: "user", Content: "how are you"},
			},
			wantInput:  3, // 5/4=1 + 0 + 11/4=2
			wantOutput: 1, // 0 + 5/4=1 + 0
		},
		{
			name: "system messages ignored",
			messages: []state.Message{
				{Role: "system", Content: "system prompt here"},
				{Role: "user", Content: "hello"},
			},
			wantInput:  1, // 5/4=1
			wantOutput: 0,
		},
		{
			name: "messages with blocks",
			messages: []state.Message{
				{Role: "user", Content: "", Blocks: []state.ContentBlock{{Text: "block text"}}},
			},
			wantInput:  0, // calculateCostTokens only looks at Content, not Blocks
			wantOutput: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotInput, gotOutput := calculateCostTokens(tt.messages)
			if gotInput != tt.wantInput {
				t.Errorf("calculateCostTokens() input = %d, want %d", gotInput, tt.wantInput)
			}
			if gotOutput != tt.wantOutput {
				t.Errorf("calculateCostTokens() output = %d, want %d", gotOutput, tt.wantOutput)
			}
		})
	}
}

func TestCostCommandExecution(t *testing.T) {
	cmd := NewCostCommand()
	ctx := context.Background()

	// The cost command reads from global state, so just verify it executes without error
	if err := cmd.Execute(ctx, []string{}); err != nil {
		t.Fatalf("cost command failed: %v", err)
	}
}

func TestCostCommandWithMessages(t *testing.T) {
	// Add temporary messages to global state
	oldMessages := state.GlobalState.GetMessages()
	defer state.GlobalState.SetMessages(oldMessages)

	state.GlobalState.SetMessages([]state.Message{
		{Role: "user", Content: "What is the weather today?"},
		{Role: "assistant", Content: "The weather is sunny and warm."},
		{Role: "user", Content: "Thank you!"},
	})

	cmd := NewCostCommand()
	ctx := context.Background()

	if err := cmd.Execute(ctx, []string{}); err != nil {
		t.Fatalf("cost command with messages failed: %v", err)
	}
}
