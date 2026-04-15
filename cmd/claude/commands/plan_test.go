package commands

import (
	"context"
	"path/filepath"
	"testing"
)

func TestPlanCommand_CreateAndShow(t *testing.T) {
	cmd := NewPlanCommand()
	tmpDir := t.TempDir()
	cmd.getPlanFilePath = func() string {
		return filepath.Join(tmpDir, "plan.json")
	}

	ctx := context.Background()

	// Create a plan
	if err := cmd.Execute(ctx, []string{"refactor auth module"}); err != nil {
		t.Fatalf("create plan failed: %v", err)
	}

	// Show the plan
	if err := cmd.Execute(ctx, []string{}); err != nil {
		t.Fatalf("show plan failed: %v", err)
	}
}

func TestPlanCommand_AddStep(t *testing.T) {
	cmd := NewPlanCommand()
	tmpDir := t.TempDir()
	cmd.getPlanFilePath = func() string {
		return filepath.Join(tmpDir, "plan.json")
	}

	ctx := context.Background()

	_ = cmd.Execute(ctx, []string{"refactor auth module"})

	if err := cmd.Execute(ctx, []string{"add", "step 1: analyze code"}); err != nil {
		t.Fatalf("add step failed: %v", err)
	}
}

func TestPlanCommand_MarkDone(t *testing.T) {
	cmd := NewPlanCommand()
	tmpDir := t.TempDir()
	cmd.getPlanFilePath = func() string {
		return filepath.Join(tmpDir, "plan.json")
	}

	ctx := context.Background()

	_ = cmd.Execute(ctx, []string{"refactor auth module"})
	_ = cmd.Execute(ctx, []string{"add", "step 1"})

	if err := cmd.Execute(ctx, []string{"done", "1"}); err != nil {
		t.Fatalf("mark done failed: %v", err)
	}
}

func TestPlanCommand_RemoveStep(t *testing.T) {
	cmd := NewPlanCommand()
	tmpDir := t.TempDir()
	cmd.getPlanFilePath = func() string {
		return filepath.Join(tmpDir, "plan.json")
	}

	ctx := context.Background()

	_ = cmd.Execute(ctx, []string{"refactor auth module"})
	_ = cmd.Execute(ctx, []string{"add", "step 1"})

	if err := cmd.Execute(ctx, []string{"remove", "1"}); err != nil {
		t.Fatalf("remove step failed: %v", err)
	}
}

func TestPlanCommand_ClearPlan(t *testing.T) {
	cmd := NewPlanCommand()
	tmpDir := t.TempDir()
	cmd.getPlanFilePath = func() string {
		return filepath.Join(tmpDir, "plan.json")
	}

	ctx := context.Background()

	_ = cmd.Execute(ctx, []string{"refactor auth module"})

	if err := cmd.Execute(ctx, []string{"clear"}); err != nil {
		t.Fatalf("clear plan failed: %v", err)
	}
}

func TestPlanCommand_NoPlan(t *testing.T) {
	cmd := NewPlanCommand()
	tmpDir := t.TempDir()
	cmd.getPlanFilePath = func() string {
		return filepath.Join(tmpDir, "plan.json")
	}

	ctx := context.Background()

	// Show when no plan exists
	if err := cmd.Execute(ctx, []string{}); err != nil {
		t.Fatalf("show no plan should not error: %v", err)
	}
}

func TestPlanCommand_InvalidArgs(t *testing.T) {
	cmd := NewPlanCommand()
	tmpDir := t.TempDir()
	cmd.getPlanFilePath = func() string {
		return filepath.Join(tmpDir, "plan.json")
	}

	ctx := context.Background()

	if err := cmd.Execute(ctx, []string{"add"}); err != nil {
		t.Fatalf("add without description should not error: %v", err)
	}

	if err := cmd.Execute(ctx, []string{"done"}); err != nil {
		t.Fatalf("done without id should not error: %v", err)
	}

	if err := cmd.Execute(ctx, []string{"remove"}); err != nil {
		t.Fatalf("remove without id should not error: %v", err)
	}
}
