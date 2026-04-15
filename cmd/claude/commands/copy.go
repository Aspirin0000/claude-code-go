package commands

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/Aspirin0000/claude-code-go/internal/state"
)

// CopyCommand copies the last assistant message to the system clipboard
type CopyCommand struct {
	*BaseCommand
}

// NewCopyCommand creates the /copy command
func NewCopyCommand() *CopyCommand {
	return &CopyCommand{
		BaseCommand: NewBaseCommand(
			"copy",
			"Copy the last assistant message to clipboard",
			CategorySession,
		).WithHelp(`Usage: /copy

Copy the most recent assistant message to the system clipboard.

Supported platforms:
  macOS  - uses pbcopy
  Linux  - uses xclip or xsel
  Windows- uses clip or PowerShell

Examples:
  /copy

Note: If no assistant message is found, nothing is copied.`),
	}
}

// Execute runs the copy command
func (c *CopyCommand) Execute(ctx context.Context, args []string) error {
	messages := state.GlobalState.GetMessages()

	var content string
	for i := len(messages) - 1; i >= 0; i-- {
		role := messages[i].Role
		if role == "" {
			role = messages[i].Type
		}
		if role == "assistant" {
			content = messages[i].Content
			break
		}
	}

	if content == "" {
		fmt.Println("ℹ️  No assistant message found to copy.")
		return nil
	}

	if err := copyToClipboard(content); err != nil {
		return fmt.Errorf("failed to copy to clipboard: %w", err)
	}

	preview := content
	if len(preview) > 80 {
		preview = preview[:77] + "..."
	}
	fmt.Printf("✅ Copied to clipboard: %s\n", preview)
	return nil
}

func copyToClipboard(text string) error {
	switch runtime.GOOS {
	case "darwin":
		cmd := exec.Command("pbcopy")
		cmd.Stdin = strings.NewReader(text)
		return cmd.Run()
	case "windows":
		cmd := exec.Command("clip")
		cmd.Stdin = strings.NewReader(text)
		if err := cmd.Run(); err != nil {
			// Fallback to PowerShell
			ps := exec.Command("powershell", "-command", fmt.Sprintf("Set-Clipboard -Value '%s'", strings.ReplaceAll(text, "'", "''")))
			return ps.Run()
		}
		return nil
	default: // linux and others
		// Try xclip first
		cmd := exec.Command("xclip", "-selection", "clipboard")
		cmd.Stdin = strings.NewReader(text)
		if err := cmd.Run(); err != nil {
			// Fallback to xsel
			cmd = exec.Command("xsel", "--clipboard", "--input")
			cmd.Stdin = strings.NewReader(text)
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("clipboard tools not available (try installing xclip or xsel)")
			}
		}
		return nil
	}
}

func init() {
	Register(NewCopyCommand())
}
