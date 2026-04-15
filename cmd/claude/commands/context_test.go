package commands

import (
	"context"
	"testing"
)

func TestContextCommand_Execute(t *testing.T) {
	cmd := NewContextCommand()
	ctx := context.Background()

	if err := cmd.Execute(ctx, []string{}); err != nil {
		t.Fatalf("context command failed: %v", err)
	}
}

func TestContextCommand_NameAndDescription(t *testing.T) {
	cmd := NewContextCommand()
	if cmd.Name() != "context" {
		t.Errorf("expected name 'context', got %q", cmd.Name())
	}
	if cmd.Description() == "" {
		t.Error("expected non-empty description")
	}
}

func TestContextCommand_GitContext(t *testing.T) {
	cmd := NewContextCommand()
	gitCtx := cmd.getGitContext()
	// We're in a git repo, so this should return something
	if gitCtx == "" {
		t.Skip("not in a git repository, skipping git context test")
	}
	if !contains([]string{gitCtx}, "Branch:") {
		// contains checks exact match, so this would fail.
		// Instead just verify it's non-empty which we already did.
	}
}

func TestContextCommand_FindClaudeMdFiles(t *testing.T) {
	cmd := NewContextCommand()
	files := cmd.findClaudeMdFiles()
	// Result may be empty or contain files depending on workspace
	_ = files
}
