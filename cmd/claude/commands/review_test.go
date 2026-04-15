package commands

import (
	"context"
	"testing"
)

func TestReviewCommand_Default(t *testing.T) {
	cmd := NewReviewCommand()
	ctx := context.Background()

	if err := cmd.Execute(ctx, []string{}); err != nil {
		t.Fatalf("default review failed: %v", err)
	}
}

func TestReviewCommand_Changes(t *testing.T) {
	cmd := NewReviewCommand()
	ctx := context.Background()

	if err := cmd.Execute(ctx, []string{"changes"}); err != nil {
		t.Fatalf("review changes failed: %v", err)
	}
}

func TestReviewCommand_Plan(t *testing.T) {
	cmd := NewReviewCommand()
	ctx := context.Background()

	if err := cmd.Execute(ctx, []string{"plan"}); err != nil {
		t.Fatalf("review plan failed: %v", err)
	}
}

func TestReviewCommand_Git(t *testing.T) {
	cmd := NewReviewCommand()
	ctx := context.Background()

	if err := cmd.Execute(ctx, []string{"git", "5"}); err != nil {
		t.Fatalf("review git failed: %v", err)
	}
}

func TestReviewCommand_Summary(t *testing.T) {
	cmd := NewReviewCommand()
	ctx := context.Background()

	if err := cmd.Execute(ctx, []string{"summary"}); err != nil {
		t.Fatalf("review summary failed: %v", err)
	}
}

func TestReviewCommand_InvalidSubcommand(t *testing.T) {
	cmd := NewReviewCommand()
	ctx := context.Background()

	// Unknown subcommand falls back to comprehensive review
	if err := cmd.Execute(ctx, []string{"unknown"}); err != nil {
		t.Fatalf("review unknown should not error: %v", err)
	}
}
