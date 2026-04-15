package commands

import (
	"context"
	"testing"
)

// testCommand is a concrete command implementation for testing
type testCommand struct {
	*BaseCommand
}

func newTestCommand(name, desc string) *testCommand {
	return &testCommand{
		BaseCommand: NewBaseCommand(name, desc, CategoryGeneral),
	}
}

func (c *testCommand) Execute(ctx context.Context, args []string) error {
	return nil
}

func TestRegistry(t *testing.T) {
	// Create a new registry for testing
	reg := NewRegistry()

	// Create a test command
	testCmd := newTestCommand("test", "Test command")

	// Test registration
	err := reg.Register(testCmd)
	if err != nil {
		t.Errorf("Failed to register command: %v", err)
	}

	// Test duplicate registration
	err = reg.Register(testCmd)
	if err != ErrCommandAlreadyExists {
		t.Errorf("Expected ErrCommandAlreadyExists, got: %v", err)
	}

	// Test Get
	cmd, found := reg.Get("test")
	if !found {
		t.Error("Command not found after registration")
	}
	if cmd.Name() != "test" {
		t.Errorf("Expected command name 'test', got '%s'", cmd.Name())
	}

	// Test Get with non-existent command
	_, found = reg.Get("nonexistent")
	if found {
		t.Error("Found non-existent command")
	}

	// Test List
	cmds := reg.List()
	if len(cmds) != 1 {
		t.Errorf("Expected 1 command, got %d", len(cmds))
	}
}

func TestBaseCommand(t *testing.T) {
	cmd := NewBaseCommand("test", "Test description", CategoryGeneral).
		WithAliases("t", "tst").
		WithHelp("Test help text")

	if cmd.Name() != "test" {
		t.Errorf("Expected name 'test', got '%s'", cmd.Name())
	}

	if cmd.Description() != "Test description" {
		t.Errorf("Expected description 'Test description', got '%s'", cmd.Description())
	}

	if cmd.Category() != CategoryGeneral {
		t.Errorf("Expected category CategoryGeneral, got %v", cmd.Category())
	}

	aliases := cmd.Aliases()
	if len(aliases) != 2 {
		t.Errorf("Expected 2 aliases, got %d", len(aliases))
	}

	if cmd.Help() != "Test help text" {
		t.Errorf("Expected help 'Test help text', got '%s'", cmd.Help())
	}
}
