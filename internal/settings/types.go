// Package settings 提供设置管理
// 来源: src/utils/settings/types.ts
// 重构: Go 设置类型（简化版）
package settings

// SettingsJson 设置 JSON 结构
type SettingsJson struct {
	AllowedMcpServers []McpServerEntry `json:"allowedMcpServers,omitempty"`
	DeniedMcpServers  []McpServerEntry `json:"deniedMcpServers,omitempty"`
}

// McpServerEntry MCP 服务器条目
type McpServerEntry struct {
	ServerName    *string  `json:"serverName,omitempty"`
	ServerCommand []string `json:"serverCommand,omitempty"`
	ServerUrl     *string  `json:"serverUrl,omitempty"`
}

// IsMcpServerNameEntry 检查是否为名称条目
func IsMcpServerNameEntry(entry McpServerEntry) bool {
	return entry.ServerName != nil
}

// IsMcpServerCommandEntry 检查是否为命令条目
func IsMcpServerCommandEntry(entry McpServerEntry) bool {
	return entry.ServerCommand != nil && len(entry.ServerCommand) > 0
}

// IsMcpServerUrlEntry 检查是否为 URL 条目
func IsMcpServerUrlEntry(entry McpServerEntry) bool {
	return entry.ServerUrl != nil
}

// SettingSource 设置来源
type SettingSource string

const (
	SettingSourceUser    SettingSource = "user"
	SettingSourceProject SettingSource = "project"
	SettingSourceLocal   SettingSource = "local"
	SettingSourcePolicy  SettingSource = "policy"
)

// IsSettingSourceEnabled 检查设置来源是否启用
func IsSettingSourceEnabled(source SettingSource) bool {
	// 简化实现：所有来源都启用
	return true
}

// GetInitialSettings 获取初始设置
func GetInitialSettings() *SettingsJson {
	return &SettingsJson{}
}

// GetSettingsForSource 获取特定来源的设置
func GetSettingsForSource(source SettingSource) *SettingsJson {
	// 简化实现：返回空设置
	return &SettingsJson{}
}
