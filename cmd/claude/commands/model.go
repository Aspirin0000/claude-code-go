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
	"claude-3-opus-20240229": {
		ID:          "claude-3-opus-20240229",
		Name:        "Claude 3 Opus",
		Description: "最强大的Claude模型，适合复杂任务",
		ContextWin:  200000,
		InputPrice:  15.0,
		OutputPrice: 75.0,
		Strengths: []string{
			"复杂的推理和分析",
			"创意写作",
			"代码生成",
			"详细的技术文档",
		},
	},
	"claude-3-sonnet-20240229": {
		ID:          "claude-3-sonnet-20240229",
		Name:        "Claude 3 Sonnet",
		Description: "平衡性能和速度，适合大多数任务",
		ContextWin:  200000,
		InputPrice:  3.0,
		OutputPrice: 15.0,
		Strengths: []string{
			"日常对话",
			"代码辅助",
			"内容摘要",
			"多语言支持",
		},
	},
	"claude-3-haiku-20240307": {
		ID:          "claude-3-haiku-20240307",
		Name:        "Claude 3 Haiku",
		Description: "最快最经济的模型，适合简单任务",
		ContextWin:  200000,
		InputPrice:  0.25,
		OutputPrice: 1.25,
		Strengths: []string{
			"快速响应",
			"简单问答",
			"分类任务",
			"轻量级任务",
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
			"切换或查看AI模型",
			CategoryConfig,
		).WithAliases("m", "switch-model").
			WithHelp(`使用: /model [模型名|list]

切换使用的AI模型或查看当前模型信息。

可用模型:
  • claude-3-opus-20240229   - 最强大的模型，适合复杂任务
  • claude-3-sonnet-20240229 - 平衡性能和速度，适合大多数任务
  • claude-3-haiku-20240307  - 最快最经济，适合简单任务

用法:
  /model                    - 显示当前模型和可用模型列表
  /model <模型名>           - 切换到指定模型
  /model list               - 列出所有可用模型及详细信息

别名: /m, /switch-model`),
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
	fmt.Println("║                   当前模型 (Current Model)                ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Println()

	if exists {
		fmt.Printf("🤖 %s\n", modelInfo.Name)
		fmt.Printf("   ID:          %s\n", modelInfo.ID)
		fmt.Printf("   描述:        %s\n", modelInfo.Description)
		fmt.Printf("   上下文窗口:  %s tokens\n", formatNumber(modelInfo.ContextWin))
		fmt.Printf("\n   💰 定价 (每1M tokens):\n")
		fmt.Printf("      输入:  $%.2f\n", modelInfo.InputPrice)
		fmt.Printf("      输出:  $%.2f\n", modelInfo.OutputPrice)
		fmt.Printf("\n   ✨ 擅长:\n")
		for _, strength := range modelInfo.Strengths {
			fmt.Printf("      • %s\n", strength)
		}
	} else {
		fmt.Printf("🤖 %s\n", currentModel)
		fmt.Println("   (此模型不在内置列表中)")
	}

	fmt.Println()
	fmt.Println("💡 提示: 使用 /model list 查看所有可用模型")
	fmt.Println("        使用 /model <模型名> 切换模型")
	fmt.Println()

	return nil
}

// listAllModels displays all available models with details
func (c *ModelCommand) listAllModels() error {
	currentModel := c.getCurrentModel()

	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║              可用模型列表 (Available Models)              ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Println()

	for _, model := range AvailableModels {
		if model.ID == currentModel {
			fmt.Printf("▶ %s (当前使用中)\n", model.Name)
		} else {
			fmt.Printf("  %s\n", model.Name)
		}
		fmt.Printf("   ID:          %s\n", model.ID)
		fmt.Printf("   描述:        %s\n", model.Description)
		fmt.Printf("   上下文窗口:  %s tokens\n", formatNumber(model.ContextWin))
		fmt.Printf("   输入价格:    $%.2f/1M tokens\n", model.InputPrice)
		fmt.Printf("   输出价格:    $%.2f/1M tokens\n", model.OutputPrice)
		fmt.Printf("   擅长:        %s\n", strings.Join(model.Strengths, ", "))
		fmt.Println()
	}

	fmt.Println("💡 使用 /model <模型ID> 切换模型")
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

	if targetModel == nil {
		fmt.Printf("\n❌ 错误: 未知的模型 '%s'\n\n", modelName)
		fmt.Println("可用模型:")
		for _, model := range AvailableModels {
			fmt.Printf("  • %s (%s)\n", model.ID, model.Name)
		}
		fmt.Println()
		fmt.Println("提示: 可以使用部分名称匹配，例如:")
		fmt.Println("  /model opus  → 切换到 claude-3-opus-20240229")
		fmt.Println("  /model sonnet → 切换到 claude-3-sonnet-20240229")
		fmt.Println("  /model haiku → 切换到 claude-3-haiku-20240307")
		fmt.Println()
		return fmt.Errorf("unknown model: %s", modelName)
	}

	// Get current model for comparison
	currentModel := c.getCurrentModel()
	if currentModel == targetModel.ID {
		fmt.Printf("\n✅ 已经在使用 %s\n\n", targetModel.Name)
		return nil
	}

	// Load config
	configPath := config.GetConfigPath()
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("加载配置失败: %w", err)
	}

	// Update model
	cfg.Model = targetModel.ID

	// Save config
	if err := cfg.Save(configPath); err != nil {
		return fmt.Errorf("保存配置失败: %w", err)
	}

	// Print success message with model details
	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║                  模型切换成功 (Model Switched)             ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Printf("✅ 已切换到: %s\n", targetModel.Name)
	fmt.Printf("   模型 ID:  %s\n", targetModel.ID)
	fmt.Printf("\n   能力概览:\n")
	for _, strength := range targetModel.Strengths {
		fmt.Printf("   • %s\n", strength)
	}
	fmt.Printf("\n   💰 使用成本:\n")
	fmt.Printf("   输入: $%.2f / 1M tokens\n", targetModel.InputPrice)
	fmt.Printf("   输出: $%.2f / 1M tokens\n", targetModel.OutputPrice)
	fmt.Printf("\n   📊 上下文窗口: %s tokens\n", formatNumber(targetModel.ContextWin))
	fmt.Println()
	fmt.Println("📝 配置已保存，将在下次对话生效")
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
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}
	if n < 1000000 {
		return fmt.Sprintf("%d,%03d", n/1000, n%1000)
	}
	return fmt.Sprintf("%d,%03d,%03d", n/1000000, (n%1000000)/1000, n%1000)
}

func init() {
	// Register the command
	Register(NewModelCommand())
}
