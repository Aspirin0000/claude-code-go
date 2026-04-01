// Package commands 提供所有 CLI 命令实现
// 来源: src/commands/ (207个 TS 文件)
// 重构: Go 命令系统 - 批量生产模式
package commands

import (
	"context"
	"fmt"
)

// Command 命令接口
// 对应 TS: Command 类型定义
type Command interface {
	// Name 命令名称 (如 "help", "bash")
	Name() string

	// Aliases 命令别名
	Aliases() []string

	// Description 命令描述
	Description() string

	// Category 命令分类
	Category() CommandCategory

	// Execute 执行命令
	Execute(ctx context.Context, args []string) error

	// Help 获取帮助文本
	Help() string
}

// CommandCategory 命令分类
type CommandCategory string

const (
	CategoryGeneral  CommandCategory = "general"  // 通用命令
	CategorySession  CommandCategory = "session"  // 会话管理
	CategoryConfig   CommandCategory = "config"   // 配置管理
	CategoryMCP      CommandCategory = "mcp"      // MCP管理
	CategoryTools    CommandCategory = "tools"    // 工具命令
	CategoryFiles    CommandCategory = "files"    // 文件操作
	CategoryAdvanced CommandCategory = "advanced" // 高级功能
	CategoryPlugins  CommandCategory = "plugins"  // 插件管理
)

// String 返回分类显示名称
func (c CommandCategory) String() string {
	switch c {
	case CategoryGeneral:
		return "通用命令"
	case CategorySession:
		return "会话管理"
	case CategoryConfig:
		return "配置管理"
	case CategoryMCP:
		return "MCP管理"
	case CategoryTools:
		return "工具命令"
	case CategoryFiles:
		return "文件操作"
	case CategoryAdvanced:
		return "高级功能"
	case CategoryPlugins:
		return "插件管理"
	default:
		return string(c)
	}
}

// BaseCommand 基础命令实现
type BaseCommand struct {
	name        string
	aliases     []string
	description string
	category    CommandCategory
	helpText    string
}

// Name 返回命令名称
func (c *BaseCommand) Name() string {
	return c.name
}

// Aliases 返回命令别名
func (c *BaseCommand) Aliases() []string {
	return c.aliases
}

// Description 返回命令描述
func (c *BaseCommand) Description() string {
	return c.description
}

// Category 返回命令分类
func (c *BaseCommand) Category() CommandCategory {
	return c.category
}

// Help 返回帮助文本
func (c *BaseCommand) Help() string {
	return c.helpText
}

// NewBaseCommand 创建基础命令
func NewBaseCommand(name, description string, category CommandCategory) *BaseCommand {
	return &BaseCommand{
		name:        name,
		description: description,
		category:    category,
		helpText: fmt.Sprintf("/%s - %s\n\n使用: /%s [参数]",
			name, description, name),
	}
}

// WithAliases 设置别名
func (c *BaseCommand) WithAliases(aliases ...string) *BaseCommand {
	c.aliases = aliases
	return c
}

// WithHelp 设置帮助文本
func (c *BaseCommand) WithHelp(help string) *BaseCommand {
	c.helpText = help
	return c
}
