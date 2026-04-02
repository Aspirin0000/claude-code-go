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
			"显示成本追踪信息",
			CategoryAdvanced,
		).
			WithHelp(`显示当前会话的成本追踪信息。

包括:
- 输入 token 使用量
- 输出 token 使用量  
- 估算成本
- 模型定价信息

使用: /cost`),
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
	fmt.Println("💰 成本追踪 (Cost Tracking)")
	fmt.Println("═══════════════════════════════════")
	fmt.Printf("\n📊 Token 使用量:\n")
	fmt.Printf("   输入 tokens:  %d\n", inputTokens)
	fmt.Printf("   输出 tokens:  %d\n", outputTokens)
	fmt.Printf("   总计:         %d\n", totalTokens)

	fmt.Printf("\n💵 估算成本 (USD):\n")
	fmt.Printf("   输入成本:     $%.4f\n", inputCost)
	fmt.Printf("   输出成本:     $%.4f\n", outputCost)
	fmt.Printf("   总成本:       $%.4f\n", totalCost)

	fmt.Printf("\n📈 平均每条消息: %.0f tokens\n",
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
