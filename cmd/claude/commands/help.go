package commands

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

// HelpCommand 显示帮助信息
type HelpCommand struct {
	*BaseCommand
}

// NewHelpCommand 创建帮助命令
func NewHelpCommand() *HelpCommand {
	base := NewBaseCommand(
		"help",
		"显示所有命令列表或特定命令的详细帮助",
		CategoryGeneral,
	).WithAliases("h", "?").WithHelp(`/help - 显示所有命令列表或特定命令的详细帮助

使用: /help [command]

参数:
  command  可选。要查看详细帮助的命令名称

示例:
  /help          显示所有可用命令列表
  /help clear    显示 clear 命令的详细帮助
  /h             等同于 /help
  /?             等同于 /help`)

	return &HelpCommand{BaseCommand: base}
}

// Execute 执行帮助命令
func (c *HelpCommand) Execute(ctx context.Context, args []string) error {
	if len(args) > 0 {
		return c.showCommandHelp(args[0])
	}
	return c.showAllCommands()
}

// showAllCommands 显示所有命令（按分类分组）
func (c *HelpCommand) showAllCommands() error {
	registry := GetRegistry()

	// 获取所有分类并排序
	categories := []CommandCategory{
		CategoryGeneral,
		CategorySession,
		CategoryConfig,
		CategoryFiles,
		CategoryTools,
		CategoryMCP,
		CategoryPlugins,
		CategoryAdvanced,
	}

	fmt.Println("\n可用命令:")
	fmt.Println(strings.Repeat("=", 50))

	for _, cat := range categories {
		cmds := registry.ListByCategory(cat)
		if len(cmds) == 0 {
			continue
		}

		// 按名称排序命令
		sort.Slice(cmds, func(i, j int) bool {
			return cmds[i].Name() < cmds[j].Name()
		})

		fmt.Printf("\n%s:\n", cat.String())
		fmt.Println(strings.Repeat("-", 30))

		for _, cmd := range cmds {
			name := cmd.Name()
			desc := cmd.Description()
			aliases := cmd.Aliases()

			// 格式化显示
			if len(aliases) > 0 {
				aliasStr := fmt.Sprintf(" (别名: %s)", strings.Join(aliases, ", "))
				fmt.Printf("  /%-15s %s%s\n", name, desc, aliasStr)
			} else {
				fmt.Printf("  /%-15s %s\n", name, desc)
			}
		}
	}

	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("提示: 使用 /help <command> 查看特定命令的详细帮助")
	fmt.Println()

	return nil
}

// showCommandHelp 显示特定命令的详细帮助
func (c *HelpCommand) showCommandHelp(cmdName string) error {
	registry := GetRegistry()

	// 去掉可能的前导斜杠
	cmdName = strings.TrimPrefix(cmdName, "/")

	// 先尝试直接获取命令
	cmd, exists := registry.Get(cmdName)
	if !exists {
		// 尝试通过别名查找
		cmd, exists = registry.GetByAlias(cmdName)
	}

	if !exists {
		// 列出匹配的命令建议
		allCmds := registry.List()
		var suggestions []string
		for _, c := range allCmds {
			if strings.Contains(c.Name(), cmdName) {
				suggestions = append(suggestions, c.Name())
			}
		}

		fmt.Printf("\n错误: 未知命令 '/%s'\n", cmdName)
		if len(suggestions) > 0 {
			fmt.Printf("\n您是否想查找:\n")
			for _, s := range suggestions {
				fmt.Printf("  /%s\n", s)
			}
		}
		fmt.Println("\n使用 /help 查看所有可用命令")
		return nil
	}

	// 显示命令详细信息
	fmt.Println()
	fmt.Println(cmd.Help())
	fmt.Println()

	// 额外信息
	aliases := cmd.Aliases()
	if len(aliases) > 0 {
		fmt.Printf("别名: %s\n", strings.Join(aliases, ", "))
	}
	fmt.Printf("分类: %s\n", cmd.Category().String())
	fmt.Println()

	return nil
}

func init() {
	// 注册帮助命令
	Register(NewHelpCommand())
}
