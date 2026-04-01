// Package plugins 提供插件管理
// 来源: src/utils/plugins/mcpPluginIntegration.ts
// 重构: Go MCP 插件集成（简化版）
package plugins

import (
	"github.com/Aspirin0000/claude-code-go/internal/mcp"
)

// GetPluginMcpServers 获取插件提供的 MCP 服务器
// 对应 TS: export function getPluginMcpServers(): Record<string, ScopedMcpServerConfig>
func GetPluginMcpServers() map[string]mcp.ScopedMcpServerConfig {
	// 简化实现：返回空映射
	// 实际实现需要加载插件并提取 MCP 配置
	return make(map[string]mcp.ScopedMcpServerConfig)
}

// LoadAllPluginsCacheOnly 仅从缓存加载所有插件
// 对应 TS: export function loadAllPluginsCacheOnly(): Promise<void>
func LoadAllPluginsCacheOnly() error {
	// 简化实现：无操作
	// 实际实现需要从缓存加载插件
	return nil
}
