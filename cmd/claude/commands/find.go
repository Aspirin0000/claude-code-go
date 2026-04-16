package commands

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// FindCommand finds files by name in the project
type FindCommand struct {
	*BaseCommand
}

// NewFindCommand creates the /find command
func NewFindCommand() *FindCommand {
	return &FindCommand{
		BaseCommand: NewBaseCommand(
			"find",
			"Find files by name in the current directory",
			CategoryTools,
		).WithAliases("fd").
			WithHelp(`Usage: /find <name>

Search for files and directories matching the given name substring.
Performs a recursive search starting from the current working directory.

Examples:
  /find main.go     - Find files containing "main.go" in their name
  /find test        - Find all files/directories with "test" in the name

Aliases: /fd`),
	}
}

// Execute runs the find command
func (c *FindCommand) Execute(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("missing search term; usage: /find <name>")
	}

	term := strings.ToLower(args[0])
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	var matches []string
	err = filepath.WalkDir(cwd, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // skip inaccessible paths
		}
		rel, _ := filepath.Rel(cwd, path)
		if rel == "." {
			return nil
		}
		// Skip common hidden directories
		if d.IsDir() {
			name := d.Name()
			if name == ".git" || name == "node_modules" || name == ".DS_Store" || name == "vendor" {
				return filepath.SkipDir
			}
		}
		if strings.Contains(strings.ToLower(d.Name()), term) {
			matches = append(matches, rel)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	fmt.Println()
	fmt.Printf("Find results for %q (%d matches):\n", term, len(matches))
	fmt.Println(strings.Repeat("─", 50))
	for _, m := range matches {
		fmt.Printf("  %s\n", m)
	}
	fmt.Println()
	return nil
}

func init() {
	Register(NewFindCommand())
}
