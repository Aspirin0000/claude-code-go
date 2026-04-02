package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// TasksCommand manages tasks
type TasksCommand struct {
	*BaseCommand
}

// Task represents a single task
type Task struct {
	ID          string    `json:"id"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	Priority    string    `json:"priority"`
	CreatedAt   time.Time `json:"created_at"`
	CompletedAt time.Time `json:"completed_at,omitempty"`
	Tags        []string  `json:"tags,omitempty"`
}

// TaskList holds all tasks
type TaskList struct {
	Tasks     []Task    `json:"tasks"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewTasksCommand creates a new tasks command
func NewTasksCommand() *TasksCommand {
	cmd := &TasksCommand{
		BaseCommand: NewBaseCommand(
			"tasks",
			"列出和管理任务",
			CategoryAdvanced,
		),
	}
	cmd.WithAliases("task")
	cmd.WithHelp(`使用: /tasks [add|done|list|remove] [task]

任务管理系统，用于跟踪和管理待办任务。

子命令:
  add <description>    添加新任务
  done <id>           标记任务为完成
  list                列出所有任务 (默认)
  remove <id>         删除任务
  clear               清除所有已完成任务
  priority <id> <p>   设置优先级 (high/medium/low)
  tag <id> <tags...>  添加标签

示例:
  /tasks add 实现用户登录功能
  /tasks add "修复登录页面的bug" --priority high
  /tasks list
  /tasks done 1
  /tasks priority 1 high
  /tasks tag 1 bug frontend`)
	return cmd
}

// Execute runs the tasks command
func (t *TasksCommand) Execute(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return t.listTasks()
	}

	subcommand := strings.ToLower(args[0])

	switch subcommand {
	case "add", "new", "create":
		if len(args) < 2 {
			fmt.Println("❌ 错误: 请提供任务描述")
			fmt.Println("用法: /tasks add <description>")
			return nil
		}
		return t.addTask(args[1:])
	case "done", "complete", "finish":
		if len(args) < 2 {
			fmt.Println("❌ 错误: 请提供任务ID")
			fmt.Println("用法: /tasks done <id>")
			return nil
		}
		return t.completeTask(args[1])
	case "list", "ls", "show":
		return t.listTasks()
	case "remove", "rm", "delete":
		if len(args) < 2 {
			fmt.Println("❌ 错误: 请提供任务ID")
			fmt.Println("用法: /tasks remove <id>")
			return nil
		}
		return t.removeTask(args[1])
	case "clear":
		return t.clearCompleted()
	case "priority":
		if len(args) < 3 {
			fmt.Println("❌ 错误: 请提供任务ID和优先级")
			fmt.Println("用法: /tasks priority <id> <high|medium|low>")
			return nil
		}
		return t.setPriority(args[1], args[2])
	case "tag":
		if len(args) < 3 {
			fmt.Println("❌ 错误: 请提供任务ID和标签")
			fmt.Println("用法: /tasks tag <id> <tag1> [tag2...]")
			return nil
		}
		return t.addTags(args[1], args[2:])
	default:
		// Treat as adding a task with the whole args as description
		return t.addTask(args)
	}
}

// getTasksFilePath returns the path to the tasks file
func (t *TasksCommand) getTasksFilePath() string {
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".claude-code", "tasks.json")
}

// loadTasks loads tasks from disk
func (t *TasksCommand) loadTasks() (*TaskList, error) {
	tasksPath := t.getTasksFilePath()

	data, err := os.ReadFile(tasksPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &TaskList{
				Tasks:     []Task{},
				UpdatedAt: time.Now(),
			}, nil
		}
		return nil, err
	}

	var taskList TaskList
	if err := json.Unmarshal(data, &taskList); err != nil {
		return nil, err
	}

	return &taskList, nil
}

// saveTasks saves tasks to disk
func (t *TasksCommand) saveTasks(taskList *TaskList) error {
	tasksPath := t.getTasksFilePath()

	// Ensure directory exists
	dir := filepath.Dir(tasksPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	taskList.UpdatedAt = time.Now()
	data, err := json.MarshalIndent(taskList, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(tasksPath, data, 0644)
}

// addTask adds a new task
func (t *TasksCommand) addTask(args []string) error {
	taskList, err := t.loadTasks()
	if err != nil {
		return fmt.Errorf("加载任务失败: %w", err)
	}

	// Parse arguments for options
	var description string
	var priority string = "medium"
	var tags []string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--priority" || arg == "-p" {
			if i+1 < len(args) {
				priority = strings.ToLower(args[i+1])
				i++
			}
		} else if arg == "--tag" || arg == "-t" {
			if i+1 < len(args) {
				tags = append(tags, strings.Split(args[i+1], ",")...)
				i++
			}
		} else if description == "" {
			description = arg
		} else {
			description += " " + arg
		}
	}

	if description == "" {
		fmt.Println("❌ 错误: 请提供任务描述")
		return nil
	}

	// Validate priority
	if priority != "high" && priority != "medium" && priority != "low" {
		priority = "medium"
	}

	// Generate task ID
	taskID := strconv.Itoa(len(taskList.Tasks) + 1)

	// Check if ID already exists (in case of deletions)
	for {
		exists := false
		for _, task := range taskList.Tasks {
			if task.ID == taskID {
				exists = true
				break
			}
		}
		if !exists {
			break
		}
		nextID, _ := strconv.Atoi(taskID)
		taskID = strconv.Itoa(nextID + 1)
	}

	task := Task{
		ID:          taskID,
		Description: description,
		Status:      "pending",
		Priority:    priority,
		CreatedAt:   time.Now(),
		Tags:        tags,
	}

	taskList.Tasks = append(taskList.Tasks, task)

	if err := t.saveTasks(taskList); err != nil {
		return fmt.Errorf("保存任务失败: %w", err)
	}

	fmt.Println()
	fmt.Println("✅ 任务已添加")
	fmt.Printf("   ID: %s\n", taskID)
	fmt.Printf("   描述: %s\n", description)
	fmt.Printf("   优先级: %s\n", priority)
	if len(tags) > 0 {
		fmt.Printf("   标签: %s\n", strings.Join(tags, ", "))
	}
	fmt.Println()

	return nil
}

// completeTask marks a task as completed
func (t *TasksCommand) completeTask(taskID string) error {
	taskList, err := t.loadTasks()
	if err != nil {
		return fmt.Errorf("加载任务失败: %w", err)
	}

	found := false
	for i := range taskList.Tasks {
		if taskList.Tasks[i].ID == taskID {
			taskList.Tasks[i].Status = "completed"
			taskList.Tasks[i].CompletedAt = time.Now()
			found = true

			fmt.Println()
			fmt.Printf("✅ 任务 %s 已完成\n", taskID)
			fmt.Printf("   %s\n", taskList.Tasks[i].Description)
			fmt.Println()
			break
		}
	}

	if !found {
		fmt.Printf("❌ 未找到任务 ID: %s\n", taskID)
		return nil
	}

	return t.saveTasks(taskList)
}

// removeTask removes a task
func (t *TasksCommand) removeTask(taskID string) error {
	taskList, err := t.loadTasks()
	if err != nil {
		return fmt.Errorf("加载任务失败: %w", err)
	}

	found := false
	var newTasks []Task
	for _, task := range taskList.Tasks {
		if task.ID == taskID {
			found = true
		} else {
			newTasks = append(newTasks, task)
		}
	}

	if !found {
		fmt.Printf("❌ 未找到任务 ID: %s\n", taskID)
		return nil
	}

	taskList.Tasks = newTasks

	if err := t.saveTasks(taskList); err != nil {
		return fmt.Errorf("删除任务失败: %w", err)
	}

	fmt.Printf("✅ 任务 %s 已删除\n", taskID)

	return nil
}

// listTasks lists all tasks
func (t *TasksCommand) listTasks() error {
	taskList, err := t.loadTasks()
	if err != nil {
		return fmt.Errorf("加载任务失败: %w", err)
	}

	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║              ✅ 任务列表 (Task List)                      ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Println()

	if len(taskList.Tasks) == 0 {
		fmt.Println("   (暂无任务)")
		fmt.Println()
		fmt.Println("💡 使用 /tasks add <描述> 添加任务")
		fmt.Println()
		return nil
	}

	// Separate pending and completed tasks
	var pending []Task
	var completed []Task

	for _, task := range taskList.Tasks {
		if task.Status == "completed" {
			completed = append(completed, task)
		} else {
			pending = append(pending, task)
		}
	}

	// Display pending tasks
	if len(pending) > 0 {
		fmt.Printf("📋 待处理任务 (%d):\n", len(pending))
		fmt.Println("  " + strings.Repeat("─", 50))
		for _, task := range pending {
			t.printTask(task)
		}
		fmt.Println()
	}

	// Display completed tasks
	if len(completed) > 0 {
		fmt.Printf("✅ 已完成任务 (%d):\n", len(completed))
		fmt.Println("  " + strings.Repeat("─", 50))
		for _, task := range completed {
			t.printTask(task)
		}
		fmt.Println()
	}

	// Summary
	fmt.Printf("总计: %d 待处理, %d 已完成\n", len(pending), len(completed))
	fmt.Println()

	return nil
}

// printTask prints a single task
func (t *TasksCommand) printTask(task Task) {
	status := "⬜"
	if task.Status == "completed" {
		status = "✅"
	}

	priorityIcon := ""
	switch task.Priority {
	case "high":
		priorityIcon = "🔴"
	case "medium":
		priorityIcon = "🟡"
	case "low":
		priorityIcon = "🟢"
	}

	fmt.Printf("  %s [%s] %s %s\n", status, task.ID, priorityIcon, task.Description)

	if len(task.Tags) > 0 {
		fmt.Printf("      标签: %s\n", strings.Join(task.Tags, ", "))
	}

	if task.Status == "completed" && !task.CompletedAt.IsZero() {
		fmt.Printf("      完成于: %s\n", task.CompletedAt.Format("01-02 15:04"))
	}
}

// clearCompleted removes all completed tasks
func (t *TasksCommand) clearCompleted() error {
	taskList, err := t.loadTasks()
	if err != nil {
		return fmt.Errorf("加载任务失败: %w", err)
	}

	var pending []Task
	completedCount := 0

	for _, task := range taskList.Tasks {
		if task.Status != "completed" {
			pending = append(pending, task)
		} else {
			completedCount++
		}
	}

	if completedCount == 0 {
		fmt.Println("没有已完成的任务需要清除")
		return nil
	}

	fmt.Printf("确定要清除 %d 个已完成的任务? (y/N): ", completedCount)

	var response string
	fmt.Scanln(&response)
	if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
		fmt.Println("取消清除")
		return nil
	}

	taskList.Tasks = pending

	if err := t.saveTasks(taskList); err != nil {
		return fmt.Errorf("清除任务失败: %w", err)
	}

	fmt.Printf("✅ 已清除 %d 个已完成的任务\n", completedCount)

	return nil
}

// setPriority sets task priority
func (t *TasksCommand) setPriority(taskID, priority string) error {
	priority = strings.ToLower(priority)
	if priority != "high" && priority != "medium" && priority != "low" {
		fmt.Println("❌ 无效的优先级，请使用: high, medium, 或 low")
		return nil
	}

	taskList, err := t.loadTasks()
	if err != nil {
		return fmt.Errorf("加载任务失败: %w", err)
	}

	found := false
	for i := range taskList.Tasks {
		if taskList.Tasks[i].ID == taskID {
			taskList.Tasks[i].Priority = priority
			found = true
			break
		}
	}

	if !found {
		fmt.Printf("❌ 未找到任务 ID: %s\n", taskID)
		return nil
	}

	if err := t.saveTasks(taskList); err != nil {
		return fmt.Errorf("保存任务失败: %w", err)
	}

	fmt.Printf("✅ 任务 %s 优先级已设置为 %s\n", taskID, priority)

	return nil
}

// addTags adds tags to a task
func (t *TasksCommand) addTags(taskID string, tags []string) error {
	taskList, err := t.loadTasks()
	if err != nil {
		return fmt.Errorf("加载任务失败: %w", err)
	}

	found := false
	for i := range taskList.Tasks {
		if taskList.Tasks[i].ID == taskID {
			taskList.Tasks[i].Tags = append(taskList.Tasks[i].Tags, tags...)
			found = true
			break
		}
	}

	if !found {
		fmt.Printf("❌ 未找到任务 ID: %s\n", taskID)
		return nil
	}

	if err := t.saveTasks(taskList); err != nil {
		return fmt.Errorf("保存任务失败: %w", err)
	}

	fmt.Printf("✅ 已为任务 %s 添加标签: %s\n", taskID, strings.Join(tags, ", "))

	return nil
}

func init() {
	Register(NewTasksCommand())
}
