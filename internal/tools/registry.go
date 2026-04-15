// Package tools provides the tool system implementation
package tools

import (
	"context"
	"encoding/json"
	"fmt"
)

// Tool tool interface
type Tool interface {
	Name() string
	Description() string
	InputSchema() json.RawMessage
	Call(ctx context.Context, input json.RawMessage) (json.RawMessage, error)
	IsReadOnly() bool
	IsDestructive() bool
}

// Result tool execution result
type Result struct {
	Data    json.RawMessage `json:"data"`
	Success bool            `json:"success"`
	Error   string          `json:"error,omitempty"`
}

// Registry tool registry
type Registry struct {
	tools map[string]Tool
}

// NewRegistry creates a new tool registry
func NewRegistry() *Registry {
	return &Registry{
		tools: make(map[string]Tool),
	}
}

// Register registers a tool
func (r *Registry) Register(tool Tool) {
	r.tools[tool.Name()] = tool
}

// Get gets a tool by name
func (r *Registry) Get(name string) (Tool, bool) {
	tool, ok := r.tools[name]
	return tool, ok
}

// List lists all tools
func (r *Registry) List() []Tool {
	list := make([]Tool, 0, len(r.tools))
	for _, tool := range r.tools {
		list = append(list, tool)
	}
	return list
}

// Call invokes a tool
func (r *Registry) Call(ctx context.Context, name string, input json.RawMessage) (*Result, error) {
	tool, ok := r.Get(name)
	if !ok {
		return nil, fmt.Errorf("tool not found: %s", name)
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

// GetToolSchemas returns JSON Schemas for all tools
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
