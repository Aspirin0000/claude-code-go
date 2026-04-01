package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Conversation defines the interface for accessing conversation data
type Conversation interface {
	GetMessages() []Message
}

// Message represents a single message in a conversation
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// SaveCommand saves conversations to files
type SaveCommand struct {
	*BaseCommand
	conversation Conversation
}

// NewSaveCommand creates a new save command
func NewSaveCommand(conv Conversation) *SaveCommand {
	cmd := &SaveCommand{
		BaseCommand:  NewBaseCommand("save", "Save current conversation to a file", CategorySession),
		conversation: conv,
	}
	cmd.WithAliases("export", "backup")
	cmd.WithHelp(`/save - Save conversation to file

Usage: /save [filename] [--format json|markdown]

Options:
  filename      Output file name (default: conversation_<timestamp>.json)
  --format      Output format: json or markdown (default: json)

Examples:
  /save                          Save as JSON with timestamp
  /save my_chat.json             Save with specific name
  /save chat.md --format md      Save as Markdown
  /export backup.json            Use export alias`)
	return cmd
}

// Execute runs the save command
func (s *SaveCommand) Execute(ctx context.Context, args []string) error {
	filename := s.generateDefaultFilename()
	format := "json"

	// Parse arguments
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--format":
			if i+1 < len(args) {
				format = args[i+1]
				i++
			}
		default:
			if len(args[i]) == 0 || args[i][0] != '-' {
				filename = args[i]
			}
		}
	}

	// Add extension based on format
	extension := ".json"
	if format == "markdown" || format == "md" {
		extension = ".md"
	}
	if !hasExtension(filename, ".json") && !hasExtension(filename, ".md") {
		filename = filename + extension
	}

	// Get conversation messages
	messages := s.conversation.GetMessages()
	if messages == nil {
		messages = []Message{}
	}

	var content []byte
	var err error

	switch format {
	case "markdown", "md":
		content, err = s.exportAsMarkdown(messages)
	default:
		content, err = s.exportAsJSON(messages)
	}

	if err != nil {
		return fmt.Errorf("failed to format conversation: %w", err)
	}

	// Create directory if needed
	dir := filepath.Dir(filename)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	// Write file
	if err := os.WriteFile(filename, content, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	fmt.Printf("Conversation saved to: %s\n", filename)
	return nil
}

func (s *SaveCommand) generateDefaultFilename() string {
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	return fmt.Sprintf("conversation_%s", timestamp)
}

func (s *SaveCommand) exportAsJSON(messages []Message) ([]byte, error) {
	export := struct {
		Timestamp string    `json:"timestamp"`
		Messages  []Message `json:"messages"`
	}{
		Timestamp: time.Now().Format(time.RFC3339),
		Messages:  messages,
	}

	return json.MarshalIndent(export, "", "  ")
}

func (s *SaveCommand) exportAsMarkdown(messages []Message) ([]byte, error) {
	var md string
	md += fmt.Sprintf("# Conversation Export\n\n")
	md += fmt.Sprintf("**Exported:** %s\n\n", time.Now().Format("2006-01-02 15:04:05"))
	md += "---\n\n"

	for _, msg := range messages {
		role := msg.Role
		if role == "" {
			role = "unknown"
		}

		switch role {
		case "user":
			md += fmt.Sprintf("## User\n\n%s\n\n", msg.Content)
		case "assistant":
			md += fmt.Sprintf("## Assistant\n\n%s\n\n", msg.Content)
		case "system":
			md += fmt.Sprintf("## System\n\n%s\n\n", msg.Content)
		default:
			md += fmt.Sprintf("## %s\n\n%s\n\n", capitalize(role), msg.Content)
		}
	}

	return []byte(md), nil
}

// Helper functions
func hasExtension(filename, ext string) bool {
	if len(filename) < len(ext) {
		return false
	}
	return filename[len(filename)-len(ext):] == ext
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	first := s[0]
	if first >= 'a' && first <= 'z' {
		first = first - 'a' + 'A'
	}
	return string(first) + s[1:]
}
