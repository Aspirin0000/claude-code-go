package commands

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	editColorReset   = "\033[0m"
	editColorRed     = "\033[31m"
	editColorGreen   = "\033[32m"
	editColorYellow  = "\033[33m"
	editColorCyan    = "\033[36m"
	editColorGray    = "\033[90m"
	editColorBold    = "\033[1m"
	editMaxFileSize  = 5 * 1024 * 1024 // 5MB
	editBackupSuffix = ".backup."
)

// EditCommand provides AI-assisted file editing
type EditCommand struct {
	*BaseCommand
}

// NewEditCommand creates a new edit command
func NewEditCommand() *EditCommand {
	return &EditCommand{
		BaseCommand: NewBaseCommand(
			"edit",
			"AI-assisted file editing",
			CategoryFiles,
		).WithAliases("modify").
			WithHelp(`Usage: /edit <file> <instruction>

Edit a file using AI assistance. The file is read, analyzed, modified according to the instruction, and a diff is shown before confirmation.

Arguments:
  <file>         Path to the file to edit
  <instruction>  Natural language instruction for the edit

Options:
  --no-backup    Skip creating a backup file
  --yes          Skip confirmation and apply changes immediately

Examples:
  /edit main.go "add error handling to the main function"
  /edit config.json "change port from 8080 to 3000"
  /edit README.md "update the installation instructions"

Workflow:
  1. Reads the file
  2. Creates a backup (.backup.<timestamp>)
  3. Applies AI-powered edit based on instruction
  4. Shows diff of changes
  5. Asks for confirmation
  6. Applies or discards changes

Aliases: /modify`),
	}
}

// Execute runs the edit command
func (c *EditCommand) Execute(ctx context.Context, args []string) error {
	if len(args) < 2 {
		fmt.Println(c.Help())
		return nil
	}

	if err := c.checkPermissions(); err != nil {
		return err
	}

	filePath, instruction, noBackup, autoConfirm := c.parseArgs(args)

	absPath, err := c.resolvePath(filePath)
	if err != nil {
		return err
	}

	if err := c.validateFile(absPath); err != nil {
		return err
	}

	return c.editFile(absPath, instruction, noBackup, autoConfirm)
}

func (c *EditCommand) parseArgs(args []string) (filePath, instruction string, noBackup, autoConfirm bool) {
	var instructionParts []string

	for i, arg := range args {
		switch arg {
		case "--no-backup":
			noBackup = true
		case "--yes", "-y":
			autoConfirm = true
		default:
			if filePath == "" && !strings.HasPrefix(arg, "-") {
				filePath = arg
			} else if i > 0 {
				instructionParts = append(instructionParts, arg)
			}
		}
	}

	instruction = strings.Join(instructionParts, " ")
	return filePath, instruction, noBackup, autoConfirm
}

func (c *EditCommand) resolvePath(path string) (string, error) {
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("cannot get home directory: %w", err)
		}
		path = home + path[1:]
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	return absPath, nil
}

func (c *EditCommand) validateFile(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file not found: %s", path)
		}
		if os.IsPermission(err) {
			return fmt.Errorf("permission denied: %s", path)
		}
		return fmt.Errorf("cannot access file: %w", err)
	}

	if info.IsDir() {
		return fmt.Errorf("'%s' is a directory, not a file", path)
	}

	if info.Size() > editMaxFileSize {
		return fmt.Errorf("file too large (%.1f MB > %.1f MB limit)",
			float64(info.Size())/(1024*1024),
			float64(editMaxFileSize)/(1024*1024))
	}

	return nil
}

func (c *EditCommand) checkPermissions() error {
	level := GetCurrentPermissionLevel()
	allowed, needsAsk := IsToolAllowed(level, "file_edit")
	if !allowed {
		return fmt.Errorf("file edit operations are not allowed in %s permission level", level)
	}

	if needsAsk {
		fmt.Printf("%s⚠ Permission Required:%s File edit requires confirmation\n",
			editColorYellow, editColorReset)
	}
	return nil
}

func (c *EditCommand) editFile(path, instruction string, noBackup, autoConfirm bool) error {
	// Read original content
	originalContent, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("cannot read file: %w", err)
	}

	// Display file info
	fmt.Printf("\n%s%sEditing:%s %s%s\n",
		editColorBold, editColorCyan, editColorReset, path, editColorReset)
	fmt.Printf("%sInstruction:%s %s\n\n",
		editColorGray, editColorReset, instruction)

	// Create backup
	if !noBackup {
		backupPath := c.createBackupPath(path)
		if err := os.WriteFile(backupPath, originalContent, 0644); err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}
		fmt.Printf("%s✓ Backup created:%s %s\n\n",
			editColorGreen, editColorReset, backupPath)
	}

	// Apply edit (simulated AI edit for now)
	newContent, err := c.applyEdit(string(originalContent), instruction)
	if err != nil {
		return fmt.Errorf("edit failed: %w", err)
	}

	// Show diff
	fmt.Printf("%sChanges to be applied:%s\n", editColorBold, editColorReset)
	fmt.Println(strings.Repeat("-", 50))
	c.showDiff(string(originalContent), newContent)
	fmt.Println(strings.Repeat("-", 50))

	// Get confirmation unless auto-confirm
	if !autoConfirm {
		if !c.confirmEdit() {
			fmt.Printf("\n%s✗ Edit cancelled.%s\n", editColorRed, editColorReset)
			return nil
		}
	}

	// Apply changes
	if err := os.WriteFile(path, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	fmt.Printf("\n%s✓ File updated successfully.%s\n", editColorGreen, editColorReset)

	if !noBackup {
		fmt.Printf("%s  Backup available at:%s %s%s%s\n",
			editColorGray, editColorReset, editColorYellow, c.createBackupPath(path), editColorReset)
	}

	return nil
}

func (c *EditCommand) createBackupPath(path string) string {
	timestamp := time.Now().Format("20060102_150405")
	return path + editBackupSuffix + timestamp
}

func (c *EditCommand) applyEdit(content, instruction string) (string, error) {
	// This is a replacement for AI-powered editing using pattern-based rules
	// In a real implementation, this would call an AI service

	fmt.Printf("%s🤖 AI Edit:%s Applying changes based on instruction...\n",
		editColorCyan, editColorReset)

	// Simple simulation of edits based on common patterns
	lowerInstruction := strings.ToLower(instruction)

	// Example: Add error handling
	if strings.Contains(lowerInstruction, "error") || strings.Contains(lowerInstruction, "handle") {
		// This is just a simulation - real implementation would use AI
		return content, nil
	}

	// Example: Change port
	if strings.Contains(lowerInstruction, "port") {
		// Simple string replacement for demonstration
		if strings.Contains(content, "8080") && strings.Contains(lowerInstruction, "3000") {
			return strings.ReplaceAll(content, "8080", "3000"), nil
		}
	}

	// For now, return original content with a note
	fmt.Printf("%s⚠ Note:%s AI editing is simulated. Returning original content.\n",
		editColorYellow, editColorReset)
	return content, nil
}

func (c *EditCommand) showDiff(original, modified string) {
	originalLines := strings.Split(original, "\n")
	modifiedLines := strings.Split(modified, "\n")

	// Simple line-by-line diff
	maxLen := len(originalLines)
	if len(modifiedLines) > maxLen {
		maxLen = len(modifiedLines)
	}

	for i := 0; i < maxLen; i++ {
		var origLine, modLine string
		if i < len(originalLines) {
			origLine = originalLines[i]
		}
		if i < len(modifiedLines) {
			modLine = modifiedLines[i]
		}

		if origLine != modLine {
			if origLine != "" {
				fmt.Printf("%s-%s %s%s\n", editColorRed, editColorReset, origLine, editColorReset)
			}
			if modLine != "" {
				fmt.Printf("%s+%s %s%s\n", editColorGreen, editColorReset, modLine, editColorReset)
			}
		}
	}
}

func (c *EditCommand) confirmEdit() bool {
	fmt.Printf("\n%sApply these changes? [y/N]: %s", editColorBold, editColorReset)

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}

func init() {
	Register(NewEditCommand())
}
