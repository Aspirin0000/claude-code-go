// Package tools provides the Agent tool
// Source: src/tools/AgentTool/
// Refactor: Go Agent tool (real implementation)
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/Aspirin0000/claude-code-go/internal/api"
)

type apiClientCtxKey struct{}

// APIClientContextKey is used to pass the API client through context to tools
var APIClientContextKey = apiClientCtxKey{}

// AgentTool creates and manages sub-agents to execute tasks
type AgentTool struct{}

func (a *AgentTool) Name() string        { return "agent" }
func (a *AgentTool) Description() string { return "Create and manage sub-agents to execute tasks" }
func (a *AgentTool) IsReadOnly() bool    { return false }
func (a *AgentTool) IsDestructive() bool { return true }

func (a *AgentTool) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"agent_type": {"type": "string", "description": "Type of agent (e.g., coder, reviewer, researcher)"},
			"task": {"type": "string", "description": "Task description for the sub-agent"},
			"files": {"type": "array", "items": {"type": "string"}, "description": "Optional list of file paths to provide as context"}
		},
		"required": ["agent_type", "task"]
	}`)
}

func (a *AgentTool) Call(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	var params struct {
		AgentType string   `json:"agent_type"`
		Task      string   `json:"task"`
		Files     []string `json:"files"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("failed to parse agent parameters: %w", err)
	}

	if params.Task == "" {
		return nil, fmt.Errorf("task is required")
	}

	// Read any provided files for context
	var fileContexts []map[string]string
	for _, path := range params.Files {
		content, err := os.ReadFile(path)
		if err != nil {
			fileContexts = append(fileContexts, map[string]string{
				"path":    path,
				"content": fmt.Sprintf("Error reading file: %v", err),
			})
			continue
		}
		fileContexts = append(fileContexts, map[string]string{
			"path":    path,
			"content": string(content),
		})
	}

	// Build the prompt for the sub-agent
	var prompt strings.Builder
	prompt.WriteString(fmt.Sprintf("You are a specialized %s agent. "+
		"You have been given a specific task. Complete it to the best of your ability. "+
		"Provide a clear, concise result. Do not mention that you are a sub-agent.\n\n", params.AgentType))
	prompt.WriteString(fmt.Sprintf("Task: %s\n", params.Task))

	if len(fileContexts) > 0 {
		prompt.WriteString("\nProvided files:\n")
		for _, fc := range fileContexts {
			prompt.WriteString(fmt.Sprintf("\n--- %s ---\n%s\n", fc["path"], fc["content"]))
		}
	}

	// Try to get API client from context
	client, ok := ctx.Value(APIClientContextKey).(*api.Client)
	if !ok || client == nil {
		return json.Marshal(map[string]interface{}{
			"status":     "error",
			"message":    "API client not available for agent execution",
			"agent_type": params.AgentType,
			"task":       params.Task,
			"files_read": len(fileContexts),
		})
	}

	messages := []api.Message{
		{Role: "user", Content: prompt.String()},
	}

	resp, err := client.ChatWithBlocks(ctx, messages, nil)
	if err != nil {
		return nil, fmt.Errorf("agent execution failed: %w", err)
	}

	var textParts []string
	for _, block := range resp.Content {
		if block.Type == "text" {
			textParts = append(textParts, block.Text)
		}
	}

	return json.Marshal(map[string]interface{}{
		"status":     "success",
		"agent_type": params.AgentType,
		"result":     strings.Join(textParts, "\n"),
		"files_read": len(fileContexts),
	})
}
