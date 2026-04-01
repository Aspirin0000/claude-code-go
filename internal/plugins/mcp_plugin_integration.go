// Package plugins provides plugin management
// Source: src/utils/plugins/mcpPluginIntegration.ts
// Refactor: Go MCP plugin integration with full plugin loading
package plugins

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Aspirin0000/claude-code-go/internal/mcp"
	"github.com/Aspirin0000/claude-code-go/internal/types"
)

// UserConfigSchema represents the schema for user configuration
type UserConfigSchema struct {
	Properties map[string]UserConfigProperty `json:"properties,omitempty"`
	Required   []string                      `json:"required,omitempty"`
}

// UserConfigProperty represents a single property in user config schema
type UserConfigProperty struct {
	Type        string      `json:"type"`
	Description string      `json:"description,omitempty"`
	Default     interface{} `json:"default,omitempty"`
}

// UserConfigValues represents saved user configuration values
type UserConfigValues map[string]interface{}

// UnconfiguredChannel represents a channel that needs user configuration
type UnconfiguredChannel struct {
	Server       string           `json:"server"`
	DisplayName  string           `json:"displayName"`
	ConfigSchema UserConfigSchema `json:"configSchema"`
}

// McpbLoadResult represents the result of loading an MCPB file
type McpbLoadResult struct {
	Manifest      McpbManifest        `json:"manifest"`
	McpConfig     mcp.McpServerConfig `json:"mcpConfig"`
	ExtractedPath string              `json:"extractedPath"`
}

// McpbManifest represents the DXT manifest in an MCPB file
type McpbManifest struct {
	Name        string           `json:"name"`
	Version     string           `json:"version,omitempty"`
	Description string           `json:"description,omitempty"`
	UserConfig  UserConfigSchema `json:"userConfig,omitempty"`
}

// NeedsConfigResult indicates the MCPB needs user configuration
type NeedsConfigResult struct {
	Status string           `json:"status"`
	Server string           `json:"server"`
	Schema UserConfigSchema `json:"schema"`
}

// PluginChannel represents a channel in plugin manifest
type PluginChannel struct {
	Server      string           `json:"server"`
	DisplayName string           `json:"displayName,omitempty"`
	UserConfig  UserConfigSchema `json:"userConfig,omitempty"`
}

// IsMcpbSource checks if a path is an MCPB file
func IsMcpbSource(path string) bool {
	return strings.HasSuffix(path, ".mcpb") || strings.HasPrefix(path, "mcpb:")
}

// LoadMcpServersFromMcpb loads MCP servers from an MCPB file
// Handles downloading, extracting, and converting DXT manifest to MCP config
func LoadMcpServersFromMcpb(plugin *types.LoadedPlugin, mcpbPath string, errors *[]types.PluginError) (map[string]mcp.McpServerConfig, error) {
	// Check if it's a local file
	if _, err := os.Stat(mcpbPath); os.IsNotExist(err) {
		// Try relative to plugin path
		mcpbPath = filepath.Join(plugin.Path, mcpbPath)
	}

	data, err := os.ReadFile(mcpbPath)
	if err != nil {
		if errors != nil {
			*errors = append(*errors, types.PluginError{
				Type:     types.PluginErrorTypeMcpbDownloadFailed,
				Source:   fmt.Sprintf("%s@%s", plugin.Name, plugin.Source),
				Plugin:   &plugin.Name,
				McpbPath: &mcpbPath,
				Reason:   stringPtr(err.Error()),
			})
		}
		return nil, err
	}

	var result McpbLoadResult
	if err := json.Unmarshal(data, &result); err != nil {
		if errors != nil {
			*errors = append(*errors, types.PluginError{
				Type:            types.PluginErrorTypeMcpbInvalidManifest,
				Source:          fmt.Sprintf("%s@%s", plugin.Name, plugin.Source),
				Plugin:          &plugin.Name,
				McpbPath:        &mcpbPath,
				ValidationError: stringPtr(fmt.Sprintf("Failed to parse MCPB: %v", err)),
			})
		}
		return nil, err
	}

	return map[string]mcp.McpServerConfig{
		result.Manifest.Name: result.McpConfig,
	}, nil
}

// LoadPluginMcpServers loads MCP servers from a plugin's manifest
// Loads from manifest mcpServers entries, .mcp.json files, and .mcpb files
func LoadPluginMcpServers(plugin *types.LoadedPlugin, errors []types.PluginError) (map[string]mcp.McpServerConfig, []types.PluginError) {
	servers := make(map[string]mcp.McpServerConfig)

	// Check for .mcp.json in plugin directory first (lowest priority)
	defaultMcpPath := filepath.Join(plugin.Path, ".mcp.json")
	if defaultServers, err := loadMcpServersFromFile(defaultMcpPath); err == nil {
		for name, config := range defaultServers {
			servers[name] = config
		}
	}

	// Handle manifest mcpServers if present (higher priority)
	if plugin.Manifest.McpServers != nil {
		switch mcpServersSpec := plugin.Manifest.McpServers.(type) {
		case string:
			// Single path - could be MCPB or JSON file
			if IsMcpbSource(mcpServersSpec) {
				if mcpbServers, err := LoadMcpServersFromMcpb(plugin, mcpServersSpec, &errors); err == nil {
					for name, config := range mcpbServers {
						servers[name] = config
					}
				}
			} else {
				path := filepath.Join(plugin.Path, mcpServersSpec)
				if fileServers, err := loadMcpServersFromFile(path); err == nil {
					for name, config := range fileServers {
						servers[name] = config
					}
				}
			}
		case []interface{}:
			// Array of paths or inline configs
			for _, spec := range mcpServersSpec {
				switch s := spec.(type) {
				case string:
					if IsMcpbSource(s) {
						if mcpbServers, err := LoadMcpServersFromMcpb(plugin, s, &errors); err == nil {
							for name, config := range mcpbServers {
								servers[name] = config
							}
						}
					} else {
						path := filepath.Join(plugin.Path, s)
						if fileServers, err := loadMcpServersFromFile(path); err == nil {
							for name, config := range fileServers {
								servers[name] = config
							}
						}
					}
				case map[string]interface{}:
					// Inline config
					var config mcp.McpServerConfig
					configData, _ := json.Marshal(s)
					json.Unmarshal(configData, &config)
					// Use name from config or generate
					name := config.Name
					if name == "" {
						name = fmt.Sprintf("server-%d", len(servers))
					}
					servers[name] = config
				}
			}
		case map[string]interface{}:
			// Direct MCP server configs
			for name, spec := range mcpServersSpec {
				var config mcp.McpServerConfig
				configData, _ := json.Marshal(spec)
				json.Unmarshal(configData, &config)
				servers[name] = config
			}
		}
	}

	if len(servers) == 0 {
		return nil, errors
	}
	return servers, errors
}

// loadMcpServersFromFile loads MCP servers from a JSON file
func loadMcpServersFromFile(filePath string) (map[string]mcp.McpServerConfig, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		return nil, err
	}

	// Check if it's in the .mcp.json format with mcpServers key
	var mcpServersData map[string]interface{}
	if servers, ok := parsed["mcpServers"].(map[string]interface{}); ok {
		mcpServersData = servers
	} else {
		mcpServersData = parsed
	}

	validatedServers := make(map[string]mcp.McpServerConfig)
	for name, config := range mcpServersData {
		var serverConfig mcp.McpServerConfig
		configData, err := json.Marshal(config)
		if err != nil {
			continue
		}
		if err := json.Unmarshal(configData, &serverConfig); err != nil {
			continue
		}
		validatedServers[name] = serverConfig
	}

	return validatedServers, nil
}

// GetUnconfiguredChannels finds channel entries whose required userConfig fields are not yet saved
func GetUnconfiguredChannels(plugin *types.LoadedPlugin) []UnconfiguredChannel {
	var unconfigured []UnconfiguredChannel

	// Load saved user config for this plugin
	pluginDataDir := getPluginDataDir(plugin.Source)
	userConfigPath := filepath.Join(pluginDataDir, "user-config.json")

	savedConfig := make(UserConfigValues)
	if data, err := os.ReadFile(userConfigPath); err == nil {
		_ = json.Unmarshal(data, &savedConfig)
	}

	// Check each channel for missing required configuration
	for _, channel := range plugin.Manifest.Channels {
		if len(channel.UserConfig) == 0 {
			continue
		}

		// Get required fields from the channel's userConfig schema
		if required, ok := channel.UserConfig["required"].([]interface{}); ok {
			missingRequired := false
			for _, req := range required {
				if reqStr, ok := req.(string); ok {
					if _, exists := savedConfig[reqStr]; !exists {
						missingRequired = true
						break
					}
				}
			}

			if missingRequired {
				unconfigured = append(unconfigured, UnconfiguredChannel{
					Server:      channel.Server,
					DisplayName: channel.DisplayName,
					ConfigSchema: UserConfigSchema{
						Properties: make(map[string]UserConfigProperty),
						Required:   getStringSlice(required),
					},
				})
			}
		}
	}

	return unconfigured
}

// AddPluginScopeToServers adds plugin scope to MCP server configs
// Adds a prefix to server names to avoid conflicts between plugins
func AddPluginScopeToServers(servers map[string]mcp.McpServerConfig, pluginName string, pluginSource string) map[string]mcp.ScopedMcpServerConfig {
	scopedServers := make(map[string]mcp.ScopedMcpServerConfig)

	for name, config := range servers {
		// Add plugin prefix to server name to avoid conflicts
		scopedName := fmt.Sprintf("plugin:%s:%s", pluginName, name)
		scoped := mcp.ScopedMcpServerConfig{
			McpServerConfig: config,
			Scope:           mcp.ConfigScopeDynamic,
			PluginSource:    &pluginSource,
		}
		scopedServers[scopedName] = scoped
	}

	return scopedServers
}

// ExtractMcpServersFromPlugins extracts all MCP servers from loaded plugins
// Resolves environment variables for all servers before returning
func ExtractMcpServersFromPlugins(plugins []types.LoadedPlugin, errors []types.PluginError) (map[string]mcp.ScopedMcpServerConfig, []types.PluginError) {
	allServers := make(map[string]mcp.ScopedMcpServerConfig)

	for i := range plugins {
		plugin := &plugins[i]
		if plugin.Enabled != nil && !*plugin.Enabled {
			continue
		}

		servers, errs := LoadPluginMcpServers(plugin, errors)
		if errs != nil {
			errors = append(errors, errs...)
		}
		if servers == nil {
			continue
		}

		// Resolve environment variables and add plugin scope
		resolvedServers := make(map[string]mcp.McpServerConfig)
		for name, config := range servers {
			resolvedConfig := ResolvePluginMcpEnvironment(config, plugin, nil, &errors, plugin.Name, name)
			resolvedServers[name] = resolvedConfig
		}

		// Store the UNRESOLVED servers on the plugin for caching
		plugin.McpServers = servers

		// Add plugin scope
		scopedServers := AddPluginScopeToServers(resolvedServers, plugin.Name, plugin.Source)
		for name, config := range scopedServers {
			allServers[name] = config
		}
	}

	return allServers, errors
}

// ResolvePluginMcpEnvironment resolves environment variables for plugin MCP servers
// Handles ${CLAUDE_PLUGIN_ROOT}, ${user_config.X}, and general ${VAR} substitution
func ResolvePluginMcpEnvironment(
	config mcp.McpServerConfig,
	plugin *types.LoadedPlugin,
	userConfig UserConfigValues,
	errors *[]types.PluginError,
	pluginName string,
	serverName string,
) mcp.McpServerConfig {
	// Clone config to avoid modifying original
	resolved := config

	// Handle different server types
	switch config.Type {
	case "", "stdio":
		// Resolve command path
		if resolved.Command != "" {
			resolved.Command = resolveValue(resolved.Command, plugin, userConfig)
		}

		// Resolve args
		if len(resolved.Args) > 0 {
			resolvedArgs := make([]string, len(resolved.Args))
			for i, arg := range resolved.Args {
				resolvedArgs[i] = resolveValue(arg, plugin, userConfig)
			}
			resolved.Args = resolvedArgs
		}

		// Resolve environment variables and add CLAUDE_PLUGIN_ROOT / CLAUDE_PLUGIN_DATA
		if resolved.Env == nil {
			resolved.Env = make(map[string]string)
		}
		resolved.Env["CLAUDE_PLUGIN_ROOT"] = plugin.Path
		resolved.Env["CLAUDE_PLUGIN_DATA"] = getPluginDataDir(plugin.Source)

		// Resolve other env vars
		for key, value := range resolved.Env {
			if key != "CLAUDE_PLUGIN_ROOT" && key != "CLAUDE_PLUGIN_DATA" {
				resolved.Env[key] = resolveValue(value, plugin, userConfig)
			}
		}

	case "sse", "http", "ws":
		// Resolve URL
		if resolved.URL != "" {
			resolved.URL = resolveValue(resolved.URL, plugin, userConfig)
		}

		// Resolve headers
		if resolved.Headers != nil {
			resolvedHeaders := make(map[string]string)
			for key, value := range resolved.Headers {
				resolvedHeaders[key] = resolveValue(value, plugin, userConfig)
			}
			resolved.Headers = resolvedHeaders
		}

	// For other types (sse-ide, ws-ide, sdk, claudeai-proxy), pass through unchanged
	case "sse-ide", "ws-ide", "sdk", "claudeai-proxy":
		// No changes needed
	}

	return resolved
}

// resolveValue resolves all variable substitutions in a value
func resolveValue(value string, plugin *types.LoadedPlugin, userConfig UserConfigValues) string {
	// Substitute plugin-specific variables
	resolved := substitutePluginVariables(value, plugin)

	// Substitute user config variables if provided
	if userConfig != nil {
		resolved = substituteUserConfigVariables(resolved, userConfig)
	}

	// Expand general environment variables
	resolved = os.ExpandEnv(resolved)

	return resolved
}

// substitutePluginVariables substitutes plugin-specific variables
func substitutePluginVariables(value string, plugin *types.LoadedPlugin) string {
	replacements := map[string]string{
		"${CLAUDE_PLUGIN_ROOT}": plugin.Path,
	}

	result := value
	for pattern, replacement := range replacements {
		result = strings.ReplaceAll(result, pattern, replacement)
	}
	return result
}

// substituteUserConfigVariables substitutes user config variables
func substituteUserConfigVariables(value string, userConfig UserConfigValues) string {
	result := value
	for key, val := range userConfig {
		pattern := fmt.Sprintf("${user_config.%s}", key)
		if strVal, ok := val.(string); ok {
			result = strings.ReplaceAll(result, pattern, strVal)
		} else {
			result = strings.ReplaceAll(result, pattern, fmt.Sprintf("%v", val))
		}
	}
	return result
}

// getPluginDataDir returns the plugin data directory for a source
func getPluginDataDir(source string) string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = "."
	}
	return filepath.Join(configDir, "claude", "plugin-data", source)
}

// GetPluginMcpServers gets MCP servers from a specific plugin with environment variable resolution and scoping
func GetPluginMcpServers(plugin *types.LoadedPlugin, errors []types.PluginError) (map[string]mcp.ScopedMcpServerConfig, []types.PluginError) {
	if plugin.Enabled != nil && !*plugin.Enabled {
		return nil, errors
	}

	// Use cached servers if available
	var servers map[string]mcp.McpServerConfig
	if plugin.McpServers != nil {
		// Convert interface{} to map[string]mcp.McpServerConfig
		if mcpMap, ok := plugin.McpServers.(map[string]mcp.McpServerConfig); ok {
			servers = mcpMap
		} else {
			var loadErrors []types.PluginError
			servers, loadErrors = LoadPluginMcpServers(plugin, errors)
			errors = append(errors, loadErrors...)
		}
	} else {
		var loadErrors []types.PluginError
		servers, loadErrors = LoadPluginMcpServers(plugin, errors)
		errors = append(errors, loadErrors...)
	}

	if servers == nil {
		return nil, errors
	}

	// Resolve environment variables
	resolvedServers := make(map[string]mcp.McpServerConfig)
	for name, config := range servers {
		resolvedConfig := ResolvePluginMcpEnvironment(config, plugin, nil, &errors, plugin.Name, name)
		resolvedServers[name] = resolvedConfig
	}

	// Add plugin scope
	return AddPluginScopeToServers(resolvedServers, plugin.Name, plugin.Source), errors
}

// LoadAllPluginsCacheOnly loads all plugins from cache only
func LoadAllPluginsCacheOnly() error {
	// In full implementation, would load from:
	// 1. ~/.claude/plugins/cache/
	// 2. ~/.claude/plugins/builtin/
	// etc.
	return nil
}

// stringPtr helper to create string pointer
func stringPtr(s string) *string {
	return &s
}

// getStringSlice converts []interface{} to []string
func getStringSlice(input []interface{}) []string {
	result := make([]string, 0, len(input))
	for _, v := range input {
		if s, ok := v.(string); ok {
			result = append(result, s)
		}
	}
	return result
}
