package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Model != "claude-sonnet-4-20250514" {
		t.Errorf("unexpected default model: %s", cfg.Model)
	}
	if cfg.Theme != "dark" {
		t.Errorf("unexpected default theme: %s", cfg.Theme)
	}
	if cfg.Provider != "anthropic" {
		t.Errorf("unexpected default provider: %s", cfg.Provider)
	}
	if !cfg.AutoSave {
		t.Error("expected auto_save to be true by default")
	}
	if cfg.Projects == nil {
		t.Error("expected projects map to be initialized")
	}
}

func TestLoadSaveRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config.json")

	cfg := &Config{
		APIKey:   "test-key",
		Model:    "claude-test",
		Theme:    "light",
		Provider: "bedrock",
		AutoSave: false,
		Projects: map[string]ProjectConfig{
			"proj1": {
				AllowedTools: []string{"bash"},
			},
		},
	}

	if err := cfg.Save(path); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if loaded.APIKey != "test-key" {
		t.Errorf("unexpected api key: %s", loaded.APIKey)
	}
	if loaded.Model != "claude-test" {
		t.Errorf("unexpected model: %s", loaded.Model)
	}
	if loaded.Provider != "bedrock" {
		t.Errorf("unexpected provider: %s", loaded.Provider)
	}
	if loaded.AutoSave != false {
		t.Errorf("expected auto_save false, got %v", loaded.AutoSave)
	}
	if len(loaded.Projects["proj1"].AllowedTools) != 1 {
		t.Errorf("unexpected allowed tools: %+v", loaded.Projects["proj1"].AllowedTools)
	}
}

func TestLoadMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nonexistent.json")
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error for missing file: %v", err)
	}
	if cfg.Model != "claude-sonnet-4-20250514" {
		t.Error("expected default config for missing file")
	}
}

func TestLoadInvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "bad.json")
	os.WriteFile(path, []byte("not json"), 0644)

	_, err := Load(path)
	if err == nil {
		t.Fatal("expected error for invalid json")
	}
}

func TestGetConfigPath(t *testing.T) {
	path := GetConfigPath()
	if path == "" {
		t.Error("expected non-empty config path")
	}
}

func TestGetProjectConfig(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Projects["/tmp/proj"] = ProjectConfig{
		AllowedTools: []string{"bash"},
	}

	pc := cfg.GetProjectConfig("/tmp/proj")
	if len(pc.AllowedTools) != 1 || pc.AllowedTools[0] != "bash" {
		t.Errorf("unexpected project config: %+v", pc)
	}

	pc2 := cfg.GetProjectConfig("/nonexistent")
	if pc2.AllowedTools == nil {
		t.Error("expected non-nil AllowedTools for missing project")
	}
}

func TestGetAutoSaveDir(t *testing.T) {
	cfg := DefaultConfig()
	cfg.AutoSaveDir = "/custom/sessions"
	if cfg.GetAutoSaveDir() != "/custom/sessions" {
		t.Errorf("unexpected auto-save dir: %s", cfg.GetAutoSaveDir())
	}

	cfg2 := DefaultConfig()
	cfg2.AutoSaveDir = ""
	if cfg2.GetAutoSaveDir() == "" {
		t.Error("expected default auto-save dir to be non-empty")
	}
}
