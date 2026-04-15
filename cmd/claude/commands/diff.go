package commands

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// DiffCommand shows git diff or file differences
// Source: src/commands/diff/
type DiffCommand struct {
	*BaseCommand
}

// NewDiffCommand creates the diff command
func NewDiffCommand() *DiffCommand {
	return &DiffCommand{
		BaseCommand: NewBaseCommand(
			"diff",
			"Show code differences",
			CategoryTools,
		).
			WithHelp(`Show Git diff or file comparisons.

Usage:
  /diff              Show unstaged changes
  /diff --staged     Show staged changes
  /diff <file>       Show diff for a specific file

Examples:
  /diff              # View working directory changes
  /diff --staged     # View staged changes
  /diff main.go      # View changes for a specific file`),
	}
}

// Execute shows diff
func (c *DiffCommand) Execute(ctx context.Context, args []string) error {
	var cmdArgs []string

	// Parse arguments
	if len(args) > 0 {
		if args[0] == "--staged" || args[0] == "-s" {
			cmdArgs = []string{"diff", "--staged"}
		} else {
			// Diff specific file
			cmdArgs = append([]string{"diff"}, args...)
		}
	} else {
		// Default: show unstaged changes
		cmdArgs = []string{"diff"}
	}

	cmd := exec.CommandContext(ctx, "git", cmdArgs...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		if len(output) > 0 {
			// Git diff returns exit code 1 when there are differences
			fmt.Println(string(output))
			return nil
		}
		return fmt.Errorf("git diff failed: %w", err)
	}

	if len(output) == 0 {
		fmt.Println("✓ No differences")
		return nil
	}

	// Colorize output (simple version)
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			fmt.Printf("\033[32m%s\033[0m\n", line) // Green for additions
		} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
			fmt.Printf("\033[31m%s\033[0m\n", line) // Red for deletions
		} else {
			fmt.Println(line)
		}
	}

	return nil
}

func init() {
	Register(NewDiffCommand())
}
