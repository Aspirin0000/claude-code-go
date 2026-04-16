package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Aspirin0000/claude-code-go/internal/state"
)

func TestRollbackCommand_NoEdits(t *testing.T) {
	orig := state.GlobalState
	state.GlobalState = state.NewAppState()
	defer func() { state.GlobalState = orig }()

	cmd := NewRollbackCommand()
	out := captureOutput(func() {
		_ = cmd.Execute(nil, nil)
	})

	if !strings.Contains(out, "No file modifications to rollback") {
		t.Errorf("expected no-edits message, got: %s", out)
	}
}

func TestRollbackCommand_RollbackWrite(t *testing.T) {
	orig := state.GlobalState
	state.GlobalState = state.NewAppState()
	defer func() { state.GlobalState = orig }()

	tmpFile := filepath.Join(t.TempDir(), "test.txt")
	_ = os.WriteFile(tmpFile, []byte("original"), 0644)

	state.GlobalState.AddEdit(state.Edit{
		Timestamp:     time.Now(),
		Tool:          "file_write",
		FilePath:      tmpFile,
		Operation:     "write",
		Description:   "Wrote file",
		BeforeContent: []byte("original"),
	})
	// Simulate the write
	_ = os.WriteFile(tmpFile, []byte("modified"), 0644)

	cmd := NewRollbackCommand()
	out := captureOutput(func() {
		_ = cmd.Execute(nil, nil)
	})

	if !strings.Contains(out, "Rolled back") {
		t.Errorf("expected rollback confirmation, got: %s", out)
	}

	content, _ := os.ReadFile(tmpFile)
	if string(content) != "original" {
		t.Errorf("expected restored content 'original', got %q", string(content))
	}

	if len(state.GlobalState.GetEdits()) != 0 {
		t.Error("expected edits list to be empty after rollback")
	}
}

func TestRollbackCommand_RollbackMove(t *testing.T) {
	orig := state.GlobalState
	state.GlobalState = state.NewAppState()
	defer func() { state.GlobalState = orig }()

	tmpDir := t.TempDir()
	src := filepath.Join(tmpDir, "a.txt")
	dst := filepath.Join(tmpDir, "b.txt")
	_ = os.WriteFile(src, []byte("hello"), 0644)
	_ = os.Rename(src, dst)

	state.GlobalState.AddEdit(state.Edit{
		Timestamp:   time.Now(),
		Tool:        "file_move",
		FilePath:    src,
		Operation:   "move",
		Description: "Moved file",
		ExtraPath:   dst,
	})

	cmd := NewRollbackCommand()
	_ = cmd.Execute(nil, nil)

	if _, err := os.Stat(src); err != nil {
		t.Errorf("expected source file to be restored, got error: %v", err)
	}
	if _, err := os.Stat(dst); !os.IsNotExist(err) {
		t.Error("expected destination file to no longer exist")
	}
}

func TestRollbackCommand_Registered(t *testing.T) {
	reg := NewRegistry()
	_ = reg.Register(NewRollbackCommand())
	if _, ok := reg.Get("rollback"); !ok {
		t.Error("rollback command not registered")
	}
}
