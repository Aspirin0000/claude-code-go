package tools

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"
)

func setupTestTaskManager(t *testing.T) {
	t.Helper()
	tmpDir := t.TempDir()
	GlobalTaskManager.dataDir = tmpDir
	GlobalTaskManager.mu.Lock()
	GlobalTaskManager.tasks = make(map[string]*Task)
	GlobalTaskManager.mu.Unlock()
}

func TestTaskManagerCreateAndGet(t *testing.T) {
	setupTestTaskManager(t)

	task, err := GlobalTaskManager.CreateTask("Test task", TaskPriorityHigh, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if task.Content != "Test task" {
		t.Errorf("unexpected content: %s", task.Content)
	}
	if task.Priority != TaskPriorityHigh {
		t.Errorf("unexpected priority: %s", task.Priority)
	}
	if task.Status != TaskStatusPending {
		t.Errorf("unexpected status: %s", task.Status)
	}

	got, err := GlobalTaskManager.GetTask(task.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != task.ID {
		t.Errorf("unexpected task id: %s", got.ID)
	}

	_, err = GlobalTaskManager.GetTask("nonexistent")
	if err == nil {
		t.Error("expected error for missing task")
	}
}

func TestTaskManagerUpdate(t *testing.T) {
	setupTestTaskManager(t)

	task, _ := GlobalTaskManager.CreateTask("Original", TaskPriorityMedium, "")
	updated, err := GlobalTaskManager.UpdateTask(task.ID, map[string]interface{}{
		"status":   "done",
		"content":  "Updated",
		"priority": "low",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Status != TaskStatusDone {
		t.Errorf("expected status done, got %s", updated.Status)
	}
	if updated.Content != "Updated" {
		t.Errorf("expected content Updated, got %s", updated.Content)
	}
	if updated.Priority != TaskPriorityLow {
		t.Errorf("expected priority low, got %s", updated.Priority)
	}
	if updated.CompletedAt == nil {
		t.Error("expected CompletedAt to be set for done task")
	}
}

func TestTaskManagerStop(t *testing.T) {
	setupTestTaskManager(t)

	task, _ := GlobalTaskManager.CreateTask("To cancel", TaskPriorityMedium, "")
	stopped, err := GlobalTaskManager.StopTask(task.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stopped.Status != TaskStatusCancelled {
		t.Errorf("expected status cancelled, got %s", stopped.Status)
	}
}

func TestTaskManagerList(t *testing.T) {
	setupTestTaskManager(t)

	GlobalTaskManager.CreateTask("Task 1", TaskPriorityMedium, "")
	t2, _ := GlobalTaskManager.CreateTask("Task 2", TaskPriorityMedium, "")
	GlobalTaskManager.StopTask(t2.ID)

	all := GlobalTaskManager.ListTasks("")
	if len(all) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(all))
	}

	cancelled := GlobalTaskManager.ListTasks("cancelled")
	if len(cancelled) != 1 {
		t.Fatalf("expected 1 cancelled task, got %d", len(cancelled))
	}
}

func TestTaskManagerPersistence(t *testing.T) {
	setupTestTaskManager(t)

	task, _ := GlobalTaskManager.CreateTask("Persistent", TaskPriorityMedium, "")
	id := task.ID

	// Create a new manager pointing to the same directory
	tm2 := NewTaskManager()
	tm2.dataDir = GlobalTaskManager.dataDir
	tm2.loadTasks()

	got, err := tm2.GetTask(id)
	if err != nil {
		t.Fatalf("unexpected error loading persisted task: %v", err)
	}
	if got.Content != "Persistent" {
		t.Errorf("unexpected persisted content: %s", got.Content)
	}
}

func TestTaskCreateTool(t *testing.T) {
	setupTestTaskManager(t)

	tool := &TaskCreateTool{}
	input, _ := json.Marshal(map[string]interface{}{
		"content":  "Tool task",
		"priority": "high",
	})

	result, err := tool.Call(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if !parsed["success"].(bool) {
		t.Fatalf("expected success, got: %+v", parsed)
	}
}

func TestTaskGetTool(t *testing.T) {
	setupTestTaskManager(t)

	task, _ := GlobalTaskManager.CreateTask("Get me", TaskPriorityMedium, "")

	tool := &TaskGetTool{}
	input, _ := json.Marshal(map[string]interface{}{
		"task_id": task.ID,
	})

	result, err := tool.Call(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if !parsed["success"].(bool) {
		t.Fatalf("expected success, got: %+v", parsed)
	}
}

func TestTaskUpdateTool(t *testing.T) {
	setupTestTaskManager(t)

	task, _ := GlobalTaskManager.CreateTask("Update me", TaskPriorityMedium, "")

	tool := &TaskUpdateTool{}
	input, _ := json.Marshal(map[string]interface{}{
		"task_id": task.ID,
		"status":  "done",
	})

	result, err := tool.Call(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if !parsed["success"].(bool) {
		t.Fatalf("expected success, got: %+v", parsed)
	}
}

func TestTaskStopTool(t *testing.T) {
	setupTestTaskManager(t)

	task, _ := GlobalTaskManager.CreateTask("Stop me", TaskPriorityMedium, "")

	tool := &TaskStopTool{}
	input, _ := json.Marshal(map[string]interface{}{
		"task_id": task.ID,
	})

	result, err := tool.Call(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if !parsed["success"].(bool) {
		t.Fatalf("expected success, got: %+v", parsed)
	}
}

func TestTaskListTool(t *testing.T) {
	setupTestTaskManager(t)

	GlobalTaskManager.CreateTask("List task 1", TaskPriorityMedium, "")
	GlobalTaskManager.CreateTask("List task 2", TaskPriorityMedium, "")

	tool := &TaskListTool{}
	input, _ := json.Marshal(map[string]interface{}{})

	result, err := tool.Call(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if parsed["count"].(float64) != 2 {
		t.Errorf("expected count 2, got %v", parsed["count"])
	}
}

func TestTaskManagerCreateEmptyContent(t *testing.T) {
	setupTestTaskManager(t)

	_, err := GlobalTaskManager.CreateTask("", TaskPriorityMedium, "")
	if err == nil {
		t.Error("expected error for empty content")
	}
}

func TestTaskDataDir(t *testing.T) {
	path := getTaskDataDir()
	if path == "" {
		t.Error("expected non-empty task data dir")
	}
	if !filepath.IsAbs(path) {
		t.Logf("task data dir may be relative: %s", path)
	}
}
