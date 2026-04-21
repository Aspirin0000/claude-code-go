package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Aspirin0000/claude-code-go/internal/state"
)

func TestExportCommand_NoMessages(t *testing.T) {
	orig := state.GlobalState
	state.GlobalState = state.NewAppState()
	defer func() { state.GlobalState = orig }()

	cmd := NewExportCommand()
	out := captureOutput(func() {
		_ = cmd.Execute(nil, nil)
	})

	if !strings.Contains(out, "No messages to export") {
		t.Errorf("expected no messages message, got: %s", out)
	}
}

func TestExportCommand_TextFormat(t *testing.T) {
	orig := state.GlobalState
	state.GlobalState = state.NewAppState()
	defer func() { state.GlobalState = orig }()

	state.GlobalState.AddMessage(state.Message{
		Role:      "user",
		Content:   "Hello",
		Timestamp: time.Now(),
	})
	state.GlobalState.AddMessage(state.Message{
		Role:      "assistant",
		Content:   "Hi there",
		Timestamp: time.Now(),
	})

	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "test.txt")

	cmd := NewExportCommand()
	out := captureOutput(func() {
		_ = cmd.Execute(nil, []string{"text", filename})
	})

	if !strings.Contains(out, "Exported 2 message(s)") {
		t.Errorf("expected export confirmation, got: %s", out)
	}

	content, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("failed to read exported file: %v", err)
	}

	if !strings.Contains(string(content), "Hello") {
		t.Error("expected 'Hello' in exported content")
	}
	if !strings.Contains(string(content), "Claude") {
		t.Error("expected 'Claude' in exported content")
	}
}

func TestExportCommand_UnknownFormat(t *testing.T) {
	cmd := NewExportCommand()
	err := cmd.Execute(nil, []string{"xml"})
	if err == nil || !strings.Contains(err.Error(), "unknown format") {
		t.Errorf("expected unknown format error, got: %v", err)
	}
}

func TestExportCommand_Registered(t *testing.T) {
	reg := NewRegistry()
	_ = reg.Register(NewExportCommand())
	if _, ok := reg.Get("export"); !ok {
		t.Error("export command not registered")
	}
}
