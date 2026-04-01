// Package settings 提供设置管理
// 来源: src/utils/settings/pluginOnlyPolicy.ts
// 重构: Go 插件独占策略
package settings

// IsRestrictedToPluginOnly 检查是否仅限插件
// 对应 TS: export function isRestrictedToPluginOnly(): boolean
func IsRestrictedToPluginOnly() bool {
	// 默认不限制
	return false
}

// ShouldAllowManagedMcpServersOnly 检查是否仅允许管理的 MCP 服务器
func ShouldAllowManagedMcpServersOnly() bool {
	// 默认允许所有 MCP 服务器
	return false
}
