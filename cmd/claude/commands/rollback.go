package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/Aspirin0000/claude-code-go/internal/state"
)

// RollbackCommand undoes the last AI file modification
type RollbackCommand struct {
	*BaseCommand
}

// NewRollbackCommand creates the /rollback command
func NewRollbackCommand() *RollbackCommand {
	return &RollbackCommand{
		BaseCommand: NewBaseCommand(
			"rollback",
			"Undo the last AI file modification",
			CategorySession,
		).WithAliases("undo").
			WithHelp(`Usage: /rollback

Undo the most recent file modification made by an AI tool.
This restores the file to its state before the last edit.

Supported operations:
  • file_write  - Restores previous file content
  • file_edit   - Reverts the search-and-replace change
  • sed_replace - Reverts the regex replacement
  • file_delete - Restores the deleted file
  • file_move   - Moves the file back to its original location
  • notebook_edit - Restores the notebook to its previous state

Aliases: /undo`),
	}
}

// Execute runs the rollback command
func (c *RollbackCommand) Execute(ctx context.Context, args []string) error {
	edits := state.GlobalState.GetEdits()
	if len(edits) == 0 {
		fmt.Println("ℹ️  No file modifications to rollback.")
		return nil
	}

	lastEdit := edits[len(edits)-1]
	if err := c.rollbackEdit(lastEdit); err != nil {
		return fmt.Errorf("rollback failed: %w", err)
	}

	// Remove the last edit
	state.GlobalState.SetEdits(edits[:len(edits)-1])

	fmt.Printf("✅ Rolled back: %s\n", lastEdit.Description)
	fmt.Printf("   File: %s\n", lastEdit.FilePath)
	return nil
}

func (c *RollbackCommand) rollbackEdit(edit state.Edit) error {
	switch edit.Operation {
	case "write", "edit":
		if err := os.WriteFile(edit.FilePath, edit.BeforeContent, 0644); err != nil {
			return fmt.Errorf("failed to restore file: %w", err)
		}
	case "delete":
		if len(edit.BeforeContent) > 0 {
			if err := os.WriteFile(edit.FilePath, edit.BeforeContent, 0644); err != nil {
				return fmt.Errorf("failed to restore deleted file: %w", err)
			}
		} else {
			return fmt.Errorf("cannot restore deleted directory or empty file (no backup available)")
		}
	case "move":
		if edit.ExtraPath == "" {
			return fmt.Errorf("missing destination path for move rollback")
		}
		if err := os.Rename(edit.ExtraPath, edit.FilePath); err != nil {
			return fmt.Errorf("failed to move file back: %w", err)
		}
	case "mkdir":
		if err := os.Remove(edit.FilePath); err != nil {
			return fmt.Errorf("failed to remove created directory: %w", err)
		}
	default:
		return fmt.Errorf("unsupported operation for rollback: %s", edit.Operation)
	}
	return nil
}

func init() {
	Register(NewRollbackCommand())
}
