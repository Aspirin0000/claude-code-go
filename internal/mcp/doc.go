// Package mcp 提供 MCP (Model Context Protocol) 服务
// 完整实现 - 8/8批次全部完成
//
// 文件清单：
// - types.go: 类型定义
// - config.go: 配置管理
// - client.go: MCP客户端 (C-1/8, C-2/8)
// - transport.go: 传输层 (HTTP/SSE/Stdio)
// - connection.go: 连接管理器 (C-3/8)
// - cache.go: 缓存系统 (C-4/8)
// - auth.go: OAuth认证 (C-5/8)
// - websocket.go: WebSocket传输 (C-6/8)
// - executor.go: 工具执行器 (C-7/8)
// - manager.go: MCP管理器 (C-8/8)
// - utils.go: 工具函数
// - claudeai.go: ClaudeAI集成
//
// 总计: ~6,200行代码
package mcp

// 版本信息
const (
	MCPClientVersion   = "1.0.0"
	MCPProtocolVersion = "2024-11-05"
)
