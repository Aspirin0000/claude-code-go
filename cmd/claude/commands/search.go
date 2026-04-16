package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Aspirin0000/claude-code-go/internal/state"
)

// SearchCommand searches the current conversation history
type SearchCommand struct {
	*BaseCommand
}

// NewSearchCommand creates the search command
func NewSearchCommand() *SearchCommand {
	return &SearchCommand{
		BaseCommand: NewBaseCommand(
			"search",
			"Search conversation history for keywords",
			CategorySession,
		).WithAliases("grep-history").
			WithHelp(`Usage: /search <keyword>

Search through the current conversation history for messages containing the given keyword.

Examples:
  /search docker
  /search function
  /grep-history "error message"`),
	}
}

// Execute runs the search command
func (c *SearchCommand) Execute(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: /search <keyword>")
	}

	keyword := strings.Join(args, " ")
	keywordLower := strings.ToLower(keyword)
	messages := state.GlobalState.GetMessages()

	var matches []struct {
		Index     int
		Role      string
		Content   string
		Timestamp time.Time
	}

	for i, msg := range messages {
		content := msg.Content
		if content == "" && len(msg.Blocks) > 0 {
			for _, block := range msg.Blocks {
				if block.Content != "" {
					content += block.Content + " "
				}
				if block.Text != "" {
					content += block.Text + " "
				}
			}
		}
		if strings.Contains(strings.ToLower(content), keywordLower) {
			role := msg.Role
			if role == "" {
				role = msg.Type
			}
			matches = append(matches, struct {
				Index     int
				Role      string
				Content   string
				Timestamp time.Time
			}{
				Index:     i + 1,
				Role:      role,
				Content:   content,
				Timestamp: msg.Timestamp,
			})
		}
	}

	if len(matches) == 0 {
		fmt.Printf("No messages found containing '%s'.\n", keyword)
		return nil
	}

	fmt.Printf("\n🔍 Found %d message(s) containing '%s':\n\n", len(matches), keyword)
	for _, m := range matches {
		roleIcon := "💬"
		switch m.Role {
		case "user":
			roleIcon = "👤"
		case "assistant":
			roleIcon = "🤖"
		case "system":
			roleIcon = "ℹ️ "
		}
		preview := m.Content
		if len(preview) > 120 {
			preview = preview[:117] + "..."
		}
		timeStr := ""
		if !m.Timestamp.IsZero() {
			timeStr = "(" + m.Timestamp.Format("15:04") + ") "
		}
		fmt.Printf("%s [#%d %s] %s%s\n", roleIcon, m.Index, capitalize(m.Role), timeStr, preview)
	}
	fmt.Println()
	return nil
}

func init() { Register(NewSearchCommand()) }
