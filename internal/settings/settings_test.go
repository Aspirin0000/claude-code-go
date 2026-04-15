package settings

import (
	"testing"
)

func TestIsMcpServerNameEntry(t *testing.T) {
	name := "test"
	entry := McpServerEntry{ServerName: &name}
	if !IsMcpServerNameEntry(entry) {
		t.Error("expected true for name entry")
	}
	if IsMcpServerCommandEntry(entry) {
		t.Error("expected false for command entry check on name entry")
	}
}

func TestIsMcpServerCommandEntry(t *testing.T) {
	entry := McpServerEntry{ServerCommand: []string{"node", "server.js"}}
	if !IsMcpServerCommandEntry(entry) {
		t.Error("expected true for command entry")
	}
	if IsMcpServerUrlEntry(entry) {
		t.Error("expected false for URL entry check on command entry")
	}
}

func TestIsMcpServerUrlEntry(t *testing.T) {
	url := "http://localhost:3000"
	entry := McpServerEntry{ServerUrl: &url}
	if !IsMcpServerUrlEntry(entry) {
		t.Error("expected true for URL entry")
	}
}

func TestIsSettingSourceEnabled(t *testing.T) {
	if !IsSettingSourceEnabled(SettingSourceUser) {
		t.Error("expected user source to be enabled")
	}
	if !IsSettingSourceEnabled(SettingSourceProject) {
		t.Error("expected project source to be enabled")
	}
}

func TestIsSettingSourceEnabledConst(t *testing.T) {
	if !IsSettingSourceEnabledConst("user") {
		t.Error("expected 'user' to be valid source")
	}
	if !IsSettingSourceEnabledConst("policy") {
		t.Error("expected 'policy' to be valid source")
	}
	if IsSettingSourceEnabledConst("invalid") {
		t.Error("expected 'invalid' to be rejected")
	}
}

func TestGetInitialSettings(t *testing.T) {
	s := GetInitialSettings()
	if s == nil {
		t.Fatal("expected non-nil settings")
	}
	if s.AllowedMcpServers != nil {
		t.Error("expected nil allowed servers in initial settings")
	}
}

func TestGetSettingsForSource(t *testing.T) {
	s := GetSettingsForSource(SettingSourceLocal)
	if s == nil {
		t.Fatal("expected non-nil settings")
	}
}

func TestGetManagedFilePath(t *testing.T) {
	path := GetManagedFilePath()
	if path == "" {
		t.Error("expected non-empty managed file path")
	}
}
