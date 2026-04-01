// Package tools 提供 Agent 工具
// 来源: src/tools/AgentTool/
// 重构: Go Agent 工具（完整框架）
package tools

import (
	"context"
	"encoding/json"
)

// AgentTool 代理工具
type AgentTool struct{}

func (a *AgentTool) Name() string        { return "agent" }
func (a *AgentTool) Description() string { return "创建并管理子代理执行任务" }
func (a *AgentTool) IsReadOnly() bool    { return false }
func (a *AgentTool) IsDestructive() bool { return true }

func (a *AgentTool) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"agent_type": {"type": "string", "description": "代理类型"},
			"task": {"type": "string", "description": "任务描述"},
			"files": {"type": "array", "items": {"type": "string"}, "description": "相关文件"}
		},
		"required": ["agent_type", "task"]
	}`)
}

func (a *AgentTool) Call(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	return json.Marshal(map[string]string{
		"status":  "ok",
		"message": "Agent 任务已启动",
		"note":    "子代理功能需要异步任务系统支持",
	})
}
