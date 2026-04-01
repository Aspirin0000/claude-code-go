// Package config 提供配置管理功能
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config 应用配置
type Config struct {
	APIKey   string                   `json:"api_key"`
	Model    string                   `json:"model"`
	Theme    string                   `json:"theme"`
	Verbose  bool                     `json:"verbose"`
	Provider string                   `json:"provider"`
	Projects map[string]ProjectConfig `json:"projects"`
	Env      map[string]string        `json:"env"`
}

// ProjectConfig 项目级配置
type ProjectConfig struct {
	AllowedTools []string                   `json:"allowed_tools"`
	MCPServers   map[string]MCPServerConfig `json:"mcp_servers"`
}

// MCPServerConfig MCP 服务器配置
type MCPServerConfig struct {
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env"`
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Model:    "claude-sonnet-4-20250514",
		Theme:    "dark",
		Provider: "anthropic",
		Projects: make(map[string]ProjectConfig),
		Env:      make(map[string]string),
	}
}

// Load 从文件加载配置
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	return &cfg, nil
}

// Save 保存配置到文件
func (c *Config) Save(path string) error {
	// 确保目录存在
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	return nil
}

// GetConfigPath 返回配置文件路径
func GetConfigPath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = "."
	}
	return filepath.Join(configDir, "claude", "config.json")
}

// GetProjectConfig 获取当前项目的配置
func (c *Config) GetProjectConfig(projectPath string) *ProjectConfig {
	if cfg, ok := c.Projects[projectPath]; ok {
		return &cfg
	}
	return &ProjectConfig{
		AllowedTools: []string{},
		MCPServers:   make(map[string]MCPServerConfig),
	}
}
