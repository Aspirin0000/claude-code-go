package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Aspirin0000/claude-code-go/internal/state"
)

// SaveCommand saves conversations to files
type SaveCommand struct {
	*BaseCommand
}

// NewSaveCommand creates a new save command
func NewSaveCommand() *SaveCommand {
	cmd := &SaveCommand{
		BaseCommand: NewBaseCommand("save", "Save current conversation to a file", CategorySession),
	}
	cmd.WithAliases("export", "backup")
	cmd.WithHelp(`Usage: /save [filename] [--format json|markdown]

Save the current conversation to a file.

Options:
  filename      Output file name (default: auto-generated)
  --format      Output format: json or markdown (default: json)

Examples:
  /save                          Save as JSON with auto-generated name
  /save my_chat.json             Save with specific name
  /save chat.md --format md      Save as Markdown
  /export backup.json            Use export alias

Note: If auto-save is enabled, sessions are automatically saved to the
sessions directory. Use /save for manual exports to custom locations.`)
	return cmd
}

// Execute runs the save command
func (s *SaveCommand) Execute(ctx context.Context, args []string) error {
	filename := ""
	format := "json"

	// Parse arguments
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--format", "-f":
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

	// Get conversation messages
	messages := state.GlobalState.GetMessages()
	if len(messages) == 0 {
		fmt.Println("No messages to save.")
		return nil
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

	// If filename not provided, use auto-save directory
	if filename == "" {
		if state.GlobalSessionStorage != nil {
			// Use session storage
			if err := state.GlobalSessionStorage.SaveSession(state.GlobalState, ""); err != nil {
				return fmt.Errorf("failed to auto-save session: %w", err)
			}
			fmt.Println("Session auto-saved successfully.")
			return nil
		}
		// Fallback to current directory
		filename = s.generateDefaultFilename()
	}

	// Add extension based on format
	extension := ".json"
	if format == "markdown" || format == "md" {
		extension = ".md"
	}
	if !strings.HasSuffix(filename, ".json") && !strings.HasSuffix(filename, ".md") {
		filename = filename + extension
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
	return fmt.Sprintf("conversation_%s.json", timestamp)
}

func (s *SaveCommand) exportAsJSON(messages []state.Message) ([]byte, error) {
	export := struct {
		Timestamp    string          `json:"timestamp"`
		SessionID    string          `json:"session_id"`
		CWD          string          `json:"cwd"`
		MessageCount int             `json:"message_count"`
		Messages     []state.Message `json:"messages"`
	}{
		Timestamp:    time.Now().Format(time.RFC3339),
		SessionID:    state.GlobalState.SessionID,
		CWD:          state.GlobalState.CWD,
		MessageCount: len(messages),
		Messages:     messages,
	}

	return json.MarshalIndent(export, "", "  ")
}

func (s *SaveCommand) exportAsMarkdown(messages []state.Message) ([]byte, error) {
	var md strings.Builder
	md.WriteString(fmt.Sprintf("# Conversation Export\n\n"))
	md.WriteString(fmt.Sprintf("**Exported:** %s\n\n", time.Now().Format("2006-01-02 15:04:05")))
	md.WriteString(fmt.Sprintf("**Session ID:** %s\n\n", state.GlobalState.SessionID))
	md.WriteString("---\n\n")

	for _, msg := range messages {
		role := msg.Role
		if role == "" {
			role = msg.Type
		}
		if role == "" {
			role = "unknown"
		}

		timestampStr := ""
		if !msg.Timestamp.IsZero() {
			timestampStr = msg.Timestamp.Format("2006-01-02 15:04:05")
		}

		switch role {
		case "user":
			if timestampStr != "" {
				md.WriteString(fmt.Sprintf("## User (%s)\n\n%s\n\n", timestampStr, msg.Content))
			} else {
				md.WriteString(fmt.Sprintf("## User\n\n%s\n\n", msg.Content))
			}
		case "assistant":
			if timestampStr != "" {
				md.WriteString(fmt.Sprintf("## Assistant (%s)\n\n%s\n\n", timestampStr, msg.Content))
			} else {
				md.WriteString(fmt.Sprintf("## Assistant\n\n%s\n\n", msg.Content))
			}
		case "system":
			if timestampStr != "" {
				md.WriteString(fmt.Sprintf("## System (%s)\n\n%s\n\n", timestampStr, msg.Content))
			} else {
				md.WriteString(fmt.Sprintf("## System\n\n%s\n\n", msg.Content))
			}
		default:
			if timestampStr != "" {
				md.WriteString(fmt.Sprintf("## %s (%s)\n\n%s\n\n", capitalize(role), timestampStr, msg.Content))
			} else {
				md.WriteString(fmt.Sprintf("## %s\n\n%s\n\n", capitalize(role), msg.Content))
			}
		}
	}

	return []byte(md.String()), nil
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

func init() { Register(NewSaveCommand()) }
