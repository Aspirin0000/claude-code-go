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
	getTasksFilePath func() string
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
			"List and manage tasks",
			CategoryAdvanced,
		),
		getTasksFilePath: func() string {
			homeDir, _ := os.UserHomeDir()
			return filepath.Join(homeDir, ".claude-code", "tasks.json")
		},
	}
	cmd.WithAliases("task", "todos", "todo")
	cmd.WithHelp(`Usage: /tasks [add|done|list|remove] [task]

Task management system for tracking and managing to-do tasks.

Subcommands:
  add <description>    Add a new task
  done <id>           Mark a task as completed
  list                List all tasks (default)
  remove <id>         Remove a task
  clear               Clear all completed tasks
  priority <id> <p>   Set priority (high/medium/low)
  tag <id> <tags...>  Add tags

Examples:
  /tasks add implement user login
  /tasks add "fix login page bug" --priority high
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
			fmt.Println("❌ Error: Please provide a task description")
			fmt.Println("Usage: /tasks add <description>")
			return nil
		}
		return t.addTask(args[1:])
	case "done", "complete", "finish":
		if len(args) < 2 {
			fmt.Println("❌ Error: Please provide a task ID")
			fmt.Println("Usage: /tasks done <id>")
			return nil
		}
		return t.completeTask(args[1])
	case "list", "ls", "show":
		return t.listTasks()
	case "remove", "rm", "delete":
		if len(args) < 2 {
			fmt.Println("❌ Error: Please provide a task ID")
			fmt.Println("Usage: /tasks remove <id>")
			return nil
		}
		return t.removeTask(args[1])
	case "clear":
		return t.clearCompleted()
	case "priority":
		if len(args) < 3 {
			fmt.Println("❌ Error: Please provide a task ID and priority")
			fmt.Println("Usage: /tasks priority <id> <high|medium|low>")
			return nil
		}
		return t.setPriority(args[1], args[2])
	case "tag":
		if len(args) < 3 {
			fmt.Println("❌ Error: Please provide a task ID and tags")
			fmt.Println("Usage: /tasks tag <id> <tag1> [tag2...]")
			return nil
		}
		return t.addTags(args[1], args[2:])
	default:
		// Treat as adding a task with the whole args as description
		return t.addTask(args)
	}
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
		return fmt.Errorf("failed to load tasks: %w", err)
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
		fmt.Println("❌ Error: Please provide a task description")
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
		return fmt.Errorf("failed to save task: %w", err)
	}

	fmt.Println()
	fmt.Println("✅ Task added")
	fmt.Printf("   ID: %s\n", taskID)
	fmt.Printf("   Description: %s\n", description)
	fmt.Printf("   Priority: %s\n", priority)
	if len(tags) > 0 {
		fmt.Printf("   Tags: %s\n", strings.Join(tags, ", "))
	}
	fmt.Println()

	return nil
}

// completeTask marks a task as completed
func (t *TasksCommand) completeTask(taskID string) error {
	taskList, err := t.loadTasks()
	if err != nil {
		return fmt.Errorf("failed to load tasks: %w", err)
	}

	found := false
	for i := range taskList.Tasks {
		if taskList.Tasks[i].ID == taskID {
			taskList.Tasks[i].Status = "completed"
			taskList.Tasks[i].CompletedAt = time.Now()
			found = true

			fmt.Println()
			fmt.Printf("✅ Task %s completed\n", taskID)
			fmt.Printf("   %s\n", taskList.Tasks[i].Description)
			fmt.Println()
			break
		}
	}

	if !found {
		fmt.Printf("❌ Task not found ID: %s\n", taskID)
		return nil
	}

	return t.saveTasks(taskList)
}

// removeTask removes a task
func (t *TasksCommand) removeTask(taskID string) error {
	taskList, err := t.loadTasks()
	if err != nil {
		return fmt.Errorf("failed to load tasks: %w", err)
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
		fmt.Printf("❌ Task not found ID: %s\n", taskID)
		return nil
	}

	taskList.Tasks = newTasks

	if err := t.saveTasks(taskList); err != nil {
		return fmt.Errorf("failed to remove task: %w", err)
	}

	fmt.Printf("✅ Task %s removed\n", taskID)

	return nil
}

// listTasks lists all tasks
func (t *TasksCommand) listTasks() error {
	taskList, err := t.loadTasks()
	if err != nil {
		return fmt.Errorf("failed to load tasks: %w", err)
	}

	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║                 ✅ Task List                             ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Println()

	if len(taskList.Tasks) == 0 {
		fmt.Println("   (No tasks)")
		fmt.Println()
		fmt.Println("💡 Use /tasks add <description> to add a task")
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
		fmt.Printf("📋 Pending tasks (%d):\n", len(pending))
		fmt.Println("  " + strings.Repeat("─", 50))
		for _, task := range pending {
			t.printTask(task)
		}
		fmt.Println()
	}

	// Display completed tasks
	if len(completed) > 0 {
		fmt.Printf("✅ Completed tasks (%d):\n", len(completed))
		fmt.Println("  " + strings.Repeat("─", 50))
		for _, task := range completed {
			t.printTask(task)
		}
		fmt.Println()
	}

	// Summary
	fmt.Printf("Total: %d pending, %d completed\n", len(pending), len(completed))
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
		fmt.Printf("      Tags: %s\n", strings.Join(task.Tags, ", "))
	}

	if task.Status == "completed" && !task.CompletedAt.IsZero() {
		fmt.Printf("      Completed at: %s\n", task.CompletedAt.Format("01-02 15:04"))
	}
}

// clearCompleted removes all completed tasks
func (t *TasksCommand) clearCompleted() error {
	taskList, err := t.loadTasks()
	if err != nil {
		return fmt.Errorf("failed to load tasks: %w", err)
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
		fmt.Println("No completed tasks to clear")
		return nil
	}

	fmt.Printf("Are you sure you want to clear %d completed tasks? (y/N): ", completedCount)

	var response string
	fmt.Scanln(&response)
	if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
		fmt.Println("Clear cancelled")
		return nil
	}

	taskList.Tasks = pending

	if err := t.saveTasks(taskList); err != nil {
		return fmt.Errorf("failed to clear tasks: %w", err)
	}

	fmt.Printf("✅ Cleared %d completed tasks\n", completedCount)

	return nil
}

// setPriority sets task priority
func (t *TasksCommand) setPriority(taskID, priority string) error {
	priority = strings.ToLower(priority)
	if priority != "high" && priority != "medium" && priority != "low" {
		fmt.Println("❌ Invalid priority, please use: high, medium, or low")
		return nil
	}

	taskList, err := t.loadTasks()
	if err != nil {
		return fmt.Errorf("failed to load tasks: %w", err)
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
		fmt.Printf("❌ Task not found ID: %s\n", taskID)
		return nil
	}

	if err := t.saveTasks(taskList); err != nil {
		return fmt.Errorf("failed to save task: %w", err)
	}

	fmt.Printf("✅ Task %s priority set to %s\n", taskID, priority)

	return nil
}

// addTags adds tags to a task
func (t *TasksCommand) addTags(taskID string, tags []string) error {
	taskList, err := t.loadTasks()
	if err != nil {
		return fmt.Errorf("failed to load tasks: %w", err)
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
		fmt.Printf("❌ Task not found ID: %s\n", taskID)
		return nil
	}

	if err := t.saveTasks(taskList); err != nil {
		return fmt.Errorf("failed to save task: %w", err)
	}

	fmt.Printf("✅ Added tags to task %s: %s\n", taskID, strings.Join(tags, ", "))

	return nil
}

func init() {
	Register(NewTasksCommand())
}
