// Package state provides application state management.
package state

import (
	"encoding/json"
	"sync"
	"time"
)

// ContentBlock represents a structured content block for multi-turn tool use.
type ContentBlock struct {
	Type      string          `json:"type"`
	Text      string          `json:"text,omitempty"`
	Content   string          `json:"content,omitempty"`
	ID        string          `json:"id,omitempty"`
	Name      string          `json:"name,omitempty"`
	Input     json.RawMessage `json:"input,omitempty"`
	ToolUseID string          `json:"tool_use_id,omitempty"`
}

// Message represents a conversation message.
type Message struct {
	UUID      string
	Type      string // user, assistant, system
	Role      string
	Content   string
	Blocks    []ContentBlock
	Timestamp time.Time `json:"timestamp"`
}

// Edit represents a file modification made by an AI tool.
type Edit struct {
	Timestamp     time.Time `json:"timestamp"`
	Tool          string    `json:"tool"`
	FilePath      string    `json:"file_path"`
	Operation     string    `json:"operation"`
	Description   string    `json:"description"`
	BeforeContent []byte    `json:"before_content,omitempty"`
	ExtraPath     string    `json:"extra_path,omitempty"`
}

// AppState holds the global application state.
type AppState struct {
	mu sync.RWMutex

	Messages       []Message
	Tools          []string
	SessionID      string
	CWD            string
	ProjectRoot    string
	ConversationID string
	TurnCount      int
	Edits          []Edit
}

// NewAppState creates a new application state.
func NewAppState() *AppState {
	return &AppState{
		Messages: make([]Message, 0),
		Tools:    make([]string, 0),
		Edits:    make([]Edit, 0),
	}
}

// AddMessage appends a message.
func (s *AppState) AddMessage(msg Message) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if msg.Timestamp.IsZero() {
		msg.Timestamp = time.Now()
	}
	s.Messages = append(s.Messages, msg)
}

// GetMessages returns a copy of all messages.
func (s *AppState) GetMessages() []Message {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]Message, len(s.Messages))
	copy(result, s.Messages)
	return result
}

// ClearMessages removes all messages.
func (s *AppState) ClearMessages() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Messages = make([]Message, 0)
}

// SetMessages replaces the message list.
func (s *AppState) SetMessages(messages []Message) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Messages = make([]Message, len(messages))
	copy(s.Messages, messages)
}

// SetSessionID sets the session ID.
func (s *AppState) SetSessionID(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.SessionID = id
}

// SetCWD sets the current working directory.
func (s *AppState) SetCWD(cwd string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.CWD = cwd
}

// IncrementTurn increments the turn counter.
func (s *AppState) IncrementTurn() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.TurnCount++
}

// AddEdit records a file modification.
func (s *AppState) AddEdit(edit Edit) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if edit.Timestamp.IsZero() {
		edit.Timestamp = time.Now()
	}
	s.Edits = append(s.Edits, edit)
}

// GetEdits returns a copy of all recorded edits.
func (s *AppState) GetEdits() []Edit {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]Edit, len(s.Edits))
	copy(result, s.Edits)
	return result
}

// ClearEdits removes all recorded edits.
func (s *AppState) ClearEdits() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Edits = make([]Edit, 0)
}

// SetEdits replaces the recorded edits list.
func (s *AppState) SetEdits(edits []Edit) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Edits = make([]Edit, len(edits))
	copy(s.Edits, edits)
}

// GlobalState is the global application state instance.
var GlobalState = NewAppState()
