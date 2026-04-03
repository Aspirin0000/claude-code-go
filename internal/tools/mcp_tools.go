// Package tools provides MCP (Model Context Protocol) tools
package tools

import (
	"context"
	"encoding/json"
	"fmt"
)

// ListMcpResourcesTool lists available resources from MCP servers
type ListMcpResourcesTool struct{}

func (t *ListMcpResourcesTool) Name() string { return "list_mcp_resources" }
func (t *ListMcpResourcesTool) Description() string {
	return "List available resources from connected MCP servers"
}
func (t *ListMcpResourcesTool) IsReadOnly() bool    { return true }
func (t *ListMcpResourcesTool) IsDestructive() bool { return false }

func (t *ListMcpResourcesTool) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"server_name": {
				"type": "string",
				"description": "Optional server name to filter resources"
			}
		}
	}`)
}

func (t *ListMcpResourcesTool) Call(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	var params struct {
		ServerName string `json:"server_name,omitempty"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("failed to parse parameters: %w", err)
	}

	// Get MCP manager instance (this would be set up during initialization)
	// For now, return a placeholder response
	result := struct {
		Resources []McpResourceInfo `json:"resources"`
		Count     int               `json:"count"`
		Note      string            `json:"note"`
	}{
		Resources: []McpResourceInfo{},
		Count:     0,
		Note:      "MCP resource listing requires active MCP server connections. Use /mcp-list to see connected servers.",
	}

	return json.Marshal(result)
}

// McpResourceInfo represents an MCP resource
type McpResourceInfo struct {
	Name        string `json:"name"`
	Server      string `json:"server"`
	URI         string `json:"uri"`
	Description string `json:"description,omitempty"`
}

// ReadMcpResourceTool reads content from an MCP resource
type ReadMcpResourceTool struct{}

func (t *ReadMcpResourceTool) Name() string        { return "read_mcp_resource" }
func (t *ReadMcpResourceTool) Description() string { return "Read content from an MCP resource by URI" }
func (t *ReadMcpResourceTool) IsReadOnly() bool    { return true }
func (t *ReadMcpResourceTool) IsDestructive() bool { return false }

func (t *ReadMcpResourceTool) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"uri": {
				"type": "string",
				"description": "Resource URI to read"
			}
		},
		"required": ["uri"]
	}`)
}

func (t *ReadMcpResourceTool) Call(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	var params struct {
		URI string `json:"uri"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("failed to parse parameters: %w", err)
	}

	if params.URI == "" {
		return nil, fmt.Errorf("uri parameter is required")
	}

	// Placeholder implementation
	result := struct {
		URI     string `json:"uri"`
		Content string `json:"content"`
		Note    string `json:"note"`
	}{
		URI:     params.URI,
		Content: "",
		Note:    "MCP resource reading requires active MCP server connections. Configure MCP servers with /mcp-add.",
	}

	return json.Marshal(result)
}

// McpTool calls a tool on an MCP server
type McpTool struct{}

func (t *McpTool) Name() string        { return "mcp_tool" }
func (t *McpTool) Description() string { return "Execute a tool on a connected MCP server" }
func (t *McpTool) IsReadOnly() bool    { return false }
func (t *McpTool) IsDestructive() bool { return true }

func (t *McpTool) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"server_name": {
				"type": "string",
				"description": "Name of the MCP server"
			},
			"tool_name": {
				"type": "string",
				"description": "Name of the tool to execute"
			},
			"arguments": {
				"type": "object",
				"description": "Tool arguments"
			}
		},
		"required": ["server_name", "tool_name"]
	}`)
}

func (t *McpTool) Call(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	var params struct {
		ServerName string          `json:"server_name"`
		ToolName   string          `json:"tool_name"`
		Arguments  json.RawMessage `json:"arguments,omitempty"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("failed to parse parameters: %w", err)
	}

	if params.ServerName == "" || params.ToolName == "" {
		return nil, fmt.Errorf("server_name and tool_name are required")
	}

	// Placeholder implementation
	result := struct {
		ServerName string `json:"server_name"`
		ToolName   string `json:"tool_name"`
		Success    bool   `json:"success"`
		Note       string `json:"note"`
	}{
		ServerName: params.ServerName,
		ToolName:   params.ToolName,
		Success:    false,
		Note:       "MCP tool execution requires configured MCP servers. Use /mcp-add to add servers, then /mcp-list to verify connection.",
	}

	return json.Marshal(result)
}

func init() {
	// These will be registered when the tools package is initialized
	// The actual registration happens in tools.go's NewDefaultRegistry
}
