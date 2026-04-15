package commands

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/Aspirin0000/claude-code-go/internal/state"
)

func TestResumeCommandListEmpty(t *testing.T) {
	cmd := NewResumeCommand()
	// Use a temp session store to avoid real filesystem
	cmd.sessionStore = &SessionStore{state: state.NewAppState()}
	err := cmd.Execute(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestResumeCommandFormatSessionList(t *testing.T) {
	cmd := NewResumeCommand()
	sessions := []*Session{
		{
			ID:           "sess-1",
			StartTime:    time.Now().Add(-30 * time.Minute),
			LastActive:   time.Now(),
			MessageCount: 5,
			Preview:      "Hello world",
		},
	}
	output := cmd.formatSessionList(sessions)
	if !strings.Contains(output, "sess-1") {
		t.Errorf("expected session ID in output")
	}
	if !strings.Contains(output, "Messages: 5") {
		t.Errorf("expected message count in output")
	}
}

func TestResumeCommandFormatTimestamp(t *testing.T) {
	cmd := NewResumeCommand()
	if cmd.formatTimestamp(time.Now()) != "just now" {
		t.Errorf("expected just now")
	}
	if !strings.Contains(cmd.formatTimestamp(time.Now().Add(-2*time.Hour)), "hours ago") {
		t.Errorf("expected hours ago")
	}
}

func TestSessionStoreListSessionsEmpty(t *testing.T) {
	store := &SessionStore{state: state.NewAppState()}
	sessions, err := store.ListSessions()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sessions) != 0 {
		t.Errorf("expected 0 sessions, got %d", len(sessions))
	}
}

func TestSessionStoreLoadSessionNotExist(t *testing.T) {
	store := &SessionStore{state: state.NewAppState()}
	_, err := store.LoadSession("nonexistent_session_xyz")
	if err == nil {
		t.Error("expected error for nonexistent session")
	}
}

func TestSessionStoreRestoreSession(t *testing.T) {
	s := state.NewAppState()
	store := &SessionStore{state: s}
	session := &Session{ID: "test-session"}
	if err := store.RestoreSession(session); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.SessionID != "test-session" {
		t.Errorf("expected session ID to be restored")
	}
}
