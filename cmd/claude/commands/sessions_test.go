package commands

import (
	"context"
	"testing"

	"github.com/Aspirin0000/claude-code-go/internal/config"
	"github.com/Aspirin0000/claude-code-go/internal/state"
)

func TestSessionsCommandStorageNotInitialized(t *testing.T) {
	origStorage := state.GlobalSessionStorage
	state.GlobalSessionStorage = nil
	defer func() { state.GlobalSessionStorage = origStorage }()

	cmd := NewSessionsCommand()
	err := cmd.Execute(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error when session storage not initialized")
	}
}

func makeTestSessionStorage(t *testing.T) *state.SessionStorage {
	cfg := config.DefaultConfig()
	cfg.AutoSave = true
	cfg.AutoSaveDir = t.TempDir()
	return state.NewSessionStorage(cfg)
}

func TestSessionsCommandListEmpty(t *testing.T) {
	origStorage := state.GlobalSessionStorage
	state.GlobalSessionStorage = makeTestSessionStorage(t)
	defer func() { state.GlobalSessionStorage = origStorage }()

	cmd := NewSessionsCommand()
	err := cmd.Execute(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSessionsCommandLoadMissingArg(t *testing.T) {
	origStorage := state.GlobalSessionStorage
	state.GlobalSessionStorage = makeTestSessionStorage(t)
	defer func() { state.GlobalSessionStorage = origStorage }()

	cmd := NewSessionsCommand()
	err := cmd.Execute(context.Background(), []string{"load"})
	if err == nil {
		t.Fatal("expected error for missing load arg")
	}
}

func TestSessionsCommandDeleteMissingArg(t *testing.T) {
	origStorage := state.GlobalSessionStorage
	state.GlobalSessionStorage = makeTestSessionStorage(t)
	defer func() { state.GlobalSessionStorage = origStorage }()

	cmd := NewSessionsCommand()
	err := cmd.Execute(context.Background(), []string{"delete"})
	if err == nil {
		t.Fatal("expected error for missing delete arg")
	}
}

func TestSessionsCommandCleanEmpty(t *testing.T) {
	origStorage := state.GlobalSessionStorage
	state.GlobalSessionStorage = makeTestSessionStorage(t)
	defer func() { state.GlobalSessionStorage = origStorage }()

	cmd := NewSessionsCommand()
	err := cmd.Execute(context.Background(), []string{"clean"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTruncate(t *testing.T) {
	if truncate("hello", 10) != "hello" {
		t.Errorf("unexpected truncate result for short string")
	}
	if truncate("hello world this is long", 10) != "hello w..." {
		t.Errorf("unexpected truncate result for long string")
	}
}
