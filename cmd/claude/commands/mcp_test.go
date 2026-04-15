package commands

import (
	"context"
	"testing"
)

func TestMCPCommand_Overview(t *testing.T) {
	cmd := NewMCPCommand()
	ctx := context.Background()

	if err := cmd.Execute(ctx, []string{}); err != nil {
		t.Fatalf("mcp overview failed: %v", err)
	}
}

func TestMCPCommand_List(t *testing.T) {
	cmd := NewMCPCommand()
	ctx := context.Background()

	if err := cmd.Execute(ctx, []string{"list"}); err != nil {
		t.Fatalf("mcp list failed: %v", err)
	}
}

func TestMCPCommand_StatusMissingName(t *testing.T) {
	cmd := NewMCPCommand()
	ctx := context.Background()

	if err := cmd.Execute(ctx, []string{"status"}); err == nil {
		t.Fatal("expected error for missing server name")
	}
}

func TestMCPCommand_UnknownSubcommand(t *testing.T) {
	cmd := NewMCPCommand()
	ctx := context.Background()

	if err := cmd.Execute(ctx, []string{"unknown"}); err == nil {
		t.Fatal("expected error for unknown subcommand")
	}
}

func TestMCPListCommand_Empty(t *testing.T) {
	cmd := NewMCPListCommand()
	ctx := context.Background()

	if err := cmd.Execute(ctx, []string{}); err != nil {
		t.Fatalf("mcp-list failed: %v", err)
	}
}

func TestMCPListCommand_Alias(t *testing.T) {
	cmd := NewMCPListCommand()
	aliases := cmd.Aliases()

	found := false
	for _, alias := range aliases {
		if alias == "mcps" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'mcps' alias")
	}
}

func TestMCPAddCommand_Interactive(t *testing.T) {
	cmd := NewMCPAddCommand()
	ctx := context.Background()

	// No args enters interactive mode, but with no stdin input it will error
	// We just verify it doesn't panic
	_ = cmd.Execute(ctx, []string{})
}

func TestMCPAddCommand_MissingName(t *testing.T) {
	cmd := NewMCPAddCommand()
	ctx := context.Background()

	if err := cmd.Execute(ctx, []string{""}); err == nil {
		t.Fatal("expected error for empty server name")
	}
}

func TestMCPRemoveCommand_NoArgs(t *testing.T) {
	cmd := NewMCPRemoveCommand()
	ctx := context.Background()

	if err := cmd.Execute(ctx, []string{}); err == nil {
		t.Fatal("expected error for missing server name")
	}
}

func TestMCPRemoveCommand_Nonexistent(t *testing.T) {
	cmd := NewMCPRemoveCommand()
	ctx := context.Background()

	// Removing a nonexistent server should return an error
	if err := cmd.Execute(ctx, []string{"nonexistent-server"}); err == nil {
		t.Fatal("expected error for nonexistent server")
	}
}
