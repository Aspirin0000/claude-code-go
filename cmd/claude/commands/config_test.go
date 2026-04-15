package commands

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/Aspirin0000/claude-code-go/internal/config"
)

func TestMaskAPIKey(t *testing.T) {
	cmd := NewConfigCommand()

	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"short", "****"},
		{"12345678", "****"},
		{"123456789", "1234****6789"},
		{"sk-ant-api03-xxxxxxxx", "sk-a****xxxx"},
	}

	for _, tt := range tests {
		result := cmd.maskAPIKey(tt.input)
		if result != tt.expected {
			t.Errorf("maskAPIKey(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestFindConfigMeta(t *testing.T) {
	cmd := NewConfigCommand()

	meta := cmd.findConfigMeta("model")
	if meta == nil {
		t.Fatal("expected to find meta for 'model'")
	}
	if meta.Key != "model" {
		t.Errorf("expected key 'model', got %q", meta.Key)
	}

	notFound := cmd.findConfigMeta("nonexistent")
	if notFound != nil {
		t.Error("expected nil for nonexistent key")
	}
}

func TestGetBasicConfigValue(t *testing.T) {
	cmd := NewConfigCommand()
	cfg := config.DefaultConfig()

	val, err := cmd.getBasicConfigValue(cfg, "model")
	if err != nil {
		t.Fatalf("getBasicConfigValue failed: %v", err)
	}
	if val == "" {
		t.Error("expected non-empty model value")
	}

	_, err = cmd.getBasicConfigValue(cfg, "nonexistent")
	if err == nil {
		t.Error("expected error for unknown key")
	}
}

func TestConfigCommand_ShowAll(t *testing.T) {
	cmd := NewConfigCommand()
	ctx := context.Background()

	// Use temp config dir
	tmpDir := t.TempDir()
	t.Setenv("CLAUDE_CONFIG_DIR", tmpDir)

	cfg := config.DefaultConfig()
	cp := config.GetConfigPath()
	_ = cfg.Save(cp)

	if err := cmd.Execute(ctx, []string{}); err != nil {
		t.Fatalf("show all config failed: %v", err)
	}
}

func TestConfigCommand_GetAndSet(t *testing.T) {
	cmd := NewConfigCommand()
	ctx := context.Background()

	tmpDir := t.TempDir()
	t.Setenv("CLAUDE_CONFIG_DIR", tmpDir)

	cp := config.GetConfigPath()
	_ = os.MkdirAll(filepath.Dir(cp), 0755)
	cfg := config.DefaultConfig()
	_ = cfg.Save(cp)

	// Get a config value
	if err := cmd.Execute(ctx, []string{"get", "model"}); err != nil {
		t.Fatalf("get config failed: %v", err)
	}

	// Set a config value
	if err := cmd.Execute(ctx, []string{"set", "theme", "light"}); err != nil {
		t.Fatalf("set config failed: %v", err)
	}

	// Verify it was saved
	loaded, err := config.Load(cp)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
	if loaded.Theme != "light" {
		t.Errorf("expected theme 'light', got %q", loaded.Theme)
	}
}

func TestConfigCommand_List(t *testing.T) {
	cmd := NewConfigCommand()
	ctx := context.Background()

	if err := cmd.Execute(ctx, []string{"list"}); err != nil {
		t.Fatalf("list config failed: %v", err)
	}
}

func TestConfigCommand_InvalidSubcommand(t *testing.T) {
	cmd := NewConfigCommand()
	ctx := context.Background()

	if err := cmd.Execute(ctx, []string{"invalid"}); err != nil {
		t.Fatalf("invalid subcommand should not error: %v", err)
	}
}

func TestConfigCommand_MissingGetArg(t *testing.T) {
	cmd := NewConfigCommand()
	ctx := context.Background()

	if err := cmd.Execute(ctx, []string{"get"}); err != nil {
		t.Fatalf("missing get arg should not error: %v", err)
	}
}

func TestConfigCommand_MissingSetArg(t *testing.T) {
	cmd := NewConfigCommand()
	ctx := context.Background()

	if err := cmd.Execute(ctx, []string{"set", "model"}); err != nil {
		t.Fatalf("missing set arg should not error: %v", err)
	}
}

func TestConfigCommand_SetNestedEnv(t *testing.T) {
	cmd := NewConfigCommand()
	ctx := context.Background()

	tmpDir := t.TempDir()
	t.Setenv("CLAUDE_CONFIG_DIR", tmpDir)

	cp := config.GetConfigPath()
	_ = os.MkdirAll(filepath.Dir(cp), 0755)
	cfg := config.DefaultConfig()
	_ = cfg.Save(cp)

	if err := cmd.Execute(ctx, []string{"set", "env.TEST_VAR", "test_value"}); err != nil {
		t.Fatalf("set nested env failed: %v", err)
	}

	loaded, err := config.Load(cp)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
	if loaded.Env["TEST_VAR"] != "test_value" {
		t.Errorf("expected env TEST_VAR='test_value', got %q", loaded.Env["TEST_VAR"])
	}
}

func TestConfigCommand_GetNestedEnv(t *testing.T) {
	cmd := NewConfigCommand()
	ctx := context.Background()

	tmpDir := t.TempDir()
	t.Setenv("CLAUDE_CONFIG_DIR", tmpDir)

	cp := config.GetConfigPath()
	_ = os.MkdirAll(filepath.Dir(cp), 0755)
	cfg := config.DefaultConfig()
	cfg.Env["EXISTING_VAR"] = "existing_value"
	_ = cfg.Save(cp)

	if err := cmd.Execute(ctx, []string{"get", "env.EXISTING_VAR"}); err != nil {
		t.Fatalf("get nested env failed: %v", err)
	}
}
