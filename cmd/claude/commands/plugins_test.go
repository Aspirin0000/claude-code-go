package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Aspirin0000/claude-code-go/internal/plugins"
)

func TestPluginsCommand_Empty(t *testing.T) {
	cmd := NewPluginsCommand()
	out := captureOutput(func() {
		_ = cmd.Execute(nil, nil)
	})

	if !strings.Contains(out, "Plugin Manager") {
		t.Errorf("expected header, got: %s", out)
	}
	if !strings.Contains(out, plugins.GetPluginsDirectory()) {
		t.Errorf("expected plugin directory, got: %s", out)
	}
}

func TestPluginsCommand_WithInstalled(t *testing.T) {
	// Create a temporary plugin cache structure
	cacheDir := t.TempDir()
	_ = os.MkdirAll(filepath.Join(cacheDir, "marketplace1", "plugin-a"), 0755)
	_ = os.MkdirAll(filepath.Join(cacheDir, "marketplace1", "plugin-b"), 0755)

	cmd := NewPluginsCommand()
	installed := cmd.listInstalledPlugins(cacheDir)

	if len(installed) != 2 {
		t.Fatalf("expected 2 plugins, got %d", len(installed))
	}

	found := map[string]bool{}
	for _, p := range installed {
		found[p] = true
	}
	if !found["plugin-a@marketplace1"] || !found["plugin-b@marketplace1"] {
		t.Errorf("unexpected plugins: %v", installed)
	}
}

func TestPluginsCommand_Registered(t *testing.T) {
	reg := NewRegistry()
	_ = reg.Register(NewPluginsCommand())
	if _, ok := reg.Get("plugins"); !ok {
		t.Error("plugins command not registered")
	}
}
