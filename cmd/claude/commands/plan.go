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
	getPlanFilePath func() string
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
			"Create or view execution plans",
			CategoryAdvanced,
		),
		getPlanFilePath: func() string {
			homeDir, _ := os.UserHomeDir()
			return filepath.Join(homeDir, ".claude-code", "plan.json")
		},
	}
	cmd.WithHelp(`Usage: /plan [description]

Create and manage execution plans to help organize and track tasks.

Usage:
  /plan                    Show current plan
  /plan <description>      Create a new execution plan
  /plan add <step>         Add a step to the current plan
  /plan done <step_id>     Mark a step as done
  /plan clear              Clear the current plan

Examples:
  /plan Refactor user authentication module
  /plan add 1. Analyze existing code
  /plan add 2. Create new interface
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
			fmt.Println("❌ Error: Please provide a step description")
			fmt.Println("Usage: /plan add <step_description>")
			return nil
		}
		return p.addStep(strings.Join(args[1:], " "))
	case "done":
		if len(args) < 2 {
			fmt.Println("❌ Error: Please provide a step ID")
			fmt.Println("Usage: /plan done <step_id>")
			return nil
		}
		return p.markStepDone(args[1])
	case "clear":
		return p.clearPlan()
	case "remove", "rm":
		if len(args) < 2 {
			fmt.Println("❌ Error: Please provide a step ID")
			fmt.Println("Usage: /plan remove <step_id>")
			return nil
		}
		return p.removeStep(args[1])
	default:
		// Create new plan with description
		description := strings.Join(args, " ")
		return p.createPlan(description)
	}
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
		return fmt.Errorf("failed to load plan: %w", err)
	}

	if plan == nil || plan.ID == "" {
		fmt.Println()
		fmt.Println("📋 No active execution plan")
		fmt.Println()
		fmt.Println("💡 Use /plan <description> to create a new plan")
		fmt.Println()
		return nil
	}

	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Printf("║  📋 Plan: %-44s ║\n", truncateString(plan.Description, 44))
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Println()

	if len(plan.Steps) == 0 {
		fmt.Println("   (No steps yet)")
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
		fmt.Printf("   Progress: %d/%d completed\n", completed, len(plan.Steps))
	}

	fmt.Printf("\n   Created: %s\n", plan.CreatedAt.Format("2006-01-02 15:04"))
	fmt.Println()

	return nil
}

// createPlan creates a new plan
func (p *PlanCommand) createPlan(description string) error {
	// Archive existing plan if exists
	existingPlan, err := p.loadPlan()
	if err != nil {
		return fmt.Errorf("failed to load existing plan: %w", err)
	}

	if existingPlan != nil && existingPlan.ID != "" {
		// Ask if user wants to replace
		fmt.Printf("⚠️  A plan already exists: %s\n", existingPlan.Description)
		fmt.Print("Replace it? (y/N): ")

		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			fmt.Println("Cancelled creating new plan")
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
		return fmt.Errorf("failed to save plan: %w", err)
	}

	fmt.Println()
	fmt.Println("✅ Created new plan:")
	fmt.Printf("   %s\n", description)
	fmt.Println()
	fmt.Println("💡 Use /plan add <step> to add steps")
	fmt.Println()

	return nil
}

// addStep adds a step to the current plan
func (p *PlanCommand) addStep(description string) error {
	plan, err := p.loadPlan()
	if err != nil {
		return fmt.Errorf("failed to load plan: %w", err)
	}

	if plan == nil || plan.ID == "" {
		fmt.Println("❌ No active plan")
		fmt.Println("Use /plan <description> to create a plan first")
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
		return fmt.Errorf("failed to save plan: %w", err)
	}

	fmt.Printf("✅ Added step %s: %s\n", stepID, description)

	return nil
}

// markStepDone marks a step as completed
func (p *PlanCommand) markStepDone(stepID string) error {
	plan, err := p.loadPlan()
	if err != nil {
		return fmt.Errorf("failed to load plan: %w", err)
	}

	if plan == nil || plan.ID == "" {
		fmt.Println("❌ No active plan")
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
		fmt.Printf("❌ Step not found: %s\n", stepID)
		return nil
	}

	plan.UpdatedAt = time.Now()

	if err := p.savePlan(plan); err != nil {
		return fmt.Errorf("failed to save plan: %w", err)
	}

	fmt.Printf("✅ Step %s completed\n", stepID)

	// Show progress
	completed := 0
	for _, step := range plan.Steps {
		if step.Status == "done" {
			completed++
		}
	}
	fmt.Printf("   Progress: %d/%d\n", completed, len(plan.Steps))

	return nil
}

// removeStep removes a step from the plan
func (p *PlanCommand) removeStep(stepID string) error {
	plan, err := p.loadPlan()
	if err != nil {
		return fmt.Errorf("failed to load plan: %w", err)
	}

	if plan == nil || plan.ID == "" {
		fmt.Println("❌ No active plan")
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
		fmt.Printf("❌ Step not found: %s\n", stepID)
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
		return fmt.Errorf("failed to save plan: %w", err)
	}

	fmt.Printf("✅ Removed step %s\n", stepID)

	return nil
}

// clearPlan clears the current plan
func (p *PlanCommand) clearPlan() error {
	plan, err := p.loadPlan()
	if err != nil {
		return fmt.Errorf("failed to load plan: %w", err)
	}

	if plan == nil || plan.ID == "" {
		fmt.Println("📋 No active plan to clear")
		return nil
	}

	fmt.Printf("⚠️  Are you sure you want to clear the plan '%s'? (y/N): ", plan.Description)

	var response string
	fmt.Scanln(&response)
	if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
		fmt.Println("Clear cancelled")
		return nil
	}

	// Clear plan by saving empty plan
	emptyPlan := &Plan{}
	if err := p.savePlan(emptyPlan); err != nil {
		return fmt.Errorf("failed to clear plan: %w", err)
	}

	fmt.Println("✅ Plan cleared")

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
