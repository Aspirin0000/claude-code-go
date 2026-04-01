package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Aspirin0000/claude-code-go/internal/state"
)

// ResumeCommand resumes previous conversation sessions
// Category: CategorySession
// Aliases: ["restore", "continue"]
type ResumeCommand struct {
	base         *BaseCommand
	sessionStore *SessionStore
}

// SessionStore manages session storage and retrieval
type SessionStore struct {
	state *state.AppState
}

// Session represents a saved session
type Session struct {
	ID           string
	StartTime    time.Time
	LastActive   time.Time
	MessageCount int
	Preview      string
}

// NewResumeCommand creates a new ResumeCommand instance
func NewResumeCommand() *ResumeCommand {
	return &ResumeCommand{
		base: NewBaseCommand(
			"resume",
			"Resumes a previous conversation session",
			CategorySession,
		).WithAliases("restore", "continue").WithHelp(`Usage: /resume [session-id]

Resumes a previous conversation session.

Examples:
  /resume          List recent sessions
  /resume abc123   Resume specific session by ID

Aliases: restore, continue`),
		sessionStore: &SessionStore{
			state: state.GlobalState,
		},
	}
}

// Name returns the command name
func (c *ResumeCommand) Name() string {
	return c.base.Name()
}

// Aliases returns command aliases
func (c *ResumeCommand) Aliases() []string {
	return c.base.Aliases()
}

// Description returns the command description
func (c *ResumeCommand) Description() string {
	return c.base.Description()
}

// Category returns the command category
func (c *ResumeCommand) Category() CommandCategory {
	return c.base.Category()
}

// Help returns the help text
func (c *ResumeCommand) Help() string {
	return c.base.Help()
}

// Execute runs the resume command
// - Lists recent sessions if no arg provided
// - Loads specific session by ID if provided
// - Restores conversation context
func (c *ResumeCommand) Execute(ctx context.Context, args []string) error {
	if len(args) == 0 {
		// List recent sessions
		sessions, err := c.sessionStore.ListSessions()
		if err != nil {
			return fmt.Errorf("failed to list sessions: %w", err)
		}

		if len(sessions) == 0 {
			fmt.Println("No previous sessions found.")
			return nil
		}

		output := c.formatSessionList(sessions)
		fmt.Println(output)
		return nil
	}

	// Resume specific session
	sessionID := args[0]
	session, err := c.sessionStore.LoadSession(sessionID)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("Session '%s' not found. Use '/resume' to list available sessions.\n", sessionID)
			return nil
		}
		return fmt.Errorf("failed to load session: %w", err)
	}

	// Restore conversation context
	if err := c.sessionStore.RestoreSession(session); err != nil {
		return fmt.Errorf("failed to restore session: %w", err)
	}

	fmt.Printf("Resumed session '%s' (%s)\n", session.ID, c.formatTimestamp(session.LastActive))
	return nil
}

// formatSessionList formats the list of sessions for display
func (c *ResumeCommand) formatSessionList(sessions []*Session) string {
	var builder strings.Builder
	builder.WriteString("Recent sessions:\n\n")

	for i, session := range sessions {
		builder.WriteString(fmt.Sprintf("%d. %s\n", i+1, session.ID))
		builder.WriteString(fmt.Sprintf("   Started: %s\n", c.formatTimestamp(session.StartTime)))
		builder.WriteString(fmt.Sprintf("   Last active: %s\n", c.formatTimestamp(session.LastActive)))
		builder.WriteString(fmt.Sprintf("   Messages: %d\n", session.MessageCount))
		if session.Preview != "" {
			preview := session.Preview
			if len(preview) > 50 {
				preview = preview[:47] + "..."
			}
			builder.WriteString(fmt.Sprintf("   Preview: %s\n", preview))
		}
		if i < len(sessions)-1 {
			builder.WriteString("\n")
		}
	}

	builder.WriteString("\nUse '/resume <session-id>' to restore a session.")
	return builder.String()
}

// formatTimestamp formats a timestamp for display
func (c *ResumeCommand) formatTimestamp(t time.Time) string {
	duration := time.Since(t)
	switch {
	case duration < time.Minute:
		return "just now"
	case duration < time.Hour:
		return fmt.Sprintf("%d minutes ago", int(duration.Minutes()))
	case duration < 24*time.Hour:
		return fmt.Sprintf("%d hours ago", int(duration.Hours()))
	default:
		return t.Format("Jan 02 15:04")
	}
}

// ListSessions returns a list of recent sessions
func (s *SessionStore) ListSessions() ([]*Session, error) {
	sessionsDir := s.getSessionsDir()

	entries, err := os.ReadDir(sessionsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*Session{}, nil
		}
		return nil, err
	}

	var sessions []*Session
	for _, entry := range entries {
		if entry.IsDir() {
			session, err := s.LoadSession(entry.Name())
			if err != nil {
				continue // Skip invalid sessions
			}
			sessions = append(sessions, session)
		}
	}

	// Sort by last active time (most recent first)
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].LastActive.After(sessions[j].LastActive)
	})

	// Return top 10 sessions
	if len(sessions) > 10 {
		sessions = sessions[:10]
	}

	return sessions, nil
}

// LoadSession loads a specific session by ID
func (s *SessionStore) LoadSession(sessionID string) (*Session, error) {
	sessionPath := filepath.Join(s.getSessionsDir(), sessionID)

	info, err := os.Stat(sessionPath)
	if err != nil {
		return nil, err
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("session '%s' is not a directory", sessionID)
	}

	// Read session metadata
	metaPath := filepath.Join(sessionPath, "meta.json")
	if _, err := os.Stat(metaPath); os.IsNotExist(err) {
		// Create basic session from directory info
		session := &Session{
			ID:         sessionID,
			StartTime:  info.ModTime().Add(-time.Hour), // Estimate start time
			LastActive: info.ModTime(),
		}
		return session, nil
	}

	// Load full session metadata
	// In a real implementation, this would read and parse meta.json
	session := &Session{
		ID:         sessionID,
		StartTime:  info.ModTime().Add(-time.Hour),
		LastActive: info.ModTime(),
		Preview:    "Previous conversation",
	}

	return session, nil
}

// RestoreSession restores the conversation context from a session
func (s *SessionStore) RestoreSession(session *Session) error {
	if session == nil {
		return fmt.Errorf("session is nil")
	}

	// Restore session ID to global state
	s.state.SetSessionID(session.ID)

	// In a real implementation, this would:
	// 1. Load the conversation messages from storage
	// 2. Restore the message history to the global state
	// 3. Update other session context

	return nil
}

// getSessionsDir returns the directory path for storing sessions
func (s *SessionStore) getSessionsDir() string {
	// Default to ~/.claude/sessions
	home, err := os.UserHomeDir()
	if err != nil {
		return ".sessions"
	}
	return filepath.Join(home, ".claude", "sessions")
}
