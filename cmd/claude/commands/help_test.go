package commands

import (
	"context"
	"strings"
	"testing"
)

func TestHelpCommand_ShowAllCommands(t *testing.T) {
	cmd := NewHelpCommand()
	ctx := context.Background()

	if err := cmd.Execute(ctx, []string{}); err != nil {
		t.Fatalf("show all commands failed: %v", err)
	}
}

func TestHelpCommand_ShowSpecificCommand(t *testing.T) {
	cmd := NewHelpCommand()
	ctx := context.Background()

	if err := cmd.Execute(ctx, []string{"clear"}); err != nil {
		t.Fatalf("show specific command help failed: %v", err)
	}
}

func TestHelpCommand_UnknownCommand(t *testing.T) {
	cmd := NewHelpCommand()
	ctx := context.Background()

	// Should handle unknown command gracefully
	if err := cmd.Execute(ctx, []string{"nonexistent-command"}); err != nil {
		t.Fatalf("help for unknown command should not error: %v", err)
	}
}

func TestHelpCommand_Aliases(t *testing.T) {
	cmd := NewHelpCommand()
	aliases := cmd.Aliases()

	expected := map[string]bool{"h": true, "?": true}
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

func TestHelpCommand_NameAndDescription(t *testing.T) {
	cmd := NewHelpCommand()
	if cmd.Name() != "help" {
		t.Errorf("expected name 'help', got %q", cmd.Name())
	}
	if !strings.Contains(cmd.Description(), "help") {
		t.Errorf("expected description to contain 'help', got %q", cmd.Description())
	}
}

func TestHelpCommand_Category(t *testing.T) {
	cmd := NewHelpCommand()
	if cmd.Category() != CategoryGeneral {
		t.Errorf("expected category %v, got %v", CategoryGeneral, cmd.Category())
	}
}
