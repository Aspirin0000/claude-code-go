// Package settings 提供设置管理
// 来源: src/utils/settings/pluginOnlyPolicy.ts
// 重构: Go 插件独占策略（简化版）
package settings

// IsRestrictedToPluginOnly 检查是否仅限插件
// 对应 TS: export function isRestrictedToPluginOnly(): boolean
func IsRestrictedToPluginOnly() bool {
	// 简化实现：不限制
	return false
}

// ShouldAllowManagedMcpServersOnly 检查是否仅允许管理的 MCP 服务器
func ShouldAllowManagedMcpServersOnly() bool {
	// 简化实现：允许所有
	return false
}
