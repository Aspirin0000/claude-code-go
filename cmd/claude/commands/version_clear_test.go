package commands

import (
	"context"
	"testing"
)

func TestVersionCommand(t *testing.T) {
	cmd := NewVersionCommand()
	ctx := context.Background()

	if err := cmd.Execute(ctx, []string{}); err != nil {
		t.Fatalf("version command failed: %v", err)
	}
}

func TestClearCommand(t *testing.T) {
	cmd := NewClearCommand()
	ctx := context.Background()

	// This should not error; actual screen clearing is hard to verify in tests
	if err := cmd.Execute(ctx, []string{}); err != nil {
		t.Fatalf("clear command failed: %v", err)
	}
}
