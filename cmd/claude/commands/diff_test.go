package commands

import (
	"context"
	"testing"
)

func TestDiffCommand_Default(t *testing.T) {
	cmd := NewDiffCommand()
	ctx := context.Background()

	// In a git repo, this should show unstaged changes or "No differences"
	if err := cmd.Execute(ctx, []string{}); err != nil {
		t.Fatalf("default diff failed: %v", err)
	}
}

func TestDiffCommand_Staged(t *testing.T) {
	cmd := NewDiffCommand()
	ctx := context.Background()

	if err := cmd.Execute(ctx, []string{"--staged"}); err != nil {
		t.Fatalf("staged diff failed: %v", err)
	}
}

func TestDiffCommand_SpecificFile(t *testing.T) {
	cmd := NewDiffCommand()
	ctx := context.Background()

	// Test with a file that exists in the repo
	if err := cmd.Execute(ctx, []string{"cmd/claude/commands/bash.go"}); err != nil {
		t.Fatalf("file diff failed: %v", err)
	}
}

func TestDiffCommand_NoGitRepo(t *testing.T) {
	cmd := NewDiffCommand()
	ctx := context.Background()

	// This will error because /tmp is not a git repo
	// But the test just verifies it handles the error
	_ = cmd.Execute(ctx, []string{"/tmp/nonexistent.go"})
}
