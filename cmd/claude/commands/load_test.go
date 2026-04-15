package commands

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/Aspirin0000/claude-code-go/internal/state"
)

func TestLoadCommand_JSON(t *testing.T) {
	cmd := NewLoadCommand()
	ctx := context.Background()

	// Save old messages and restore after test
	oldMessages := state.GlobalState.GetMessages()
	defer state.GlobalState.SetMessages(oldMessages)

	tmpDir := t.TempDir()
	jsonFile := filepath.Join(tmpDir, "test.json")
	content := `{
		"sessionId": "test-session",
		"messages": [
			{"uuid": "1", "type": "user", "role": "user", "content": "hello"},
			{"uuid": "2", "type": "assistant", "role": "assistant", "content": "hi"}
		]
	}`
	_ = os.WriteFile(jsonFile, []byte(content), 0644)

	if err := cmd.Execute(ctx, []string{jsonFile}); err != nil {
		t.Fatalf("load json failed: %v", err)
	}

	msgs := state.GlobalState.GetMessages()
	if len(msgs) != 2 {
		t.Errorf("expected 2 messages, got %d", len(msgs))
	}
}

func TestLoadCommand_Markdown(t *testing.T) {
	cmd := NewLoadCommand()
	ctx := context.Background()

	oldMessages := state.GlobalState.GetMessages()
	defer state.GlobalState.SetMessages(oldMessages)

	tmpDir := t.TempDir()
	mdFile := filepath.Join(tmpDir, "test.md")
	content := "# Session\n\n## User\n\nhello\n\n## Assistant\n\nhi\n"
	_ = os.WriteFile(mdFile, []byte(content), 0644)

	if err := cmd.Execute(ctx, []string{mdFile}); err != nil {
		t.Fatalf("load markdown failed: %v", err)
	}

	msgs := state.GlobalState.GetMessages()
	if len(msgs) != 2 {
		t.Errorf("expected 2 messages, got %d", len(msgs))
	}
}

func TestLoadCommand_NoArgs(t *testing.T) {
	cmd := NewLoadCommand()
	ctx := context.Background()

	if err := cmd.Execute(ctx, []string{}); err == nil {
		t.Fatal("expected error for missing filename")
	}
}

func TestLoadCommand_FileNotFound(t *testing.T) {
	cmd := NewLoadCommand()
	ctx := context.Background()

	if err := cmd.Execute(ctx, []string{"/nonexistent/file.json"}); err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoadCommand_InvalidJSON(t *testing.T) {
	cmd := NewLoadCommand()
	ctx := context.Background()

	tmpDir := t.TempDir()
	jsonFile := filepath.Join(tmpDir, "invalid.json")
	_ = os.WriteFile(jsonFile, []byte("not json"), 0644)

	if err := cmd.Execute(ctx, []string{jsonFile}); err == nil {
		t.Fatal("expected error for invalid json")
	}
}

func TestLoadCommand_EmptyMessages(t *testing.T) {
	cmd := NewLoadCommand()
	ctx := context.Background()

	tmpDir := t.TempDir()
	jsonFile := filepath.Join(tmpDir, "empty.json")
	content := `{"sessionId": "test", "messages": []}`
	_ = os.WriteFile(jsonFile, []byte(content), 0644)

	if err := cmd.Execute(ctx, []string{jsonFile}); err == nil {
		t.Fatal("expected error for empty messages")
	}
}

func TestLoadCommand_AutoDetect(t *testing.T) {
	cmd := NewLoadCommand()
	ctx := context.Background()

	oldMessages := state.GlobalState.GetMessages()
	defer state.GlobalState.SetMessages(oldMessages)

	tmpDir := t.TempDir()
	file := filepath.Join(tmpDir, "conversation.txt")
	content := `{"sessionId": "test", "messages": [{"uuid": "1", "type": "user", "content": "hello"}]}`
	_ = os.WriteFile(file, []byte(content), 0644)

	if err := cmd.Execute(ctx, []string{file}); err != nil {
		t.Fatalf("auto detect load failed: %v", err)
	}
}

func TestLoadCommand_Aliases(t *testing.T) {
	cmd := NewLoadCommand()
	aliases := cmd.Aliases()

	expected := map[string]bool{"import": true, "restore-file": true}
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
