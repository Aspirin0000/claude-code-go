// Package mcp 提供 MCP 服务
// 来源: src/services/mcp/utils.ts
// 重构: Go MCP 工具函数
package mcp

// GetProjectMcpServerStatus 获取项目 MCP 服务器状态
// 对应 TS: export function getProjectMcpServerStatus(...)
func GetProjectMcpServerStatus(serverName string) string {
	// 返回默认启用状态
	// 实际实现需要检查项目配置
	return "enabled"
}

// IsMcpServerDisabled 检查 MCP 服务器是否禁用
func IsMcpServerDisabled(serverName string) bool {
	// 默认假设所有服务器都启用
	return false
}
