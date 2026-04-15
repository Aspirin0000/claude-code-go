package commands

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Aspirin0000/claude-code-go/internal/state"
)

// LoadCommand loads conversations from files or auto-saved sessions
type LoadCommand struct {
	*BaseCommand
}

// ConversationData represents persisted conversation data
type ConversationData struct {
	SessionID      string                 `json:"sessionId,omitempty"`
	ConversationID string                 `json:"conversationId,omitempty"`
	Messages       []ConversationMessage  `json:"messages"`
	CreatedAt      time.Time              `json:"createdAt,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// ConversationMessage represents a single saved message
type ConversationMessage struct {
	UUID      string    `json:"uuid"`
	Type      string    `json:"type"`
	Role      string    `json:"role,omitempty"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp,omitempty"`
}

// NewLoadCommand creates a new load command
func NewLoadCommand() *LoadCommand {
	return &LoadCommand{
		BaseCommand: NewBaseCommand(
			"load",
			"Load a conversation from a file or saved session",
			CategorySession,
		).WithAliases("import", "restore-file").
			WithHelp(`Usage: /load <filename|session-name>

Load conversation history from a file or from the auto-save sessions directory.
Supported file formats: JSON (.json) and Markdown (.md)

Arguments:
  filename/session-name  File path or saved session name to load

Aliases: /import, /restore-file

Examples:
  /load my-session.json
  /load backup.md
  /load project_2026-04-15_10-30

Tip:
  Use /sessions to list available auto-saved sessions.`),
	}
}

// Execute runs the load operation
func (c *LoadCommand) Execute(ctx context.Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("please provide a file path or session name: /load <filename|session-name>")
	}

	filename := args[0]

	// If a direct path does not exist, try loading from session storage.
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		if state.GlobalSessionStorage != nil {
			loadedState, loadErr := state.GlobalSessionStorage.LoadSession(filename)
			if loadErr == nil {
				state.GlobalState.SetMessages(loadedState.GetMessages())
				state.GlobalState.SetSessionID(loadedState.SessionID)
				state.GlobalState.SetCWD(loadedState.CWD)
				state.GlobalState.ProjectRoot = loadedState.ProjectRoot
				fmt.Printf("✓ Session '%s' loaded successfully.\n", filename)
				fmt.Printf("  - Messages: %d\n", len(loadedState.GetMessages()))
				fmt.Printf("  - Session ID: %s\n", loadedState.SessionID)
				return nil
			}
		}
		return fmt.Errorf("file or saved session not found: %s", filename)
	}

	// Choose loader based on file extension
	ext := strings.ToLower(filepath.Ext(filename))
	var data *ConversationData
	var err error

	switch ext {
	case ".json":
		data, err = c.loadJSON(filename)
	case ".md", ".markdown":
		data, err = c.loadMarkdown(filename)
	default:
		// Try to auto-detect the format
		data, err = c.autoDetectAndLoad(filename)
	}

	if err != nil {
		return fmt.Errorf("failed to load file: %w", err)
	}

	// Validate data
	if err := c.validateData(data); err != nil {
		return fmt.Errorf("data validation failed: %w", err)
	}

	// Restore session state
	if err := c.restoreSession(data); err != nil {
		return fmt.Errorf("failed to restore session: %w", err)
	}

	fmt.Printf("✓ Conversation loaded successfully from %s\n", filename)
	fmt.Printf("  - Session ID: %s\n", data.SessionID)
	fmt.Printf("  - Message count: %d\n", len(data.Messages))

	return nil
}

// loadJSON loads from a JSON file
func (c *LoadCommand) loadJSON(filename string) (*ConversationData, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var data ConversationData
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		return nil, fmt.Errorf("JSON parsing failed: %w", err)
	}

	return &data, nil
}

// loadMarkdown loads from a Markdown file
func (c *LoadCommand) loadMarkdown(filename string) (*ConversationData, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data := &ConversationData{
		Messages: make([]ConversationMessage, 0),
		Metadata: make(map[string]interface{}),
	}

	scanner := bufio.NewScanner(file)
	var currentMsg *ConversationMessage
	var contentLines []string

	for scanner.Scan() {
		line := scanner.Text()

		// Detect role headings (## User, ## Assistant, ## System)
		if strings.HasPrefix(line, "## ") {
			// Save the previous message
			if currentMsg != nil && len(contentLines) > 0 {
				currentMsg.Content = strings.Join(contentLines, "\n")
				data.Messages = append(data.Messages, *currentMsg)
			}

			// Create a new message
			role := strings.TrimPrefix(line, "## ")
			role = strings.TrimSpace(role)
			role = strings.ToLower(role)

			currentMsg = &ConversationMessage{
				UUID:      generateUUID(),
				Type:      role,
				Role:      role,
				Timestamp: time.Now(),
			}
			contentLines = []string{}
		} else if currentMsg != nil {
			// Accumulate content
			contentLines = append(contentLines, line)
		} else {
			// The file header may contain metadata
			if strings.HasPrefix(line, "# ") {
				data.SessionID = strings.TrimPrefix(line, "# ")
				data.SessionID = strings.TrimSpace(data.SessionID)
			}
		}
	}

	// Save the last message
	if currentMsg != nil && len(contentLines) > 0 {
		currentMsg.Content = strings.Join(contentLines, "\n")
		data.Messages = append(data.Messages, *currentMsg)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return data, nil
}

// autoDetectAndLoad automatically detects and loads a file
func (c *LoadCommand) autoDetectAndLoad(filename string) (*ConversationData, error) {
	// Read file content for format detection
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	// Try JSON parsing
	var jsonData ConversationData
	if err := json.Unmarshal(content, &jsonData); err == nil && len(jsonData.Messages) > 0 {
		return &jsonData, nil
	}

	// Try Markdown parsing
	return c.loadMarkdown(filename)
}

// validateData validates loaded data
func (c *LoadCommand) validateData(data *ConversationData) error {
	if data == nil {
		return fmt.Errorf("data is nil")
	}

	if len(data.Messages) == 0 {
		return fmt.Errorf("message list is empty")
	}

	// Validate each message
	for i, msg := range data.Messages {
		if msg.Type == "" {
			return fmt.Errorf("message %d is missing type", i)
		}
		if msg.Content == "" {
			return fmt.Errorf("message %d is missing content", i)
		}
	}

	return nil
}

// restoreSession restores session state
func (c *LoadCommand) restoreSession(data *ConversationData) error {
	// Clear current messages
	state.GlobalState.ClearMessages()

	// Restore session ID
	if data.SessionID != "" {
		state.GlobalState.SetSessionID(data.SessionID)
	}

	// Set conversation ID
	state.GlobalState.ConversationID = data.ConversationID

	// Restore messages
	for _, msg := range data.Messages {
		stateMsg := state.Message{
			UUID:    msg.UUID,
			Type:    msg.Type,
			Role:    msg.Role,
			Content: msg.Content,
		}
		state.GlobalState.AddMessage(stateMsg)
	}

	// Update turn count
	state.GlobalState.TurnCount = len(data.Messages) / 2

	return nil
}

func init() { Register(NewLoadCommand()) }
