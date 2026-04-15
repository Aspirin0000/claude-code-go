package claudeinchrome

import "testing"

func TestIsClaudeInChromeMCPServer(t *testing.T) {
	if !IsClaudeInChromeMCPServer("claude-in-chrome") {
		t.Error("expected claude-in-chrome to match")
	}
	if !IsClaudeInChromeMCPServer("claude-in-chrome-mcp") {
		t.Error("expected claude-in-chrome-mcp to match")
	}
	if IsClaudeInChromeMCPServer("other-server") {
		t.Error("expected other-server to not match")
	}
	if IsClaudeInChromeMCPServer("") {
		t.Error("expected empty string to not match")
	}
}
