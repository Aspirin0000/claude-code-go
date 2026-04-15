package commands

import (
	"context"
	"strings"
	"testing"
)

func TestToolsCommand_Default(t *testing.T) {
	cmd := NewToolsCommand()
	ctx := context.Background()

	if err := cmd.Execute(ctx, []string{}); err != nil {
		t.Fatalf("tools command failed: %v", err)
	}
}

func TestToolsCommand_CoreFilter(t *testing.T) {
	cmd := NewToolsCommand()
	ctx := context.Background()

	if err := cmd.Execute(ctx, []string{"core"}); err != nil {
		t.Fatalf("tools core failed: %v", err)
	}
}

func TestToolsCommand_SearchFilter(t *testing.T) {
	cmd := NewToolsCommand()
	ctx := context.Background()

	if err := cmd.Execute(ctx, []string{"search"}); err != nil {
		t.Fatalf("tools search failed: %v", err)
	}
}

func TestToolsCommand_TaskFilter(t *testing.T) {
	cmd := NewToolsCommand()
	ctx := context.Background()

	if err := cmd.Execute(ctx, []string{"task"}); err != nil {
		t.Fatalf("tools task failed: %v", err)
	}
}

func TestToolsCommand_WebFilter(t *testing.T) {
	cmd := NewToolsCommand()
	ctx := context.Background()

	if err := cmd.Execute(ctx, []string{"web"}); err != nil {
		t.Fatalf("tools web failed: %v", err)
	}
}

func TestToolsCommand_MCPFilter(t *testing.T) {
	cmd := NewToolsCommand()
	ctx := context.Background()

	if err := cmd.Execute(ctx, []string{"mcp"}); err != nil {
		t.Fatalf("tools mcp failed: %v", err)
	}
}

func TestToolsCommand_Aliases(t *testing.T) {
	cmd := NewToolsCommand()
	aliases := cmd.Aliases()

	expected := map[string]bool{"list-tools": true, "t": true}
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

func TestToolsCommand_NameAndDescription(t *testing.T) {
	cmd := NewToolsCommand()
	if cmd.Name() != "tools" {
		t.Errorf("expected name 'tools', got %q", cmd.Name())
	}
	if !strings.Contains(cmd.Description(), "tools") {
		t.Errorf("expected description to contain 'tools', got %q", cmd.Description())
	}
}
