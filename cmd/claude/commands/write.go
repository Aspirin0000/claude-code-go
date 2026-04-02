// Package commands provides file operation commands
// Source: src/commands/files/
package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// WriteCommand writes content to a file
// Usage: /write <file> <content>
type WriteCommand struct {
	*BaseCommand
}

// NewWriteCommand creates the /write command
func NewWriteCommand() *WriteCommand {
	return &WriteCommand{
		BaseCommand: NewBaseCommand(
			"write",
			"Write content to a file",
			CategoryFiles,
		).WithAliases("w", "create").
			WithHelp(`Write content to a file.

Usage:
  /write <file> <content>     - Write content to file
  /write <file>               - Interactive mode (read from stdin)
  /write --append <file>      - Append to file

Examples:
  /write hello.txt "Hello World"
  /write config.json '{"key": "value"}'
  echo "content" | /write file.txt

Safety:
  - Creates parent directories if needed
  - Overwrites existing files (use --append to append)
  - Backs up existing files with .backup suffix`),
	}
}

// Execute writes content to file
func (c *WriteCommand) Execute(ctx context.Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: /write <file> [content]")
	}

	appendMode := false
	if args[0] == "--append" || args[0] == "-a" {
		appendMode = true
		args = args[1:]
		if len(args) < 1 {
			return fmt.Errorf("usage: /write --append <file> [content]")
		}
	}

	filePath := args[0]
	content := ""
	if len(args) > 1 {
		content = strings.Join(args[1:], " ")
	}

	// Expand ~ to home directory
	if strings.HasPrefix(filePath, "~") {
		home, _ := os.UserHomeDir()
		filePath = filepath.Join(home, filePath[1:])
	}

	// Create parent directories
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Backup existing file
	if _, err := os.Stat(filePath); err == nil && !appendMode {
		backupPath := filePath + ".backup"
		os.Rename(filePath, backupPath)
		fmt.Printf("Backed up existing file to: %s\n", backupPath)
	}

	// Write content
	flags := os.O_CREATE | os.O_WRONLY
	if appendMode {
		flags |= os.O_APPEND
	} else {
		flags |= os.O_TRUNC
	}

	file, err := os.OpenFile(filePath, flags, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	if content != "" {
		if _, err := file.WriteString(content); err != nil {
			return fmt.Errorf("failed to write content: %w", err)
		}
		if !strings.HasSuffix(content, "\n") {
			file.WriteString("\n")
		}
	}

	action := "written"
	if appendMode {
		action = "appended"
	}
	fmt.Printf("✓ File %s: %s\n", action, filePath)
	return nil
}

func init() {
	Register(NewWriteCommand())
}
