package commands

import (
	"strings"
	"testing"

	"github.com/Aspirin0000/claude-code-go/internal/plugins"
)

func TestPluginInstallCommand_MissingArgs(t *testing.T) {
	cmd := NewPluginInstallCommand()
	err := cmd.Execute(nil, nil)
	if err == nil || !strings.Contains(err.Error(), "usage: /plugin-install") {
		t.Errorf("expected usage error, got: %v", err)
	}
}

func TestPluginInstallCommand_ParseSource(t *testing.T) {
	cmd := NewPluginInstallCommand()

	tests := []struct {
		input    string
		wantType string
		wantErr  bool
	}{
		{"local:/path/to/plugin", "local", false},
		{"npm:package-name", "npm", false},
		{"github:user/repo", "github", false},
		{"github:user/repo@v1.0", "github", false},
		{"url:https://example.com/plugin.git", "url", false},
		{"invalid", "", true},
		{"unknown:value", "", true},
	}

	for _, tt := range tests {
		source, err := cmd.parseSource(tt.input)
		if tt.wantErr {
			if err == nil {
				t.Errorf("expected error for %q", tt.input)
			}
			continue
		}
		if err != nil {
			t.Errorf("unexpected error for %q: %v", tt.input, err)
			continue
		}
		if source.Type != tt.wantType {
			t.Errorf("parseSource(%q).Type = %q, want %q", tt.input, source.Type, tt.wantType)
		}
	}
}

func TestPluginInstallCommand_ExtractName(t *testing.T) {
	cmd := NewPluginInstallCommand()

	tests := []struct {
		source plugins.PluginSource
		want   string
	}{
		{plugins.PluginSource{Type: "local", Path: "/path/to/my-plugin"}, "my-plugin"},
		{plugins.PluginSource{Type: "npm", Package: "test-package"}, "test-package"},
		{plugins.PluginSource{Type: "github", Repo: "user/repo-name"}, "repo-name"},
		{plugins.PluginSource{Type: "url", URL: "https://github.com/user/repo.git"}, "repo"},
	}

	for _, tt := range tests {
		got := cmd.extractPluginName(tt.source)
		if got != tt.want {
			t.Errorf("extractPluginName(%+v) = %q, want %q", tt.source, got, tt.want)
		}
	}
}

func TestPluginInstallCommand_Registered(t *testing.T) {
	reg := NewRegistry()
	_ = reg.Register(NewPluginInstallCommand())
	if _, ok := reg.Get("plugin-install"); !ok {
		t.Error("plugin-install command not registered")
	}
}
