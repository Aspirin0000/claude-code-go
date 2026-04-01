// Package tools 提供工具系统实现
package tools

import (
	"context"
	"encoding/json"
	"fmt"
)

// Tool 工具接口
type Tool interface {
	Name() string
	Description() string
	InputSchema() json.RawMessage
	Call(ctx context.Context, input json.RawMessage) (json.RawMessage, error)
	IsReadOnly() bool
	IsDestructive() bool
}

// Result 工具执行结果
type Result struct {
	Data    json.RawMessage `json:"data"`
	Success bool            `json:"success"`
	Error   string          `json:"error,omitempty"`
}

// Registry 工具注册表
type Registry struct {
	tools map[string]Tool
}

// NewRegistry 创建新的工具注册表
func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]Tool),
	}
}

// Register 注册工具
func (r *Registry) Register(tool Tool) {
	r.tools[tool.Name()] = tool
}

// Get 获取工具
func (r *Registry) Get(name string) (Tool, bool) {
	tool, ok := r.tools[name]
	return tool, ok
}

// List 列出所有工具
func (r *Registry) List() []Tool {
	list := make([]Tool, 0, len(r.tools))
	for _, tool := range r.tools {
		list = append(list, tool)
	}
	return list
}

// Call 调用工具
func (r *Registry) Call(ctx context.Context, name string, input json.RawMessage) (*Result, error) {
	tool, ok := r.Get(name)
	if !ok {
		return nil, fmt.Errorf("工具未找到: %s", name)
	}

	output, err := tool.Call(ctx, input)
	if err != nil {
		return &Result{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &Result{
		Data:    output,
		Success: true,
	}, nil
}

// GetToolSchemas 获取所有工具的 JSON Schema
func (r *Registry) GetToolSchemas() []map[string]interface{} {
	schemas := make([]map[string]interface{}, 0, len(r.tools))
	for _, tool := range r.tools {
		var inputSchema map[string]interface{}
		json.Unmarshal(tool.InputSchema(), &inputSchema)

		schemas = append(schemas, map[string]interface{}{
			"name":         tool.Name(),
			"description":  tool.Description(),
			"input_schema": inputSchema,
		})
	}
	return schemas
}
