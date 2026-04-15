package commands

import (
	"context"
	"os"
	"testing"

	"github.com/Aspirin0000/claude-code-go/internal/config"
)

func TestInitCommand_CreatesConfig(t *testing.T) {
	cmd := NewInitCommand()
	ctx := context.Background()

	tmpDir := t.TempDir()
	t.Setenv("CLAUDE_CONFIG_DIR", tmpDir)

	if err := cmd.Execute(ctx, []string{}); err != nil {
		t.Fatalf("init command failed: %v", err)
	}

	// Verify config file was created
	cp := config.GetConfigPath()
	if _, err := os.Stat(cp); os.IsNotExist(err) {
		t.Errorf("expected config file to be created at %s", cp)
	}
}

func TestInitCommand_Idempotent(t *testing.T) {
	cmd := NewInitCommand()
	ctx := context.Background()

	tmpDir := t.TempDir()
	t.Setenv("CLAUDE_CONFIG_DIR", tmpDir)

	// First init
	if err := cmd.Execute(ctx, []string{}); err != nil {
		t.Fatalf("first init failed: %v", err)
	}

	// Second init should be idempotent (warns but doesn't error)
	if err := cmd.Execute(ctx, []string{}); err != nil {
		t.Fatalf("second init should not error: %v", err)
	}
}
