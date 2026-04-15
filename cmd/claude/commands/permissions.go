// Package commands provides CLI command implementations
package commands

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/Aspirin0000/claude-code-go/internal/config"
	"github.com/Aspirin0000/claude-code-go/internal/tools"
)

// ============================================================================
// Permission Level Definitions
// ============================================================================

// PermissionLevel permission level type
type PermissionLevel string

const (
	// PermissionLevelAsk ask mode - confirm before each use
	PermissionLevelAsk PermissionLevel = "ask"
	// PermissionLevelReadOnly read-only mode - only allow read operations
	PermissionLevelReadOnly PermissionLevel = "read-only"
	// PermissionLevelStandard standard mode - allow most tools, dangerous ops require confirmation
	PermissionLevelStandard PermissionLevel = "standard"
	// PermissionLevelFull full mode - allow all tools without asking
	PermissionLevelFull PermissionLevel = "full"
)

// AllPermissionLevels all available permission levels.
var AllPermissionLevels = []PermissionLevel{
	PermissionLevelAsk,
	PermissionLevelReadOnly,
	PermissionLevelStandard,
	PermissionLevelFull,
}

// PermissionLevelInfo permission level metadata.
type PermissionLevelInfo struct {
	Level       PermissionLevel
	Name        string
	Description string
	Color       string
}

// PermissionLevelDetails detailed info for each permission level.
var PermissionLevelDetails = map[PermissionLevel]PermissionLevelInfo{
	PermissionLevelAsk: {
		Level:       PermissionLevelAsk,
		Name:        "Ask",
		Description: "Ask for confirmation before each tool use",
		Color:       "\033[33m", // Yellow
	},
	PermissionLevelReadOnly: {
		Level:       PermissionLevelReadOnly,
		Name:        "Read-Only",
		Description: "Only allow file reading and search operations",
		Color:       "\033[32m", // Green
	},
	PermissionLevelStandard: {
		Level:       PermissionLevelStandard,
		Name:        "Standard",
		Description: "Allow most tools; dangerous operations require confirmation",
		Color:       "\033[36m", // Cyan
	},
	PermissionLevelFull: {
		Level:       PermissionLevelFull,
		Name:        "Full",
		Description: "Allow all tools without confirmation (use with caution)",
		Color:       "\033[31m", // Red
	},
}

// ResetColor ANSI reset code.
const ResetColor = "\033[0m"

// ============================================================================
// Tool Categories
// ============================================================================

// ToolCategory tool category.
type ToolCategory string

const (
	// ToolCategoryRead read tools
	ToolCategoryRead ToolCategory = "read"
	// ToolCategoryWrite write tools
	ToolCategoryWrite ToolCategory = "write"
	// ToolCategorySystem system tools
	ToolCategorySystem ToolCategory = "system"
	// ToolCategoryNetwork network tools
	ToolCategoryNetwork ToolCategory = "network"
	// ToolCategoryTask task tools
	ToolCategoryTask ToolCategory = "task"
)

// ToolInfo metadata for a tool used in permission checks.
type ToolInfo struct {
	Name        string
	Category    ToolCategory
	Description string
	IsDangerous bool
}

// ToolRegistry tool registry for permission management.
var ToolRegistry = []ToolInfo{
	// Read tools
	{Name: "file_read", Category: ToolCategoryRead, Description: "Read file contents", IsDangerous: false},
	{Name: "grep", Category: ToolCategoryRead, Description: "Search file contents", IsDangerous: false},
	{Name: "glob", Category: ToolCategoryRead, Description: "Find files", IsDangerous: false},

	// Write tools
	{Name: "file_write", Category: ToolCategoryWrite, Description: "Write or create files", IsDangerous: true},
	{Name: "file_edit", Category: ToolCategoryWrite, Description: "Edit file contents", IsDangerous: true},

	// System tools
	{Name: "bash", Category: ToolCategorySystem, Description: "Execute shell commands", IsDangerous: true},

	// Network tools
	{Name: "web_search", Category: ToolCategoryNetwork, Description: "Web search", IsDangerous: false},
	{Name: "web_fetch", Category: ToolCategoryNetwork, Description: "Fetch web page content", IsDangerous: false},

	// Task tools
	{Name: "todo_write", Category: ToolCategoryTask, Description: "Manage task lists", IsDangerous: false},
	{Name: "task_get", Category: ToolCategoryTask, Description: "Get task information", IsDangerous: false},
	{Name: "task_create", Category: ToolCategoryTask, Description: "Create new task", IsDangerous: false},
	{Name: "task_update", Category: ToolCategoryTask, Description: "Update task", IsDangerous: false},
	{Name: "task_stop", Category: ToolCategoryTask, Description: "Stop task", IsDangerous: false},
	{Name: "agent", Category: ToolCategoryTask, Description: "Create sub-agent", IsDangerous: true},
	{Name: "notebook_edit", Category: ToolCategoryTask, Description: "Edit notebook", IsDangerous: true},
}

// ============================================================================
// Permission logic
// ============================================================================

// IsToolAllowed checks if a tool is allowed at the current permission level
func IsToolAllowed(level PermissionLevel, toolName string) (allowed bool, needsAsk bool) {
	tool := findTool(toolName)
	if tool == nil {
		// Unknown tools: not allowed in ask/read-only, allowed in standard/full
		switch level {
		case PermissionLevelAsk:
			return false, true
		case PermissionLevelReadOnly:
			return false, false
		case PermissionLevelStandard, PermissionLevelFull:
			return true, level == PermissionLevelStandard
		}
		return false, false
	}

	switch level {
	case PermissionLevelAsk:
		// All tools require confirmation
		return true, true

	case PermissionLevelReadOnly:
		// Only read tools allowed
		return tool.Category == ToolCategoryRead, false

	case PermissionLevelStandard:
		// Allow all tools, but dangerous ones require confirmation
		if tool.IsDangerous {
			return true, true
		}
		return true, false

	case PermissionLevelFull:
		// Allow all tools without confirmation
		return true, false

	default:
		return false, false
	}
}

// findTool looks up tool info
func findTool(name string) *ToolInfo {
	for i := range ToolRegistry {
		if ToolRegistry[i].Name == name {
			return &ToolRegistry[i]
		}
	}
	return nil
}

// GetAllowedTools returns the list of tools allowed at a given permission level
func GetAllowedTools(level PermissionLevel) []ToolInfo {
	var allowed []ToolInfo
	for _, tool := range ToolRegistry {
		isAllowed, _ := IsToolAllowed(level, tool.Name)
		if isAllowed {
			allowed = append(allowed, tool)
		}
	}
	return allowed
}

// GetToolsNeedingAsk returns the list of tools that require confirmation
func GetToolsNeedingAsk(level PermissionLevel) []ToolInfo {
	var needsAsk []ToolInfo
	for _, tool := range ToolRegistry {
		isAllowed, ask := IsToolAllowed(level, tool.Name)
		if isAllowed && ask {
			needsAsk = append(needsAsk, tool)
		}
	}
	return needsAsk
}

// ============================================================================
// Config Integration
// ============================================================================

// ConfigKeyPermissionLevel config key for permission level
const ConfigKeyPermissionLevel = "permission_level"

// GetCurrentPermissionLevel reads the current permission level from config
func GetCurrentPermissionLevel() PermissionLevel {
	cfg, err := config.Load(config.GetConfigPath())
	if err != nil {
		return PermissionLevelStandard // default
	}

	if levelStr, ok := cfg.Env[ConfigKeyPermissionLevel]; ok {
		level := PermissionLevel(levelStr)
		if isValidPermissionLevel(level) {
			return level
		}
	}

	return PermissionLevelStandard
}

// SetPermissionLevel sets the permission level and saves it to config
func SetPermissionLevel(level PermissionLevel) error {
	if !isValidPermissionLevel(level) {
		return fmt.Errorf("invalid permission level: %s", level)
	}

	cfg, err := config.Load(config.GetConfigPath())
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.Env == nil {
		cfg.Env = make(map[string]string)
	}
	cfg.Env[ConfigKeyPermissionLevel] = string(level)

	if err := cfg.Save(config.GetConfigPath()); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

// isValidPermissionLevel checks if a permission level is valid
func isValidPermissionLevel(level PermissionLevel) bool {
	for _, valid := range AllPermissionLevels {
		if level == valid {
			return true
		}
	}
	return false
}

// ============================================================================
// Command Implementation
// ============================================================================

// PermissionsCommand manages tool permissions
type PermissionsCommand struct {
	BaseCommand
}

// NewPermissionsCommand creates the /permissions command
func NewPermissionsCommand() *PermissionsCommand {
	cmd := &PermissionsCommand{
		BaseCommand: *NewBaseCommand(
			"permissions",
			"Manage tool permissions and access levels",
			CategoryConfig,
		),
	}
	cmd.WithAliases("perms", "access").
		WithHelp(`Manage tool permissions and access levels

Usage:
  /permissions           Show current permission level
  /permissions <level>   Set permission level
  /permissions list      List all available levels

Permission Levels:
  ask       - Ask before each tool use
  read-only - Only allow read operations
  standard  - Allow most tools, ask for dangerous ones
  full      - Allow all tools without asking`)
	return cmd
}

// Execute runs the command
func (c *PermissionsCommand) Execute(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return c.showCurrentLevel()
	}

	switch args[0] {
	case "list", "ls", "all":
		return c.listLevels()
	case "help", "-h", "--help":
		return c.showHelp()
	default:
		return c.setLevel(args[0])
	}
}

// showCurrentLevel displays the current permission level
func (c *PermissionsCommand) showCurrentLevel() error {
	current := GetCurrentPermissionLevel()
	info := PermissionLevelDetails[current]

	fmt.Println()
	fmt.Printf("%sCurrent Permission Level:%s %s%s%s\n",
		ColorBold,
		ResetColor,
		info.Color,
		info.Name,
		ResetColor,
	)
	fmt.Printf("Description: %s\n", info.Description)
	fmt.Println()

	fmt.Println("Allowed Tools:")
	allowed := GetAllowedTools(current)
	if len(allowed) == 0 {
		fmt.Println("  (none)")
	} else {
		for _, tool := range allowed {
			_, needsAsk := IsToolAllowed(current, tool.Name)
			status := "✓"
			if needsAsk {
				status = "?"
			}
			fmt.Printf("  %s %s - %s\n", status, tool.Name, tool.Description)
		}
	}

	needsAsk := GetToolsNeedingAsk(current)
	if len(needsAsk) > 0 {
		fmt.Println("\nTools requiring confirmation:")
		for _, tool := range needsAsk {
			fmt.Printf("  ? %s - %s\n", tool.Name, tool.Description)
		}
	}

	fmt.Println()
	fmt.Printf("Use '%s <level>' to change permission level.\n", c.Name())
	fmt.Printf("Use '%s list' to see all available levels.\n", c.Name())
	fmt.Println()

	return nil
}

// listLevels lists all permission levels
func (c *PermissionsCommand) listLevels() error {
	current := GetCurrentPermissionLevel()

	fmt.Println()
	fmt.Println("Available Permission Levels:")
	fmt.Println()

	for _, level := range AllPermissionLevels {
		info := PermissionLevelDetails[level]
		marker := "  "
		if level == current {
			marker = "→ "
		}

		fmt.Printf("%s%s%s%s%s\n",
			marker,
			info.Color,
			info.Name,
			ResetColor,
			func() string {
				if level == current {
					return " (current)"
				}
				return ""
			}(),
		)
		fmt.Printf("   %s\n", info.Description)

		allowed := GetAllowedTools(level)
		needsAsk := GetToolsNeedingAsk(level)

		fmt.Printf("   Tools: %d allowed", len(allowed))
		if len(needsAsk) > 0 {
			fmt.Printf(", %d require confirmation", len(needsAsk))
		}
		fmt.Println()
		fmt.Println()
	}

	fmt.Printf("Use '%s <level>' to change your permission level.\n", c.Name())
	fmt.Println()

	return nil
}

// setLevel sets the permission level
func (c *PermissionsCommand) setLevel(levelStr string) error {
	levelStr = strings.ToLower(strings.TrimSpace(levelStr))

	// Handle aliases
	switch levelStr {
	case "ask", "a":
		levelStr = "ask"
	case "read-only", "readonly", "read", "r":
		levelStr = "read-only"
	case "standard", "std", "s":
		levelStr = "standard"
	case "full", "f", "unrestricted":
		levelStr = "full"
	}

	level := PermissionLevel(levelStr)

	if !isValidPermissionLevel(level) {
		fmt.Fprintf(os.Stderr, "Error: Invalid permission level '%s'\n", levelStr)
		fmt.Fprintf(os.Stderr, "\nValid levels are:\n")
		for _, l := range AllPermissionLevels {
			fmt.Fprintf(os.Stderr, "  - %s\n", l)
		}
		fmt.Fprintf(os.Stderr, "\nUse '%s list' to see detailed information.\n", c.Name())
		return fmt.Errorf("invalid permission level")
	}

	current := GetCurrentPermissionLevel()

	if level == current {
		info := PermissionLevelDetails[level]
		fmt.Printf("\nPermission level is already set to %s%s%s.\n\n",
			info.Color,
			info.Name,
			ResetColor,
		)
		return nil
	}

	if err := SetPermissionLevel(level); err != nil {
		return fmt.Errorf("failed to set permission level: %w", err)
	}

	info := PermissionLevelDetails[level]
	oldInfo := PermissionLevelDetails[current]

	fmt.Println()
	fmt.Printf("Permission level changed:\n")
	fmt.Printf("  From: %s%s%s\n", oldInfo.Color, oldInfo.Name, ResetColor)
	fmt.Printf("  To:   %s%s%s\n", info.Color, info.Name, ResetColor)
	fmt.Println()
	fmt.Printf("Description: %s\n", info.Description)
	fmt.Println()

	if level == PermissionLevelFull {
		fmt.Printf("%s⚠️  Warning:%s You have enabled full permissions.\n",
			"\033[31m", ResetColor)
		fmt.Println("   All tools will be executed without confirmation.")
		fmt.Println("   Use with caution!")
		fmt.Println()
	} else if level == PermissionLevelAsk {
		fmt.Println("ℹ️  You will be asked before each tool use.")
		fmt.Println()
	}

	return nil
}

// showHelp shows help text
func (c *PermissionsCommand) showHelp() error {
	fmt.Println()
	fmt.Printf("Usage: %s [command|level]\n", c.Name())
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  (no args)    Show current permission level")
	fmt.Println("  list         List all permission levels")
	fmt.Println("  help         Show this help message")
	fmt.Println()
	fmt.Println("Permission Levels:")
	for _, level := range AllPermissionLevels {
		info := PermissionLevelDetails[level]
		fmt.Printf("  %s%-12s%s %s\n",
			info.Color,
			level,
			ResetColor,
			info.Description,
		)
	}
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Printf("  %s              Show current level\n", c.Name())
	fmt.Printf("  %s read-only    Set to read-only mode\n", c.Name())
	fmt.Printf("  %s list         List all levels\n", c.Name())
	fmt.Println()
	fmt.Println("Aliases:", strings.Join(c.Aliases(), ", "))
	fmt.Println()
	return nil
}

// GetAllowedToolsForRegistry filters the tool registry by permission level
func GetAllowedToolsForRegistry(registry *tools.Registry, level PermissionLevel) []tools.Tool {
	allTools := registry.List()
	var allowed []tools.Tool

	for _, tool := range allTools {
		isAllowed, _ := IsToolAllowed(level, tool.Name())
		if isAllowed {
			allowed = append(allowed, tool)
		}
	}

	return allowed
}

// ShouldAskBeforeToolUse checks whether to prompt before using a tool
func ShouldAskBeforeToolUse(toolName string) bool {
	level := GetCurrentPermissionLevel()
	_, needsAsk := IsToolAllowed(level, toolName)
	return needsAsk
}

func init() { Register(NewPermissionsCommand()) }
