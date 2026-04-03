// Package config provides configuration management
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config application configuration
type Config struct {
	APIKey      string                   `json:"api_key"`
	Model       string                   `json:"model"`
	Theme       string                   `json:"theme"`
	Verbose     bool                     `json:"verbose"`
	Provider    string                   `json:"provider"`
	AutoSave    bool                     `json:"auto_save"`
	AutoSaveDir string                   `json:"auto_save_dir,omitempty"`
	Projects    map[string]ProjectConfig `json:"projects"`
	Env         map[string]string        `json:"env"`
}

// ProjectConfig project-level configuration
type ProjectConfig struct {
	AllowedTools []string                   `json:"allowed_tools"`
	MCPServers   map[string]MCPServerConfig `json:"mcp_servers"`
}

// MCPServerConfig MCP server configuration
type MCPServerConfig struct {
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env"`
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		Model:    "claude-sonnet-4-20250514",
		Theme:    "dark",
		Provider: "anthropic",
		AutoSave: true, // Auto-save enabled by default
		Projects: make(map[string]ProjectConfig),
		Env:      make(map[string]string),
	}
}

// Load loads configuration from file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &cfg, nil
}

// Save saves configuration to file
func (c *Config) Save(path string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize config: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetConfigPath returns the configuration file path
func GetConfigPath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = "."
	}
	return filepath.Join(configDir, "claude", "config.json")
}

// GetProjectConfig gets configuration for the current project
func (c *Config) GetProjectConfig(projectPath string) *ProjectConfig {
	if cfg, ok := c.Projects[projectPath]; ok {
		return &cfg
	}
	return &ProjectConfig{
		AllowedTools: []string{},
		MCPServers:   make(map[string]MCPServerConfig),
	}
}

// GetAutoSaveDir returns the auto-save directory
func (c *Config) GetAutoSaveDir() string {
	if c.AutoSaveDir != "" {
		return c.AutoSaveDir
	}
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = "."
	}
	return filepath.Join(configDir, "claude", "sessions")
}
