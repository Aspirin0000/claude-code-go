package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

// GlobCommand finds files by glob pattern
type GlobCommand struct {
	*BaseCommand
}

// NewGlobCommand creates a new glob command
func NewGlobCommand() *GlobCommand {
	return &GlobCommand{
		BaseCommand: NewBaseCommand(
			"glob",
			"Find files by glob pattern",
			CategoryTools,
		).WithAliases("find", "files").
			WithHelp(`Usage: /glob <pattern>
       /glob <pattern> [directory]

Find files matching a glob pattern. Supports ** wildcards for recursive search.

Arguments:
  <pattern>    Glob pattern to match (e.g., "*.go", "**/*.json")
  [directory]  Directory to search (default: current directory)

Pattern Syntax:
  *       Match any sequence of non-separator characters
  **      Match any sequence of characters including separators (recursive)
  ?       Match any single character
  [...]   Match any character in the bracket

Examples:
  /glob "*.go"                    Find all Go files in current directory
  /glob "**/*.go"                 Find all Go files recursively
  /glob "*.md" ./docs             Find all markdown files in ./docs
  /glob "**/*test*.go"            Find all test files recursively
  /glob "src/**/*.ts"             Find all TypeScript files in src/

Aliases: /find, /files`),
	}
}

// Execute runs the glob command
func (c *GlobCommand) Execute(ctx context.Context, args []string) error {
	if len(args) == 0 {
		fmt.Println(c.Help())
		return nil
	}

	pattern := args[0]
	searchDir := "."

	if len(args) > 1 {
		searchDir = args[1]
	}

	// Validate directory
	info, err := os.Stat(searchDir)
	if err != nil {
		return fmt.Errorf("cannot access directory '%s': %w", searchDir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("'%s' is not a directory", searchDir)
	}

	// Check for ** wildcard (recursive)
	if strings.Contains(pattern, "**") {
		return c.globRecursive(pattern, searchDir)
	}

	// Simple glob pattern
	return c.globSimple(pattern, searchDir)
}

// globSimple handles simple glob patterns without **
func (c *GlobCommand) globSimple(pattern, dir string) error {
	fullPattern := filepath.Join(dir, pattern)
	matches, err := filepath.Glob(fullPattern)
	if err != nil {
		return fmt.Errorf("invalid pattern '%s': %w", pattern, err)
	}

	if len(matches) == 0 {
		fmt.Printf("No files match pattern '%s'\n", pattern)
		return nil
	}

	// Sort for consistent output
	sort.Strings(matches)

	fmt.Printf("\n%sFound %d file(s) matching '%s':%s\n", "\033[1m", len(matches), pattern, "\033[0m")
	fmt.Println(strings.Repeat("-", 50))

	for _, match := range matches {
		// Get relative path if in current directory
		rel, err := filepath.Rel(dir, match)
		if err != nil {
			rel = match
		}

		// Get file info
		info, err := os.Stat(match)
		if err != nil {
			fmt.Printf("  %s\n", rel)
			continue
		}

		if info.IsDir() {
			fmt.Printf("  %s%s/%s\n", "\033[34m", rel, "\033[0m")
		} else {
			fmt.Printf("  %s\n", rel)
		}
	}

	fmt.Println()
	return nil
}

// globRecursive handles ** wildcard patterns
func (c *GlobCommand) globRecursive(pattern, dir string) error {
	// Parse the pattern to extract the base directory and file pattern
	parts := strings.Split(pattern, "**")
	if len(parts) != 2 {
		return fmt.Errorf("invalid recursive pattern")
	}

	prefix := strings.TrimSuffix(parts[0], "/")
	suffix := strings.TrimPrefix(parts[1], "/")

	// Build the starting directory
	startDir := dir
	if prefix != "" {
		startDir = filepath.Join(dir, prefix)
	}

	// Validate start directory
	info, err := os.Stat(startDir)
	if err != nil {
		return fmt.Errorf("cannot access directory '%s': %w", startDir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("'%s' is not a directory", startDir)
	}

	// Collect all matches
	var matches []string
	var mu sync.Mutex
	var wg sync.WaitGroup

	errChan := make(chan error, 1)

	// Walk the directory tree
	wg.Add(1)
	go func() {
		defer wg.Done()
		c.walkAndMatch(startDir, suffix, &matches, &mu, errChan)
	}()

	wg.Wait()
	close(errChan)

	if err := <-errChan; err != nil {
		return err
	}

	if len(matches) == 0 {
		fmt.Printf("No files match pattern '%s'\n", pattern)
		return nil
	}

	// Sort for consistent output
	sort.Strings(matches)

	fmt.Printf("\n%sFound %d file(s) matching '%s':%s\n", "\033[1m", len(matches), pattern, "\033[0m")
	fmt.Println(strings.Repeat("-", 50))

	for _, match := range matches {
		// Get relative path from original dir
		rel, err := filepath.Rel(dir, match)
		if err != nil {
			rel = match
		}

		info, err := os.Stat(match)
		if err != nil {
			fmt.Printf("  %s\n", rel)
			continue
		}

		if info.IsDir() {
			fmt.Printf("  %s%s/%s\n", "\033[34m", rel, "\033[0m")
		} else {
			fmt.Printf("  %s\n", rel)
		}
	}

	fmt.Println()
	return nil
}

// walkAndMatch walks directory and matches files
func (c *GlobCommand) walkAndMatch(dir, pattern string, matches *[]string, mu *sync.Mutex, errChan chan<- error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		select {
		case errChan <- fmt.Errorf("error reading directory '%s': %w", dir, err):
		default:
		}
		return
	}

	for _, entry := range entries {
		fullPath := filepath.Join(dir, entry.Name())

		if entry.IsDir() {
			// Skip hidden directories and common non-source directories
			name := entry.Name()
			if strings.HasPrefix(name, ".") ||
				name == "node_modules" ||
				name == "vendor" ||
				name == "dist" ||
				name == "build" ||
				name == "__pycache__" ||
				name == ".git" {
				continue
			}

			// Recurse into subdirectory
			c.walkAndMatch(fullPath, pattern, matches, mu, errChan)
		} else {
			// Check if file matches pattern
			matched, err := filepath.Match(pattern, entry.Name())
			if err != nil {
				continue
			}
			if matched {
				mu.Lock()
				*matches = append(*matches, fullPath)
				mu.Unlock()
			}
		}
	}
}

func init() {
	Register(NewGlobCommand())
}
