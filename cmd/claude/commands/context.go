package commands

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/Aspirin0000/claude-code-go/internal/bootstrap"
	"github.com/Aspirin0000/claude-code-go/internal/state"
)

// ContextCommand shows the current conversation context
type ContextCommand struct {
	*BaseCommand
}

// NewContextCommand creates the /context command
func NewContextCommand() *ContextCommand {
	return &ContextCommand{
		BaseCommand: NewBaseCommand(
			"context",
			"Show the current conversation context sent to the AI",
			CategorySession,
		).WithHelp(`Usage: /context

Display the current conversation context that is sent to the AI model.
This includes:
  - Current working directory
  - Git repository status
  - Current date and time
  - Message count and model info
  - Available tools summary

Useful for understanding what the AI knows about your current session.`),
	}
}

// Execute runs the context command
func (c *ContextCommand) Execute(ctx context.Context, args []string) error {
	fmt.Println()
	fmt.Println("╔════════════════════════════════════════════════════════╗")
	fmt.Println("║           Current Conversation Context                 ║")
	fmt.Println("╚════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Session info
	fmt.Printf("📅 Date/Time:   %s\n", time.Now().Format("2006-01-02 15:04:05 MST"))
	fmt.Printf("📁 Working Dir: %s\n", bootstrap.GetCwdState())
	fmt.Printf("🔧 Project Root: %s\n", bootstrap.GetProjectRoot())
	fmt.Printf("🆔 Session ID:  %s\n", bootstrap.GetSessionId())
	fmt.Println()

	// Model info
	modelInfo := getCurrentModelInfo()
	fmt.Printf("🤖 Model:       %s\n", modelInfo.Name)
	if modelInfo.Source != "" {
		fmt.Printf("   Source:      %s\n", modelInfo.Source)
	}
	fmt.Println()

	// Git info
	gitInfo := c.getGitContext()
	if gitInfo != "" {
		fmt.Println("🔀 Git Context:")
		fmt.Println(gitInfo)
		fmt.Println()
	}

	// Conversation stats
	messages := state.GlobalState.GetMessages()
	fmt.Printf("💬 Conversation:\n")
	fmt.Printf("   Messages:    %d\n", len(messages))
	fmt.Printf("   Turns:       %d\n", state.GlobalState.TurnCount)
	fmt.Println()

	// Tools summary
	fmt.Printf("🛠️  Tools available:\n")
	fmt.Printf("   Core tools + MCP tools are active\n")
	fmt.Printf("   Use /tools to see the full list\n")
	fmt.Println()

	// Files context (CLAUDE.md)
	claudeMdFiles := c.findClaudeMdFiles()
	if len(claudeMdFiles) > 0 {
		fmt.Printf("📄 CLAUDE.md files found:\n")
		for _, f := range claudeMdFiles {
			fmt.Printf("   • %s\n", f)
		}
		fmt.Println()
	}

	return nil
}

func (c *ContextCommand) getGitContext() string {
	cwd := bootstrap.GetCwdState()
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Dir = cwd
	if err := cmd.Run(); err != nil {
		return ""
	}

	var parts []string

	branchCmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	branchCmd.Dir = cwd
	if out, err := branchCmd.Output(); err == nil {
		parts = append(parts, fmt.Sprintf("   Branch: %s", strings.TrimSpace(string(out))))
	}

	statusCmd := exec.Command("git", "status", "--porcelain")
	statusCmd.Dir = cwd
	if out, err := statusCmd.Output(); err == nil {
		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		modified, staged, untracked := 0, 0, 0
		for _, line := range lines {
			if line == "" {
				continue
			}
			if len(line) >= 2 {
				indexStatus := line[0]
				workTreeStatus := line[1]
				if indexStatus != ' ' && indexStatus != '?' {
					staged++
				}
				if workTreeStatus != ' ' {
					modified++
				}
				if indexStatus == '?' {
					untracked++
				}
			}
		}
		if modified > 0 || staged > 0 || untracked > 0 {
			parts = append(parts, fmt.Sprintf("   Changes: %d modified, %d staged, %d untracked", modified, staged, untracked))
		} else {
			parts = append(parts, "   Working tree clean")
		}
	}

	return strings.Join(parts, "\n")
}

func (c *ContextCommand) findClaudeMdFiles() []string {
	var results []string
	cwd := bootstrap.GetCwdState()
	root := bootstrap.GetProjectRoot()

	for _, dir := range []string{cwd, root} {
		if dir == "" {
			continue
		}
		path := filepath.Join(dir, "CLAUDE.md")
		if _, err := os.Stat(path); err == nil {
			results = append(results, path)
		}
	}

	// Also check parent directories up to root
	if root != "" && strings.HasPrefix(cwd, root) {
		for dir := cwd; dir != root && strings.HasPrefix(dir, root); dir = filepath.Dir(dir) {
			path := filepath.Join(dir, "CLAUDE.md")
			if _, err := os.Stat(path); err == nil {
				alreadyFound := false
				for _, existing := range results {
					if existing == path {
						alreadyFound = true
						break
					}
				}
				if !alreadyFound {
					results = append(results, path)
				}
			}
		}
	}

	return results
}

func init() {
	Register(NewContextCommand())
}
