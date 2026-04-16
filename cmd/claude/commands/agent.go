package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/Aspirin0000/claude-code-go/internal/api"
	"github.com/Aspirin0000/claude-code-go/internal/config"
	"github.com/Aspirin0000/claude-code-go/internal/tools"
)

// AgentCommand invokes a specialized AI agent directly
type AgentCommand struct {
	*BaseCommand
}

// NewAgentCommand creates the /agent command
func NewAgentCommand() *AgentCommand {
	return &AgentCommand{
		BaseCommand: NewBaseCommand(
			"agent",
			"Spawn a specialized AI agent to complete a task",
			CategoryAdvanced,
		).
			WithHelp(`Usage: /agent <type> <task> [file1] [file2] ...

Spawn a specialized AI agent to work on a specific task.
Optional file paths can be provided as context.

Examples:
  /agent coder "Refactor this function to use generics" main.go
  /agent reviewer "Review the error handling" handler.go utils.go
  /agent researcher "Find best practices for Go context cancellation"
`),
	}
}

// Execute runs the agent command
func (c *AgentCommand) Execute(ctx context.Context, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: /agent <type> <task> [files...]")
	}

	agentType := args[0]
	task := args[1]
	files := args[2:]

	// Build tool input
	input := map[string]interface{}{
		"agent_type": agentType,
		"task":       task,
		"files":      files,
	}
	inputJSON, _ := json.Marshal(input)

	// Create API client
	cfgPath := config.GetConfigPath()
	cfg, err := config.Load(cfgPath)
	if err != nil {
		cfg = config.DefaultConfig()
	}
	if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
		cfg.APIKey = apiKey
	}
	if model := os.Getenv("CLAUDE_MODEL"); model != "" {
		cfg.Model = model
	}

	if cfg.APIKey == "" {
		return fmt.Errorf("API key not configured; run /login or set ANTHROPIC_API_KEY")
	}

	client := api.NewClient(cfg.APIKey, cfg.Model)
	client.SetProvider(cfg.Provider)

	ctxWithClient := context.WithValue(ctx, tools.APIClientContextKey, client)

	agent := &tools.AgentTool{}
	result, err := agent.Call(ctxWithClient, inputJSON)
	if err != nil {
		return fmt.Errorf("agent failed: %w", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(result, &parsed); err != nil {
		return fmt.Errorf("failed to parse agent result: %w", err)
	}

	status, _ := parsed["status"].(string)
	res, _ := parsed["result"].(string)
	if status == "error" {
		msg, _ := parsed["message"].(string)
		return fmt.Errorf("agent error: %s", msg)
	}

	fmt.Println()
	fmt.Printf("🤖 Agent (%s) result:\n", agentType)
	fmt.Println(strings.Repeat("─", 50))
	fmt.Println(res)
	fmt.Println()
	return nil
}

func init() {
	Register(NewAgentCommand())
}
