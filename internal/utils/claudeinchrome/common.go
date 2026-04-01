// Package claudeinchrome 提供 Claude in Chrome 支持
// 来源: src/utils/claudeInChrome/common.ts
// 重构: Go Claude in Chrome 工具（简化版）
package claudeinchrome

// IsClaudeInChromeMCPServer 检查是否为 Claude in Chrome MCP 服务器
// 对应 TS: export function isClaudeInChromeMCPServer(name: string): boolean
func IsClaudeInChromeMCPServer(name string) bool {
	// 简化实现：检查名称前缀
	return name == "claude-in-chrome" || name == "claude-in-chrome-mcp"
}
