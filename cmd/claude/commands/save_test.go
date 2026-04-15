package commands

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Aspirin0000/claude-code-go/internal/state"
)

func TestSaveCommandExportAsJSON(t *testing.T) {
	origState := state.GlobalState
	defer func() { state.GlobalState = origState }()
	state.GlobalState = state.NewAppState()
	state.GlobalState.AddMessage(state.Message{Role: "user", Content: "Hello"})
	state.GlobalState.AddMessage(state.Message{Role: "assistant", Content: "Hi there"})

	cmd := NewSaveCommand()
	data, err := cmd.exportAsJSON(state.GlobalState.GetMessages())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var export map[string]interface{}
	if err := json.Unmarshal(data, &export); err != nil {
		t.Fatalf("failed to unmarshal export: %v", err)
	}
	if export["message_count"].(float64) != 2 {
		t.Errorf("expected 2 messages, got %v", export["message_count"])
	}
}

func TestSaveCommandExportAsMarkdown(t *testing.T) {
	origState := state.GlobalState
	defer func() { state.GlobalState = origState }()
	state.GlobalState = state.NewAppState()
	state.GlobalState.AddMessage(state.Message{Role: "user", Content: "Hello"})
	state.GlobalState.AddMessage(state.Message{Role: "assistant", Content: "Hi there"})

	cmd := NewSaveCommand()
	data, err := cmd.exportAsMarkdown(state.GlobalState.GetMessages())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	md := string(data)
	if !strings.Contains(md, "# Conversation Export") {
		t.Errorf("expected markdown title")
	}
	if !strings.Contains(md, "## User") {
		t.Errorf("expected user section")
	}
	if !strings.Contains(md, "## Assistant") {
		t.Errorf("expected assistant section")
	}
}

func TestSaveCommandNoMessages(t *testing.T) {
	origState := state.GlobalState
	defer func() { state.GlobalState = origState }()
	state.GlobalState = state.NewAppState()

	cmd := NewSaveCommand()
	err := cmd.Execute(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSaveCommandExecuteMarkdown(t *testing.T) {
	origState := state.GlobalState
	origStorage := state.GlobalSessionStorage
	defer func() {
		state.GlobalState = origState
		state.GlobalSessionStorage = origStorage
	}()
	state.GlobalState = state.NewAppState()
	state.GlobalState.AddMessage(state.Message{Role: "user", Content: "test"})

	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "chat.md")

	cmd := NewSaveCommand()
	err := cmd.Execute(context.Background(), []string{filename, "--format", "md"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	if !strings.Contains(string(data), "## User") {
		t.Errorf("expected markdown content")
	}
}

func TestSaveCommandCapitalize(t *testing.T) {
	if capitalize("hello") != "Hello" {
		t.Errorf("unexpected capitalize result")
	}
	if capitalize("") != "" {
		t.Errorf("unexpected empty capitalize result")
	}
}
