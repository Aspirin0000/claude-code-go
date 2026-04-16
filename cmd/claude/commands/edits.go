package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/Aspirin0000/claude-code-go/internal/state"
)

// EditsCommand shows file modifications made by AI tools during the session
type EditsCommand struct {
	*BaseCommand
}

// NewEditsCommand creates the /edits command
func NewEditsCommand() *EditsCommand {
	return &EditsCommand{
		BaseCommand: NewBaseCommand(
			"edits",
			"Show file modifications made by AI tools during this session",
			CategorySession,
		).WithAliases("changes", "mods").
			WithHelp(`Usage: /edits

Display all file modifications made by AI tools during the current session.
This includes files written, edited, deleted, moved, or created by tools.

Aliases: /changes, /mods`),
	}
}

// Execute runs the edits command
func (c *EditsCommand) Execute(ctx context.Context, args []string) error {
	edits := state.GlobalState.GetEdits()
	if len(edits) == 0 {
		fmt.Println("ℹ️  No file modifications recorded in this session yet.")
		fmt.Println("   AI tools like file_write, file_edit, and file_delete will be tracked here.")
		return nil
	}

	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║              Session File Modifications                 ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Println()

	for i, edit := range edits {
		timestamp := edit.Timestamp.Format("15:04:05")
		opSymbol := operationSymbol(edit.Operation)
		fmt.Printf("%d. [%s] %s %s\n", i+1, timestamp, opSymbol, edit.FilePath)
		if edit.Description != "" {
			fmt.Printf("   %s (%s)\n", edit.Description, edit.Tool)
		} else {
			fmt.Printf("   Tool: %s\n", edit.Tool)
		}
		fmt.Println()
	}

	fmt.Printf("Total modifications: %d\n", len(edits))
	fmt.Println()
	return nil
}

func operationSymbol(op string) string {
	switch strings.ToLower(op) {
	case "write", "create":
		return "✍️ "
	case "edit":
		return "📝"
	case "delete":
		return "🗑️ "
	case "mkdir":
		return "📁"
	case "move":
		return "📦"
	default:
		return "🔧"
	}
}

func init() {
	Register(NewEditsCommand())
}
