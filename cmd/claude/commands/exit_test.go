package commands

import (
	"testing"
)

func TestExitCommand_NameAndDescription(t *testing.T) {
	cmd := NewExitCommand()
	if cmd.Name() != "exit" {
		t.Errorf("expected name 'exit', got %q", cmd.Name())
	}
	if cmd.Description() == "" {
		t.Error("expected non-empty description")
	}
}

func TestExitCommand_Aliases(t *testing.T) {
	cmd := NewExitCommand()
	aliases := cmd.Aliases()

	expected := map[string]bool{"quit": true, "q": true, "bye": true}
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

func TestExitCommand_Category(t *testing.T) {
	cmd := NewExitCommand()
	if cmd.Category() != CategoryGeneral {
		t.Errorf("expected category %v, got %v", CategoryGeneral, cmd.Category())
	}
}
