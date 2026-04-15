package commands

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"
)

const (
	defaultTimeout = 30 * time.Second
	maxOutputLines = 1000
	maxOutputBytes = 100 * 1024 // 100KB
	resetColor     = "\033[0m"
	redColor       = "\033[31m"
	yellowColor    = "\033[33m"
	greenColor     = "\033[32m"
	cyanColor      = "\033[36m"
	grayColor      = "\033[90m"
)

var (
	dangerousPatterns = []string{
		`rm\s+-rf?\s+/`,
		`rm\s+-rf?\s+~`,
		`rm\s+-rf?\s+\*`,
		`dd\s+if=.*of=/dev/`,
		`>\s*/dev/`,
		`mkfs\.`,
		`fdisk`,
		`format\s+c:`,
		`del\s+/f\s+/s\s+/q`,
		`rmdir\s+/s\s+/q`,
		`:(){ :|:& };:`, // Fork bomb
	}

	readonlySafeCommands = []string{
		"ls", "cat", "pwd", "echo", "head", "tail", "less", "more",
		"grep", "find", "which", "whereis", "file", "stat", "du", "df",
		"ps", "top", "htop", "free", "uptime", "whoami", "id", "groups",
		"date", "cal", "clear", "man", "help", "history",
	}
)

// BashCommand executes bash commands
type BashCommand struct {
	*BaseCommand
	timeout time.Duration
	dryRun  bool
}

// NewBashCommand creates a bash command
func NewBashCommand() *BashCommand {
	return &BashCommand{
		BaseCommand: NewBaseCommand(
			"bash",
			"Execute bash commands",
			CategoryTools,
		).WithAliases("sh", "shell", "exec").
			WithHelp(`Execute bash commands in the current working directory.

Usage:
  /bash <command>           - Execute a bash command
  /bash -c <command>        - Execute with -c flag
  /bash --dry-run <command> - Show what would be executed without running
  /bash --timeout=<seconds> - Set custom timeout (default: 30s)

Examples:
  /bash ls -la              - List files
  /bash pwd                 - Show current directory
  /bash "cd .. && ls"       - Change directory and list
  /bash -c "echo hello"     - Execute with -c flag
  /bash --dry-run rm file   - Preview before execution

Security:
  Commands are validated against your current permission level.
  Dangerous commands will require confirmation.
  Use /permissions to check or change your permission level.

Aliases: /sh, /shell, /exec`),
		timeout: defaultTimeout,
		dryRun:  false,
	}
}

// Execute executes bash commands
func (c *BashCommand) Execute(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return c.showHelp()
	}

	// Parse flags
	command, err := c.parseArgs(args)
	if err != nil {
		return err
	}

	if command == "" {
		return fmt.Errorf("no command specified")
	}

	// Check permissions
	if err := c.checkPermissions(command); err != nil {
		return err
	}

	// Validate command for security
	if err := c.validateCommand(command); err != nil {
		return err
	}

	// Dry run mode
	if c.dryRun {
		fmt.Printf("%s[DRY RUN] Would execute:%s %s\n", yellowColor, resetColor, command)
		return nil
	}

	// Ask for confirmation if needed
	if c.needsConfirmation(command) {
		if !c.askConfirmation(command) {
			fmt.Println("Command cancelled.")
			return nil
		}
	}

	// Execute the command
	return c.executeCommand(ctx, command)
}

// parseArgs parses command arguments
func (c *BashCommand) parseArgs(args []string) (string, error) {
	var commandParts []string
	useCFlag := false

	for i := 0; i < len(args); i++ {
		arg := args[i]

		switch arg {
		case "-c":
			useCFlag = true
		case "--dry-run":
			c.dryRun = true
		case "--timeout":
			if i+1 >= len(args) {
				return "", fmt.Errorf("--timeout requires a value")
			}
			duration, err := time.ParseDuration(args[i+1] + "s")
			if err != nil {
				return "", fmt.Errorf("invalid timeout value: %s", args[i+1])
			}
			c.timeout = duration
			i++
		case "-h", "--help":
			c.showHelp()
			return "", nil
		default:
			if strings.HasPrefix(arg, "--timeout=") {
				timeoutStr := strings.TrimPrefix(arg, "--timeout=")
				duration, err := time.ParseDuration(timeoutStr + "s")
				if err != nil {
					return "", fmt.Errorf("invalid timeout value: %s", timeoutStr)
				}
				c.timeout = duration
			} else {
				commandParts = append(commandParts, arg)
			}
		}
	}

	if len(commandParts) == 0 {
		return "", nil
	}

	command := strings.Join(commandParts, " ")

	// Handle -c flag: join all remaining args
	if useCFlag {
		command = strings.Join(commandParts, " ")
	}

	return command, nil
}

// checkPermissions checks if the current permission level allows bash command execution
func (c *BashCommand) checkPermissions(command string) error {
	level := GetCurrentPermissionLevel()

	// Check if bash tool is allowed
	allowed, _ := IsToolAllowed(level, "bash")
	if !allowed {
		return fmt.Errorf("bash command execution is not allowed in %s permission level", level)
	}

	// Check if command is safe for readonly mode
	if level == PermissionLevelReadOnly {
		if !c.isReadonlySafe(command) {
			return fmt.Errorf("command '%s' is not allowed in read-only mode", command)
		}
	}

	return nil
}

// isReadonlySafe checks if the command is read-only safe
func (c *BashCommand) isReadonlySafe(command string) bool {
	// Extract the first command from the command string
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return false
	}

	firstCmd := filepath.Base(parts[0])

	for _, safe := range readonlySafeCommands {
		if firstCmd == safe {
			return true
		}
	}

	return false
}

// validateCommand validates if the command is safe
func (c *BashCommand) validateCommand(command string) error {
	// Check for dangerous patterns
	for _, pattern := range dangerousPatterns {
		matched, err := regexp.MatchString(pattern, command)
		if err != nil {
			continue
		}
		if matched {
			return fmt.Errorf("dangerous command pattern detected: '%s' - command blocked for safety", pattern)
		}
	}

	return nil
}

// needsConfirmation checks if the command needs confirmation
func (c *BashCommand) needsConfirmation(command string) bool {
	level := GetCurrentPermissionLevel()

	// Always ask in ask mode
	if level == PermissionLevelAsk {
		return true
	}

	// In standard mode, check if dangerous
	if level == PermissionLevelStandard {
		return c.isDangerous(command)
	}

	// In full mode, no confirmation needed
	return false
}

// isDangerous checks if the command is dangerous
func (c *BashCommand) isDangerous(command string) bool {
	// Check for write/modify operations
	dangerousPrefixes := []string{
		"rm", "del", "rmdir",
		"mv", "move", "ren", "rename",
		"cp", "copy", "xcopy", "robocopy",
		"chmod", "chown", "chgrp",
		"mkfs", "fdisk", "format",
		"dd",
	}

	parts := strings.Fields(command)
	if len(parts) == 0 {
		return false
	}

	firstCmd := filepath.Base(strings.ToLower(parts[0]))

	for _, prefix := range dangerousPrefixes {
		if firstCmd == prefix || strings.HasPrefix(firstCmd, prefix) {
			return true
		}
	}

	// Check for pipe to file (redirection)
	if strings.Contains(command, ">") || strings.Contains(command, "2>") {
		return true
	}

	return false
}

// askConfirmation asks the user for confirmation
func (c *BashCommand) askConfirmation(command string) bool {
	level := GetCurrentPermissionLevel()
	info := PermissionLevelDetails[level]

	fmt.Printf("\n%sPermission Level:%s %s%s%s\n",
		grayColor, info.Color, info.Name, resetColor, grayColor)
	fmt.Printf("Command: %s%s%s\n", cyanColor, command, resetColor)
	fmt.Printf("Execute this command? [y/N]: %s", resetColor)

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}

// executeCommand executes a bash command
func (c *BashCommand) executeCommand(ctx context.Context, command string) error {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	// Print command info
	fmt.Printf("\n%s$ %s%s\n", greenColor, command, resetColor)
	fmt.Println(strings.Repeat("-", 50))

	// Handle special commands
	if handled, err := c.handleSpecialCommand(command); handled {
		if err != nil {
			fmt.Printf("%sError: %v%s\n", redColor, err, resetColor)
			return err
		}
		return nil
	}

	// Execute using bash
	cmd := exec.CommandContext(ctx, "bash", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	// Set working directory
	wd, err := os.Getwd()
	if err == nil {
		cmd.Dir = wd
	}

	// Execute command
	startTime := time.Now()
	err = cmd.Run()
	duration := time.Since(startTime)

	if err != nil {
		// Handle timeout
		if ctx.Err() == context.DeadlineExceeded {
			fmt.Printf("\n%s⚠ Command timed out after %v%s\n", yellowColor, c.timeout, resetColor)
			return fmt.Errorf("command timed out")
		}

		// Handle other errors
		exitCode := 1
		if exitError, ok := err.(*exec.ExitError); ok {
			if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
				exitCode = status.ExitStatus()
			}
		}

		fmt.Printf("\n%s✗ Command failed (exit code %d) after %v%s\n",
			redColor, exitCode, duration, resetColor)
		return err
	}

	fmt.Printf("\n%s✓ Command completed in %v%s\n", greenColor, duration, resetColor)
	return nil
}

// handleSpecialCommand handles special commands (cd, pwd, etc.)
func (c *BashCommand) handleSpecialCommand(command string) (bool, error) {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return false, nil
	}

	firstCmd := strings.ToLower(parts[0])

	switch firstCmd {
	case "cd":
		return c.handleCd(parts)
	case "pwd":
		return c.handlePwd()
	case "ls":
		return false, nil // Let bash handle ls
	case "exit", "quit":
		return c.handleExit(parts)
	default:
		return false, nil
	}
}

// handleCd handles the cd command
func (c *BashCommand) handleCd(parts []string) (bool, error) {
	var dir string
	if len(parts) < 2 {
		home, err := os.UserHomeDir()
		if err != nil {
			return true, fmt.Errorf("cannot get home directory: %w", err)
		}
		dir = home
	} else {
		dir = parts[1]
		// Handle ~ expansion
		if strings.HasPrefix(dir, "~") {
			home, err := os.UserHomeDir()
			if err != nil {
				return true, fmt.Errorf("cannot get home directory: %w", err)
			}
			dir = home + dir[1:]
		}
	}

	err := os.Chdir(dir)
	if err != nil {
		return true, fmt.Errorf("cannot change to directory '%s': %w", dir, err)
	}

	newDir, _ := os.Getwd()
	fmt.Printf("%s→ %s%s\n", cyanColor, newDir, resetColor)
	return true, nil
}

// handlePwd handles the pwd command
func (c *BashCommand) handlePwd() (bool, error) {
	dir, err := os.Getwd()
	if err != nil {
		return true, fmt.Errorf("cannot get current directory: %w", err)
	}
	fmt.Println(dir)
	return true, nil
}

// handleExit handles exit/quit commands
func (c *BashCommand) handleExit(parts []string) (bool, error) {
	code := 0
	if len(parts) > 1 {
		fmt.Sscanf(parts[1], "%d", &code)
	}
	os.Exit(code)
	return true, nil
}

// showHelp shows help information
func (c *BashCommand) showHelp() error {
	fmt.Println(c.Help())
	return nil
}

// SetDryRun sets dry-run mode (for testing)
func (c *BashCommand) SetDryRun(dryRun bool) {
	c.dryRun = dryRun
}

// SetTimeout sets the timeout (for testing)
func (c *BashCommand) SetTimeout(timeout time.Duration) {
	c.timeout = timeout
}

func init() {
	Register(NewBashCommand())
}
