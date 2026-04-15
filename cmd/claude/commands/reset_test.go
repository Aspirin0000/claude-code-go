package commands

import (
	"context"
	"testing"

	"github.com/Aspirin0000/claude-code-go/internal/state"
)

func TestResetCommand_WithMessages(t *testing.T) {
	cmd := NewResetCommand()
	ctx := context.Background()

	oldMessages := state.GlobalState.GetMessages()
	defer state.GlobalState.SetMessages(oldMessages)

	state.GlobalState.SetMessages([]state.Message{
		{Role: "user", Content: "hello"},
		{Role: "assistant", Content: "hi"},
	})
	state.GlobalState.TurnCount = 1

	if err := cmd.Execute(ctx, []string{}); err != nil {
		t.Fatalf("reset failed: %v", err)
	}

	if len(state.GlobalState.GetMessages()) != 0 {
		t.Errorf("expected 0 messages after reset, got %d", len(state.GlobalState.GetMessages()))
	}
	if state.GlobalState.TurnCount != 0 {
		t.Errorf("expected turn count 0, got %d", state.GlobalState.TurnCount)
	}
}

func TestResetCommand_NoMessages(t *testing.T) {
	cmd := NewResetCommand()
	ctx := context.Background()

	oldMessages := state.GlobalState.GetMessages()
	defer state.GlobalState.SetMessages(oldMessages)

	state.GlobalState.ClearMessages()

	if err := cmd.Execute(ctx, []string{}); err != nil {
		t.Fatalf("reset with no messages failed: %v", err)
	}
}

func TestResetCommand_Aliases(t *testing.T) {
	cmd := NewResetCommand()
	aliases := cmd.Aliases()

	expected := map[string]bool{"clear-history": true, "clear-chat": true}
	for _, alias := range aliases {
		if !expected[alias] {
			t.Errorf("unexpected alias: %s", alias)
		}
		delete(expected, alias)
	}
	if len(expected) > 0 {
		t.Errorf("missing expected aliases: %v", expected)
	}
}
