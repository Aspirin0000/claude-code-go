package commands

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/Aspirin0000/claude-code-go/internal/bootstrap"
	"github.com/Aspirin0000/claude-code-go/internal/config"
	"github.com/Aspirin0000/claude-code-go/internal/mcp"
	"github.com/Aspirin0000/claude-code-go/internal/state"
)

// StatusCommand shows current session status
type StatusCommand struct {
	*BaseCommand
}

// NewStatusCommand creates a new status command
func NewStatusCommand() *StatusCommand {
	cmd := &StatusCommand{
		BaseCommand: NewBaseCommand(
			"status",
			"显示当前会话状态",
			CategoryGeneral,
		),
	}
	cmd.WithAliases("st")
	cmd.WithHelp(`显示当前会话状态信息，包括:
  - 当前使用的模型
  - 会话持续时间
  - 消息数量
  - 当前目录
  - Git 状态
  - MCP 服务器连接状态
  - Token 使用情况`)
	return cmd
}

// Execute runs the status command
func (c *StatusCommand) Execute(ctx context.Context, args []string) error {
	// Get session info
	sessionID := bootstrap.GetSessionId()
	sessionStart := bootstrap.GetSessionStartTime()
	sessionDuration := time.Since(sessionStart)

	// Get message count and calculate tokens
	messages := state.GlobalState.GetMessages()
	messageCount := len(messages)
	inputTokens, outputTokens := calculateTokenUsage(messages)
	totalTokens := inputTokens + outputTokens

	// Get current directory
	cwd := bootstrap.GetCwdState()
	originalCwd := bootstrap.GetOriginalCwd()

	// Get model info from config
	modelInfo := getCurrentModelInfo()

	// Print status header
	fmt.Println()
	fmt.Println("╔════════════════════════════════════════╗")
	fmt.Println("║           会话状态 (Session Status)      ║")
	fmt.Println("╚════════════════════════════════════════╝")
	fmt.Println()

	// Session info
	fmt.Printf("📊 会话信息\n")
	fmt.Printf("   会话 ID:       %s\n", sessionID)
	fmt.Printf("   会话时长:      %s\n", formatDuration(sessionDuration))
	fmt.Printf("   消息数量:      %d\n", messageCount)
	fmt.Println()

	// Model info with config source
	fmt.Printf("🤖 模型信息\n")
	fmt.Printf("   当前模型:      %s\n", modelInfo.Name)
	if modelInfo.Source != "" {
		fmt.Printf("   配置来源:      %s\n", modelInfo.Source)
	}
	fmt.Println()

	// Directory info
	fmt.Printf("📁 目录信息\n")
	fmt.Printf("   当前目录:      %s\n", cwd)
	if cwd != originalCwd {
		fmt.Printf("   启动目录:      %s\n", originalCwd)
	}
	fmt.Println()

	// Git status
	gitInfo := getGitStatus(cwd)
	if gitInfo != "" {
		fmt.Printf("🔀 Git 状态\n%s", gitInfo)
		fmt.Println()
	}

	// MCP servers status
	mcpInfo := getMCPStatus()
	fmt.Printf("🔌 MCP 服务器\n%s", mcpInfo)
	fmt.Println()

	// Token usage with actual calculations
	fmt.Printf("📝 Token 使用\n")
	fmt.Printf("   输入 Token:    %d\n", inputTokens)
	fmt.Printf("   输出 Token:    %d\n", outputTokens)
	fmt.Printf("   总计:          %d\n", totalTokens)
	if messageCount > 0 {
		avgPerMessage := totalTokens / messageCount
		fmt.Printf("   平均每消息:    %d tokens\n", avgPerMessage)
	}
	fmt.Println()

	return nil
}

// formatDuration formats duration in human-readable format
func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%d小时 %d分钟 %d秒", hours, minutes, seconds)
	} else if minutes > 0 {
		return fmt.Sprintf("%d分钟 %d秒", minutes, seconds)
	}
	return fmt.Sprintf("%d秒", seconds)
}

// ModelInfo contains model name and configuration source
type ModelInfo struct {
	Name   string
	Source string
}

// getCurrentModelInfo returns the current model info from config with fallback
func getCurrentModelInfo() ModelInfo {
	// Priority 1: Environment variable
	if envModel := os.Getenv("CLAUDE_CODE_MODEL"); envModel != "" {
		return ModelInfo{
			Name:   envModel,
			Source: "环境变量 (CLAUDE_CODE_MODEL)",
		}
	}

	// Priority 2: Config file
	configPath := config.GetConfigPath()
	if cfg, err := config.Load(configPath); err == nil && cfg.Model != "" {
		return ModelInfo{
			Name:   cfg.Model,
			Source: fmt.Sprintf("配置文件 (%s)", configPath),
		}
	}

	// Priority 3: Default model
	defaultCfg := config.DefaultConfig()
	return ModelInfo{
		Name:   defaultCfg.Model,
		Source: "默认配置",
	}
}

// calculateTokenUsage calculates input/output tokens from messages
func calculateTokenUsage(messages []state.Message) (inputTokens, outputTokens int) {
	for _, msg := range messages {
		tokenCount := estimateTokensForText(msg.Content)

		switch msg.Role {
		case "user", "system":
			inputTokens += tokenCount
		case "assistant":
			outputTokens += tokenCount
		default:
			// Unknown role, count as input
			inputTokens += tokenCount
		}
	}

	return inputTokens, outputTokens
}

// estimateTokensForText estimates token count for text content
// Uses a hybrid approach: 1 token ≈ 4 characters for English, 1 token ≈ 2 characters for CJK
func estimateTokensForText(text string) int {
	if text == "" {
		return 0
	}

	// Estimate based on character types
	var asciiCount, cjkCount int
	for _, r := range text {
		if r < 128 {
			asciiCount++
		} else {
			// Assume non-ASCII is CJK or other Unicode
			cjkCount++
		}
	}

	// ASCII: ~4 chars per token
	// CJK: ~2 chars per token
	asciiTokens := asciiCount / 4
	if asciiCount%4 > 0 {
		asciiTokens++
	}

	cjkTokens := cjkCount / 2
	if cjkCount%2 > 0 {
		cjkTokens++
	}

	// Base estimation
	estimatedTokens := asciiTokens + cjkTokens

	// Add overhead for message structure (role, formatting, etc.)
	overhead := 4

	return estimatedTokens + overhead
}

// getGitStatus returns git status information
func getGitStatus(cwd string) string {
	// Check if in git repo
	gitDir := filepath.Join(cwd, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		// Try to find git root
		cmd := exec.Command("git", "rev-parse", "--git-dir")
		cmd.Dir = cwd
		output, err := cmd.Output()
		if err != nil {
			return ""
		}
		gitDir = strings.TrimSpace(string(output))
	}

	var result strings.Builder

	// Get branch name
	branchCmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	branchCmd.Dir = cwd
	branch, err := branchCmd.Output()
	if err == nil {
		result.WriteString(fmt.Sprintf("   分支:          %s", strings.TrimSpace(string(branch))))
	}

	// Get status
	statusCmd := exec.Command("git", "status", "--porcelain")
	statusCmd.Dir = cwd
	status, err := statusCmd.Output()
	if err == nil {
		lines := strings.Split(strings.TrimSpace(string(status)), "\n")
		modified := 0
		staged := 0
		untracked := 0

		for _, line := range lines {
			if line == "" {
				continue
			}
			if len(line) >= 2 {
				indexStatus := line[0]
				workTreeStatus := line[1]

				if indexStatus != ' ' && indexStatus != '?' {
					staged++
				}
				if workTreeStatus != ' ' {
					modified++
				}
				if indexStatus == '?' {
					untracked++
				}
			}
		}

		if modified > 0 || staged > 0 || untracked > 0 {
			result.WriteString(fmt.Sprintf(" (%d修改, %d暂存, %d未跟踪)", modified, staged, untracked))
		} else {
			result.WriteString(" (工作区干净)")
		}
		result.WriteString("\n")
	}

	return result.String()
}

// getMCPStatus returns MCP server status
func getMCPStatus() string {
	manager := mcp.GetGlobalMCPManager()
	statuses := manager.GetStatus()

	if len(statuses) == 0 {
		return "   无配置的服务器\n"
	}

	var result strings.Builder
	connected := 0
	for _, status := range statuses {
		icon := "❌"
		if status.Connected {
			icon = "✅"
			connected++
		}
		result.WriteString(fmt.Sprintf("   %s %s (%s)\n", icon, status.Name, status.Type))
	}

	result.WriteString(fmt.Sprintf("   总计: %d/%d 已连接\n", connected, len(statuses)))

	return result.String()
}

func init() {
	// Register the command
	Register(NewStatusCommand())
}
