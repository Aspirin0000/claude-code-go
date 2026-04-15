// Package ui provides terminal user interface helpers
package ui

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// TerminalUI terminal user interface
type TerminalUI struct {
	width  int
	height int
}

// NewTerminalUI creates a new terminal UI
func NewTerminalUI() *TerminalUI {
	ui := &TerminalUI{
		width:  80,
		height: 24,
	}

	// Get actual terminal size
	if width, height, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
		ui.width = width
		ui.height = height
	}

	return ui
}

// PrintWelcome prints the welcome banner
func (ui *TerminalUI) PrintWelcome() {
	fmt.Println("╔════════════════════════════════════════╗")
	fmt.Printf("║     Claude Code Go %-19s ║\n", "v2.1.88")
	fmt.Println("║     Interactive AI Coding Assistant    ║")
	fmt.Println("╚════════════════════════════════════════╝")
	fmt.Println()
}

// PrintPrompt prints the input prompt
func (ui *TerminalUI) PrintPrompt() {
	fmt.Print("\n❯ ")
}

// PrintMessage prints a chat message
func (ui *TerminalUI) PrintMessage(role, content string) {
	switch role {
	case "user":
		fmt.Printf("\n👤 You:\n%s\n", content)
	case "assistant":
		fmt.Printf("\n🤖 Claude:\n%s\n", content)
	case "system":
		fmt.Printf("\nℹ️  %s\n", content)
	default:
		fmt.Printf("\n%s: %s\n", role, content)
	}
}

// PrintError prints an error message
func (ui *TerminalUI) PrintError(err string) {
	fmt.Printf("\n❌ Error: %s\n", err)
}

// PrintSuccess prints a success message
func (ui *TerminalUI) PrintSuccess(msg string) {
	fmt.Printf("\n✅ %s\n", msg)
}

// PrintToolUse prints tool invocation info
func (ui *TerminalUI) PrintToolUse(toolName string, input string) {
	fmt.Printf("\n🔧 Using tool: %s\n", toolName)
	if input != "" {
		fmt.Printf("   Args: %s\n", input)
	}
}

// PrintToolResult prints tool execution result
func (ui *TerminalUI) PrintToolResult(success bool, output string) {
	if success {
		fmt.Printf("   Result: ✓\n")
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
		fmt.Printf("   Result: ✗ %s\n", output)
	}
}

// PrintHelp prints help text
func (ui *TerminalUI) PrintHelp() {
	fmt.Println("\nAvailable commands:")
	fmt.Println("  /help      - Show help")
	fmt.Println("  /quit      - Exit the program")
	fmt.Println("  /clear     - Clear conversation history")
	fmt.Println("  /model     - Switch model")
	fmt.Println("  /tools     - List available tools")
	fmt.Println()
	fmt.Println("Shortcuts:")
	fmt.Println("  Ctrl+C     - Exit")
	fmt.Println("  Ctrl+D     - Send message")
	fmt.Println()
}

// ClearScreen clears the screen
func (ui *TerminalUI) ClearScreen() {
	fmt.Print("\033[H\033[2J")
}

// SetSize sets terminal dimensions
func (ui *TerminalUI) SetSize(width, height int) {
	ui.width = width
	ui.height = height
}

// WrapText wraps text to terminal width
func (ui *TerminalUI) WrapText(text string) string {
	lines := strings.Split(text, "\n")
	var result []string

	for _, line := range lines {
		if len(line) <= ui.width {
			result = append(result, line)
		} else {
			// Simple wrapping
			for len(line) > ui.width {
				result = append(result, line[:ui.width])
				line = line[ui.width:]
			}
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}
