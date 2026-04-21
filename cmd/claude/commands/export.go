package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Aspirin0000/claude-code-go/internal/state"
)

// ExportCommand exports the current conversation to various formats
type ExportCommand struct {
	*BaseCommand
}

// NewExportCommand creates the /export command
func NewExportCommand() *ExportCommand {
	return &ExportCommand{
		BaseCommand: NewBaseCommand(
			"export",
			"Export conversation to Markdown, JSON, or plain text",
			CategorySession,
		).WithAliases("dump").
			WithHelp(`Usage: /export [format] [filename]

Export the current conversation to a file.

Formats:
  markdown (md)  - Markdown format with headers and code blocks
  json           - JSON format with full message structure
  text (txt)     - Plain text format (default)

Examples:
  /export                    Export as text to conversation.txt
  /export markdown           Export as Markdown to conversation.md
  /export json my-chat.json  Export as JSON to my-chat.json

Aliases: /dump`),
	}
}

// Execute runs the export command
func (c *ExportCommand) Execute(ctx context.Context, args []string) error {
	format := "text"
	filename := ""

	if len(args) > 0 {
		format = strings.ToLower(args[0])
	}
	if len(args) > 1 {
		filename = args[1]
	}

	// Validate format
	switch format {
	case "markdown", "md":
		format = "markdown"
		if filename == "" {
			filename = "conversation.md"
		}
	case "json":
		if filename == "" {
			filename = "conversation.json"
		}
	case "text", "txt":
		format = "text"
		if filename == "" {
			filename = "conversation.txt"
		}
	default:
		return fmt.Errorf("unknown format %q; supported: markdown, json, text", format)
	}

	messages := state.GlobalState.GetMessages()
	if len(messages) == 0 {
		fmt.Println("ℹ️  No messages to export.")
		return nil
	}

	var content string
	var err error

	switch format {
	case "markdown":
		content, err = c.exportMarkdown(messages)
	case "json":
		content, err = c.exportJSON(messages)
	case "text":
		content, err = c.exportText(messages)
	}

	if err != nil {
		return fmt.Errorf("export failed: %w", err)
	}

	if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	fmt.Printf("✅ Exported %d message(s) to %s (%s format)\n", len(messages), filename, format)
	return nil
}

func (c *ExportCommand) exportMarkdown(messages []state.Message) (string, error) {
	var b strings.Builder
	b.WriteString("# Conversation Export\n\n")
	b.WriteString(fmt.Sprintf("*Exported on %s*\n\n", time.Now().Format("2006-01-02 15:04:05")))
	b.WriteString("---\n\n")

	for i, msg := range messages {
		b.WriteString(fmt.Sprintf("## Message %d\n\n", i+1))
		
		role := msg.Role
		if role == "user" {
			role = "User"
		} else if role == "assistant" {
			role = "Assistant"
		} else if role == "system" {
			role = "System"
		}
		
		b.WriteString(fmt.Sprintf("**Role:** %s\n", role))
		if !msg.Timestamp.IsZero() {
			b.WriteString(fmt.Sprintf("**Time:** %s\n", msg.Timestamp.Format("2006-01-02 15:04:05")))
		}
		b.WriteString("\n")
		
		if msg.Content != "" {
			b.WriteString(msg.Content)
			b.WriteString("\n\n")
		}
		
		if len(msg.Blocks) > 0 {
			b.WriteString("**Blocks:**\n\n")
			for _, block := range msg.Blocks {
				b.WriteString(fmt.Sprintf("- Type: `%s`\n", block.Type))
				if block.Name != "" {
					b.WriteString(fmt.Sprintf("  - Name: %s\n", block.Name))
				}
				if block.Text != "" {
					b.WriteString(fmt.Sprintf("  - Text: %s\n", block.Text))
				}
			}
			b.WriteString("\n")
		}
		
		b.WriteString("---\n\n")
	}

	return b.String(), nil
}

func (c *ExportCommand) exportJSON(messages []state.Message) (string, error) {
	type exportMessage struct {
		Role      string    `json:"role"`
		Content   string    `json:"content"`
		Timestamp time.Time `json:"timestamp,omitempty"`
		Blocks    []struct {
			Type string `json:"type"`
			Text string `json:"text,omitempty"`
			Name string `json:"name,omitempty"`
		} `json:"blocks,omitempty"`
	}

	exported := make([]exportMessage, len(messages))
	for i, msg := range messages {
		exported[i] = exportMessage{
			Role:      msg.Role,
			Content:   msg.Content,
			Timestamp: msg.Timestamp,
		}
		if len(msg.Blocks) > 0 {
			exported[i].Blocks = make([]struct {
				Type string `json:"type"`
				Text string `json:"text,omitempty"`
				Name string `json:"name,omitempty"`
			}, len(msg.Blocks))
			for j, block := range msg.Blocks {
				exported[i].Blocks[j] = struct {
					Type string `json:"type"`
					Text string `json:"text,omitempty"`
					Name string `json:"name,omitempty"`
				}{
					Type: block.Type,
					Text: block.Text,
					Name: block.Name,
				}
			}
		}
	}

	data, err := json.MarshalIndent(exported, "", "  ")
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (c *ExportCommand) exportText(messages []state.Message) (string, error) {
	var b strings.Builder
	b.WriteString("Conversation Export\n")
	b.WriteString("===================\n\n")
	b.WriteString(fmt.Sprintf("Exported on: %s\n\n", time.Now().Format("2006-01-02 15:04:05")))
	b.WriteString(fmt.Sprintf("Total messages: %d\n\n", len(messages)))
	b.WriteString(strings.Repeat("-", 50))
	b.WriteString("\n\n")

	for i, msg := range messages {
		role := msg.Role
		if role == "user" {
			role = "You"
		} else if role == "assistant" {
			role = "Claude"
		}
		
		timestamp := ""
		if !msg.Timestamp.IsZero() {
			timestamp = fmt.Sprintf(" [%s]", msg.Timestamp.Format("15:04:05"))
		}
		
		b.WriteString(fmt.Sprintf("[%d] %s%s\n", i+1, role, timestamp))
		b.WriteString(strings.Repeat("-", 30))
		b.WriteString("\n")
		
		if msg.Content != "" {
			b.WriteString(msg.Content)
			b.WriteString("\n")
		}
		
		if len(msg.Blocks) > 0 {
			for _, block := range msg.Blocks {
				if block.Type == "tool_use" {
					b.WriteString(fmt.Sprintf("\n[Tool: %s]\n", block.Name))
				}
			}
		}
		
		b.WriteString("\n")
	}

	return b.String(), nil
}

func init() {
	Register(NewExportCommand())
}
