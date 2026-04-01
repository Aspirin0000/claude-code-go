package commands

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Aspirin0000/claude-code-go/internal/state"
)

// CompactCommand compacts conversation history by summarizing old messages
type CompactCommand struct {
	*BaseCommand
}

// NewCompactCommand creates a new compact command
func NewCompactCommand() *CompactCommand {
	return &CompactCommand{
		BaseCommand: NewBaseCommand(
			"compact",
			"压缩会话历史，将旧消息总结为摘要",
			CategorySession,
		).WithAliases("compress", "summary").
			WithHelp(`使用: /compact [保留消息数]

压缩会话历史记录，将较早的消息总结为摘要。
这有助于管理长对话，减少上下文窗口的使用。

参数:
  [保留消息数]  可选，保留最近的消息数量 (默认: 10)

示例:
  /compact        # 保留最近10条消息，总结其余
  /compact 5      # 保留最近5条消息，总结其余
  /compact 20     # 保留最近20条消息，总结其余

别名: /compress, /summary

注意: 压缩后会话历史将被修改，无法恢复原始消息。`),
	}
}

// Execute executes the compact command
func (c *CompactCommand) Execute(ctx context.Context, args []string) error {
	// Parse argument: number of messages to keep
	keepCount := 10 // default
	if len(args) > 0 {
		n, err := strconv.Atoi(args[0])
		if err != nil || n < 1 {
			return fmt.Errorf("invalid argument: %s (expected positive integer)", args[0])
		}
		keepCount = n
	}

	// Get current messages
	messages := state.GlobalState.GetMessages()
	if len(messages) == 0 {
		fmt.Println("ℹ️  No messages to compact")
		return nil
	}

	// If we have fewer messages than keepCount, nothing to do
	if len(messages) <= keepCount {
		fmt.Printf("ℹ️  Only %d messages in history (threshold: %d), nothing to compact\n",
			len(messages), keepCount)
		return nil
	}

	// Calculate before token count (rough estimate: 1 token ≈ 4 characters)
	beforeTokens := estimateTokens(messages)

	// Split messages: older messages to summarize, recent to keep
	oldMessages := messages[:len(messages)-keepCount]
	recentMessages := messages[len(messages)-keepCount:]

	// Generate summary of old messages
	summary := generateSummary(oldMessages)

	// Create new message list: summary + recent messages
	newMessages := make([]state.Message, 0, len(recentMessages)+1)

	// Add summary as a system message
	summaryMsg := state.Message{
		UUID:    generateUUID(),
		Type:    "system",
		Role:    "system",
		Content: summary,
	}
	newMessages = append(newMessages, summaryMsg)
	newMessages = append(newMessages, recentMessages...)

	// Update state
	state.GlobalState.SetMessages(newMessages)

	// Calculate after token count
	afterTokens := estimateTokens(newMessages)
	savedTokens := beforeTokens - afterTokens
	savedPercent := float64(savedTokens) * 100 / float64(beforeTokens)

	// Display results
	fmt.Println("✅ Session compacted successfully")
	fmt.Println()
	fmt.Printf("📊 Statistics:\n")
	fmt.Printf("   Messages before: %d\n", len(messages))
	fmt.Printf("   Messages after:  %d (1 summary + %d recent)\n", len(newMessages), keepCount)
	fmt.Printf("   Tokens before:   ~%d\n", beforeTokens)
	fmt.Printf("   Tokens after:    ~%d\n", afterTokens)
	fmt.Printf("   Tokens saved:    ~%d (%.1f%%)\n", savedTokens, savedPercent)
	fmt.Println()
	fmt.Printf("📝 Summary:\n%s\n", summary)

	return nil
}

// estimateTokens estimates token count (rough approximation: 1 token ≈ 4 characters)
func estimateTokens(messages []state.Message) int {
	totalChars := 0
	for _, msg := range messages {
		totalChars += len(msg.Content)
		// Add overhead for role/type info
		totalChars += len(msg.Role) + len(msg.Type)
	}
	return totalChars / 4
}

// generateSummary generates a summary of old messages
func generateSummary(messages []state.Message) string {
	var sb strings.Builder

	sb.WriteString("## Previous Conversation Summary\n\n")
	sb.WriteString(fmt.Sprintf("*Generated on %s*\n\n", time.Now().Format("2006-01-02 15:04:05")))
	sb.WriteString(fmt.Sprintf("*Original message count: %d*\n\n", len(messages)))

	// Extract key topics/themes
	topics := extractTopics(messages)
	if len(topics) > 0 {
		sb.WriteString("### Topics Covered\n")
		for _, topic := range topics {
			sb.WriteString(fmt.Sprintf("- %s\n", topic))
		}
		sb.WriteString("\n")
	}

	// Summarize key decisions/actions
	decisions := extractDecisions(messages)
	if len(decisions) > 0 {
		sb.WriteString("### Key Decisions & Actions\n")
		for i, decision := range decisions {
			sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, decision))
		}
		sb.WriteString("\n")
	}

	// Context about files/tools used
	fileRefs := extractFileReferences(messages)
	if len(fileRefs) > 0 {
		sb.WriteString("### Files Referenced\n")
		for _, file := range fileRefs {
			sb.WriteString(fmt.Sprintf("- `%s`\n", file))
		}
		sb.WriteString("\n")
	}

	// Current context state
	sb.WriteString("### Current Context\n")
	sb.WriteString(fmt.Sprintf("- Working directory: %s\n", state.GlobalState.CWD))
	sb.WriteString(fmt.Sprintf("- Session ID: %s\n", state.GlobalState.SessionID))

	return sb.String()
}

// extractTopics extracts main topics from messages
func extractTopics(messages []state.Message) []string {
	topics := make([]string, 0)

	// Look for explicit topic indicators in user messages
	for _, msg := range messages {
		if msg.Role == "user" {
			content := strings.ToLower(msg.Content)

			// Check for file operations
			if strings.Contains(content, "file") || strings.Contains(content, "create") ||
				strings.Contains(content, "write") || strings.Contains(content, "edit") {
				if !contains(topics, "File operations") {
					topics = append(topics, "File operations")
				}
			}

			// Check for code-related content
			if strings.Contains(content, "code") || strings.Contains(content, "function") ||
				strings.Contains(content, "implement") || strings.Contains(content, "bug") {
				if !contains(topics, "Code development") {
					topics = append(topics, "Code development")
				}
			}

			// Check for configuration
			if strings.Contains(content, "config") || strings.Contains(content, "setting") {
				if !contains(topics, "Configuration") {
					topics = append(topics, "Configuration")
				}
			}

			// Check for search/grep operations
			if strings.Contains(content, "search") || strings.Contains(content, "find") ||
				strings.Contains(content, "grep") {
				if !contains(topics, "Search operations") {
					topics = append(topics, "Search operations")
				}
			}
		}
	}

	return topics
}

// extractDecisions extracts key decisions from messages
func extractDecisions(messages []state.Message) []string {
	decisions := make([]string, 0)

	for _, msg := range messages {
		if msg.Role == "assistant" {
			content := msg.Content

			// Look for completion indicators
			if strings.Contains(content, "Created") || strings.Contains(content, "created") {
				lines := strings.Split(content, "\n")
				for _, line := range lines {
					line = strings.TrimSpace(line)
					if strings.Contains(line, "Created") || strings.Contains(line, "created") {
						decisions = append(decisions, line)
						break
					}
				}
			}

			// Look for file modifications
			if strings.Contains(content, "Modified") || strings.Contains(content, "modified") ||
				strings.Contains(content, "Updated") || strings.Contains(content, "updated") {
				lines := strings.Split(content, "\n")
				for _, line := range lines {
					line = strings.TrimSpace(line)
					if strings.Contains(line, "Modified") || strings.Contains(line, "Updated") {
						decisions = append(decisions, line)
						break
					}
				}
			}
		}
	}

	return decisions
}

// extractFileReferences extracts file paths mentioned in messages
func extractFileReferences(messages []state.Message) []string {
	files := make(map[string]bool)

	for _, msg := range messages {
		content := msg.Content

		// Look for file paths (simple heuristic)
		words := strings.Fields(content)
		for _, word := range words {
			word = strings.Trim(word, "`\"'()[]{},;:!?")

			// Check for common file extensions
			if strings.Contains(word, ".") {
				extensions := []string{".go", ".js", ".ts", ".py", ".java", ".md", ".json",
					".yaml", ".yml", ".toml", ".mod", ".sum", ".txt", ".sh", ".dockerfile"}
				for _, ext := range extensions {
					if strings.HasSuffix(strings.ToLower(word), ext) {
						// Clean up the path
						if !strings.HasPrefix(word, "http") && len(word) < 200 {
							files[word] = true
						}
						break
					}
				}
			}
		}
	}

	// Convert to slice
	result := make([]string, 0, len(files))
	for file := range files {
		result = append(result, file)
	}

	return result
}

// contains checks if a string slice contains a value
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// generateUUID generates a simple UUID-like string
func generateUUID() string {
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), time.Now().Unix())
}
