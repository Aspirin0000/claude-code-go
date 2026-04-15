package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Aspirin0000/claude-code-go/internal/config"
)

func TestFormatNumber(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{0, "0"},
		{999, "999"},
		{1000, "1,000"},
		{1234567, "1,234,567"},
		{1000000000, "1,000,000,000"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("format_%d", tt.input), func(t *testing.T) {
			result := formatNumber(tt.input)
			if result != tt.expected {
				t.Errorf("formatNumber(%d) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestModelCommandShowCurrentModel(t *testing.T) {
	cmd := NewModelCommand()
	ctx := context.Background()

	// Should work without args (shows current model)
	if err := cmd.Execute(ctx, []string{}); err != nil {
		t.Fatalf("show current model failed: %v", err)
	}
}

func TestModelCommandListModels(t *testing.T) {
	cmd := NewModelCommand()
	ctx := context.Background()

	if err := cmd.Execute(ctx, []string{"list"}); err != nil {
		t.Fatalf("list models failed: %v", err)
	}
}

func TestModelCommandSwitchModel(t *testing.T) {
	cmd := NewModelCommand()
	ctx := context.Background()

	// Use a temporary config directory
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	// Save a default config there
	cfg := config.DefaultConfig()
	if err := cfg.Save(configPath); err != nil {
		t.Fatalf("failed to save test config: %v", err)
	}

	// Override the getCurrentModel behavior by setting env
	// But the command itself calls config.GetConfigPath() which we can't override
	// So instead we test that it handles valid and invalid model names gracefully
	// by checking error behavior for invalid models

	// For a known model, it should succeed if we can point config at tmpDir
	// Since we can't easily inject config path, we test the partial match logic
	// indirectly via switchModel by testing getCurrentModel with env var
	t.Setenv("CLAUDE_CODE_MODEL", "claude-3-haiku-20240307")

	if err := cmd.Execute(ctx, []string{"haiku"}); err != nil {
		t.Fatalf("switch model failed: %v", err)
	}

	// Verify env is read back
	current := cmd.getCurrentModel()
	if current != "claude-3-haiku-20240307" {
		t.Errorf("expected current model from env, got %s", current)
	}
}

func TestModelCommandGetCurrentModelPriority(t *testing.T) {
	cmd := NewModelCommand()

	// Clean env
	t.Setenv("CLAUDE_CODE_MODEL", "")

	// Default should be returned when no env and no config
	defaultModel := config.DefaultConfig().Model
	current := cmd.getCurrentModel()
	if current != defaultModel {
		t.Errorf("expected default model %q, got %q", defaultModel, current)
	}

	// Env variable should take priority
	t.Setenv("CLAUDE_CODE_MODEL", "env-model")
	current = cmd.getCurrentModel()
	if current != "env-model" {
		t.Errorf("expected env model, got %q", current)
	}
}

func TestAvailableModelsContainExpected(t *testing.T) {
	expectedModels := []string{
		"claude-sonnet-4-20250514",
		"claude-opus-4-20250514",
		"claude-3-opus-20240229",
		"claude-3-sonnet-20240229",
		"claude-3-haiku-20240307",
	}

	for _, modelID := range expectedModels {
		if _, ok := AvailableModels[modelID]; !ok {
			t.Errorf("expected model %q to be in AvailableModels", modelID)
		}
	}
}

func TestModelCommandSwitchToSameModel(t *testing.T) {
	cmd := NewModelCommand()
	ctx := context.Background()

	// Switch to the model currently set in env/config
	current := cmd.getCurrentModel()
	if err := cmd.Execute(ctx, []string{current}); err != nil {
		t.Fatalf("switch to same model failed: %v", err)
	}
}

func TestModelCommandInvalidPartialMatch(t *testing.T) {
	cmd := NewModelCommand()
	ctx := context.Background()

	// An arbitrary unknown model ID should be accepted (custom model support)
	if err := cmd.Execute(ctx, []string{"custom-unknown-model-v99"}); err != nil {
		t.Fatalf("switch to custom model should succeed: %v", err)
	}
}

func TestModelCommandPartialMatch(t *testing.T) {
	cmd := NewModelCommand()
	ctx := context.Background()

	// "sonnet" should match one of the sonnet models
	// We can't easily verify the exact match without config injection,
	// but we can verify it doesn't error
	t.Setenv("CLAUDE_CODE_MODEL", "claude-3-haiku-20240307")
	if err := cmd.Execute(ctx, []string{"sonnet"}); err != nil {
		t.Fatalf("partial match failed: %v", err)
	}
}

func TestModelCommandHelp(t *testing.T) {
	cmd := NewModelCommand()
	help := cmd.Help()
	if help == "" {
		t.Error("expected non-empty help text")
	}
	if !strings.Contains(help, "/model") {
		t.Error("expected help to mention /model")
	}
}

func TestModelCommandAliases(t *testing.T) {
	cmd := NewModelCommand()
	aliases := cmd.Aliases()
	expectedAliases := map[string]bool{"m": true, "switch-model": true}
	for _, alias := range aliases {
		if !expectedAliases[alias] {
			t.Errorf("unexpected alias: %s", alias)
		}
		delete(expectedAliases, alias)
	}
	if len(expectedAliases) > 0 {
		t.Errorf("missing expected aliases: %v", expectedAliases)
	}
}

func TestModelCommandSwitchModelWithTempConfig(t *testing.T) {
	cmd := NewModelCommand()
	ctx := context.Background()

	// Create temp dir and config file
	tmpDir := t.TempDir()

	// We monkey-patch config loading by overriding the config path indirectly
	// config.GetConfigPath uses os.UserConfigDir, which we can't override easily
	// So we test via setting env var to bypass config file
	t.Setenv("CLAUDE_CODE_MODEL", "")

	// The command will load from real config path, which may have a previous value.
	// To ensure a clean test, write to the actual config path with our temp dir
	// by temporarily changing UserConfigDir behavior via XDG_CONFIG_HOME on Unix
	// or LOCALAPPDATA on Windows
	if os.Getenv("XDG_CONFIG_HOME") == "" {
		t.Setenv("XDG_CONFIG_HOME", tmpDir)
	} else {
		// If already set, we can't easily override for just this test without side effects
		// Skip this part of the test
		t.Skip("XDG_CONFIG_HOME already set, skipping config file test")
	}

	// Verify config path uses temp dir
	cp := config.GetConfigPath()
	if !strings.HasPrefix(cp, tmpDir) {
		t.Skipf("config path not in temp dir: %s", cp)
	}

	// Save a config with a known model
	cfg := config.DefaultConfig()
	cfg.Model = "claude-3-haiku-20240307"
	if err := cfg.Save(cp); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	// Now switch to another known model
	if err := cmd.Execute(ctx, []string{"claude-sonnet-4-20250514"}); err != nil {
		t.Fatalf("switch model failed: %v", err)
	}

	// Verify it was saved
	loaded, err := config.Load(cp)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}
	if loaded.Model != "claude-sonnet-4-20250514" {
		t.Errorf("expected model to be saved, got %q", loaded.Model)
	}
}
