// Package tools 提供 Task 工具完整实现
// 来源: src/tools/TaskTool/ (多个文件)
// 重构: Go Task 工具（完整实现，支持持久化）
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// TaskStatus 任务状态
type TaskStatus string

const (
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusDone       TaskStatus = "done"
	TaskStatusCancelled  TaskStatus = "cancelled"
	TaskStatusPending    TaskStatus = "pending"
)

// TaskPriority 任务优先级
type TaskPriority string

const (
	TaskPriorityHigh   TaskPriority = "high"
	TaskPriorityMedium TaskPriority = "medium"
	TaskPriorityLow    TaskPriority = "low"
)

// Task 任务定义
type Task struct {
	ID          string       `json:"id"`
	Content     string       `json:"content"`
	Status      TaskStatus   `json:"status"`
	Priority    TaskPriority `json:"priority"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	CompletedAt *time.Time   `json:"completed_at,omitempty"`
	ParentID    string       `json:"parent_id,omitempty"`
	Tags        []string     `json:"tags,omitempty"`
}

// TaskManager 任务管理器
type TaskManager struct {
	tasks   map[string]*Task
	mu      sync.RWMutex
	dataDir string
}

// GlobalTaskManager 全局任务管理器实例
var GlobalTaskManager *TaskManager

func init() {
	GlobalTaskManager = NewTaskManager()
}

// NewTaskManager 创建新的任务管理器
func NewTaskManager() *TaskManager {
	tm := &TaskManager{
		tasks:   make(map[string]*Task),
		dataDir: getTaskDataDir(),
	}
	tm.loadTasks()
	return tm
}

// getTaskDataDir 获取任务数据目录
func getTaskDataDir() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = "."
	}
	return filepath.Join(configDir, "claude", "tasks")
}

// getTaskFilePath 获取任务文件路径
func (tm *TaskManager) getTaskFilePath() string {
	return filepath.Join(tm.dataDir, "tasks.json")
}

// loadTasks 从文件加载任务
func (tm *TaskManager) loadTasks() error {
	taskFile := tm.getTaskFilePath()

	if _, err := os.Stat(taskFile); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(taskFile)
	if err != nil {
		return fmt.Errorf("failed to read tasks file: %w", err)
	}

	var tasks []*Task
	if err := json.Unmarshal(data, &tasks); err != nil {
		return fmt.Errorf("failed to parse tasks: %w", err)
	}

	tm.mu.Lock()
	defer tm.mu.Unlock()

	for _, task := range tasks {
		tm.tasks[task.ID] = task
	}

	return nil
}

// saveTasks 保存任务到文件
func (tm *TaskManager) saveTasks() error {
	tm.mu.RLock()
	tasks := make([]*Task, 0, len(tm.tasks))
	for _, task := range tm.tasks {
		tasks = append(tasks, task)
	}
	tm.mu.RUnlock()

	// Ensure directory exists
	if err := os.MkdirAll(tm.dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create tasks directory: %w", err)
	}

	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize tasks: %w", err)
	}

	taskFile := tm.getTaskFilePath()
	if err := os.WriteFile(taskFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write tasks file: %w", err)
	}

	return nil
}

// CreateTask 创建新任务
func (tm *TaskManager) CreateTask(content string, priority TaskPriority, parentID string) (*Task, error) {
	if content == "" {
		return nil, fmt.Errorf("task content cannot be empty")
	}

	if priority == "" {
		priority = TaskPriorityMedium
	}

	task := &Task{
		ID:        tm.generateTaskID(),
		Content:   content,
		Status:    TaskStatusPending,
		Priority:  priority,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		ParentID:  parentID,
	}

	tm.mu.Lock()
	tm.tasks[task.ID] = task
	tm.mu.Unlock()

	if err := tm.saveTasks(); err != nil {
		return nil, err
	}

	return task, nil
}

// GetTask 获取任务
func (tm *TaskManager) GetTask(taskID string) (*Task, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	task, ok := tm.tasks[taskID]
	if !ok {
		return nil, fmt.Errorf("task %s not found", taskID)
	}

	return task, nil
}

// UpdateTask 更新任务
func (tm *TaskManager) UpdateTask(taskID string, updates map[string]interface{}) (*Task, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	task, ok := tm.tasks[taskID]
	if !ok {
		return nil, fmt.Errorf("task %s not found", taskID)
	}

	// Apply updates
	if content, ok := updates["content"].(string); ok && content != "" {
		task.Content = content
	}

	if status, ok := updates["status"].(string); ok && status != "" {
		task.Status = TaskStatus(status)
		task.UpdatedAt = time.Now()

		// Handle completion
		if task.Status == TaskStatusDone {
			now := time.Now()
			task.CompletedAt = &now
		} else {
			task.CompletedAt = nil
		}
	}

	if priority, ok := updates["priority"].(string); ok && priority != "" {
		task.Priority = TaskPriority(priority)
		task.UpdatedAt = time.Now()
	}

	if err := tm.saveTasks(); err != nil {
		return nil, err
	}

	return task, nil
}

// DeleteTask 删除任务
func (tm *TaskManager) DeleteTask(taskID string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if _, ok := tm.tasks[taskID]; !ok {
		return fmt.Errorf("task %s not found", taskID)
	}

	delete(tm.tasks, taskID)

	return tm.saveTasks()
}

// ListTasks 列出所有任务
func (tm *TaskManager) ListTasks(statusFilter string) []*Task {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	tasks := make([]*Task, 0, len(tm.tasks))
	for _, task := range tm.tasks {
		if statusFilter == "" || string(task.Status) == statusFilter {
			tasks = append(tasks, task)
		}
	}

	return tasks
}

// StopTask 停止任务（将状态设为 cancelled）
func (tm *TaskManager) StopTask(taskID string) (*Task, error) {
	return tm.UpdateTask(taskID, map[string]interface{}{
		"status": TaskStatusCancelled,
	})
}

// generateTaskID 生成任务 ID
func (tm *TaskManager) generateTaskID() string {
	return fmt.Sprintf("task_%d", time.Now().UnixNano())
}

// ============ Tool Implementations ============

// TaskGetTool 获取任务工具
type TaskGetTool struct{}

func (t *TaskGetTool) Name() string        { return "task_get" }
func (t *TaskGetTool) Description() string { return "Get task information by ID" }
func (t *TaskGetTool) IsReadOnly() bool    { return true }
func (t *TaskGetTool) IsDestructive() bool { return false }

func (t *TaskGetTool) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"task_id": {
				"type": "string",
				"description": "Task ID to retrieve"
			}
		},
		"required": ["task_id"]
	}`)
}

func (t *TaskGetTool) Call(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	var params struct {
		TaskID string `json:"task_id"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("failed to parse parameters: %w", err)
	}

	task, err := GlobalTaskManager.GetTask(params.TaskID)
	if err != nil {
		return nil, err
	}

	return json.Marshal(map[string]interface{}{
		"success": true,
		"task":    task,
	})
}

// TaskCreateTool 创建任务工具
type TaskCreateTool struct{}

func (t *TaskCreateTool) Name() string        { return "task_create" }
func (t *TaskCreateTool) Description() string { return "Create a new task" }
func (t *TaskCreateTool) IsReadOnly() bool    { return false }
func (t *TaskCreateTool) IsDestructive() bool { return false }

func (t *TaskCreateTool) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"content": {
				"type": "string",
				"description": "Task content/description"
			},
			"priority": {
				"type": "string",
				"enum": ["high", "medium", "low"],
				"description": "Task priority (default: medium)"
			},
			"parent_id": {
				"type": "string",
				"description": "Parent task ID for subtasks"
			}
		},
		"required": ["content"]
	}`)
}

func (t *TaskCreateTool) Call(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	var params struct {
		Content  string `json:"content"`
		Priority string `json:"priority"`
		ParentID string `json:"parent_id"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("failed to parse parameters: %w", err)
	}

	task, err := GlobalTaskManager.CreateTask(params.Content, TaskPriority(params.Priority), params.ParentID)
	if err != nil {
		return nil, err
	}

	return json.Marshal(map[string]interface{}{
		"success": true,
		"message": "Task created successfully",
		"task":    task,
	})
}

// TaskUpdateTool 更新任务工具
type TaskUpdateTool struct{}

func (t *TaskUpdateTool) Name() string        { return "task_update" }
func (t *TaskUpdateTool) Description() string { return "Update task status or content" }
func (t *TaskUpdateTool) IsReadOnly() bool    { return false }
func (t *TaskUpdateTool) IsDestructive() bool { return true }

func (t *TaskUpdateTool) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"task_id": {
				"type": "string",
				"description": "Task ID to update"
			},
			"status": {
				"type": "string",
				"enum": ["in_progress", "done", "cancelled", "pending"],
				"description": "New task status"
			},
			"content": {
				"type": "string",
				"description": "New task content"
			},
			"priority": {
				"type": "string",
				"enum": ["high", "medium", "low"],
				"description": "New task priority"
			}
		},
		"required": ["task_id"]
	}`)
}

func (t *TaskUpdateTool) Call(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	var params struct {
		TaskID   string `json:"task_id"`
		Status   string `json:"status"`
		Content  string `json:"content"`
		Priority string `json:"priority"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("failed to parse parameters: %w", err)
	}

	updates := make(map[string]interface{})
	if params.Status != "" {
		updates["status"] = params.Status
	}
	if params.Content != "" {
		updates["content"] = params.Content
	}
	if params.Priority != "" {
		updates["priority"] = params.Priority
	}

	task, err := GlobalTaskManager.UpdateTask(params.TaskID, updates)
	if err != nil {
		return nil, err
	}

	return json.Marshal(map[string]interface{}{
		"success": true,
		"message": "Task updated successfully",
		"task":    task,
	})
}

// TaskStopTool 停止任务工具
type TaskStopTool struct{}

func (t *TaskStopTool) Name() string        { return "task_stop" }
func (t *TaskStopTool) Description() string { return "Stop/cancel a task" }
func (t *TaskStopTool) IsReadOnly() bool    { return false }
func (t *TaskStopTool) IsDestructive() bool { return true }

func (t *TaskStopTool) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"task_id": {
				"type": "string",
				"description": "Task ID to stop"
			}
		},
		"required": ["task_id"]
	}`)
}

func (t *TaskStopTool) Call(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	var params struct {
		TaskID string `json:"task_id"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("failed to parse parameters: %w", err)
	}

	task, err := GlobalTaskManager.StopTask(params.TaskID)
	if err != nil {
		return nil, err
	}

	return json.Marshal(map[string]interface{}{
		"success": true,
		"message": "Task stopped successfully",
		"task":    task,
	})
}

// TaskListTool 列出任务工具（额外添加）
type TaskListTool struct{}

func (t *TaskListTool) Name() string        { return "task_list" }
func (t *TaskListTool) Description() string { return "List all tasks with optional status filter" }
func (t *TaskListTool) IsReadOnly() bool    { return true }
func (t *TaskListTool) IsDestructive() bool { return false }

func (t *TaskListTool) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"status": {
				"type": "string",
				"enum": ["in_progress", "done", "cancelled", "pending"],
				"description": "Filter by task status"
			}
		}
	}`)
}

func (t *TaskListTool) Call(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	var params struct {
		Status string `json:"status"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("failed to parse parameters: %w", err)
	}

	tasks := GlobalTaskManager.ListTasks(params.Status)

	return json.Marshal(map[string]interface{}{
		"success": true,
		"count":   len(tasks),
		"tasks":   tasks,
	})
}
