// Package mcp 提供 MCP 服务
// 来源: src/services/mcp/claudeai.ts
// 重构: Go ClaudeAI MCP 配置获取（简化版）
package mcp

// FetchClaudeAIMcpConfigsIfEligible 获取 ClaudeAI MCP 配置
// 对应 TS: export async function fetchClaudeAIMcpConfigsIfEligible()
func FetchClaudeAIMcpConfigsIfEligible() (map[string]McpServerConfig, error) {
	// 简化实现：返回空配置
	// 实际实现需要检查用户资格并调用 ClaudeAI API
	return make(map[string]McpServerConfig), nil
}
