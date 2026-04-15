package commands

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Color codes for terminal output
const (
	ColorReset   = "\033[0m"
	ColorRed     = "\033[31m"
	ColorGreen   = "\033[32m"
	ColorYellow  = "\033[33m"
	ColorBlue    = "\033[34m"
	ColorMagenta = "\033[35m"
	ColorCyan    = "\033[36m"
	ColorWhite   = "\033[37m"
	ColorBold    = "\033[1m"
)

// GitCommand implements Git commands
type GitCommand struct {
	BaseCommand
}

// NewGitCommand creates a new Git command
func NewGitCommand() *GitCommand {
	cmd := &GitCommand{
		BaseCommand: BaseCommand{
			name:        "git",
			description: "Git operations helper",
			category:    CategoryTools,
			aliases:     []string{"g", "vcs"},
		},
	}

	cmd.helpText = fmt.Sprintf(`/%s - %s

%sUsage:%s
  /git <subcommand> [args]

%sSubcommands:%s
  %sstatus%s              Show working tree status
		%slog [n]%s            Show last n commits (default: 10)
		%sdiff%s               Show unstaged changes
  %sbranch%s               List branches
  %scommit <message>%s   Commit staged changes
  %sadd <files>%s        Stage files (use . for all)
  %spush%s               Push to remote
  %spull%s               Pull from remote
  %sstash [command]%s    Stash changes (list, pop, drop)

%sAliases:%s %s

%sTips:%s
  - Run without args in subcommands for interactive mode
  - Destructive operations require confirmation
  - Colored output for better readability
`,
		cmd.name, cmd.description,
		ColorBold, ColorReset,
		ColorBold, ColorReset,
		ColorGreen, ColorReset,
		ColorGreen, ColorReset,
		ColorGreen, ColorReset,
		ColorGreen, ColorReset,
		ColorGreen, ColorReset,
		ColorGreen, ColorReset,
		ColorGreen, ColorReset,
		ColorGreen, ColorReset,
		ColorGreen, ColorReset,
		ColorBold, ColorReset,
		strings.Join(cmd.aliases, ", "),
		ColorBold, ColorReset,
	)

	return cmd
}

// Execute executes the Git command
func (g *GitCommand) Execute(ctx context.Context, args []string) error {
	// Check if we are in a git repository
	if !g.IsGitRepo() {
		return fmt.Errorf("not a git repository")
	}

	if len(args) == 0 {
		return g.showHelp()
	}

	subcommand := args[0]
	var cmdArgs []string
	if len(args) > 1 {
		cmdArgs = args[1:]
	}

	switch subcommand {
	case "status":
		return g.status()
	case "log":
		return g.log(cmdArgs)
	case "diff":
		return g.diff(cmdArgs)
	case "branch":
		return g.branch(cmdArgs)
	case "commit":
		return g.commit(cmdArgs)
	case "add":
		return g.add(cmdArgs)
	case "push":
		return g.push(cmdArgs)
	case "pull":
		return g.pull(cmdArgs)
	case "stash":
		return g.stash(cmdArgs)
	case "help", "--help", "-h":
		return g.showHelp()
	default:
		return g.runGitCommand(subcommand, cmdArgs...)
	}
}

// IsGitRepo checks if the current directory is a Git repository
func (g *GitCommand) IsGitRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Dir = g.getWorkDir()
	err := cmd.Run()
	return err == nil
}

// GetCurrentBranch gets the current branch name
func (g *GitCommand) GetCurrentBranch() string {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = g.getWorkDir()
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(output))
}

// GetLastCommit gets the most recent commit info
func (g *GitCommand) GetLastCommit() string {
	cmd := exec.Command("git", "log", "-1", "--format=%h %s")
	cmd.Dir = g.getWorkDir()
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(output))
}

// GetRepoStatus gets the repository status summary
func (g *GitCommand) GetRepoStatus() string {
	branch := g.GetCurrentBranch()
	commit := g.GetLastCommit()
	return fmt.Sprintf("%s[%s]%s %s@%s", ColorCyan, branch, ColorReset, ColorYellow, commit)
}

// status shows git status
func (g *GitCommand) status() error {
	fmt.Printf("%sGit Status%s - %s\n\n", ColorBold, ColorReset, g.GetRepoStatus())

	cmd := exec.Command("git", "status", "-sb")
	cmd.Dir = g.getWorkDir()
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to get status: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		// Colorize status output
		g.colorizeStatusLine(line)
	}

	return nil
}

// colorizeStatusLine adds color to a status output line
func (g *GitCommand) colorizeStatusLine(line string) {
	if strings.HasPrefix(line, "##") {
		// Branch info
		fmt.Printf("%s%s%s\n", ColorBlue, line, ColorReset)
	} else if strings.HasPrefix(line, " M") || strings.HasPrefix(line, "M ") {
		// Modified
		fmt.Printf("%s%s%s\n", ColorYellow, line, ColorReset)
	} else if strings.HasPrefix(line, " A") || strings.HasPrefix(line, "A ") {
		// Added
		fmt.Printf("%s%s%s\n", ColorGreen, line, ColorReset)
	} else if strings.HasPrefix(line, " D") || strings.HasPrefix(line, "D ") {
		// Deleted
		fmt.Printf("%s%s%s\n", ColorRed, line, ColorReset)
	} else if strings.HasPrefix(line, "??") {
		// Untracked
		fmt.Printf("%s%s%s\n", ColorMagenta, line, ColorReset)
	} else {
		fmt.Println(line)
	}
}

// log shows commit log
func (g *GitCommand) log(args []string) error {
	count := 10
	if len(args) > 0 {
		if n, err := strconv.Atoi(args[0]); err == nil && n > 0 {
			count = n
		}
	}

	fmt.Printf("%sGit Log%s - Last %d commits on %s[%s]%s\n\n",
		ColorBold, ColorReset, count, ColorCyan, g.GetCurrentBranch(), ColorReset)

	cmd := exec.Command("git", "log", fmt.Sprintf("-%d", count), "--color=always",
		"--format=%C(yellow)%h%C(reset) - %C(green)%ar%C(reset) %C(blue)%an%C(reset) %s%C(auto)%d")
	cmd.Dir = g.getWorkDir()
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to get log: %w", err)
	}

	fmt.Println(string(output))
	return nil
}

// diff shows differences
func (g *GitCommand) diff(args []string) error {
	fmt.Printf("%sGit Diff%s - %s\n\n", ColorBold, ColorReset, g.GetRepoStatus())

	cmdArgs := []string{"diff", "--color=always"}
	cmdArgs = append(cmdArgs, args...)

	cmd := exec.Command("git", cmdArgs...)
	cmd.Dir = g.getWorkDir()
	output, err := cmd.CombinedOutput()
	if err != nil {
		// git diff returns exit code 1 when there are differences, which is normal
		if exitErr, ok := err.(*exec.ExitError); !ok || exitErr.ExitCode() != 1 {
			return fmt.Errorf("failed to get diff: %w", err)
		}
	}

	if len(output) == 0 {
		fmt.Printf("%sNo changes to display%s\n", ColorGreen, ColorReset)
	} else {
		fmt.Println(string(output))
	}

	return nil
}

// branch handles branch operations
func (g *GitCommand) branch(args []string) error {
	if len(args) == 0 {
		// List branches
		fmt.Printf("%sGit Branches%s - Current: %s[%s]%s\n\n",
			ColorBold, ColorReset, ColorCyan, g.GetCurrentBranch(), ColorReset)

		cmd := exec.Command("git", "branch", "-vv", "--color=always")
		cmd.Dir = g.getWorkDir()
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to list branches: %w", err)
		}

		fmt.Println(string(output))
	} else {
		// Create or switch branch
		return g.runGitCommand("branch", args...)
	}

	return nil
}

// add stages files
func (g *GitCommand) add(args []string) error {
	if len(args) == 0 {
		// Interactive mode - show status and ask
		fmt.Printf("%sGit Add%s - %s\n\n", ColorBold, ColorReset, g.GetRepoStatus())

		// Show current status
		cmd := exec.Command("git", "status", "-s")
		cmd.Dir = g.getWorkDir()
		output, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("failed to get status: %w", err)
		}

		if len(output) == 0 {
			fmt.Printf("%sNo changes to add%s\n", ColorGreen, ColorReset)
			return nil
		}

		fmt.Println("Changes to add:")
		fmt.Println(string(output))

		fmt.Print("\nAdd all changes? [y/N]: ")
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))

		if response == "y" || response == "yes" {
			return g.runGitCommand("add", ".")
		}

		fmt.Println("Cancelled")
		return nil
	}

	// Add specific files
	fmt.Printf("%sAdding files:%s\n", ColorGreen, ColorReset)
	for _, file := range args {
		fmt.Printf("  - %s\n", file)
	}

	return g.runGitCommand("add", args...)
}

// commit commits changes
func (g *GitCommand) commit(args []string) error {
	if len(args) == 0 {
		// Check if there are staged changes
		cmd := exec.Command("git", "diff", "--cached", "--quiet")
		cmd.Dir = g.getWorkDir()
		err := cmd.Run()
		if err == nil {
			return fmt.Errorf("no changes staged for commit")
		}

		// Show what will be committed
		fmt.Printf("%sGit Commit%s - %s\n\n", ColorBold, ColorReset, g.GetRepoStatus())

		diffCmd := exec.Command("git", "diff", "--cached", "--stat")
		diffCmd.Dir = g.getWorkDir()
		output, _ := diffCmd.Output()
		if len(output) > 0 {
			fmt.Println("Changes to be committed:")
			fmt.Println(string(output))
		}

		// Prompt for commit message
		fmt.Print("\nCommit message: ")
		reader := bufio.NewReader(os.Stdin)
		message, _ := reader.ReadString('\n')
		message = strings.TrimSpace(message)

		if message == "" {
			return fmt.Errorf("aborting commit due to empty commit message")
		}

		return g.runGitCommand("commit", "-m", message)
	}

	// Commit with provided message
	message := strings.Join(args, " ")
	return g.runGitCommand("commit", "-m", message)
}

// push pushes to remote
func (g *GitCommand) push(args []string) error {
	// Check if there are commits to push
	cmd := exec.Command("git", "log", "@{u}..HEAD", "--oneline")
	cmd.Dir = g.getWorkDir()
	output, _ := cmd.Output()

	commits := strings.TrimSpace(string(output))
	if commits != "" {
		commitCount := len(strings.Split(commits, "\n"))
		fmt.Printf("%sWarning:%s About to push %d commit(s) to remote\n", ColorYellow, ColorReset, commitCount)

		if !g.confirm("Continue?") {
			fmt.Println("Push cancelled")
			return nil
		}
	}

	fmt.Printf("%sPushing to remote...%s\n", ColorBlue, ColorReset)
	return g.runGitCommand("push", args...)
}

// pull pulls from remote
func (g *GitCommand) pull(args []string) error {
	// Warning about potential conflicts
	fmt.Printf("%sWarning:%s This will pull changes from remote and may cause merge conflicts\n", ColorYellow, ColorReset)

	if !g.confirm("Continue?") {
		fmt.Println("Pull cancelled")
		return nil
	}

	fmt.Printf("%sPulling from remote...%s\n", ColorBlue, ColorReset)
	return g.runGitCommand("pull", args...)
}

// stash stashes changes
func (g *GitCommand) stash(args []string) error {
	if len(args) == 0 {
		// Default stash operation
		fmt.Printf("%sGit Stash%s - %s\n", ColorBold, ColorReset, g.GetRepoStatus())

		// Check if there are changes to stash
		cmd := exec.Command("git", "status", "--porcelain")
		cmd.Dir = g.getWorkDir()
		output, _ := cmd.Output()

		if len(output) == 0 {
			fmt.Printf("%sNo local changes to save%s\n", ColorGreen, ColorReset)
			return nil
		}

		fmt.Println("Changes to be stashed:")
		fmt.Println(string(output))

		if !g.confirm("Stash these changes?") {
			fmt.Println("Stash cancelled")
			return nil
		}

		return g.runGitCommand("stash", "push", "-m", fmt.Sprintf("WIP on %s", g.GetCurrentBranch()))
	}

	// Handle specific stash subcommands
	switch args[0] {
	case "list":
		cmd := exec.Command("git", "stash", "list", "--color=always")
		cmd.Dir = g.getWorkDir()
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to list stashes: %w", err)
		}

		if len(output) == 0 {
			fmt.Printf("%sNo stashes found%s\n", ColorGreen, ColorReset)
		} else {
			fmt.Printf("%sStash list:%s\n", ColorBold, ColorReset)
			fmt.Println(string(output))
		}
		return nil
	case "pop":
		if !g.confirm("Pop stash? This will apply and remove the most recent stash") {
			fmt.Println("Cancelled")
			return nil
		}
		return g.runGitCommand("stash", "pop")
	case "drop":
		if !g.confirm("Drop stash? This cannot be undone") {
			fmt.Println("Cancelled")
			return nil
		}
		return g.runGitCommand("stash", "drop")
	default:
		return g.runGitCommand("stash", args...)
	}
}

// confirm asks the user for confirmation
func (g *GitCommand) confirm(message string) bool {
	fmt.Printf("%s [y/N]: ", message)
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}

// runGitCommand runs a git command and displays output
func (g *GitCommand) runGitCommand(subcommand string, args ...string) error {
	cmdArgs := append([]string{subcommand}, args...)
	cmd := exec.Command("git", cmdArgs...)
	cmd.Dir = g.getWorkDir()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// getWorkDir gets the working directory
func (g *GitCommand) getWorkDir() string {
	if dir, err := os.Getwd(); err == nil {
		return dir
	}
	return "."
}

// showHelp shows help information
func (g *GitCommand) showHelp() error {
	fmt.Println(g.helpText)
	return nil
}

// GetInfo gets Git command information
func (g *GitCommand) GetInfo() map[string]interface{} {
	return map[string]interface{}{
		"name":        g.name,
		"description": g.description,
		"category":    g.category,
		"aliases":     g.aliases,
	}
}

// QuickStatus performs a quick status check (for prompts)
func (g *GitCommand) QuickStatus() string {
	if !g.IsGitRepo() {
		return ""
	}

	branch := g.GetCurrentBranch()

	// Check for uncommitted changes
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = g.getWorkDir()
	output, _ := cmd.Output()

	status := ""
	if len(output) > 0 {
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		modified := 0
		untracked := 0
		for _, line := range lines {
			if len(line) >= 2 {
				if line[0] == '?' && line[1] == '?' {
					untracked++
				} else {
					modified++
				}
			}
		}
		if modified > 0 {
			status += fmt.Sprintf(" %s%d modified%s", ColorYellow, modified, ColorReset)
		}
		if untracked > 0 {
			status += fmt.Sprintf(" %s%d untracked%s", ColorMagenta, untracked, ColorReset)
		}
	}

	return fmt.Sprintf("%s[%s]%s%s", ColorCyan, branch, ColorReset, status)
}

// AutoCommit automatically commits (convenience feature)
func (g *GitCommand) AutoCommit(message string) error {
	if !g.IsGitRepo() {
		return fmt.Errorf("not a git repository")
	}

	// Check for changes
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = g.getWorkDir()
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	if len(output) == 0 {
		fmt.Printf("%sNo changes to commit%s\n", ColorGreen, ColorReset)
		return nil
	}

	// Add all changes
	addCmd := exec.Command("git", "add", "-A")
	addCmd.Dir = g.getWorkDir()
	if err := addCmd.Run(); err != nil {
		return fmt.Errorf("failed to add changes: %w", err)
	}

	// Commit with timestamp if no message provided
	if message == "" {
		message = fmt.Sprintf("Auto commit at %s", time.Now().Format("2006-01-02 15:04:05"))
	}

	commitCmd := exec.Command("git", "commit", "-m", message)
	commitCmd.Dir = g.getWorkDir()
	commitCmd.Stdout = os.Stdout
	commitCmd.Stderr = os.Stderr

	if err := commitCmd.Run(); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	fmt.Printf("%s✓ Committed:%s %s\n", ColorGreen, ColorReset, message)
	return nil
}

// FindGitRoot finds the Git repository root directory
func (g *GitCommand) FindGitRoot(startDir string) (string, error) {
	dir := startDir
	if dir == "" {
		dir = g.getWorkDir()
	}

	for {
		gitDir := filepath.Join(dir, ".git")
		if info, err := os.Stat(gitDir); err == nil && info.IsDir() {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("not a git repository (or any of the parent directories)")
		}
		dir = parent
	}
}

// GetRemoteInfo gets remote repository information
func (g *GitCommand) GetRemoteInfo() (map[string]string, error) {
	if !g.IsGitRepo() {
		return nil, fmt.Errorf("not a git repository")
	}

	cmd := exec.Command("git", "remote", "-v")
	cmd.Dir = g.getWorkDir()
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	remotes := make(map[string]string)
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			name := parts[0]
			url := parts[1]
			// Only keep the fetch URL
			if strings.Contains(line, "(fetch)") {
				remotes[name] = url
			}
		}
	}

	return remotes, nil
}

func init() { Register(NewGitCommand()) }
