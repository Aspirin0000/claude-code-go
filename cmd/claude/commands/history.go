package commands

import (
	"context"
	"fmt"

	"github.com/Aspirin0000/claude-code-go/internal/state"
)

// HistoryCommand shows conversation history summary
type HistoryCommand struct {
	*BaseCommand
}

// NewHistoryCommand creates the history command
func NewHistoryCommand() *HistoryCommand {
	return &HistoryCommand{
		BaseCommand: NewBaseCommand(
			"history",
			"Show conversation history summary",
			CategorySession,
		).WithHelp(`Usage: /history

Show a summary of the current conversation history.

Displays:
  - Number of messages
  - Breakdown by role (user, assistant, system)
  - List of tools used in the session

Examples:
  /history       Show conversation summary`),
	}
}

// Execute runs the history command
func (c *HistoryCommand) Execute(ctx context.Context, args []string) error {
	messages := state.GlobalState.GetMessages()

	var userCount, assistantCount, systemCount int
	toolSet := make(map[string]int)

	for _, msg := range messages {
		switch msg.Role {
		case "user":
			userCount++
		case "assistant":
			assistantCount++
			// Count tool uses in assistant messages
			for _, block := range msg.Blocks {
				if block.Type == "tool_use" {
					toolSet[block.Name]++
				}
			}
		case "system":
			systemCount++
		}
	}

	fmt.Println()
	fmt.Println("╔════════════════════════════════════════╗")
	fmt.Println("║        Conversation History            ║")
	fmt.Println("╚════════════════════════════════════════╝")
	fmt.Println()

	fmt.Printf("Total messages: %d\n", len(messages))
	fmt.Printf("  👤 User:      %d\n", userCount)
	fmt.Printf("  🤖 Assistant: %d\n", assistantCount)
	fmt.Printf("  ℹ️  System:    %d\n", systemCount)
	fmt.Println()

	if len(toolSet) > 0 {
		fmt.Println("Tools used in this session:")
		for toolName, count := range toolSet {
			fmt.Printf("  • %-20s %d time(s)\n", toolName, count)
		}
		fmt.Println()
	}

	return nil
}

func init() {
	Register(NewHistoryCommand())
}
