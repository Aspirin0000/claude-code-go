package commands

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/Aspirin0000/claude-code-go/internal/config"
)

// ModelInfo contains information about an AI model
type AIModelInfo struct {
	ID          string
	Name        string
	Description string
	ContextWin  int     // Context window in tokens
	InputPrice  float64 // Price per 1M input tokens
	OutputPrice float64 // Price per 1M output tokens
	Strengths   []string
}

// AvailableModels defines all supported AI models
var AvailableModels = map[string]AIModelInfo{
	"claude-sonnet-4-20250514": {
		ID:          "claude-sonnet-4-20250514",
		Name:        "Claude 4 Sonnet",
		Description: "Balanced performance and speed for most tasks (latest)",
		ContextWin:  200000,
		InputPrice:  3.0,
		OutputPrice: 15.0,
		Strengths: []string{
			"Everyday conversations",
			"Code assistance",
			"Content summarization",
			"Multilingual support",
		},
	},
	"claude-opus-4-20250514": {
		ID:          "claude-opus-4-20250514",
		Name:        "Claude 4 Opus",
		Description: "Most capable Claude model for complex tasks",
		ContextWin:  200000,
		InputPrice:  15.0,
		OutputPrice: 75.0,
		Strengths: []string{
			"Complex reasoning and analysis",
			"Creative writing",
			"Code generation",
			"Detailed technical documentation",
		},
	},
	"claude-3-opus-20240229": {
		ID:          "claude-3-opus-20240229",
		Name:        "Claude 3 Opus",
		Description: "Most capable Claude 3 model for complex tasks",
		ContextWin:  200000,
		InputPrice:  15.0,
		OutputPrice: 75.0,
		Strengths: []string{
			"Complex reasoning and analysis",
			"Creative writing",
			"Code generation",
			"Detailed technical documentation",
		},
	},
	"claude-3-sonnet-20240229": {
		ID:          "claude-3-sonnet-20240229",
		Name:        "Claude 3 Sonnet",
		Description: "Balanced Claude 3 model for most tasks",
		ContextWin:  200000,
		InputPrice:  3.0,
		OutputPrice: 15.0,
		Strengths: []string{
			"Everyday conversations",
			"Code assistance",
			"Content summarization",
			"Multilingual support",
		},
	},
	"claude-3-haiku-20240307": {
		ID:          "claude-3-haiku-20240307",
		Name:        "Claude 3 Haiku",
		Description: "Fastest and most cost-effective model for simple tasks",
		ContextWin:  200000,
		InputPrice:  0.25,
		OutputPrice: 1.25,
		Strengths: []string{
			"Fast responses",
			"Simple Q&A",
			"Classification tasks",
			"Lightweight tasks",
		},
	},
}

// ModelCommand handles model switching and display
type ModelCommand struct {
	*BaseCommand
}

// NewModelCommand creates a new model command
func NewModelCommand() *ModelCommand {
	return &ModelCommand{
		BaseCommand: NewBaseCommand(
			"model",
			"Switch or view AI models",
			CategoryConfig,
		).WithAliases("m", "switch-model").
			WithHelp(`Usage: /model [model-name|list]

Switch the active AI model or view current model information.

Available models:
  • claude-3-opus-20240229   - Most capable model for complex tasks
  • claude-3-sonnet-20240229 - Balanced performance and speed for most tasks
  • claude-3-haiku-20240307  - Fastest and most economical for simple tasks

Usage:
  /model                    - Show the current model and available model list
  /model <model-name>       - Switch to the specified model
  /model list               - List all available models with details

Aliases: /m, /switch-model`),
	}
}

// Execute runs the model command
func (c *ModelCommand) Execute(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return c.showCurrentModel()
	}

	arg := strings.ToLower(strings.TrimSpace(args[0]))

	switch arg {
	case "list", "ls", "all":
		return c.listAllModels()
	default:
		return c.switchModel(arg)
	}
}

// showCurrentModel displays the current model and brief info
func (c *ModelCommand) showCurrentModel() error {
	currentModel := c.getCurrentModel()
	modelInfo, exists := AvailableModels[currentModel]

	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║                      Current Model                       ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Println()

	if exists {
		fmt.Printf("🤖 %s\n", modelInfo.Name)
		fmt.Printf("   ID:          %s\n", modelInfo.ID)
		fmt.Printf("   Description: %s\n", modelInfo.Description)
		fmt.Printf("   Context:     %s tokens\n", formatNumber(modelInfo.ContextWin))
		fmt.Printf("\n   💰 Pricing (per 1M tokens):\n")
		fmt.Printf("      Input:  $%.2f\n", modelInfo.InputPrice)
		fmt.Printf("      Output: $%.2f\n", modelInfo.OutputPrice)
		fmt.Printf("\n   ✨ Strengths:\n")
		for _, strength := range modelInfo.Strengths {
			fmt.Printf("      • %s\n", strength)
		}
	} else {
		fmt.Printf("🤖 %s\n", currentModel)
		fmt.Println("   (this model is not in the built-in list)")
	}

	fmt.Println()
	fmt.Println("💡 Tip: use /model list to see all available models")
	fmt.Println("      use /model <model-name> to switch models")
	fmt.Println()

	return nil
}

// listAllModels displays all available models with details
func (c *ModelCommand) listAllModels() error {
	currentModel := c.getCurrentModel()

	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║                    Available Models                     ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Println()

	for _, model := range AvailableModels {
		if model.ID == currentModel {
			fmt.Printf("▶ %s (currently active)\n", model.Name)
		} else {
			fmt.Printf("  %s\n", model.Name)
		}
		fmt.Printf("   ID:          %s\n", model.ID)
		fmt.Printf("   Description: %s\n", model.Description)
		fmt.Printf("   Context:     %s tokens\n", formatNumber(model.ContextWin))
		fmt.Printf("   Input price: $%.2f/1M tokens\n", model.InputPrice)
		fmt.Printf("   Output price:$%.2f/1M tokens\n", model.OutputPrice)
		fmt.Printf("   Strengths:   %s\n", strings.Join(model.Strengths, ", "))
		fmt.Println()
	}

	fmt.Println("💡 Use /model <model-id> to switch models")
	fmt.Println()

	return nil
}

// switchModel switches to a new model and persists the choice
func (c *ModelCommand) switchModel(modelName string) error {
	// Try to find the model by full ID or partial match
	var targetModel *AIModelInfo

	// First, try exact match
	if model, exists := AvailableModels[modelName]; exists {
		targetModel = &model
	} else {
		// Try partial match
		for _, model := range AvailableModels {
			if strings.Contains(model.ID, modelName) ||
				strings.Contains(strings.ToLower(model.Name), modelName) {
				targetModel = &model
				break
			}
		}
	}

	// If no match in known models, allow arbitrary model ID
	if targetModel == nil {
		targetModel = &AIModelInfo{
			ID:          modelName,
			Name:        modelName,
			Description: "Custom or newly released model",
			ContextWin:  200000,
		}
	}

	// Get current model for comparison
	currentModel := c.getCurrentModel()
	if currentModel == targetModel.ID {
		fmt.Printf("\n✅ Already using %s\n\n", targetModel.Name)
		return nil
	}

	// Load config
	configPath := config.GetConfigPath()
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Update model
	cfg.Model = targetModel.ID

	// Save config
	if err := cfg.Save(configPath); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	// Print success message with model details
	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║                     Model Switched                      ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Printf("✅ Switched to: %s\n", targetModel.Name)
	fmt.Printf("   Model ID: %s\n", targetModel.ID)
	fmt.Printf("\n   Capability overview:\n")
	for _, strength := range targetModel.Strengths {
		fmt.Printf("   • %s\n", strength)
	}
	fmt.Printf("\n   💰 Pricing:\n")
	fmt.Printf("   Input:  $%.2f / 1M tokens\n", targetModel.InputPrice)
	fmt.Printf("   Output: $%.2f / 1M tokens\n", targetModel.OutputPrice)
	fmt.Printf("\n   📊 Context window: %s tokens\n", formatNumber(targetModel.ContextWin))
	fmt.Println()
	fmt.Println("📝 Configuration saved; it will take effect for new conversations")
	fmt.Println()

	return nil
}

// getCurrentModel returns the currently configured model
func (c *ModelCommand) getCurrentModel() string {
	// Priority 1: Environment variable
	if envModel := os.Getenv("CLAUDE_CODE_MODEL"); envModel != "" {
		return envModel
	}

	// Priority 2: Config file
	configPath := config.GetConfigPath()
	if cfg, err := config.Load(configPath); err == nil && cfg.Model != "" {
		return cfg.Model
	}

	// Priority 3: Default
	defaultCfg := config.DefaultConfig()
	return defaultCfg.Model
}

// formatNumber formats a number with thousands separators
func formatNumber(n int) string {
	if n < 0 {
		return "-" + formatNumber(-n)
	}
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}
	if n < 1000000 {
		return fmt.Sprintf("%d,%03d", n/1000, n%1000)
	}
	if n < 1000000000 {
		return fmt.Sprintf("%d,%03d,%03d", n/1000000, (n%1000000)/1000, n%1000)
	}
	return fmt.Sprintf("%d,%03d,%03d,%03d", n/1000000000, (n%1000000000)/1000000, (n%1000000)/1000, n%1000)
}

func init() {
	// Register the command
	Register(NewModelCommand())
}
