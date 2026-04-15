package commands

import (
	"os"
	"strings"
	"testing"

	"github.com/Aspirin0000/claude-code-go/internal/config"
)

func TestThemeCommand_ShowCurrent(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("CLAUDE_CONFIG_DIR", tmpDir)

	cfg := config.DefaultConfig()
	cfg.Theme = "dark"
	_ = cfg.Save(config.GetConfigPath())

	cmd := NewThemeCommand()
	out := captureOutput(func() {
		_ = cmd.Execute(nil, nil)
	})

	if !strings.Contains(out, "dark") {
		t.Errorf("expected output to contain current theme 'dark', got: %s", out)
	}
}

func TestThemeCommand_Switch(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("CLAUDE_CONFIG_DIR", tmpDir)

	cfg := config.DefaultConfig()
	cfg.Theme = "dark"
	_ = cfg.Save(config.GetConfigPath())

	cmd := NewThemeCommand()
	out := captureOutput(func() {
		_ = cmd.Execute(nil, []string{"light"})
	})

	if !strings.Contains(out, "Switched to light theme") {
		t.Errorf("expected switch confirmation, got: %s", out)
	}

	cfg, err := config.Load(config.GetConfigPath())
	if err != nil {
		t.Fatalf("failed to reload config: %v", err)
	}
	if cfg.Theme != "light" {
		t.Errorf("expected theme 'light', got %q", cfg.Theme)
	}
}

func TestThemeCommand_SwitchSameTheme(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("CLAUDE_CONFIG_DIR", tmpDir)

	cfg := config.DefaultConfig()
	cfg.Theme = "dark"
	_ = cfg.Save(config.GetConfigPath())

	cmd := NewThemeCommand()
	out := captureOutput(func() {
		_ = cmd.Execute(nil, []string{"dark"})
	})

	if !strings.Contains(out, "Already using dark theme") {
		t.Errorf("expected already-using message, got: %s", out)
	}
}

func TestThemeCommand_InvalidTheme(t *testing.T) {
	cmd := NewThemeCommand()
	err := cmd.Execute(nil, []string{"neon"})
	if err == nil || !strings.Contains(err.Error(), "unknown theme") {
		t.Errorf("expected unknown theme error, got: %v", err)
	}
}

func TestThemeCommand_EnvOverride(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("CLAUDE_CONFIG_DIR", tmpDir)
	t.Setenv("CLAUDE_THEME", "light")

	cfg := config.DefaultConfig()
	cfg.Theme = "dark"
	_ = cfg.Save(config.GetConfigPath())

	cmd := NewThemeCommand()
	current := cmd.getCurrentTheme()
	if current != "light" {
		t.Errorf("expected env override 'light', got %q", current)
	}
}

func TestThemeCommand_Registered(t *testing.T) {
	reg := NewRegistry()
	_ = reg.Register(NewThemeCommand())
	if _, ok := reg.Get("theme"); !ok {
		t.Error("theme command not registered")
	}
}

func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	f()
	_ = w.Close()
	os.Stdout = old
	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	return string(buf[:n])
}
