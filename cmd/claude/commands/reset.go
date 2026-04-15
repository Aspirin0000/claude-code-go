package commands

import (
	"context"
	"fmt"

	"github.com/Aspirin0000/claude-code-go/internal/state"
)

// ResetCommand resets the conversation by clearing all messages
// This is different from /clear which clears the terminal screen
type ResetCommand struct {
	*BaseCommand
}

// NewResetCommand creates the /reset command
func NewResetCommand() *ResetCommand {
	return &ResetCommand{
		BaseCommand: NewBaseCommand(
			"reset",
			"Reset the conversation by clearing all messages",
			CategorySession,
		).WithAliases("clear-history", "clear-chat").
			WithHelp(`Usage: /reset

Clear all conversation messages and reset the session state.
This removes all messages from the current conversation history.

Aliases: /clear-history, /clear-chat

Note: This action cannot be undone. Use /save before resetting if you want to preserve the conversation.`),
	}
}

// Execute runs the reset command
func (c *ResetCommand) Execute(ctx context.Context, args []string) error {
	messageCount := len(state.GlobalState.GetMessages())
	if messageCount == 0 {
		fmt.Println("ℹ️  No messages to reset.")
		return nil
	}

	state.GlobalState.ClearMessages()
	state.GlobalState.TurnCount = 0

	fmt.Printf("✅ Conversation reset successfully. Removed %d message(s).\n", messageCount)
	fmt.Println("   You can start a fresh conversation now.")

	return nil
}

func init() {
	Register(NewResetCommand())
}
