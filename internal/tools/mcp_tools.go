// Package tools provides MCP (Model Context Protocol) tools
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Aspirin0000/claude-code-go/internal/mcp"
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

	manager := mcp.GetGlobalMCPManager()
	connected := manager.GetConnectionManager().ListConnected()

	if len(connected) == 0 {
		return json.Marshal(map[string]interface{}{
			"resources": []mcp.ResourceInfo{},
			"count":     0,
			"note":      "No MCP servers are currently connected. Use /mcp-list to check status and /mcp-add to configure servers.",
		})
	}

	var allResources []mcp.ResourceInfo
	for _, serverName := range connected {
		if params.ServerName != "" && serverName != params.ServerName {
			continue
		}

		client, exists := manager.GetConnectionManager().GetClient(serverName)
		if !exists {
			continue
		}

		resources, err := mcp.FetchResourcesForClient(client)
		if err != nil {
			continue
		}
		allResources = append(allResources, resources...)
	}

	return json.Marshal(map[string]interface{}{
		"resources": allResources,
		"count":     len(allResources),
		"servers":   len(connected),
	})
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

	manager := mcp.GetGlobalMCPManager()
	connected := manager.GetConnectionManager().ListConnected()

	if len(connected) == 0 {
		return json.Marshal(map[string]interface{}{
			"uri":     params.URI,
			"content": "",
			"note":    "No MCP servers are currently connected.",
		})
	}

	// Try each connected server to read the resource
	var lastErr error
	for _, serverName := range connected {
		client, exists := manager.GetConnectionManager().GetClient(serverName)
		if !exists {
			continue
		}

		result, err := client.ReadResource(params.URI)
		if err != nil {
			lastErr = err
			continue
		}

		var contentParts []string
		for _, c := range result.Contents {
			if c.Text != "" {
				contentParts = append(contentParts, c.Text)
			}
		}

		return json.Marshal(map[string]interface{}{
			"uri":      params.URI,
			"server":   serverName,
			"content":  strings.Join(contentParts, "\n"),
			"mimeType": result.Contents[0].MIMEType,
		})
	}

	errMsg := "Resource not found on any connected server"
	if lastErr != nil {
		errMsg = fmt.Sprintf("%s: %v", errMsg, lastErr)
	}

	return nil, fmt.Errorf("%s", errMsg)
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
		ServerName string                 `json:"server_name"`
		ToolName   string                 `json:"tool_name"`
		Arguments  map[string]interface{} `json:"arguments,omitempty"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("failed to parse parameters: %w", err)
	}

	if params.ServerName == "" || params.ToolName == "" {
		return nil, fmt.Errorf("server_name and tool_name are required")
	}

	manager := mcp.GetGlobalMCPManager()

	// Construct full tool name: mcp__<server>__<tool>
	fullToolName := fmt.Sprintf("mcp__%s__%s", params.ServerName, params.ToolName)

	result, err := manager.ExecuteTool(ctx, fullToolName, params.Arguments)
	if err != nil {
		return nil, fmt.Errorf("tool execution failed: %w", err)
	}

	// Format result content
	var contentParts []string
	for _, block := range result.Content {
		if block.Text != "" {
			contentParts = append(contentParts, block.Text)
		}
	}

	return json.Marshal(map[string]interface{}{
		"server_name": params.ServerName,
		"tool_name":   params.ToolName,
		"success":     !result.IsError,
		"content":     strings.Join(contentParts, "\n"),
		"is_error":    result.IsError,
	})
}

func init() {
	// These will be registered when the tools package is initialized
	// The actual registration happens in tools.go's NewDefaultRegistry
}
