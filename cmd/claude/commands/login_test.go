package commands

import (
	"context"
	"testing"

	"github.com/Aspirin0000/claude-code-go/internal/config"
)

func TestLoginCommand_WithArg(t *testing.T) {
	cmd := NewLoginCommand()
	ctx := context.Background()

	tmpDir := t.TempDir()
	t.Setenv("CLAUDE_CONFIG_DIR", tmpDir)

	apiKey := "sk-ant-test-key-12345"
	if err := cmd.Execute(ctx, []string{apiKey}); err != nil {
		t.Fatalf("login with arg failed: %v", err)
	}

	cfg, err := config.Load(config.GetConfigPath())
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
	if cfg.APIKey != apiKey {
		t.Errorf("expected API key %q, got %q", apiKey, cfg.APIKey)
	}
}

func TestLoginCommand_EmptyArg(t *testing.T) {
	cmd := NewLoginCommand()
	ctx := context.Background()

	if err := cmd.Execute(ctx, []string{""}); err == nil {
		t.Fatal("expected error for empty API key")
	}
}

func TestLogoutCommand_RemovesKey(t *testing.T) {
	cmd := NewLogoutCommand()
	ctx := context.Background()

	tmpDir := t.TempDir()
	t.Setenv("CLAUDE_CONFIG_DIR", tmpDir)

	// Pre-populate config with API key
	cfg := config.DefaultConfig()
	cfg.APIKey = "sk-ant-test-key-12345"
	_ = cfg.Save(config.GetConfigPath())

	if err := cmd.Execute(ctx, []string{}); err != nil {
		t.Fatalf("logout failed: %v", err)
	}

	loaded, err := config.Load(config.GetConfigPath())
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
	if loaded.APIKey != "" {
		t.Errorf("expected API key to be cleared, got %q", loaded.APIKey)
	}
}

func TestLogoutCommand_NoKey(t *testing.T) {
	cmd := NewLogoutCommand()
	ctx := context.Background()

	tmpDir := t.TempDir()
	t.Setenv("CLAUDE_CONFIG_DIR", tmpDir)

	if err := cmd.Execute(ctx, []string{}); err != nil {
		t.Fatalf("logout with no key should not error: %v", err)
	}
}

func TestMin(t *testing.T) {
	if min(3, 5) != 3 {
		t.Error("min(3, 5) should be 3")
	}
	if min(10, 2) != 2 {
		t.Error("min(10, 2) should be 2")
	}
}
