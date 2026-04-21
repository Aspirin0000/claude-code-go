package settings

import (
	"os"
	"path/filepath"
	"testing"
)

// ============================================================================
// Plugin Only Policy Tests
// ============================================================================

func TestIsValidCustomizationSurface(t *testing.T) {
	if !IsValidCustomizationSurface("agents") {
		t.Error("expected 'agents' to be valid")
	}
	if !IsValidCustomizationSurface("commands") {
		t.Error("expected 'commands' to be valid")
	}
	if !IsValidCustomizationSurface("hooks") {
		t.Error("expected 'hooks' to be valid")
	}
	if !IsValidCustomizationSurface("outputStyles") {
		t.Error("expected 'outputStyles' to be valid")
	}
	if IsValidCustomizationSurface("invalid") {
		t.Error("expected 'invalid' to be invalid")
	}
}

func TestIsSourceAdminTrusted(t *testing.T) {
	if !IsSourceAdminTrusted("plugin") {
		t.Error("expected 'plugin' to be admin trusted")
	}
	if !IsSourceAdminTrusted("policySettings") {
		t.Error("expected 'policySettings' to be admin trusted")
	}
	if !IsSourceAdminTrusted("built-in") {
		t.Error("expected 'built-in' to be admin trusted")
	}
	if !IsSourceAdminTrusted("builtin") {
		t.Error("expected 'builtin' to be admin trusted")
	}
	if !IsSourceAdminTrusted("bundled") {
		t.Error("expected 'bundled' to be admin trusted")
	}
	if IsSourceAdminTrusted("") {
		t.Error("expected empty string to not be admin trusted")
	}
	if IsSourceAdminTrusted("user") {
		t.Error("expected 'user' to not be admin trusted")
	}
}

func TestIsAllowedByPolicy(t *testing.T) {
	// When not restricted, everything should be allowed
	if !IsAllowedByPolicy(CustomizationSurfaceAgents, "user") {
		t.Error("expected allowed when not restricted")
	}

	// When restricted, only admin-trusted sources should be allowed
	// This depends on policy settings, so we test the logic
	if !IsAllowedByPolicy(CustomizationSurfaceCommands, "plugin") {
		t.Error("expected plugin to be allowed")
	}
}

func TestGetRestrictedSurfaces(t *testing.T) {
	surfaces := GetRestrictedSurfaces()
	// Should return a slice (possibly empty)
	if surfaces == nil {
		t.Skip("surfaces is nil, skipping")
	}
}

func TestIsStrictPluginOnlyEnabled(t *testing.T) {
	// This depends on policy settings
	_ = IsStrictPluginOnlyEnabled()
}

func TestShouldAllowManagedMcpServersOnly(t *testing.T) {
	// This depends on policy settings
	_ = ShouldAllowManagedMcpServersOnly()
}

// ============================================================================
// Settings Loading Tests
// ============================================================================

func TestLoadSettingsFromFile(t *testing.T) {
	// Test with non-existent file
	settings := loadSettingsFromFile("/non/existent/path.json")
	if settings != nil {
		t.Error("expected nil for non-existent file")
	}

	// Test with invalid JSON
	tmpDir := t.TempDir()
	invalidPath := filepath.Join(tmpDir, "invalid.json")
	os.WriteFile(invalidPath, []byte("not json"), 0644)

	settings = loadSettingsFromFile(invalidPath)
	if settings != nil {
		t.Error("expected nil for invalid JSON")
	}

	// Test with valid JSON
	validPath := filepath.Join(tmpDir, "valid.json")
	validJSON := `{"strictPluginOnlyCustomization": true}`
	os.WriteFile(validPath, []byte(validJSON), 0644)

	settings = loadSettingsFromFile(validPath)
	if settings == nil {
		t.Fatal("expected non-nil settings")
	}
}

func TestLoadPolicySettings(t *testing.T) {
	// Test with environment variable
	tmpDir := t.TempDir()
	policyPath := filepath.Join(tmpDir, "policy.json")
	os.WriteFile(policyPath, []byte(`{"strictPluginOnlyCustomization": true}`), 0644)

	os.Setenv("CLAUDE_POLICY_SETTINGS_PATH", policyPath)
	defer os.Unsetenv("CLAUDE_POLICY_SETTINGS_PATH")

	settings := loadPolicySettings()
	if settings == nil {
		t.Fatal("expected non-nil settings")
	}
}

func TestLoadUserSettings(t *testing.T) {
	// This will try to load from user config dir
	settings := loadUserSettings()
	// May be nil if file doesn't exist, which is fine
	_ = settings
}

func TestLoadProjectSettings(t *testing.T) {
	// This will try to load from current directory
	settings := loadProjectSettings()
	// May be nil if file doesn't exist, which is fine
	_ = settings
}

// ============================================================================
// IsRestrictedToPluginOnly Tests
// ============================================================================

func TestIsRestrictedToPluginAll(t *testing.T) {
	// Create a temporary policy settings file
	tmpDir := t.TempDir()
	policyPath := filepath.Join(tmpDir, "policy.json")
	os.WriteFile(policyPath, []byte(`{"strictPluginOnlyCustomization": true}`), 0644)

	os.Setenv("CLAUDE_POLICY_SETTINGS_PATH", policyPath)
	defer os.Unsetenv("CLAUDE_POLICY_SETTINGS_PATH")

	if !IsRestrictedToPluginOnly(CustomizationSurfaceAgents) {
		t.Error("expected agents to be restricted")
	}
	if !IsRestrictedToPluginOnly(CustomizationSurfaceCommands) {
		t.Error("expected commands to be restricted")
	}
}

func TestIsRestrictedToPluginOnlyArray(t *testing.T) {
	// Create a temporary policy settings file with array
	tmpDir := t.TempDir()
	policyPath := filepath.Join(tmpDir, "policy.json")
	os.WriteFile(policyPath, []byte(`{"strictPluginOnlyCustomization": ["agents", "hooks"]}`), 0644)

	os.Setenv("CLAUDE_POLICY_SETTINGS_PATH", policyPath)
	defer os.Unsetenv("CLAUDE_POLICY_SETTINGS_PATH")

	if !IsRestrictedToPluginOnly(CustomizationSurfaceAgents) {
		t.Error("expected agents to be restricted")
	}
	if !IsRestrictedToPluginOnly(CustomizationSurfaceHooks) {
		t.Error("expected hooks to be restricted")
	}
	if IsRestrictedToPluginOnly(CustomizationSurfaceCommands) {
		t.Error("expected commands to not be restricted")
	}
}

func TestIsRestrictedToPluginOnlyNotRestricted(t *testing.T) {
	// Create a temporary policy settings file with false
	tmpDir := t.TempDir()
	policyPath := filepath.Join(tmpDir, "policy.json")
	os.WriteFile(policyPath, []byte(`{"strictPluginOnlyCustomization": false}`), 0644)

	os.Setenv("CLAUDE_POLICY_SETTINGS_PATH", policyPath)
	defer os.Unsetenv("CLAUDE_POLICY_SETTINGS_PATH")

	if IsRestrictedToPluginOnly(CustomizationSurfaceAgents) {
		t.Error("expected agents to not be restricted")
	}
}

// ============================================================================
// Managed Path Tests
// ============================================================================

func TestGetManagedFilePathExtended(t *testing.T) {
	path := GetManagedFilePath()
	if path == "" {
		t.Error("expected non-empty managed file path")
	}
	if !filepath.IsAbs(path) {
		t.Error("expected absolute path")
	}
}
