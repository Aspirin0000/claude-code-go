package commands

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ReviewCommand reviews recent changes or plan
type ReviewCommand struct {
	*BaseCommand
}

// NewReviewCommand creates a new review command
func NewReviewCommand() *ReviewCommand {
	cmd := &ReviewCommand{
		BaseCommand: NewBaseCommand(
			"review",
			"Review recent changes or plan",
			CategoryAdvanced,
		),
	}
	cmd.WithHelp(`Usage: /review

Review recent changes, execution plan, or session summary.

Subcommands:
  /review             Show comprehensive review report
  /review changes     View recent code changes
  /review plan        View current execution plan
  /review git         View recent Git commits
  /review summary     View session summary

Examples:
  /review
  /review changes
  /review git 5`)
	return cmd
}

// Execute runs the review command
func (r *ReviewCommand) Execute(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return r.showComprehensiveReview()
	}

	subcommand := strings.ToLower(args[0])

	switch subcommand {
	case "changes", "diff":
		return r.reviewChanges()
	case "plan":
		return r.reviewPlan()
	case "git", "commits":
		count := 10
		if len(args) > 1 {
			if n, err := parseInt(args[1]); err == nil && n > 0 {
				count = n
			}
		}
		return r.reviewGitCommits(count)
	case "summary", "session":
		return r.reviewSessionSummary()
	default:
		return r.showComprehensiveReview()
	}
}

// showComprehensiveReview shows a comprehensive review
func (r *ReviewCommand) showComprehensiveReview() error {
	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║              📊 Review Report                  ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Git status
	if r.isGitRepo() {
		fmt.Println("🔀 Git Status")
		fmt.Println("  " + strings.Repeat("─", 50))
		r.showGitStatus()
		fmt.Println()
	}

	// Current plan
	fmt.Println("📋 Execution Plan")
	fmt.Println("  " + strings.Repeat("─", 50))
	if err := r.showCurrentPlanBrief(); err != nil {
		fmt.Println("  (Unable to load plan)")
	}
	fmt.Println()

	// Recent changes summary
	fmt.Println("📝 Recent Activity")
	fmt.Println("  " + strings.Repeat("─", 50))
	r.showRecentActivity()
	fmt.Println()

	fmt.Println("💡 Tip: Use /review <subcommand> for more details")
	fmt.Println("        /review changes, /review plan, /review git")
	fmt.Println()

	return nil
}

// reviewChanges reviews recent file changes
func (r *ReviewCommand) reviewChanges() error {
	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║              📝 Recent Changes               ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Check git status
	if r.isGitRepo() {
		fmt.Println("Uncommitted changes:")
		cmd := exec.Command("git", "status", "-s")
		output, err := cmd.Output()
		if err != nil {
			fmt.Println("  Unable to get Git status")
		} else if len(output) == 0 {
			fmt.Println("  ✅ Working tree clean, no uncommitted changes")
		} else {
			lines := strings.Split(strings.TrimSpace(string(output)), "\n")
			for _, line := range lines {
				if line != "" {
					fmt.Printf("  %s\n", line)
				}
			}
		}
		fmt.Println()

		// Show recent commits
		fmt.Println("Recent commits:")
		logCmd := exec.Command("git", "log", "--oneline", "-5", "--color=never")
		logOutput, err := logCmd.Output()
		if err == nil && len(logOutput) > 0 {
			lines := strings.Split(strings.TrimSpace(string(logOutput)), "\n")
			for _, line := range lines {
				fmt.Printf("  • %s\n", line)
			}
		}
		fmt.Println()
	}

	// Show recently modified files
	fmt.Println("Recently modified files:")
	r.showRecentFiles()
	fmt.Println()

	return nil
}

// reviewPlan reviews the current plan
func (r *ReviewCommand) reviewPlan() error {
	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║              📋 Plan Review                    ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Println()

	planCmd := NewPlanCommand()
	return planCmd.Execute(context.Background(), []string{})
}

// reviewGitCommits reviews recent git commits
func (r *ReviewCommand) reviewGitCommits(count int) error {
	if !r.isGitRepo() {
		fmt.Println("❌ Current directory is not a Git repository")
		return nil
	}

	fmt.Println()
	fmt.Printf("╔══════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║         🔀 Recent %d Commits              ║\n", count)
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Get current branch
	branchCmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	branch, _ := branchCmd.Output()
	fmt.Printf("Branch: %s\n\n", strings.TrimSpace(string(branch)))

	// Get commits with details
	logCmd := exec.Command("git", "log",
		fmt.Sprintf("-%d", count),
		"--format=%h|%an|%ar|%s",
		"--color=never")
	output, err := logCmd.Output()
	if err != nil {
		return fmt.Errorf("unable to get commit history: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for i, line := range lines {
		parts := strings.SplitN(line, "|", 4)
		if len(parts) == 4 {
			fmt.Printf("%d. %s %s\n", i+1, parts[0], parts[3])
			fmt.Printf("   Author: %s | %s\n", parts[1], parts[2])
			fmt.Println()
		}
	}

	// Show stats
	statCmd := exec.Command("git", "diff", "HEAD~1", "--stat")
	statOutput, _ := statCmd.Output()
	if len(statOutput) > 0 {
		fmt.Println("Recent commit stats:")
		fmt.Println(string(statOutput))
	}

	return nil
}

// reviewSessionSummary reviews session summary
func (r *ReviewCommand) reviewSessionSummary() error {
	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║           📊 Session Summary                   ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Session duration
	fmt.Println("⏱️  Session Info")
	fmt.Println("  " + strings.Repeat("─", 50))
	fmt.Printf("  Start time: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println()

	// Working directory
	cwd, _ := os.Getwd()
	fmt.Println("📁 Working Directory")
	fmt.Println("  " + strings.Repeat("─", 50))
	fmt.Printf("  %s\n", cwd)
	fmt.Println()

	// Git info
	if r.isGitRepo() {
		fmt.Println("🔀 Git Info")
		fmt.Println("  " + strings.Repeat("─", 50))

		branchCmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
		branch, _ := branchCmd.Output()
		fmt.Printf("  Branch: %s\n", strings.TrimSpace(string(branch)))

		// Count commits today
		today := time.Now().Format("2006-01-02")
		countCmd := exec.Command("git", "rev-list", "--count", "--since", today+" 00:00:00", "HEAD")
		count, _ := countCmd.Output()
		fmt.Printf("  Commits today: %s\n", strings.TrimSpace(string(count)))
		fmt.Println()
	}

	// File statistics
	fmt.Println("📊 File Statistics")
	fmt.Println("  " + strings.Repeat("─", 50))
	r.showFileStats()
	fmt.Println()

	return nil
}

// Helper functions

func (r *ReviewCommand) isGitRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	err := cmd.Run()
	return err == nil
}

func (r *ReviewCommand) showGitStatus() {
	cmd := exec.Command("git", "status", "-s")
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("  Unable to get status")
		return
	}

	if len(output) == 0 {
		fmt.Println("  ✅ Working tree clean")
		return
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	modified := 0
	added := 0
	deleted := 0
	untracked := 0

	for _, line := range lines {
		if len(line) >= 2 {
			indexStatus := line[0]
			workTreeStatus := line[1]

			if indexStatus == 'A' || workTreeStatus == 'A' {
				added++
			}
			if indexStatus == 'D' || workTreeStatus == 'D' {
				deleted++
			}
			if indexStatus == 'M' || workTreeStatus == 'M' {
				modified++
			}
			if indexStatus == '?' {
				untracked++
			}
		}
	}

	if modified > 0 {
		fmt.Printf("  📝 Modified: %d\n", modified)
	}
	if added > 0 {
		fmt.Printf("  ➕ Added: %d\n", added)
	}
	if deleted > 0 {
		fmt.Printf("  ➖ Deleted: %d\n", deleted)
	}
	if untracked > 0 {
		fmt.Printf("  ❓ Untracked: %d\n", untracked)
	}
}

func (r *ReviewCommand) showCurrentPlanBrief() error {
	planPath := filepath.Join(getHomeDir(), ".claude-code", "plan.json")

	data, err := os.ReadFile(planPath)
	if err != nil || len(data) == 0 {
		fmt.Println("  📋 No active plan")
		return nil
	}

	// Simple parsing to check if plan exists
	if strings.Contains(string(data), `"id"`) {
		fmt.Println("  📋 Active plan exists")

		// Extract description if available
		if idx := strings.Index(string(data), `"description"`); idx != -1 {
			start := idx + len(`"description"`) + 2
			if start < len(data) {
				end := strings.Index(string(data[start:]), `"`)
				if end > 0 {
					desc := string(data[start : start+end])
					fmt.Printf("     %s\n", desc)
				}
			}
		}
	} else {
		fmt.Println("  📋 No active plan")
	}

	return nil
}

func (r *ReviewCommand) showRecentFiles() {
	// Find recently modified files in current directory
	files := []string{}

	filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			if info.Name() == ".git" || info.Name() == "node_modules" || info.Name() == ".claude-code" {
				return filepath.SkipDir
			}
			return nil
		}

		// Only include files modified in last 24 hours
		if time.Since(info.ModTime()) < 24*time.Hour {
			files = append(files, path)
		}
		return nil
	})

	if len(files) == 0 {
		fmt.Println("  (No files modified in the past 24 hours)")
		return
	}

	// Show first 10 files
	limit := 10
	if len(files) < limit {
		limit = len(files)
	}

	for i := 0; i < limit; i++ {
		fmt.Printf("  • %s\n", files[i])
	}

	if len(files) > limit {
		fmt.Printf("  ... and %d more files\n", len(files)-limit)
	}
}

func (r *ReviewCommand) showRecentActivity() {
	// Show some activity statistics
	fmt.Println("  • Current directory: " + getCurrentDir())

	if r.isGitRepo() {
		cmd := exec.Command("git", "rev-list", "--count", "--since", "24.hours.ago", "HEAD")
		output, _ := cmd.Output()
		count := strings.TrimSpace(string(output))
		if count != "0" && count != "" {
			fmt.Printf("  • %s commits in the past 24 hours\n", count)
		}
	}

	// Check for plan
	planPath := filepath.Join(getHomeDir(), ".claude-code", "plan.json")
	if _, err := os.Stat(planPath); err == nil {
		fmt.Println("  • Active execution plan")
	}
}

func (r *ReviewCommand) showFileStats() {
	// Count files by extension
	extCounts := make(map[string]int)
	totalFiles := 0

	filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			if info.Name() == ".git" || info.Name() == "node_modules" || info.Name() == ".claude-code" {
				return filepath.SkipDir
			}
			return nil
		}

		totalFiles++
		ext := filepath.Ext(info.Name())
		if ext == "" {
			ext = "(no extension)"
		}
		extCounts[ext]++
		return nil
	})

	fmt.Printf("  Total files: %d\n", totalFiles)

	if len(extCounts) > 0 {
		fmt.Println("  By type:")
		// Show top 5 extensions
		type extCount struct {
			ext   string
			count int
		}
		var counts []extCount
		for ext, count := range extCounts {
			counts = append(counts, extCount{ext, count})
		}

		// Sort by count (simple bubble sort)
		for i := 0; i < len(counts); i++ {
			for j := i + 1; j < len(counts); j++ {
				if counts[j].count > counts[i].count {
					counts[i], counts[j] = counts[j], counts[i]
				}
			}
		}

		limit := 5
		if len(counts) < limit {
			limit = len(counts)
		}
		for i := 0; i < limit; i++ {
			fmt.Printf("    %s: %d\n", counts[i].ext, counts[i].count)
		}
	}
}

func getHomeDir() string {
	home, _ := os.UserHomeDir()
	return home
}

func getCurrentDir() string {
	dir, _ := os.Getwd()
	return dir
}

func parseInt(s string) (int, error) {
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	return n, err
}

func init() {
	Register(NewReviewCommand())
}
