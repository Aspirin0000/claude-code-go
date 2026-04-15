package commands

import (
	"context"
	"testing"

	"github.com/Aspirin0000/claude-code-go/internal/state"
)

func TestHistoryCommand(t *testing.T) {
	// Save and restore global state
	origState := state.GlobalState
	defer func() { state.GlobalState = origState }()

	state.GlobalState = state.NewAppState()
	state.GlobalState.AddMessage(state.Message{Role: "user", Content: "Hello"})
	state.GlobalState.AddMessage(state.Message{Role: "assistant", Content: "Hi"})

	cmd := NewHistoryCommand()
	ctx := context.Background()

	if err := cmd.Execute(ctx, []string{}); err != nil {
		t.Fatalf("history command failed: %v", err)
	}
}
