package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/Aspirin0000/claude-code-go/internal/state"
)

// SessionsCommand manages auto-saved sessions
type SessionsCommand struct {
	*BaseCommand
}

// NewSessionsCommand creates the sessions command
func NewSessionsCommand() *SessionsCommand {
	return &SessionsCommand{
		BaseCommand: NewBaseCommand(
			"sessions",
			"List and manage saved sessions",
			CategorySession,
		).WithAliases("list-sessions", "saved").
			WithHelp(`Usage: /sessions [action] [options]

Manage auto-saved conversation sessions.

Actions:
  list       List all saved sessions (default)
  load <n>   Load session by number or name
  delete <n> Delete session by number or name
  clean      Remove all sessions

Examples:
  /sessions              List all saved sessions
  /sessions list         Same as above
  /sessions load 1       Load the most recent session
  /sessions load my_chat Load session named "my_chat"
  /sessions delete 3     Delete session #3
  /sessions clean        Delete all saved sessions

Note: Sessions are automatically saved when auto-save is enabled.
Use /config to check or modify auto-save settings.`),
	}
}

// Execute runs the sessions command
func (c *SessionsCommand) Execute(ctx context.Context, args []string) error {
	if state.GlobalSessionStorage == nil {
		return fmt.Errorf("session storage not initialized")
	}

	action := "list"
	if len(args) > 0 {
		action = args[0]
	}

	switch action {
	case "list", "ls":
		return c.listSessions()
	case "load", "restore":
		if len(args) < 2 {
			return fmt.Errorf("usage: /sessions load <number|name>")
		}
		return c.loadSession(args[1])
	case "delete", "rm", "remove":
		if len(args) < 2 {
			return fmt.Errorf("usage: /sessions delete <number|name>")
		}
		return c.deleteSession(args[1])
	case "clean", "clear", "purge":
		return c.cleanSessions()
	default:
		// If first arg is a number, treat it as "load <n>"
		return c.loadSession(action)
	}
}

func (c *SessionsCommand) listSessions() error {
	sessions, err := state.GlobalSessionStorage.ListSessions()
	if err != nil {
		return fmt.Errorf("failed to list sessions: %w", err)
	}

	if len(sessions) == 0 {
		fmt.Println("No saved sessions found.")
		fmt.Println("\nSessions are automatically saved when auto-save is enabled.")
		fmt.Println("Use '/config get auto_save' to check the current setting.")
		return nil
	}

	fmt.Printf("\n%-4s %-30s %-20s %s\n", "#", "Name", "Modified", "Size")
	fmt.Println(strings.Repeat("-", 80))

	for i, session := range sessions {
		fmt.Printf("%-4d %-30s %-20s %s\n",
			i+1,
			truncate(session.Name, 28),
			session.Modified.Format("2006-01-02 15:04"),
			session.FormatSize(),
		)
	}

	fmt.Println()
	fmt.Printf("Total: %d session(s)\n", len(sessions))
	fmt.Println("\nUse '/sessions load <number>' to restore a session")

	return nil
}

func (c *SessionsCommand) loadSession(identifier string) error {
	sessions, err := state.GlobalSessionStorage.ListSessions()
	if err != nil {
		return fmt.Errorf("failed to list sessions: %w", err)
	}

	if len(sessions) == 0 {
		return fmt.Errorf("no saved sessions found")
	}

	var targetName string

	// Try to parse as number first
	var num int
	if _, err := fmt.Sscanf(identifier, "%d", &num); err == nil {
		if num < 1 || num > len(sessions) {
			return fmt.Errorf("invalid session number: %d (valid range: 1-%d)", num, len(sessions))
		}
		targetName = sessions[num-1].Name
	} else {
		// Treat as name
		targetName = identifier
	}

	// Load the session
	loadedState, err := state.GlobalSessionStorage.LoadSession(targetName)
	if err != nil {
		return fmt.Errorf("failed to load session: %w", err)
	}

	// Copy to global state
	state.GlobalState.SetMessages(loadedState.GetMessages())
	state.GlobalState.SetSessionID(loadedState.SessionID)
	state.GlobalState.SetCWD(loadedState.CWD)
	state.GlobalState.ProjectRoot = loadedState.ProjectRoot

	fmt.Printf("✓ Session '%s' loaded successfully.\n", targetName)
	fmt.Printf("  - Messages: %d\n", len(loadedState.GetMessages()))
	fmt.Printf("  - Session ID: %s\n", loadedState.SessionID)

	return nil
}

func (c *SessionsCommand) deleteSession(identifier string) error {
	sessions, err := state.GlobalSessionStorage.ListSessions()
	if err != nil {
		return fmt.Errorf("failed to list sessions: %w", err)
	}

	if len(sessions) == 0 {
		return fmt.Errorf("no saved sessions found")
	}

	var targetName string

	// Try to parse as number first
	var num int
	if _, err := fmt.Sscanf(identifier, "%d", &num); err == nil {
		if num < 1 || num > len(sessions) {
			return fmt.Errorf("invalid session number: %d (valid range: 1-%d)", num, len(sessions))
		}
		targetName = sessions[num-1].Name
	} else {
		// Treat as name
		targetName = identifier
	}

	// Confirm deletion
	fmt.Printf("Delete session '%s'? [y/N]: ", targetName)
	var response string
	fmt.Scanln(&response)

	if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
		fmt.Println("Cancelled.")
		return nil
	}

	// Delete the session
	if err := state.GlobalSessionStorage.DeleteSession(targetName); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	fmt.Printf("✓ Session '%s' deleted.\n", targetName)
	return nil
}

func (c *SessionsCommand) cleanSessions() error {
	sessions, err := state.GlobalSessionStorage.ListSessions()
	if err != nil {
		return fmt.Errorf("failed to list sessions: %w", err)
	}

	if len(sessions) == 0 {
		fmt.Println("No sessions to clean.")
		return nil
	}

	fmt.Printf("WARNING: This will delete all %d saved sessions.\n", len(sessions))
	fmt.Print("Are you sure? Type 'yes' to confirm: ")

	var response string
	fmt.Scanln(&response)

	if response != "yes" {
		fmt.Println("Cancelled.")
		return nil
	}

	// Delete all sessions
	deleted := 0
	for _, session := range sessions {
		if err := state.GlobalSessionStorage.DeleteSession(session.Name); err == nil {
			deleted++
		}
	}

	fmt.Printf("✓ Cleaned %d session(s).\n", deleted)
	return nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func init() {
	Register(NewSessionsCommand())
}
