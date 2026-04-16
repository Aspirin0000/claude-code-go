package commands

import (
	"strings"
	"testing"

	"github.com/Aspirin0000/claude-code-go/internal/config"
)

func TestAgentCommand_MissingArgs(t *testing.T) {
	cmd := NewAgentCommand()
	err := cmd.Execute(nil, nil)
	if err == nil || !strings.Contains(err.Error(), "usage: /agent") {
		t.Errorf("expected usage error, got: %v", err)
	}
}

func TestAgentCommand_NoAPIKey(t *testing.T) {
	// Ensure no API key is available
	t.Setenv("ANTHROPIC_API_KEY", "")
	tmpDir := t.TempDir()
	t.Setenv("CLAUDE_CONFIG_DIR", tmpDir)
	_ = config.DefaultConfig().Save(config.GetConfigPath())

	cmd := NewAgentCommand()
	err := cmd.Execute(nil, []string{"coder", "write a function"})
	if err == nil || !strings.Contains(err.Error(), "API key not configured") {
		t.Errorf("expected API key error, got: %v", err)
	}
}

func TestAgentCommand_Registered(t *testing.T) {
	reg := NewRegistry()
	_ = reg.Register(NewAgentCommand())
	if _, ok := reg.Get("agent"); !ok {
		t.Error("agent command not registered")
	}
}
