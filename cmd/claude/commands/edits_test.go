package commands

import (
	"strings"
	"testing"
	"time"

	"github.com/Aspirin0000/claude-code-go/internal/state"
)

func TestEditsCommand_Empty(t *testing.T) {
	orig := state.GlobalState
	state.GlobalState = state.NewAppState()
	defer func() { state.GlobalState = orig }()

	cmd := NewEditsCommand()
	out := captureOutput(func() {
		_ = cmd.Execute(nil, nil)
	})

	if !strings.Contains(out, "No file modifications recorded") {
		t.Errorf("expected empty edits message, got: %s", out)
	}
}

func TestEditsCommand_WithEdits(t *testing.T) {
	orig := state.GlobalState
	state.GlobalState = state.NewAppState()
	defer func() { state.GlobalState = orig }()

	state.GlobalState.AddEdit(state.Edit{
		Timestamp:   time.Now(),
		Tool:        "file_write",
		FilePath:    "/tmp/test.go",
		Operation:   "write",
		Description: "Created file",
	})
	state.GlobalState.AddEdit(state.Edit{
		Timestamp:   time.Now(),
		Tool:        "file_edit",
		FilePath:    "/tmp/test.go",
		Operation:   "edit",
		Description: "Edited file (search and replace)",
	})

	cmd := NewEditsCommand()
	out := captureOutput(func() {
		_ = cmd.Execute(nil, nil)
	})

	if !strings.Contains(out, "Session File Modifications") {
		t.Errorf("expected header, got: %s", out)
	}
	if !strings.Contains(out, "/tmp/test.go") {
		t.Errorf("expected file path in output, got: %s", out)
	}
	if !strings.Contains(out, "Total modifications: 2") {
		t.Errorf("expected total count, got: %s", out)
	}
}

func TestEditsCommand_Registered(t *testing.T) {
	reg := NewRegistry()
	_ = reg.Register(NewEditsCommand())
	if _, ok := reg.Get("edits"); !ok {
		t.Error("edits command not registered")
	}
}
