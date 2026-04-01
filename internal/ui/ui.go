// Package ui 提供终端用户界面
package ui

import (
	"fmt"
	"strings"
)

// TerminalUI 终端界面
type TerminalUI struct {
	width  int
	height int
}

// NewTerminalUI 创建新的终端 UI
func NewTerminalUI() *TerminalUI {
	return &TerminalUI{
		width:  80,
		height: 24,
	}
	// TODO: 获取实际终端尺寸
}

// PrintWelcome 打印欢迎信息
func (ui *TerminalUI) PrintWelcome() {
	fmt.Println("╔════════════════════════════════════════╗")
	fmt.Println("║     Claude Code Go v0.1.0              ║")
	fmt.Println("║     交互式 AI 编程助手                  ║")
	fmt.Println("╚════════════════════════════════════════╝")
	fmt.Println()
}

// PrintPrompt 打印提示符
func (ui *TerminalUI) PrintPrompt() {
	fmt.Print("\n❯ ")
}

// PrintMessage 打印消息
func (ui *TerminalUI) PrintMessage(role, content string) {
	switch role {
	case "user":
		fmt.Printf("\n👤 你:\n%s\n", content)
	case "assistant":
		fmt.Printf("\n🤖 Claude:\n%s\n", content)
	case "system":
		fmt.Printf("\nℹ️  %s\n", content)
	default:
		fmt.Printf("\n%s: %s\n", role, content)
	}
}

// PrintError 打印错误
func (ui *TerminalUI) PrintError(err string) {
	fmt.Printf("\n❌ 错误: %s\n", err)
}

// PrintSuccess 打印成功消息
func (ui *TerminalUI) PrintSuccess(msg string) {
	fmt.Printf("\n✅ %s\n", msg)
}

// PrintToolUse 打印工具使用信息
func (ui *TerminalUI) PrintToolUse(toolName string, input string) {
	fmt.Printf("\n🔧 使用工具: %s\n", toolName)
	if input != "" {
		fmt.Printf("   参数: %s\n", input)
	}
}

// PrintToolResult 打印工具结果
func (ui *TerminalUI) PrintToolResult(success bool, output string) {
	if success {
		fmt.Printf("   结果: ✓\n")
		if output != "" {
			lines := strings.Split(output, "\n")
			for i, line := range lines {
				if i >= 5 {
					fmt.Printf("   ... (%d more lines)\n", len(lines)-5)
					break
				}
				fmt.Printf("   %s\n", line)
			}
		}
	} else {
		fmt.Printf("   结果: ✗ %s\n", output)
	}
}

// PrintHelp 打印帮助信息
func (ui *TerminalUI) PrintHelp() {
	fmt.Println("\n可用命令:")
	fmt.Println("  /help      - 显示帮助")
	fmt.Println("  /quit      - 退出程序")
	fmt.Println("  /clear     - 清空对话历史")
	fmt.Println("  /model     - 切换模型")
	fmt.Println("  /tools     - 列出可用工具")
	fmt.Println()
	fmt.Println("快捷键:")
	fmt.Println("  Ctrl+C     - 退出")
	fmt.Println("  Ctrl+D     - 发送消息")
	fmt.Println()
}

// ClearScreen 清屏
func (ui *TerminalUI) ClearScreen() {
	fmt.Print("\033[H\033[2J")
}

// SetSize 设置终端尺寸
func (ui *TerminalUI) SetSize(width, height int) {
	ui.width = width
	ui.height = height
}

// WrapText 自动换行
func (ui *TerminalUI) WrapText(text string) string {
	lines := strings.Split(text, "\n")
	var result []string

	for _, line := range lines {
		if len(line) <= ui.width {
			result = append(result, line)
		} else {
			// 简单换行
			for len(line) > ui.width {
				result = append(result, line[:ui.width])
				line = line[ui.width:]
			}
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}
