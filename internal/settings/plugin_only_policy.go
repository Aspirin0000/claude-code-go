// Package settings provides settings management
// Source: src/utils/settings/pluginOnlyPolicy.ts
// Refactor: Go plugin-only policy implementation with config reading
package settings

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// CustomizationSurface represents different customization surfaces
// Corresponds to CUSTOMIZATION_SURFACES in TypeScript
type CustomizationSurface string

const (
	// CustomizationSurfaceAgents agent definitions surface
	CustomizationSurfaceAgents CustomizationSurface = "agents"
	// CustomizationSurfaceCommands slash commands surface
	CustomizationSurfaceCommands CustomizationSurface = "commands"
	// CustomizationSurfaceHooks hooks configuration surface
	CustomizationSurfaceHooks CustomizationSurface = "hooks"
	// CustomizationSurfaceOutputStyles output styles surface
	CustomizationSurfaceOutputStyles CustomizationSurface = "outputStyles"
)

// CUSTOMIZATION_SURFACES all valid customization surfaces
var CUSTOMIZATION_SURFACES = []CustomizationSurface{
	CustomizationSurfaceAgents,
	CustomizationSurfaceCommands,
	CustomizationSurfaceHooks,
	CustomizationSurfaceOutputStyles,
}

// IsValidCustomizationSurface checks if a surface is valid
func IsValidCustomizationSurface(surface string) bool {
	for _, s := range CUSTOMIZATION_SURFACES {
		if string(s) == surface {
			return true
		}
	}
	return false
}

// PolicySettings represents the strictPluginOnlyCustomization policy structure
// Can be: true (lock all), false/nil (lock none), or array of surfaces to lock
type PolicySettings struct {
	StrictPluginOnlyCustomization interface{} `json:"strictPluginOnlyCustomization"`
}

// SettingsForSource represents settings loaded from a specific source
type SettingsForSource struct {
	StrictPluginOnlyCustomization interface{} `json:"strictPluginOnlyCustomization"`
	// Add other policy settings as needed
}

// ADMIN_TRUSTED_SOURCES sources that bypass strictPluginOnlyCustomization
// Admin-trusted because:
//   - plugin — gated separately by strictKnownMarketplaces
//   - policySettings — from managed settings, admin-controlled by definition
//   - built-in / builtin / bundled — ship with the CLI, not user-authored
var ADMIN_TRUSTED_SOURCES = map[string]bool{
	"plugin":         true,
	"policySettings": true,
	"built-in":       true,
	"builtin":        true,
	"bundled":        true,
}

// IsRestrictedToPluginOnly checks whether a customization surface is locked to plugin-only sources
// by the managed strictPluginOnlyCustomization policy.
//
// "Locked" means user-level (~/.claude/*) and project-level (.claude/*)
// sources are skipped for that surface. Managed (policySettings) and
// plugin-provided sources always load regardless — the policy is admin-set,
// so managed sources are already admin-controlled, and plugins are gated
// separately via strictKnownMarketplaces.
//
// true locks all four surfaces; array form locks only those listed.
// Absent/undefined → nothing locked (the default).
func IsRestrictedToPluginOnly(surface CustomizationSurface) bool {
	policy := getSettingsForSource("policySettings")
	if policy == nil {
		return false
	}

	// Check if policy is explicitly true (lock all surfaces)
	if policyValue, ok := policy.StrictPluginOnlyCustomization.(bool); ok {
		return policyValue
	}

	// Check if policy is an array of surfaces to lock
	if policyArray, ok := policy.StrictPluginOnlyCustomization.([]interface{}); ok {
		for _, item := range policyArray {
			if itemStr, ok := item.(string); ok {
				if CustomizationSurface(itemStr) == surface {
					return true
				}
			}
		}
	}

	return false
}

// IsSourceAdminTrusted checks whether a customization's source is admin-trusted under
// strictPluginOnlyCustomization. Use this to gate frontmatter-hook
// registration and similar per-item checks where the item carries a
// source tag but the surface's filesystem loader already ran.
//
// Pattern at call sites:
//
//	allowed := !IsRestrictedToPluginOnly(surface) || IsSourceAdminTrusted(item.source)
//	if item.hooks && allowed { register(...) }
func IsSourceAdminTrusted(source string) bool {
	if source == "" {
		return false
	}
	return ADMIN_TRUSTED_SOURCES[source]
}

// IsAllowedByPolicy checks if loading from a source is allowed by current policy
// Combines surface restriction with source trust check
func IsAllowedByPolicy(surface CustomizationSurface, source string) bool {
	// If not restricted, everything is allowed
	if !IsRestrictedToPluginOnly(surface) {
		return true
	}

	// If restricted, only admin-trusted sources are allowed
	return IsSourceAdminTrusted(source)
}

// getSettingsForSource loads settings from a specific source
// In production, this would read from:
//   - policySettings: managed/admin settings file
//   - userSettings: ~/.claude/settings.json
//   - projectSettings: .claude/settings.json
//   - etc.
func getSettingsForSource(source string) *SettingsForSource {
	switch source {
	case "policySettings":
		return loadPolicySettings()
	case "userSettings":
		return loadUserSettings()
	case "projectSettings":
		return loadProjectSettings()
	default:
		return nil
	}
}

// loadPolicySettings loads policy settings from managed settings
// Looks for settings in CLAUDE_POLICY_SETTINGS_PATH or default location
func loadPolicySettings() *SettingsForSource {
	// Check environment variable override
	if policyPath := os.Getenv("CLAUDE_POLICY_SETTINGS_PATH"); policyPath != "" {
		return loadSettingsFromFile(policyPath)
	}

	// Default policy settings location
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil
	}

	policyPath := filepath.Join(configDir, "claude", "policy-settings.json")
	return loadSettingsFromFile(policyPath)
}

// loadUserSettings loads user-level settings
func loadUserSettings() *SettingsForSource {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil
	}

	settingsPath := filepath.Join(configDir, "claude", "settings.json")
	return loadSettingsFromFile(settingsPath)
}

// loadProjectSettings loads project-level settings from current directory
func loadProjectSettings() *SettingsForSource {
	cwd, err := os.Getwd()
	if err != nil {
		return nil
	}

	settingsPath := filepath.Join(cwd, ".claude", "settings.json")
	return loadSettingsFromFile(settingsPath)
}

// loadSettingsFromFile loads settings from a JSON file
func loadSettingsFromFile(path string) *SettingsForSource {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var settings SettingsForSource
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil
	}

	return &settings
}

// ShouldAllowManagedMcpServersOnly checks if only managed MCP servers should be allowed
// This is a convenience function for MCP server policy checking
func ShouldAllowManagedMcpServersOnly() bool {
	// Check if there's a specific MCP policy setting
	policy := getSettingsForSource("policySettings")
	if policy == nil {
		return false
	}

	// Could check for a specific managedMcpServersOnly field
	// For now, align with general strictPluginOnlyCustomization behavior
	return IsRestrictedToPluginOnly(CustomizationSurfaceAgents)
}

// GetRestrictedSurfaces returns the list of surfaces that are currently restricted
// Useful for debugging and UI display
func GetRestrictedSurfaces() []CustomizationSurface {
	var restricted []CustomizationSurface

	for _, surface := range CUSTOMIZATION_SURFACES {
		if IsRestrictedToPluginOnly(surface) {
			restricted = append(restricted, surface)
		}
	}

	return restricted
}

// IsStrictPluginOnlyEnabled checks if any surface has strict plugin-only mode enabled
func IsStrictPluginOnlyEnabled() bool {
	return len(GetRestrictedSurfaces()) > 0
}
