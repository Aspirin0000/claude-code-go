package commands

import (
	"context"
	"path/filepath"
	"testing"
)

func TestMemoryCommand_SetAndGet(t *testing.T) {
	cmd := NewMemoryCommand()

	// Use a temp file for memories
	tmpDir := t.TempDir()
	cmd.getMemoryFilePath = func() string {
		return filepath.Join(tmpDir, "memory.json")
	}

	ctx := context.Background()

	// Set a memory
	if err := cmd.Execute(ctx, []string{"set", "test_key", "test_value"}); err != nil {
		t.Fatalf("set memory failed: %v", err)
	}

	// Get the memory
	if err := cmd.Execute(ctx, []string{"get", "test_key"}); err != nil {
		t.Fatalf("get memory failed: %v", err)
	}
}

func TestMemoryCommand_ListAndDelete(t *testing.T) {
	cmd := NewMemoryCommand()

	tmpDir := t.TempDir()
	cmd.getMemoryFilePath = func() string {
		return filepath.Join(tmpDir, "memory.json")
	}

	ctx := context.Background()

	// Set multiple memories
	_ = cmd.Execute(ctx, []string{"set", "key1", "value1"})
	_ = cmd.Execute(ctx, []string{"set", "key2", "value2"})

	// List
	if err := cmd.Execute(ctx, []string{"list"}); err != nil {
		t.Fatalf("list memories failed: %v", err)
	}

	// Delete
	if err := cmd.Execute(ctx, []string{"delete", "key1"}); err != nil {
		t.Fatalf("delete memory failed: %v", err)
	}

	// Verify deletion by getting
	if err := cmd.Execute(ctx, []string{"get", "key1"}); err == nil {
		t.Fatal("expected error after deleting memory")
	}
}

func TestMemoryCommand_Search(t *testing.T) {
	cmd := NewMemoryCommand()

	tmpDir := t.TempDir()
	cmd.getMemoryFilePath = func() string {
		return filepath.Join(tmpDir, "memory.json")
	}

	ctx := context.Background()

	_ = cmd.Execute(ctx, []string{"set", "project_lang", "Go"})
	_ = cmd.Execute(ctx, []string{"set", "framework", "Gin"})

	if err := cmd.Execute(ctx, []string{"search", "Go"}); err != nil {
		t.Fatalf("search memories failed: %v", err)
	}
}

func TestMemoryCommand_Clear(t *testing.T) {
	cmd := NewMemoryCommand()

	tmpDir := t.TempDir()
	cmd.getMemoryFilePath = func() string {
		return filepath.Join(tmpDir, "memory.json")
	}

	ctx := context.Background()

	_ = cmd.Execute(ctx, []string{"set", "key1", "value1"})

	// Clear should work even with empty stdin
	if err := cmd.Execute(ctx, []string{"clear"}); err != nil {
		t.Fatalf("clear memories failed: %v", err)
	}
}

func TestMemoryCommand_InvalidAction(t *testing.T) {
	cmd := NewMemoryCommand()

	tmpDir := t.TempDir()
	cmd.getMemoryFilePath = func() string {
		return filepath.Join(tmpDir, "memory.json")
	}

	ctx := context.Background()

	if err := cmd.Execute(ctx, []string{"unknown"}); err == nil {
		t.Fatal("expected error for unknown action")
	}
}
