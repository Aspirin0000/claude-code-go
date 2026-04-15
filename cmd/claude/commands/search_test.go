package commands

import (
	"context"
	"testing"

	"github.com/Aspirin0000/claude-code-go/internal/state"
)

func TestSearchCommandNoArgs(t *testing.T) {
	cmd := NewSearchCommand()
	err := cmd.Execute(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for missing keyword")
	}
}

func TestSearchCommandNoMatches(t *testing.T) {
	origState := state.GlobalState
	defer func() { state.GlobalState = origState }()
	state.GlobalState = state.NewAppState()
	state.GlobalState.AddMessage(state.Message{Role: "user", Content: "hello world"})

	cmd := NewSearchCommand()
	err := cmd.Execute(context.Background(), []string{"docker"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSearchCommandWithMatches(t *testing.T) {
	origState := state.GlobalState
	defer func() { state.GlobalState = origState }()
	state.GlobalState = state.NewAppState()
	state.GlobalState.AddMessage(state.Message{Role: "user", Content: "how do I use docker?"})
	state.GlobalState.AddMessage(state.Message{Role: "assistant", Content: "You can run docker ps to list containers."})
	state.GlobalState.AddMessage(state.Message{Role: "user", Content: "thanks"})

	cmd := NewSearchCommand()
	err := cmd.Execute(context.Background(), []string{"docker"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSearchCommandMultiWord(t *testing.T) {
	origState := state.GlobalState
	defer func() { state.GlobalState = origState }()
	state.GlobalState = state.NewAppState()
	state.GlobalState.AddMessage(state.Message{Role: "user", Content: "error message here"})

	cmd := NewSearchCommand()
	err := cmd.Execute(context.Background(), []string{"error", "message"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
