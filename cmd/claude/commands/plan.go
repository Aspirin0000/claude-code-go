package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// PlanCommand manages execution plans
type PlanCommand struct {
	*BaseCommand
}

// Plan represents an execution plan
type Plan struct {
	ID          string    `json:"id"`
	Description string    `json:"description"`
	Steps       []Step    `json:"steps"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Step represents a single step in a plan
type Step struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Order       int    `json:"order"`
}

// NewPlanCommand creates a new plan command
func NewPlanCommand() *PlanCommand {
	cmd := &PlanCommand{
		BaseCommand: NewBaseCommand(
			"plan",
			"创建或查看执行计划",
			CategoryAdvanced,
		),
	}
	cmd.WithHelp(`使用: /plan [description]

创建或管理执行计划，帮助组织和跟踪任务。

用法:
  /plan                    显示当前计划
  /plan <description>      创建新的执行计划
  /plan add <step>         添加步骤到当前计划
  /plan done <step_id>     标记步骤为完成
  /plan clear              清除当前计划

示例:
  /plan 重构用户认证模块
  /plan add 1. 分析现有代码
  /plan add 2. 创建新接口
  /plan done 1`)
	return cmd
}

// Execute runs the plan command
func (p *PlanCommand) Execute(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return p.showCurrentPlan()
	}

	subcommand := strings.ToLower(args[0])

	switch subcommand {
	case "add":
		if len(args) < 2 {
			fmt.Println("❌ 错误: 请提供步骤描述")
			fmt.Println("用法: /plan add <step_description>")
			return nil
		}
		return p.addStep(strings.Join(args[1:], " "))
	case "done":
		if len(args) < 2 {
			fmt.Println("❌ 错误: 请提供步骤ID")
			fmt.Println("用法: /plan done <step_id>")
			return nil
		}
		return p.markStepDone(args[1])
	case "clear":
		return p.clearPlan()
	case "remove", "rm":
		if len(args) < 2 {
			fmt.Println("❌ 错误: 请提供步骤ID")
			fmt.Println("用法: /plan remove <step_id>")
			return nil
		}
		return p.removeStep(args[1])
	default:
		// Create new plan with description
		description := strings.Join(args, " ")
		return p.createPlan(description)
	}
}

// getPlanFilePath returns the path to the plan file
func (p *PlanCommand) getPlanFilePath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".claude-code", "plan.json")
}

// loadPlan loads the current plan from disk
func (p *PlanCommand) loadPlan() (*Plan, error) {
	planPath := p.getPlanFilePath()

	data, err := os.ReadFile(planPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var plan Plan
	if err := json.Unmarshal(data, &plan); err != nil {
		return nil, err
	}

	return &plan, nil
}

// savePlan saves the plan to disk
func (p *PlanCommand) savePlan(plan *Plan) error {
	planPath := p.getPlanFilePath()

	// Ensure directory exists
	dir := filepath.Dir(planPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(plan, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(planPath, data, 0644)
}

// showCurrentPlan displays the current plan
func (p *PlanCommand) showCurrentPlan() error {
	plan, err := p.loadPlan()
	if err != nil {
		return fmt.Errorf("加载计划失败: %w", err)
	}

	if plan == nil || plan.ID == "" {
		fmt.Println()
		fmt.Println("📋 当前没有执行计划")
		fmt.Println()
		fmt.Println("💡 使用 /plan <描述> 创建新计划")
		fmt.Println()
		return nil
	}

	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Printf("║  📋 执行计划: %-44s ║\n", truncateString(plan.Description, 44))
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Println()

	if len(plan.Steps) == 0 {
		fmt.Println("   (暂无步骤)")
	} else {
		completed := 0
		for _, step := range plan.Steps {
			status := "⬜"
			if step.Status == "done" {
				status = "✅"
				completed++
			}
			fmt.Printf("   %s %s. %s\n", status, step.ID, step.Description)
		}
		fmt.Println()
		fmt.Printf("   进度: %d/%d 完成\n", completed, len(plan.Steps))
	}

	fmt.Printf("\n   创建于: %s\n", plan.CreatedAt.Format("2006-01-02 15:04"))
	fmt.Println()

	return nil
}

// createPlan creates a new plan
func (p *PlanCommand) createPlan(description string) error {
	// Archive existing plan if exists
	existingPlan, err := p.loadPlan()
	if err != nil {
		return fmt.Errorf("加载现有计划失败: %w", err)
	}

	if existingPlan != nil && existingPlan.ID != "" {
		// Ask if user wants to replace
		fmt.Printf("⚠️  已存在计划: %s\n", existingPlan.Description)
		fmt.Print("是否替换? (y/N): ")

		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			fmt.Println("取消创建新计划")
			return nil
		}
	}

	plan := &Plan{
		ID:          generatePlanID(),
		Description: description,
		Steps:       []Step{},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := p.savePlan(plan); err != nil {
		return fmt.Errorf("保存计划失败: %w", err)
	}

	fmt.Println()
	fmt.Println("✅ 创建新计划:")
	fmt.Printf("   %s\n", description)
	fmt.Println()
	fmt.Println("💡 使用 /plan add <步骤> 添加步骤")
	fmt.Println()

	return nil
}

// addStep adds a step to the current plan
func (p *PlanCommand) addStep(description string) error {
	plan, err := p.loadPlan()
	if err != nil {
		return fmt.Errorf("加载计划失败: %w", err)
	}

	if plan == nil || plan.ID == "" {
		fmt.Println("❌ 没有活动计划")
		fmt.Println("使用 /plan <描述> 先创建计划")
		return nil
	}

	stepID := fmt.Sprintf("%d", len(plan.Steps)+1)
	step := Step{
		ID:          stepID,
		Description: description,
		Status:      "pending",
		Order:       len(plan.Steps),
	}

	plan.Steps = append(plan.Steps, step)
	plan.UpdatedAt = time.Now()

	if err := p.savePlan(plan); err != nil {
		return fmt.Errorf("保存计划失败: %w", err)
	}

	fmt.Printf("✅ 添加步骤 %s: %s\n", stepID, description)

	return nil
}

// markStepDone marks a step as completed
func (p *PlanCommand) markStepDone(stepID string) error {
	plan, err := p.loadPlan()
	if err != nil {
		return fmt.Errorf("加载计划失败: %w", err)
	}

	if plan == nil || plan.ID == "" {
		fmt.Println("❌ 没有活动计划")
		return nil
	}

	found := false
	for i := range plan.Steps {
		if plan.Steps[i].ID == stepID {
			plan.Steps[i].Status = "done"
			found = true
			break
		}
	}

	if !found {
		fmt.Printf("❌ 未找到步骤 ID: %s\n", stepID)
		return nil
	}

	plan.UpdatedAt = time.Now()

	if err := p.savePlan(plan); err != nil {
		return fmt.Errorf("保存计划失败: %w", err)
	}

	fmt.Printf("✅ 步骤 %s 已完成\n", stepID)

	// Show progress
	completed := 0
	for _, step := range plan.Steps {
		if step.Status == "done" {
			completed++
		}
	}
	fmt.Printf("   进度: %d/%d\n", completed, len(plan.Steps))

	return nil
}

// removeStep removes a step from the plan
func (p *PlanCommand) removeStep(stepID string) error {
	plan, err := p.loadPlan()
	if err != nil {
		return fmt.Errorf("加载计划失败: %w", err)
	}

	if plan == nil || plan.ID == "" {
		fmt.Println("❌ 没有活动计划")
		return nil
	}

	found := false
	newSteps := []Step{}
	for _, step := range plan.Steps {
		if step.ID == stepID {
			found = true
		} else {
			newSteps = append(newSteps, step)
		}
	}

	if !found {
		fmt.Printf("❌ 未找到步骤 ID: %s\n", stepID)
		return nil
	}

	// Reorder remaining steps
	for i := range newSteps {
		newSteps[i].ID = fmt.Sprintf("%d", i+1)
		newSteps[i].Order = i
	}

	plan.Steps = newSteps
	plan.UpdatedAt = time.Now()

	if err := p.savePlan(plan); err != nil {
		return fmt.Errorf("保存计划失败: %w", err)
	}

	fmt.Printf("✅ 已移除步骤 %s\n", stepID)

	return nil
}

// clearPlan clears the current plan
func (p *PlanCommand) clearPlan() error {
	plan, err := p.loadPlan()
	if err != nil {
		return fmt.Errorf("加载计划失败: %w", err)
	}

	if plan == nil || plan.ID == "" {
		fmt.Println("📋 没有活动计划需要清除")
		return nil
	}

	fmt.Printf("⚠️  确定要清除计划 '%s'? (y/N): ", plan.Description)

	var response string
	fmt.Scanln(&response)
	if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
		fmt.Println("取消清除")
		return nil
	}

	// Clear plan by saving empty plan
	emptyPlan := &Plan{}
	if err := p.savePlan(emptyPlan); err != nil {
		return fmt.Errorf("清除计划失败: %w", err)
	}

	fmt.Println("✅ 计划已清除")

	return nil
}

// Helper functions
func generatePlanID() string {
	return fmt.Sprintf("plan_%d", time.Now().Unix())
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func init() {
	Register(NewPlanCommand())
}
