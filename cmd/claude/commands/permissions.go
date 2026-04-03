// Package commands 提供 CLI 命令实现
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
// Permission Level 定义
// ============================================================================

// PermissionLevel 权限级别类型
type PermissionLevel string

const (
	// PermissionLevelAsk 询问模式 - 每次使用前都询问
	PermissionLevelAsk PermissionLevel = "ask"
	// PermissionLevelReadOnly 只读模式 - 只允许读操作
	PermissionLevelReadOnly PermissionLevel = "read-only"
	// PermissionLevelStandard 标准模式 - 允许大多数工具，危险操作需确认
	PermissionLevelStandard PermissionLevel = "standard"
	// PermissionLevelFull 完全模式 - 允许所有工具无需询问
	PermissionLevelFull PermissionLevel = "full"
)

// AllPermissionLevels 所有权限级别
var AllPermissionLevels = []PermissionLevel{
	PermissionLevelAsk,
	PermissionLevelReadOnly,
	PermissionLevelStandard,
	PermissionLevelFull,
}

// PermissionLevelInfo 权限级别信息
type PermissionLevelInfo struct {
	Level       PermissionLevel
	Name        string
	Description string
	Color       string
}

// PermissionLevelDetails 权限级别详细信息
var PermissionLevelDetails = map[PermissionLevel]PermissionLevelInfo{
	PermissionLevelAsk: {
		Level:       PermissionLevelAsk,
		Name:        "Ask",
		Description: "每次使用工具前都会询问确认",
		Color:       "\033[33m", // Yellow
	},
	PermissionLevelReadOnly: {
		Level:       PermissionLevelReadOnly,
		Name:        "Read-Only",
		Description: "只允许读取文件和搜索操作",
		Color:       "\033[32m", // Green
	},
	PermissionLevelStandard: {
		Level:       PermissionLevelStandard,
		Name:        "Standard",
		Description: "允许大多数工具，危险操作需要确认",
		Color:       "\033[36m", // Cyan
	},
	PermissionLevelFull: {
		Level:       PermissionLevelFull,
		Name:        "Full",
		Description: "允许所有工具无需确认（谨慎使用）",
		Color:       "\033[31m", // Red
	},
}

// ResetColor 重置颜色
const ResetColor = "\033[0m"

// ============================================================================
// Tool 分类
// ============================================================================

// ToolCategory 工具分类
type ToolCategory string

const (
	// ToolCategoryRead 读取类工具
	ToolCategoryRead ToolCategory = "read"
	// ToolCategoryWrite 写入类工具
	ToolCategoryWrite ToolCategory = "write"
	// ToolCategorySystem 系统类工具
	ToolCategorySystem ToolCategory = "system"
	// ToolCategoryNetwork 网络类工具
	ToolCategoryNetwork ToolCategory = "network"
	// ToolCategoryTask 任务类工具
	ToolCategoryTask ToolCategory = "task"
)

// ToolInfo 工具信息
type ToolInfo struct {
	Name        string
	Category    ToolCategory
	Description string
	IsDangerous bool
}

// ToolRegistry 工具注册表（用于权限管理）
var ToolRegistry = []ToolInfo{
	// 读取类工具
	{Name: "file_read", Category: ToolCategoryRead, Description: "读取文件内容", IsDangerous: false},
	{Name: "grep", Category: ToolCategoryRead, Description: "搜索文件内容", IsDangerous: false},
	{Name: "glob", Category: ToolCategoryRead, Description: "查找文件", IsDangerous: false},

	// 写入类工具
	{Name: "file_write", Category: ToolCategoryWrite, Description: "写入或创建文件", IsDangerous: true},
	{Name: "file_edit", Category: ToolCategoryWrite, Description: "编辑文件内容", IsDangerous: true},

	// 系统类工具
	{Name: "bash", Category: ToolCategorySystem, Description: "执行 shell 命令", IsDangerous: true},

	// 网络类工具
	{Name: "web_search", Category: ToolCategoryNetwork, Description: "网页搜索", IsDangerous: false},
	{Name: "web_fetch", Category: ToolCategoryNetwork, Description: "获取网页内容", IsDangerous: false},

	// 任务类工具
	{Name: "todo_write", Category: ToolCategoryTask, Description: "管理任务列表", IsDangerous: false},
	{Name: "task_get", Category: ToolCategoryTask, Description: "获取任务信息", IsDangerous: false},
	{Name: "task_create", Category: ToolCategoryTask, Description: "创建新任务", IsDangerous: false},
	{Name: "task_update", Category: ToolCategoryTask, Description: "更新任务", IsDangerous: false},
	{Name: "task_stop", Category: ToolCategoryTask, Description: "停止任务", IsDangerous: false},
	{Name: "agent", Category: ToolCategoryTask, Description: "创建子代理", IsDangerous: true},
	{Name: "notebook_edit", Category: ToolCategoryTask, Description: "编辑笔记本", IsDangerous: true},
}

// ============================================================================
// 权限判断逻辑
// ============================================================================

// IsToolAllowed 检查工具在当前权限级别下是否允许使用
func IsToolAllowed(level PermissionLevel, toolName string) (allowed bool, needsAsk bool) {
	tool := findTool(toolName)
	if tool == nil {
		// 未知工具，在 ask 和 read-only 下不允许，其他级别允许但需确认
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
		// 所有工具都需要询问
		return true, true

	case PermissionLevelReadOnly:
		// 只允许读取类工具
		return tool.Category == ToolCategoryRead, false

	case PermissionLevelStandard:
		// 允许所有工具，但危险操作需要确认
		if tool.IsDangerous {
			return true, true
		}
		return true, false

	case PermissionLevelFull:
		// 允许所有工具无需确认
		return true, false

	default:
		return false, false
	}
}

// findTool 查找工具信息
func findTool(name string) *ToolInfo {
	for i := range ToolRegistry {
		if ToolRegistry[i].Name == name {
			return &ToolRegistry[i]
		}
	}
	return nil
}

// GetAllowedTools 获取指定权限级别下允许的工具列表
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

// GetToolsNeedingAsk 获取需要询问的工具列表
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
// 配置集成
// ============================================================================

// ConfigKeyPermissionLevel 配置文件中权限级别的键名
const ConfigKeyPermissionLevel = "permission_level"

// GetCurrentPermissionLevel 从配置获取当前权限级别
func GetCurrentPermissionLevel() PermissionLevel {
	cfg, err := config.Load(config.GetConfigPath())
	if err != nil {
		return PermissionLevelStandard // 默认标准模式
	}

	// 从 Env 中读取权限级别
	if levelStr, ok := cfg.Env[ConfigKeyPermissionLevel]; ok {
		level := PermissionLevel(levelStr)
		if isValidPermissionLevel(level) {
			return level
		}
	}

	return PermissionLevelStandard
}

// SetPermissionLevel 设置权限级别并保存到配置
func SetPermissionLevel(level PermissionLevel) error {
	if !isValidPermissionLevel(level) {
		return fmt.Errorf("无效的权限级别: %s", level)
	}

	cfg, err := config.Load(config.GetConfigPath())
	if err != nil {
		return fmt.Errorf("加载配置失败: %w", err)
	}

	// 保存到 Env
	if cfg.Env == nil {
		cfg.Env = make(map[string]string)
	}
	cfg.Env[ConfigKeyPermissionLevel] = string(level)

	if err := cfg.Save(config.GetConfigPath()); err != nil {
		return fmt.Errorf("保存配置失败: %w", err)
	}

	return nil
}

// isValidPermissionLevel 检查权限级别是否有效
func isValidPermissionLevel(level PermissionLevel) bool {
	for _, valid := range AllPermissionLevels {
		if level == valid {
			return true
		}
	}
	return false
}

// ============================================================================
// 命令实现
// ============================================================================

// PermissionsCommand 权限管理命令
type PermissionsCommand struct {
	BaseCommand
}

// NewPermissionsCommand 创建权限命令
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

// Execute 执行命令
func (c *PermissionsCommand) Execute(ctx context.Context, args []string) error {
	if len(args) == 0 {
		// 显示当前权限级别
		return c.showCurrentLevel()
	}

	switch args[0] {
	case "list", "ls", "all":
		return c.listLevels()
	case "help", "-h", "--help":
		return c.showHelp()
	default:
		// 尝试设置为指定的权限级别
		return c.setLevel(args[0])
	}
}

// showCurrentLevel 显示当前权限级别
func (c *PermissionsCommand) showCurrentLevel() error {
	current := GetCurrentPermissionLevel()
	info := PermissionLevelDetails[current]

	fmt.Println()
	fmt.Printf("%sCurrent Permission Level:%s %s%s%s\n",
		ResetColor,
		info.Color,
		info.Name,
		ResetColor,
	)
	fmt.Printf("Description: %s\n", info.Description)
	fmt.Println()

	// 显示当前级别下允许的工具
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

	// 显示需要询问的工具
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

// listLevels 列出所有权限级别
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

		// 显示该级别允许的工具数量
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

// setLevel 设置权限级别
func (c *PermissionsCommand) setLevel(levelStr string) error {
	// 标准化输入
	levelStr = strings.ToLower(strings.TrimSpace(levelStr))

	// 处理别名
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

	// 设置新级别
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

	// 显示安全提示
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

// showHelp 显示帮助信息
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

// GetAllowedToolsForRegistry 根据权限级别过滤工具注册表
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

// ShouldAskBeforeToolUse 检查使用工具前是否需要询问
func ShouldAskBeforeToolUse(toolName string) bool {
	level := GetCurrentPermissionLevel()
	_, needsAsk := IsToolAllowed(level, toolName)
	return needsAsk
}

func init() { Register(NewPermissionsCommand()) }
