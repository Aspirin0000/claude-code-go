package commands

import (
	"context"
	"fmt"

	"github.com/Aspirin0000/claude-code-go/internal/state"
)

// CostCommand shows cost tracking information
// Source: src/commands/cost/
type CostCommand struct {
	*BaseCommand
}

// NewCostCommand creates the cost command
func NewCostCommand() *CostCommand {
	return &CostCommand{
		BaseCommand: NewBaseCommand(
			"cost",
			"Show cost tracking information",
			CategoryAdvanced,
		).
			WithHelp(`Show cost tracking information for the current session.

Includes:
- Input token usage
- Output token usage
- Estimated cost
- Model pricing info

Usage: /cost`),
	}
}

// Execute shows cost information
func (c *CostCommand) Execute(ctx context.Context, args []string) error {
	messages := state.GlobalState.GetMessages()

	inputTokens, outputTokens := calculateCostTokens(messages)
	totalTokens := inputTokens + outputTokens

	// Rough cost estimation (Claude 3 Sonnet pricing)
	inputCost := float64(inputTokens) * 0.000003   // $3 per million tokens
	outputCost := float64(outputTokens) * 0.000015 // $15 per million tokens
	totalCost := inputCost + outputCost

	fmt.Println()
	fmt.Println("💰 Cost Tracking")
	fmt.Println("═══════════════════════════════════")
	fmt.Printf("\n📊 Token Usage:\n")
	fmt.Printf("   Input tokens:  %d\n", inputTokens)
	fmt.Printf("   Output tokens: %d\n", outputTokens)
	fmt.Printf("   Total:         %d\n", totalTokens)

	fmt.Printf("\n💵 Estimated Cost (USD):\n")
	fmt.Printf("   Input cost:  $%.4f\n", inputCost)
	fmt.Printf("   Output cost: $%.4f\n", outputCost)
	fmt.Printf("   Total cost:  $%.4f\n", totalCost)

	fmt.Printf("\n📈 Avg per message: %.0f tokens\n",
		float64(totalTokens)/float64(len(messages)+1))
	fmt.Println()

	return nil
}

// calculateCostTokens estimates tokens for cost calculation
func calculateCostTokens(messages []state.Message) (input int, output int) {
	for _, msg := range messages {
		content := msg.Content
		tokenCount := len(content) / 4 // rough estimate

		if msg.Role == "user" {
			input += tokenCount
		} else if msg.Role == "assistant" {
			output += tokenCount
		}
	}
	return input, output
}

func init() {
	Register(NewCostCommand())
}
