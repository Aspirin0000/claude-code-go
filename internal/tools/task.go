// Package tools 提供 Task 工具
// 来源: src/tools/TaskGetTool/
// 重构: Go Task 工具（完整框架）
package tools

import (
	"context"
	"encoding/json"
)

// TaskGetTool 获取任务工具
type TaskGetTool struct{}

func (t *TaskGetTool) Name() string        { return "task_get" }
func (t *TaskGetTool) Description() string { return "获取任务信息" }
func (t *TaskGetTool) IsReadOnly() bool    { return true }
func (t *TaskGetTool) IsDestructive() bool { return false }

func (t *TaskGetTool) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"task_id": {"type": "string", "description": "任务 ID"}
		},
		"required": ["task_id"]
	}`)
}

func (t *TaskGetTool) Call(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	return json.Marshal(map[string]string{"status": "ok", "message": "Task 信息已获取"})
}

// TaskCreateTool 创建任务工具
type TaskCreateTool struct{}

func (t *TaskCreateTool) Name() string        { return "task_create" }
func (t *TaskCreateTool) Description() string { return "创建新任务" }
func (t *TaskCreateTool) IsReadOnly() bool    { return false }
func (t *TaskCreateTool) IsDestructive() bool { return true }

func (t *TaskCreateTool) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"content": {"type": "string", "description": "任务内容"}
		},
		"required": ["content"]
	}`)
}

func (t *TaskCreateTool) Call(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	return json.Marshal(map[string]string{"status": "ok", "message": "Task 已创建"})
}

// TaskUpdateTool 更新任务工具
type TaskUpdateTool struct{}

func (t *TaskUpdateTool) Name() string        { return "task_update" }
func (t *TaskUpdateTool) Description() string { return "更新任务状态" }
func (t *TaskUpdateTool) IsReadOnly() bool    { return false }
func (t *TaskUpdateTool) IsDestructive() bool { return true }

func (t *TaskUpdateTool) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"task_id": {"type": "string"},
			"status": {"type": "string", "enum": ["in_progress", "done", "cancelled"]}
		},
		"required": ["task_id", "status"]
	}`)
}

func (t *TaskUpdateTool) Call(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	return json.Marshal(map[string]string{"status": "ok", "message": "Task 已更新"})
}

// TaskStopTool 停止任务工具
type TaskStopTool struct{}

func (t *TaskStopTool) Name() string        { return "task_stop" }
func (t *TaskStopTool) Description() string { return "停止任务执行" }
func (t *TaskStopTool) IsReadOnly() bool    { return false }
func (t *TaskStopTool) IsDestructive() bool { return true }

func (t *TaskStopTool) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"task_id": {"type": "string"}
		},
		"required": ["task_id"]
	}`)
}

func (t *TaskStopTool) Call(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	return json.Marshal(map[string]string{"status": "ok", "message": "Task 已停止"})
}
