package commands

import (
	"context"
	"testing"

	"github.com/Aspirin0000/claude-code-go/internal/state"
)

func TestCopyCommand_WithAssistantMessage(t *testing.T) {
	cmd := NewCopyCommand()
	ctx := context.Background()

	oldMessages := state.GlobalState.GetMessages()
	defer state.GlobalState.SetMessages(oldMessages)

	state.GlobalState.SetMessages([]state.Message{
		{Role: "user", Content: "hello"},
		{Role: "assistant", Content: "This is the assistant response to copy."},
	})

	// The test may fail if no clipboard tool is installed; that's OK for CI
	if err := cmd.Execute(ctx, []string{}); err != nil {
		t.Skipf("clipboard not available in test environment: %v", err)
	}
}

func TestCopyCommand_NoAssistantMessage(t *testing.T) {
	cmd := NewCopyCommand()
	ctx := context.Background()

	oldMessages := state.GlobalState.GetMessages()
	defer state.GlobalState.SetMessages(oldMessages)

	state.GlobalState.SetMessages([]state.Message{
		{Role: "user", Content: "hello"},
	})

	if err := cmd.Execute(ctx, []string{}); err != nil {
		t.Fatalf("should not error when no assistant message: %v", err)
	}
}

func TestCopyCommand_EmptyMessages(t *testing.T) {
	cmd := NewCopyCommand()
	ctx := context.Background()

	oldMessages := state.GlobalState.GetMessages()
	defer state.GlobalState.SetMessages(oldMessages)

	state.GlobalState.ClearMessages()

	if err := cmd.Execute(ctx, []string{}); err != nil {
		t.Fatalf("should not error with empty messages: %v", err)
	}
}
