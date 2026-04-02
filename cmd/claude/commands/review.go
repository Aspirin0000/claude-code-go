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
			"查看最近的更改或计划",
			CategoryAdvanced,
		),
	}
	cmd.WithHelp(`使用: /review

查看最近的更改、执行计划或会话摘要。

子命令:
  /review             显示综合审查报告
  /review changes     查看最近的代码更改
  /review plan        查看当前执行计划
  /review git         查看最近的Git提交
  /review summary     查看会话摘要

示例:
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
	fmt.Println("║              📊 审查报告 (Review Report)                  ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Git status
	if r.isGitRepo() {
		fmt.Println("🔀 Git 状态")
		fmt.Println("  " + strings.Repeat("─", 50))
		r.showGitStatus()
		fmt.Println()
	}

	// Current plan
	fmt.Println("📋 执行计划")
	fmt.Println("  " + strings.Repeat("─", 50))
	if err := r.showCurrentPlanBrief(); err != nil {
		fmt.Println("  (无法加载计划)")
	}
	fmt.Println()

	// Recent changes summary
	fmt.Println("📝 最近活动")
	fmt.Println("  " + strings.Repeat("─", 50))
	r.showRecentActivity()
	fmt.Println()

	fmt.Println("💡 提示: 使用 /review <子命令> 查看详细信息")
	fmt.Println("        /review changes, /review plan, /review git")
	fmt.Println()

	return nil
}

// reviewChanges reviews recent file changes
func (r *ReviewCommand) reviewChanges() error {
	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║              📝 最近的更改 (Recent Changes)               ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Check git status
	if r.isGitRepo() {
		fmt.Println("未提交的更改:")
		cmd := exec.Command("git", "status", "-s")
		output, err := cmd.Output()
		if err != nil {
			fmt.Println("  无法获取Git状态")
		} else if len(output) == 0 {
			fmt.Println("  ✅ 工作区干净，没有未提交的更改")
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
		fmt.Println("最近的提交:")
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
	fmt.Println("最近修改的文件:")
	r.showRecentFiles()
	fmt.Println()

	return nil
}

// reviewPlan reviews the current plan
func (r *ReviewCommand) reviewPlan() error {
	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║              📋 计划审查 (Plan Review)                    ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Println()

	planCmd := NewPlanCommand()
	return planCmd.Execute(context.Background(), []string{})
}

// reviewGitCommits reviews recent git commits
func (r *ReviewCommand) reviewGitCommits(count int) error {
	if !r.isGitRepo() {
		fmt.Println("❌ 当前目录不是Git仓库")
		return nil
	}

	fmt.Println()
	fmt.Printf("╔══════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║         🔀 最近的 %d 次提交 (Recent Commits)              ║\n", count)
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Get current branch
	branchCmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	branch, _ := branchCmd.Output()
	fmt.Printf("分支: %s\n\n", strings.TrimSpace(string(branch)))

	// Get commits with details
	logCmd := exec.Command("git", "log",
		fmt.Sprintf("-%d", count),
		"--format=%h|%an|%ar|%s",
		"--color=never")
	output, err := logCmd.Output()
	if err != nil {
		return fmt.Errorf("无法获取提交历史: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for i, line := range lines {
		parts := strings.SplitN(line, "|", 4)
		if len(parts) == 4 {
			fmt.Printf("%d. %s %s\n", i+1, parts[0], parts[3])
			fmt.Printf("   作者: %s | %s\n", parts[1], parts[2])
			fmt.Println()
		}
	}

	// Show stats
	statCmd := exec.Command("git", "diff", "HEAD~1", "--stat")
	statOutput, _ := statCmd.Output()
	if len(statOutput) > 0 {
		fmt.Println("最近提交的统计:")
		fmt.Println(string(statOutput))
	}

	return nil
}

// reviewSessionSummary reviews session summary
func (r *ReviewCommand) reviewSessionSummary() error {
	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║           📊 会话摘要 (Session Summary)                   ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Session duration
	fmt.Println("⏱️  会话信息")
	fmt.Println("  " + strings.Repeat("─", 50))
	fmt.Printf("  启动时间: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Println()

	// Working directory
	cwd, _ := os.Getwd()
	fmt.Println("📁 工作目录")
	fmt.Println("  " + strings.Repeat("─", 50))
	fmt.Printf("  %s\n", cwd)
	fmt.Println()

	// Git info
	if r.isGitRepo() {
		fmt.Println("🔀 Git 信息")
		fmt.Println("  " + strings.Repeat("─", 50))

		branchCmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
		branch, _ := branchCmd.Output()
		fmt.Printf("  分支: %s\n", strings.TrimSpace(string(branch)))

		// Count commits today
		today := time.Now().Format("2006-01-02")
		countCmd := exec.Command("git", "rev-list", "--count", "--since", today+" 00:00:00", "HEAD")
		count, _ := countCmd.Output()
		fmt.Printf("  今日提交: %s\n", strings.TrimSpace(string(count)))
		fmt.Println()
	}

	// File statistics
	fmt.Println("📊 文件统计")
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
		fmt.Println("  无法获取状态")
		return
	}

	if len(output) == 0 {
		fmt.Println("  ✅ 工作区干净")
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
		fmt.Printf("  📝 修改: %d\n", modified)
	}
	if added > 0 {
		fmt.Printf("  ➕ 新增: %d\n", added)
	}
	if deleted > 0 {
		fmt.Printf("  ➖ 删除: %d\n", deleted)
	}
	if untracked > 0 {
		fmt.Printf("  ❓ 未跟踪: %d\n", untracked)
	}
}

func (r *ReviewCommand) showCurrentPlanBrief() error {
	planPath := filepath.Join(getHomeDir(), ".claude-code", "plan.json")

	data, err := os.ReadFile(planPath)
	if err != nil || len(data) == 0 {
		fmt.Println("  📋 没有活动计划")
		return nil
	}

	// Simple parsing to check if plan exists
	if strings.Contains(string(data), `"id"`) {
		fmt.Println("  📋 有活动计划")

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
		fmt.Println("  📋 没有活动计划")
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
		fmt.Println("  (过去24小时没有修改的文件)")
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
		fmt.Printf("  ... 还有 %d 个文件\n", len(files)-limit)
	}
}

func (r *ReviewCommand) showRecentActivity() {
	// Show some activity statistics
	fmt.Println("  • 当前目录: " + getCurrentDir())

	if r.isGitRepo() {
		cmd := exec.Command("git", "rev-list", "--count", "--since", "24.hours.ago", "HEAD")
		output, _ := cmd.Output()
		count := strings.TrimSpace(string(output))
		if count != "0" && count != "" {
			fmt.Printf("  • 过去24小时有 %s 次提交\n", count)
		}
	}

	// Check for plan
	planPath := filepath.Join(getHomeDir(), ".claude-code", "plan.json")
	if _, err := os.Stat(planPath); err == nil {
		fmt.Println("  • 有活动的执行计划")
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
			ext = "(无扩展名)"
		}
		extCounts[ext]++
		return nil
	})

	fmt.Printf("  总文件数: %d\n", totalFiles)

	if len(extCounts) > 0 {
		fmt.Println("  按类型:")
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
