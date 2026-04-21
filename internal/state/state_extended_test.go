package state

import (
	"encoding/json"
	"testing"
	"time"
)

// ============================================================================
// ContentBlock Tests
// ============================================================================

func TestContentBlock(t *testing.T) {
	block := ContentBlock{
		Type:      "text",
		Text:      "Hello",
		Content:   "World",
		ID:        "block-1",
		Name:      "test-block",
		Input:     json.RawMessage(`{"key":"value"}`),
		ToolUseID: "tool-1",
	}

	if block.Type != "text" {
		t.Errorf("expected type 'text', got %q", block.Type)
	}
	if block.Text != "Hello" {
		t.Errorf("expected text 'Hello', got %q", block.Text)
	}
	if block.ID != "block-1" {
		t.Errorf("expected ID 'block-1', got %q", block.ID)
	}
}

// ============================================================================
// Message Tests
// ============================================================================

func TestMessage(t *testing.T) {
	msg := Message{
		UUID:      "uuid-123",
		Type:      "user",
		Role:      "user",
		Content:   "Hello",
		Blocks:    []ContentBlock{{Type: "text", Text: "Hello"}},
		Timestamp: time.Now(),
	}

	if msg.UUID != "uuid-123" {
		t.Errorf("expected UUID 'uuid-123', got %q", msg.UUID)
	}
	if msg.Type != "user" {
		t.Errorf("expected type 'user', got %q", msg.Type)
	}
	if len(msg.Blocks) != 1 {
		t.Errorf("expected 1 block, got %d", len(msg.Blocks))
	}
}

// ============================================================================
// Edit Tests
// ============================================================================

func TestEdit(t *testing.T) {
	edit := Edit{
		Timestamp:     time.Now(),
		Tool:          "file_write",
		FilePath:      "/tmp/test.txt",
		Operation:     "write",
		Description:   "Write test file",
		BeforeContent: []byte("old content"),
		ExtraPath:     "/tmp/backup.txt",
	}

	if edit.Tool != "file_write" {
		t.Errorf("expected tool 'file_write', got %q", edit.Tool)
	}
	if edit.FilePath != "/tmp/test.txt" {
		t.Errorf("expected file path '/tmp/test.txt', got %q", edit.FilePath)
	}
	if string(edit.BeforeContent) != "old content" {
		t.Errorf("expected before content 'old content', got %q", string(edit.BeforeContent))
	}
}

// ============================================================================
// AppState Initialization Tests
// ============================================================================

func TestNewAppState(t *testing.T) {
	s := NewAppState()
	if s == nil {
		t.Fatal("expected non-nil AppState")
	}
	if s.Messages == nil {
		t.Error("expected non-nil Messages")
	}
	if s.Tools == nil {
		t.Error("expected non-nil Tools")
	}
	if s.Edits == nil {
		t.Error("expected non-nil Edits")
	}
	if len(s.Messages) != 0 {
		t.Errorf("expected 0 messages, got %d", len(s.Messages))
	}
	if len(s.Tools) != 0 {
		t.Errorf("expected 0 tools, got %d", len(s.Tools))
	}
	if len(s.Edits) != 0 {
		t.Errorf("expected 0 edits, got %d", len(s.Edits))
	}
}

// ============================================================================
// AppState Message Tests (Extended)
// ============================================================================

func TestAppStateAddMessageWithBlocks(t *testing.T) {
	s := NewAppState()
	msg := Message{
		Type:   "assistant",
		Content: "Hello",
		Blocks: []ContentBlock{
			{Type: "text", Text: "Hello"},
			{Type: "tool_use", Name: "bash", ID: "tool-1"},
		},
	}
	s.AddMessage(msg)

	msgs := s.GetMessages()
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}
	if len(msgs[0].Blocks) != 2 {
		t.Errorf("expected 2 blocks, got %d", len(msgs[0].Blocks))
	}
}

func TestAppStateGetMessagesReturnsCopy(t *testing.T) {
	s := NewAppState()
	s.AddMessage(Message{Type: "user", Content: "hello"})

	msgs1 := s.GetMessages()
	msgs1[0].Content = "modified"

	msgs2 := s.GetMessages()
	if msgs2[0].Content != "hello" {
		t.Error("expected GetMessages to return a copy")
	}
}

func TestAppStateSetMessagesReplaces(t *testing.T) {
	s := NewAppState()
	s.AddMessage(Message{Type: "user", Content: "old"})
	s.SetMessages([]Message{
		{Type: "assistant", Content: "new"},
	})

	msgs := s.GetMessages()
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}
	if msgs[0].Content != "new" {
		t.Errorf("expected 'new', got %q", msgs[0].Content)
	}
}

func TestAppStateSetMessagesReturnsCopy(t *testing.T) {
	s := NewAppState()
	original := []Message{{Type: "user", Content: "test"}}
	s.SetMessages(original)

	original[0].Content = "modified"
	msgs := s.GetMessages()
	if msgs[0].Content != "test" {
		t.Error("expected SetMessages to store a copy")
	}
}

// ============================================================================
// AppState Session Tests
// ============================================================================

func TestAppStateSetSessionID(t *testing.T) {
	s := NewAppState()
	s.SetSessionID("session-123")
	if s.SessionID != "session-123" {
		t.Errorf("expected 'session-123', got %q", s.SessionID)
	}
}

func TestAppStateSetCWD(t *testing.T) {
	s := NewAppState()
	s.SetCWD("/home/user")
	if s.CWD != "/home/user" {
		t.Errorf("expected '/home/user', got %q", s.CWD)
	}
}

func TestAppStateIncrementTurnExtended(t *testing.T) {
	s := NewAppState()
	if s.TurnCount != 0 {
		t.Errorf("expected 0, got %d", s.TurnCount)
	}

	s.IncrementTurn()
	if s.TurnCount != 1 {
		t.Errorf("expected 1, got %d", s.TurnCount)
	}

	s.IncrementTurn()
	if s.TurnCount != 2 {
		t.Errorf("expected 2, got %d", s.TurnCount)
	}
}

// ============================================================================
// AppState Edit Tests
// ============================================================================

func TestAppStateAddEdit(t *testing.T) {
	s := NewAppState()
	edit := Edit{
		Tool:    "file_write",
		FilePath: "/tmp/test.txt",
	}
	s.AddEdit(edit)

	edits := s.GetEdits()
	if len(edits) != 1 {
		t.Fatalf("expected 1 edit, got %d", len(edits))
	}
	if edits[0].Tool != "file_write" {
		t.Errorf("expected tool 'file_write', got %q", edits[0].Tool)
	}
}

func TestAppStateAddEditAutoSetsTimestamp(t *testing.T) {
	s := NewAppState()
	edit := Edit{
		Tool:    "file_write",
		FilePath: "/tmp/test.txt",
	}
	s.AddEdit(edit)

	edits := s.GetEdits()
	if len(edits) != 1 {
		t.Fatalf("expected 1 edit, got %d", len(edits))
	}
	if edits[0].Timestamp.IsZero() {
		t.Error("expected timestamp to be auto-set")
	}
}

func TestAppStateAddEditPreservesTimestamp(t *testing.T) {
	s := NewAppState()
	customTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	edit := Edit{
		Timestamp: customTime,
		Tool:      "file_write",
		FilePath:  "/tmp/test.txt",
	}
	s.AddEdit(edit)

	edits := s.GetEdits()
	if !edits[0].Timestamp.Equal(customTime) {
		t.Errorf("expected timestamp %v, got %v", customTime, edits[0].Timestamp)
	}
}

func TestAppStateGetEditsReturnsCopy(t *testing.T) {
	s := NewAppState()
	s.AddEdit(Edit{Tool: "file_write", FilePath: "/tmp/test.txt"})

	edits1 := s.GetEdits()
	edits1[0].FilePath = "/tmp/modified.txt"

	edits2 := s.GetEdits()
	if edits2[0].FilePath != "/tmp/test.txt" {
		t.Error("expected GetEdits to return a copy")
	}
}

func TestAppStateClearEdits(t *testing.T) {
	s := NewAppState()
	s.AddEdit(Edit{Tool: "file_write", FilePath: "/tmp/test.txt"})
	s.ClearEdits()

	edits := s.GetEdits()
	if len(edits) != 0 {
		t.Errorf("expected 0 edits, got %d", len(edits))
	}
}

func TestAppStateSetEdits(t *testing.T) {
	s := NewAppState()
	s.SetEdits([]Edit{
		{Tool: "file_write", FilePath: "/tmp/test1.txt"},
		{Tool: "file_edit", FilePath: "/tmp/test2.txt"},
	})

	edits := s.GetEdits()
	if len(edits) != 2 {
		t.Fatalf("expected 2 edits, got %d", len(edits))
	}
	if edits[0].Tool != "file_write" {
		t.Errorf("expected 'file_write', got %q", edits[0].Tool)
	}
	if edits[1].Tool != "file_edit" {
		t.Errorf("expected 'file_edit', got %q", edits[1].Tool)
	}
}

func TestAppStateSetEditsReturnsCopy(t *testing.T) {
	s := NewAppState()
	original := []Edit{{Tool: "file_write", FilePath: "/tmp/test.txt"}}
	s.SetEdits(original)

	original[0].FilePath = "/tmp/modified.txt"
	edits := s.GetEdits()
	if edits[0].FilePath != "/tmp/test.txt" {
		t.Error("expected SetEdits to store a copy")
	}
}

// ============================================================================
// AppState Concurrency Tests (Extended)
// ============================================================================

func TestAppStateConcurrentEdits(t *testing.T) {
	s := NewAppState()
	done := make(chan bool, 2)

	go func() {
		for i := 0; i < 100; i++ {
			s.AddEdit(Edit{Tool: "file_write", FilePath: "/tmp/test.txt"})
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			s.GetEdits()
		}
		done <- true
	}()

	<-done
	<-done

	edits := s.GetEdits()
	if len(edits) != 100 {
		t.Errorf("expected 100 edits, got %d", len(edits))
	}
}

func TestAppStateConcurrentMixedOperations(t *testing.T) {
	s := NewAppState()
	done := make(chan bool, 3)

	go func() {
		for i := 0; i < 50; i++ {
			s.AddMessage(Message{Type: "user", Content: "msg"})
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 50; i++ {
			s.AddEdit(Edit{Tool: "file_write", FilePath: "/tmp/test.txt"})
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 50; i++ {
			s.IncrementTurn()
		}
		done <- true
	}()

	<-done
	<-done
	<-done

	msgs := s.GetMessages()
	if len(msgs) != 50 {
		t.Errorf("expected 50 messages, got %d", len(msgs))
	}

	edits := s.GetEdits()
	if len(edits) != 50 {
		t.Errorf("expected 50 edits, got %d", len(edits))
	}

	if s.TurnCount != 50 {
		t.Errorf("expected turn count 50, got %d", s.TurnCount)
	}
}

// ============================================================================
// GlobalState Tests
// ============================================================================

func TestGlobalState(t *testing.T) {
	if GlobalState == nil {
		t.Fatal("GlobalState should be initialized")
	}

	// Test that GlobalState is a valid AppState
	GlobalState.AddMessage(Message{Type: "user", Content: "test"})
	msgs := GlobalState.GetMessages()
	if len(msgs) < 1 {
		t.Error("expected at least 1 message in GlobalState")
	}
}

// ============================================================================
// AppState Other Fields Tests
// ============================================================================

func TestAppStateTools(t *testing.T) {
	s := NewAppState()
	// Tools field exists but has no methods yet
	// Just verify it exists
	if s.Tools == nil {
		t.Error("expected non-nil Tools")
	}
}

func TestAppStateProjectRoot(t *testing.T) {
	s := NewAppState()
	// ProjectRoot field exists but has no setter yet
	// Just verify it exists
	_ = s.ProjectRoot
}

func TestAppStateConversationID(t *testing.T) {
	s := NewAppState()
	// ConversationID field exists but has no setter yet
	// Just verify it exists
	_ = s.ConversationID
}
