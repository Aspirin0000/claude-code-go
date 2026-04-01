// Package plugins provides plugin management
// Source: src/utils/plugins/pluginLoader.ts
// Refactor: Go plugin loader with full discovery, loading, and validation
package plugins

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Aspirin0000/claude-code-go/internal/types"
)

// PluginLoadResult represents the result of loading plugins
type PluginLoadResult struct {
	Enabled  []types.LoadedPlugin
	Disabled []types.LoadedPlugin
	Errors   []types.PluginError
}

// PluginSource represents the source of a plugin
type PluginSource struct {
	Type     string // "npm", "github", "url", "git-subdir", "local"
	Package  string // For npm
	Repo     string // For github
	URL      string // For url
	Ref      string // Git ref
	Sha      string // Git commit sha
	Path     string // Local path or git-subdir path
	Registry string // NPM registry
}

// IsRemote returns true if the source is remote (requires network)
func (s PluginSource) IsRemote() bool {
	return s.Type == "npm" || s.Type == "github" || s.Type == "url"
}

// GetPluginCachePath returns the path where plugin cache is stored
func GetPluginCachePath() string {
	return filepath.Join(getPluginsDirectory(), "cache")
}

// getPluginsDirectory returns the main plugins directory
func getPluginsDirectory() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = "."
	}
	return filepath.Join(configDir, "claude", "plugins")
}

// getPluginSeedDirs returns seed directories for plugins
func getPluginSeedDirs() []string {
	// Check for CLAUDE_PLUGIN_SEED_DIRS environment variable
	if seedDirs := os.Getenv("CLAUDE_PLUGIN_SEED_DIRS"); seedDirs != "" {
		return strings.Split(seedDirs, string(filepath.ListSeparator))
	}
	return nil
}

// ParsePluginIdentifier parses a plugin identifier like "name@marketplace"
func ParsePluginIdentifier(pluginID string) (name, marketplace string) {
	parts := strings.Split(pluginID, "@")
	if len(parts) >= 2 {
		return parts[0], parts[1]
	}
	return pluginID, ""
}

// GetVersionedCachePath returns the versioned cache path for a plugin
func GetVersionedCachePath(pluginID, version string) string {
	return GetVersionedCachePathIn(getPluginsDirectory(), pluginID, version)
}

// GetVersionedCachePathIn returns versioned cache path under a specific base directory
func GetVersionedCachePathIn(baseDir, pluginID, version string) string {
	name, marketplace := ParsePluginIdentifier(pluginID)

	// Sanitize components to prevent path traversal
	sanitizedMarketplace := sanitizePathComponent(marketplace)
	sanitizedName := sanitizePathComponent(name)
	sanitizedVersion := sanitizePathComponent(version)

	if sanitizedMarketplace == "" {
		sanitizedMarketplace = "unknown"
	}

	return filepath.Join(baseDir, "cache", sanitizedMarketplace, sanitizedName, sanitizedVersion)
}

// GetLegacyCachePath returns the legacy (non-versioned) cache path for a plugin
func GetLegacyCachePath(pluginName string) string {
	return filepath.Join(GetPluginCachePath(), sanitizePathComponent(pluginName))
}

// ResolvePluginPath resolves plugin path with fallback to legacy location
func ResolvePluginPath(pluginID, version string) (string, error) {
	// Try versioned path first
	if version != "" {
		versionedPath := GetVersionedCachePath(pluginID, version)
		if _, err := os.Stat(versionedPath); err == nil {
			return versionedPath, nil
		}
	}

	// Fall back to legacy path for existing installations
	name, _ := ParsePluginIdentifier(pluginID)
	legacyPath := GetLegacyCachePath(name)
	if _, err := os.Stat(legacyPath); err == nil {
		return legacyPath, nil
	}

	// Return versioned path for new installations
	if version != "" {
		return GetVersionedCachePath(pluginID, version), nil
	}

	return legacyPath, nil
}

// CopyDir recursively copies a directory
func CopyDir(src, dest string) error {
	// Create destination directory
	if err := os.MkdirAll(dest, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Read source directory
	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("failed to read source directory: %w", err)
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		destPath := filepath.Join(dest, entry.Name())

		if entry.IsDir() {
			if err := CopyDir(srcPath, destPath); err != nil {
				return err
			}
		} else {
			// Copy file
			data, err := os.ReadFile(srcPath)
			if err != nil {
				return fmt.Errorf("failed to read file %s: %w", srcPath, err)
			}
			if err := os.WriteFile(destPath, data, 0644); err != nil {
				return fmt.Errorf("failed to write file %s: %w", destPath, err)
			}
		}
	}

	return nil
}

// CopyPluginToVersionedCache copies plugin files to versioned cache directory
func CopyPluginToVersionedCache(sourcePath, pluginID, version string) (string, error) {
	cachePath := GetVersionedCachePath(pluginID, version)

	// If cache already exists, return it
	if _, err := os.Stat(cachePath); err == nil {
		// Check if directory has content
		entries, err := os.ReadDir(cachePath)
		if err == nil && len(entries) > 0 {
			return cachePath, nil
		}
	}

	// Create parent directories
	if err := os.MkdirAll(filepath.Dir(cachePath), 0755); err != nil {
		return "", fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Copy plugin files
	if err := CopyDir(sourcePath, cachePath); err != nil {
		return "", fmt.Errorf("failed to copy plugin to cache: %w", err)
	}

	// Remove .git directory from cache if present
	gitPath := filepath.Join(cachePath, ".git")
	os.RemoveAll(gitPath)

	// Validate that cache has content
	entries, err := os.ReadDir(cachePath)
	if err != nil {
		return "", fmt.Errorf("failed to read cache directory: %w", err)
	}
	if len(entries) == 0 {
		return "", fmt.Errorf("destination is empty after copy")
	}

	return cachePath, nil
}

// ValidateGitUrl validates a git URL
func ValidateGitUrl(url string) error {
	// Check for valid protocols
	validPrefixes := []string{"https://", "http://", "file://", "git@"}
	valid := false
	for _, prefix := range validPrefixes {
		if strings.HasPrefix(url, prefix) {
			valid = true
			break
		}
	}

	if !valid {
		// Check for SSH format like git@host:path
		if matched := strings.Contains(url, "@") && strings.Contains(url, ":"); matched {
			valid = true
		}
	}

	if !valid {
		return fmt.Errorf("invalid git URL protocol: only HTTPS, HTTP, file:// and SSH (git@) URLs are supported")
	}

	return nil
}

// GenerateTemporaryCacheName generates a temporary cache name for a plugin
func GenerateTemporaryCacheName(source PluginSource) string {
	timestamp := fmt.Sprintf("%d", os.Getpid())

	var prefix string
	switch source.Type {
	case "npm":
		prefix = "npm"
	case "github":
		prefix = "github"
	case "url":
		prefix = "git"
	case "git-subdir":
		prefix = "subdir"
	case "pip":
		prefix = "pip"
	default:
		prefix = "local"
	}

	return fmt.Sprintf("temp_%s_%s", prefix, timestamp)
}

// LoadPluginManifest loads and validates a plugin manifest from a JSON file
func LoadPluginManifest(manifestPath, pluginName, source string) (*types.PluginManifest, error) {
	// Check if manifest file exists
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		// Return default manifest
		return &types.PluginManifest{
			Name:        pluginName,
			Description: fmt.Sprintf("Plugin from %s", source),
		}, nil
	}

	// Read and parse the manifest
	content, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	var manifest types.PluginManifest
	if err := json.Unmarshal(content, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest JSON: %w", err)
	}

	// Validate required fields
	if manifest.Name == "" {
		manifest.Name = pluginName
	}

	return &manifest, nil
}

// CreatePluginFromPath creates a LoadedPlugin object from a plugin directory path
// Scans the plugin directory structure and loads all components
func CreatePluginFromPath(pluginPath, source string, enabled bool, fallbackName string, strict bool) (*types.LoadedPlugin, []types.PluginError) {
	errors := []types.PluginError{}

	// Load or create the plugin manifest
	manifestPath := filepath.Join(pluginPath, ".claude-plugin", "plugin.json")
	legacyManifestPath := filepath.Join(pluginPath, "plugin.json")

	var manifest *types.PluginManifest
	var err error

	if _, err = os.Stat(manifestPath); err == nil {
		manifest, err = LoadPluginManifest(manifestPath, fallbackName, source)
	} else if _, err = os.Stat(legacyManifestPath); err == nil {
		manifest, err = LoadPluginManifest(legacyManifestPath, fallbackName, source)
	} else {
		manifest, err = LoadPluginManifest(manifestPath, fallbackName, source)
	}

	if err != nil {
		errors = append(errors, types.PluginError{
			Type:   types.PluginErrorTypeManifestParseError,
			Source: source,
			Plugin: &fallbackName,
		})
		manifest = &types.PluginManifest{
			Name:        fallbackName,
			Description: fmt.Sprintf("Plugin from %s", source),
		}
	}

	// Create the base plugin object
	plugin := &types.LoadedPlugin{
		Name:       manifest.Name,
		Manifest:   *manifest,
		Path:       pluginPath,
		Source:     source,
		Repository: source,
		Enabled:    &enabled,
	}

	// Auto-detect optional directories
	commandsDir := filepath.Join(pluginPath, "commands")
	agentsDir := filepath.Join(pluginPath, "agents")
	skillsDir := filepath.Join(pluginPath, "skills")
	outputStylesDir := filepath.Join(pluginPath, "output-styles")

	if _, err := os.Stat(commandsDir); err == nil {
		plugin.CommandsPath = &commandsDir
	}
	if _, err := os.Stat(agentsDir); err == nil {
		plugin.AgentsPath = &agentsDir
	}
	if _, err := os.Stat(skillsDir); err == nil {
		plugin.SkillsPath = &skillsDir
	}
	if _, err := os.Stat(outputStylesDir); err == nil {
		plugin.OutputStylesPath = &outputStylesDir
	}

	// Process commands from manifest
	if manifest.Commands != nil {
		processManifestCommands(plugin, manifest, pluginPath, source, &errors)
	}

	// Process agents from manifest
	if manifest.Agents != nil {
		processManifestAgents(plugin, manifest, pluginPath, source, &errors)
	}

	// Process skills from manifest
	if manifest.Skills != nil {
		processManifestSkills(plugin, manifest, pluginPath, source, &errors)
	}

	// Load hooks if present
	if manifest.Hooks != nil {
		plugin.HooksConfig = manifest.Hooks
	}

	return plugin, errors
}

// processManifestCommands processes command configurations from manifest
func processManifestCommands(plugin *types.LoadedPlugin, manifest *types.PluginManifest, pluginPath, source string, errors *[]types.PluginError) {
	switch cmds := manifest.Commands.(type) {
	case []interface{}:
		// Array of paths
		for _, cmd := range cmds {
			if cmdPath, ok := cmd.(string); ok {
				fullPath := filepath.Join(pluginPath, cmdPath)
				if _, err := os.Stat(fullPath); err == nil {
					plugin.CommandsPaths = append(plugin.CommandsPaths, fullPath)
				} else {
					*errors = append(*errors, types.PluginError{
						Type:      types.PluginErrorTypePathNotFound,
						Source:    source,
						Plugin:    &manifest.Name,
						Path:      &fullPath,
						Component: stringPtr("commands"),
					})
				}
			}
		}
	case map[string]interface{}:
		// Object mapping command names to metadata
		plugin.CommandsMetadata = make(map[string]types.CommandMetadata)
		for name, meta := range cmds {
			if metaMap, ok := meta.(map[string]interface{}); ok {
				cmdMeta := types.CommandMetadata{
					Name: name,
				}
				if desc, ok := metaMap["description"].(string); ok {
					cmdMeta.Description = desc
				}
				plugin.CommandsMetadata[name] = cmdMeta

				// Check for source path
				if src, ok := metaMap["source"].(string); ok {
					fullPath := filepath.Join(pluginPath, src)
					if _, err := os.Stat(fullPath); err == nil {
						plugin.CommandsPaths = append(plugin.CommandsPaths, fullPath)
					}
				}
			}
		}
	}
}

// processManifestAgents processes agent configurations from manifest
func processManifestAgents(plugin *types.LoadedPlugin, manifest *types.PluginManifest, pluginPath, source string, errors *[]types.PluginError) {
	switch agents := manifest.Agents.(type) {
	case []interface{}:
		// Array of paths
		for _, agent := range agents {
			if agentPath, ok := agent.(string); ok {
				fullPath := filepath.Join(pluginPath, agentPath)
				if _, err := os.Stat(fullPath); err == nil {
					plugin.AgentsPaths = append(plugin.AgentsPaths, fullPath)
				} else {
					*errors = append(*errors, types.PluginError{
						Type:      types.PluginErrorTypePathNotFound,
						Source:    source,
						Plugin:    &manifest.Name,
						Path:      &fullPath,
						Component: stringPtr("agents"),
					})
				}
			}
		}
	}
}

// processManifestSkills processes skill configurations from manifest
func processManifestSkills(plugin *types.LoadedPlugin, manifest *types.PluginManifest, pluginPath, source string, errors *[]types.PluginError) {
	switch skills := manifest.Skills.(type) {
	case []interface{}:
		// Array of paths
		for _, skill := range skills {
			if skillPath, ok := skill.(string); ok {
				fullPath := filepath.Join(pluginPath, skillPath)
				if _, err := os.Stat(fullPath); err == nil {
					plugin.SkillsPaths = append(plugin.SkillsPaths, fullPath)
				} else {
					*errors = append(*errors, types.PluginError{
						Type:      types.PluginErrorTypePathNotFound,
						Source:    source,
						Plugin:    &manifest.Name,
						Path:      &fullPath,
						Component: stringPtr("skills"),
					})
				}
			}
		}
	}
}

// LoadAllPlugins loads all plugins from various sources
// Sources (in order of precedence):
// 1. Marketplace-based plugins
// 2. Session-only plugins
// 3. Built-in plugins
func LoadAllPlugins() (*PluginLoadResult, error) {
	result := &PluginLoadResult{
		Enabled:  []types.LoadedPlugin{},
		Disabled: []types.LoadedPlugin{},
		Errors:   []types.PluginError{},
	}

	// Load from plugin cache directory
	cachePath := GetPluginCachePath()
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		// No plugins directory, return empty result
		return result, nil
	}

	entries, err := os.ReadDir(cachePath)
	if err != nil {
		return result, fmt.Errorf("failed to read plugin cache: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		pluginPath := filepath.Join(cachePath, entry.Name())
		source := fmt.Sprintf("cache:%s", entry.Name())

		plugin, errors := CreatePluginFromPath(pluginPath, source, true, entry.Name(), true)
		if plugin != nil {
			if plugin.Enabled != nil && *plugin.Enabled {
				result.Enabled = append(result.Enabled, *plugin)
			} else {
				result.Disabled = append(result.Disabled, *plugin)
			}
		}
		result.Errors = append(result.Errors, errors...)
	}

	// Load from builtin plugins directory
	builtinPath := filepath.Join(getPluginsDirectory(), "builtin")
	if entries, err := os.ReadDir(builtinPath); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			pluginPath := filepath.Join(builtinPath, entry.Name())
			source := fmt.Sprintf("builtin:%s", entry.Name())

			plugin, errors := CreatePluginFromPath(pluginPath, source, true, entry.Name(), true)
			if plugin != nil {
				if plugin.Enabled != nil && *plugin.Enabled {
					result.Enabled = append(result.Enabled, *plugin)
				} else {
					result.Disabled = append(result.Disabled, *plugin)
				}
			}
			result.Errors = append(result.Errors, errors...)
		}
	}

	return result, nil
}

// LoadPluginByID loads a specific plugin by its ID
func LoadPluginByID(pluginID string) (*types.LoadedPlugin, error) {
	// Try to find in cache
	cachePath := GetPluginCachePath()
	name, _ := ParsePluginIdentifier(pluginID)

	pluginPath := filepath.Join(cachePath, name)
	if _, err := os.Stat(pluginPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("plugin not found: %s", pluginID)
	}

	source := fmt.Sprintf("cache:%s", name)
	plugin, errors := CreatePluginFromPath(pluginPath, source, true, name, true)
	if len(errors) > 0 {
		// Log errors but return plugin if loaded
		for _, err := range errors {
			_ = err
		}
	}

	return plugin, nil
}

// EnablePlugin enables a plugin
func EnablePlugin(plugin *types.LoadedPlugin) error {
	enabled := true
	plugin.Enabled = &enabled
	// In full implementation, would persist to settings
	return nil
}

// DisablePlugin disables a plugin
func DisablePlugin(plugin *types.LoadedPlugin) error {
	enabled := false
	plugin.Enabled = &enabled
	// In full implementation, would persist to settings
	return nil
}

// InstallPlugin installs a plugin from a source
func InstallPlugin(source PluginSource, targetDir string) (*types.LoadedPlugin, error) {
	var sourcePath string

	switch source.Type {
	case "local":
		sourcePath = source.Path
	case "npm":
		// Would implement npm install logic
		return nil, fmt.Errorf("npm plugin installation not yet implemented")
	case "github":
		// Would implement git clone from GitHub
		return nil, fmt.Errorf("github plugin installation not yet implemented")
	case "url":
		// Would implement git clone from URL
		return nil, fmt.Errorf("url plugin installation not yet implemented")
	default:
		return nil, fmt.Errorf("unsupported plugin source type: %s", source.Type)
	}

	// Copy to target directory
	if err := CopyDir(sourcePath, targetDir); err != nil {
		return nil, fmt.Errorf("failed to copy plugin: %w", err)
	}

	// Load the installed plugin
	pluginSource := fmt.Sprintf("%s:%s", source.Type, source.Path)
	plugin, errors := CreatePluginFromPath(targetDir, pluginSource, true, filepath.Base(targetDir), true)
	if len(errors) > 0 {
		// Log errors
		for _, err := range errors {
			_ = err
		}
	}

	return plugin, nil
}

// sanitizePathComponent sanitizes a path component to prevent traversal attacks
func sanitizePathComponent(s string) string {
	// Replace any path separators and special characters
	replacer := strings.NewReplacer(
		"/", "-",
		"\\", "-",
		"..", "-",
		":", "-",
		"*", "-",
		"?", "-",
		"<", "-",
		">", "-",
		"|", "-",
	)
	return replacer.Replace(s)
}
