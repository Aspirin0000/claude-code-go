package commands

import (
	"context"
	"path/filepath"
	"testing"
)

func TestTasksCommand_AddAndList(t *testing.T) {
	cmd := NewTasksCommand()
	tmpDir := t.TempDir()
	cmd.getTasksFilePath = func() string {
		return filepath.Join(tmpDir, "tasks.json")
	}

	ctx := context.Background()

	// Add a task
	if err := cmd.Execute(ctx, []string{"add", "test task"}); err != nil {
		t.Fatalf("add task failed: %v", err)
	}

	// List tasks
	if err := cmd.Execute(ctx, []string{"list"}); err != nil {
		t.Fatalf("list tasks failed: %v", err)
	}
}

func TestTasksCommand_DefaultListsTasks(t *testing.T) {
	cmd := NewTasksCommand()
	tmpDir := t.TempDir()
	cmd.getTasksFilePath = func() string {
		return filepath.Join(tmpDir, "tasks.json")
	}

	ctx := context.Background()

	// Default (no args) should list tasks
	if err := cmd.Execute(ctx, []string{}); err != nil {
		t.Fatalf("default list failed: %v", err)
	}
}

func TestTasksCommand_DoneAndRemove(t *testing.T) {
	cmd := NewTasksCommand()
	tmpDir := t.TempDir()
	cmd.getTasksFilePath = func() string {
		return filepath.Join(tmpDir, "tasks.json")
	}

	ctx := context.Background()

	// Add a task
	if err := cmd.Execute(ctx, []string{"add", "task to complete"}); err != nil {
		t.Fatalf("add task failed: %v", err)
	}

	// Mark as done (ID should be 1 for first task)
	if err := cmd.Execute(ctx, []string{"done", "1"}); err != nil {
		t.Fatalf("done task failed: %v", err)
	}

	// Remove the task
	if err := cmd.Execute(ctx, []string{"remove", "1"}); err != nil {
		t.Fatalf("remove task failed: %v", err)
	}
}

func TestTasksCommand_PriorityAndTags(t *testing.T) {
	cmd := NewTasksCommand()
	tmpDir := t.TempDir()
	cmd.getTasksFilePath = func() string {
		return filepath.Join(tmpDir, "tasks.json")
	}

	ctx := context.Background()

	// Add a task
	if err := cmd.Execute(ctx, []string{"add", "priority task"}); err != nil {
		t.Fatalf("add task failed: %v", err)
	}

	// Set priority
	if err := cmd.Execute(ctx, []string{"priority", "1", "high"}); err != nil {
		t.Fatalf("set priority failed: %v", err)
	}

	// Add tags
	if err := cmd.Execute(ctx, []string{"tag", "1", "bug", "urgent"}); err != nil {
		t.Fatalf("add tags failed: %v", err)
	}
}

func TestTasksCommand_ClearCompleted(t *testing.T) {
	cmd := NewTasksCommand()
	tmpDir := t.TempDir()
	cmd.getTasksFilePath = func() string {
		return filepath.Join(tmpDir, "tasks.json")
	}

	ctx := context.Background()

	// Add and complete a task
	_ = cmd.Execute(ctx, []string{"add", "completed task"})
	_ = cmd.Execute(ctx, []string{"done", "1"})

	// Clear completed
	if err := cmd.Execute(ctx, []string{"clear"}); err != nil {
		t.Fatalf("clear completed failed: %v", err)
	}
}

func TestTasksCommand_AddWithoutDescription(t *testing.T) {
	cmd := NewTasksCommand()
	tmpDir := t.TempDir()
	cmd.getTasksFilePath = func() string {
		return filepath.Join(tmpDir, "tasks.json")
	}

	ctx := context.Background()

	// Should handle missing description gracefully (prints error, returns nil)
	if err := cmd.Execute(ctx, []string{"add"}); err != nil {
		t.Fatalf("add without description should not error: %v", err)
	}
}

func TestTasksCommand_DoneWithoutID(t *testing.T) {
	cmd := NewTasksCommand()
	tmpDir := t.TempDir()
	cmd.getTasksFilePath = func() string {
		return filepath.Join(tmpDir, "tasks.json")
	}

	ctx := context.Background()

	if err := cmd.Execute(ctx, []string{"done"}); err != nil {
		t.Fatalf("done without id should not error: %v", err)
	}
}

func TestTasksCommand_RemoveWithoutID(t *testing.T) {
	cmd := NewTasksCommand()
	tmpDir := t.TempDir()
	cmd.getTasksFilePath = func() string {
		return filepath.Join(tmpDir, "tasks.json")
	}

	ctx := context.Background()

	if err := cmd.Execute(ctx, []string{"remove"}); err != nil {
		t.Fatalf("remove without id should not error: %v", err)
	}
}

func TestTasksCommand_ImplicitAdd(t *testing.T) {
	cmd := NewTasksCommand()
	tmpDir := t.TempDir()
	cmd.getTasksFilePath = func() string {
		return filepath.Join(tmpDir, "tasks.json")
	}

	ctx := context.Background()

	// Unknown subcommand is treated as task description
	if err := cmd.Execute(ctx, []string{"implicit", "task", "description"}); err != nil {
		t.Fatalf("implicit add failed: %v", err)
	}
}

func TestTasksCommand_Aliases(t *testing.T) {
	cmd := NewTasksCommand()
	aliases := cmd.Aliases()

	expected := map[string]bool{"task": true, "todos": true, "todo": true}
	for _, alias := range aliases {
		if !expected[alias] {
			t.Errorf("unexpected alias: %s", alias)
		}
		delete(expected, alias)
	}
	if len(expected) > 0 {
		t.Errorf("missing expected aliases: %v", expected)
	}
}
