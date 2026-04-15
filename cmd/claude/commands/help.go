package commands

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/Aspirin0000/claude-code-go/internal/tools"
)

// HelpCommand displays help information
type HelpCommand struct {
	*BaseCommand
}

// NewHelpCommand creates a new help command
func NewHelpCommand() *HelpCommand {
	base := NewBaseCommand(
		"help",
		"Display list of all commands or detailed help for a specific command",
		CategoryGeneral,
	).WithAliases("h", "?").WithHelp(`Usage: /help [command]

Display help information for commands.

Arguments:
  command  Optional. Command name to view detailed help

Examples:
  /help          Show list of all available commands
  /help clear    Show detailed help for the clear command
  /h             Same as /help
  /?             Same as /help

Related:
  /tools         List all available AI tools
  /sessions      Manage saved sessions`)

	return &HelpCommand{BaseCommand: base}
}

// Execute runs the help command
func (c *HelpCommand) Execute(ctx context.Context, args []string) error {
	if len(args) > 0 {
		return c.showCommandHelp(args[0])
	}
	return c.showAllCommands()
}

// showAllCommands displays all commands grouped by category
func (c *HelpCommand) showAllCommands() error {
	registry := GetRegistry()

	// Get all categories in order
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

	fmt.Println("\nAvailable Commands:")
	fmt.Println(strings.Repeat("=", 60))

	for _, cat := range categories {
		cmds := registry.ListByCategory(cat)
		if len(cmds) == 0 {
			continue
		}

		// Sort commands by name
		sort.Slice(cmds, func(i, j int) bool {
			return cmds[i].Name() < cmds[j].Name()
		})

		fmt.Printf("\n%s:\n", cat.String())
		fmt.Println(strings.Repeat("-", 30))

		for _, cmd := range cmds {
			name := cmd.Name()
			desc := cmd.Description()
			aliases := cmd.Aliases()

			// Format display
			if len(aliases) > 0 {
				aliasStr := fmt.Sprintf(" (aliases: %s)", strings.Join(aliases, ", "))
				fmt.Printf("  /%-15s %s%s\n", name, desc, aliasStr)
			} else {
				fmt.Printf("  /%-15s %s\n", name, desc)
			}
		}
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("Tips:")
	fmt.Println("  • Use /help <command> for detailed help on a specific command")
	fmt.Println("  • Use /tools to see available AI tools")
	fmt.Println("  • Use /sessions to manage saved conversations")
	fmt.Println()

	return nil
}

// showCommandHelp displays detailed help for a specific command
func (c *HelpCommand) showCommandHelp(cmdName string) error {
	registry := GetRegistry()

	// Remove possible leading slash
	cmdName = strings.TrimPrefix(cmdName, "/")

	// Try to get command directly first
	cmd, exists := registry.Get(cmdName)
	if !exists {
		// Try to find by alias
		cmd, exists = registry.GetByAlias(cmdName)
	}

	if !exists {
		// List matching command suggestions
		allCmds := registry.List()
		var suggestions []string
		for _, c := range allCmds {
			if strings.Contains(c.Name(), cmdName) {
				suggestions = append(suggestions, c.Name())
			}
		}

		fmt.Printf("\nError: Unknown command '/%s'\n", cmdName)
		if len(suggestions) > 0 {
			fmt.Printf("\nDid you mean:\n")
			for _, s := range suggestions {
				fmt.Printf("  /%s\n", s)
			}
		}
		fmt.Println("\nUse /help to see all available commands")
		return nil
	}

	// Display command details
	fmt.Println()
	fmt.Println(cmd.Help())
	fmt.Println()

	// Additional info
	aliases := cmd.Aliases()
	if len(aliases) > 0 {
		fmt.Printf("Aliases: %s\n", strings.Join(aliases, ", "))
	}
	fmt.Printf("Category: %s\n", cmd.Category().String())
	fmt.Println()

	return nil
}

// showQuickHelp displays a quick reference
func (c *HelpCommand) showQuickHelp() {
	fmt.Print(`
╔══════════════════════════════════════════════════════════╗
║                    QUICK REFERENCE                       ║
╠══════════════════════════════════════════════════════════╣
║  Core Commands:    /help, /exit, /clear, /status        ║
║  Conversation:     /save, /load, /sessions, /compact    ║
║  Configuration:    /config, /model, /init, /doctor      ║
║  Tools:            /bash, /git, /grep, /glob            ║
║  MCP:              /mcp, /mcp-add, /mcp-list            ║
╚══════════════════════════════════════════════════════════╝
`)
}

// ToolsCommand lists available AI tools
type ToolsCommand struct {
	*BaseCommand
}

// NewToolsCommand creates the tools command
func NewToolsCommand() *ToolsCommand {
	return &ToolsCommand{
		BaseCommand: NewBaseCommand(
			"tools",
			"List all available AI tools",
			CategoryTools,
		).WithAliases("list-tools", "t").
			WithHelp(`Usage: /tools [category]

List all available AI tools that Claude can use.

Arguments:
  category  Optional. Filter by category: core, file, search, mcp

Examples:
  /tools           List all tools
  /tools core      Show core tools only
  /tools file      Show file operation tools

Tool Categories:
  Core Tools     - bash, file_read, file_write, file_edit
  Search Tools   - grep, glob
  Task Tools     - todo_write, task_*, notebook_edit
  MCP Tools      - Available when MCP servers are connected
  Web Tools      - web_search, web_fetch

Note: Available tools depend on current permission level.
Use /permissions to check or modify permissions.`),
	}
}

// Execute runs the tools command
func (c *ToolsCommand) Execute(ctx context.Context, args []string) error {
	category := ""
	if len(args) > 0 {
		category = strings.ToLower(args[0])
	}

	registry := tools.NewDefaultRegistry()
	allTools := registry.List()

	if len(allTools) == 0 {
		fmt.Println("No tools available.")
		return nil
	}

	fmt.Println("\nAvailable AI Tools:")
	fmt.Println(strings.Repeat("=", 70))

	// Group tools by category
	coreTools := []tools.Tool{}
	searchTools := []tools.Tool{}
	taskTools := []tools.Tool{}
	webTools := []tools.Tool{}
	otherTools := []tools.Tool{}

	for _, tool := range allTools {
		name := tool.Name()
		switch {
		case name == "bash" || name == "file_read" || name == "file_write" || name == "file_edit":
			coreTools = append(coreTools, tool)
		case name == "grep" || name == "glob":
			searchTools = append(searchTools, tool)
		case strings.HasPrefix(name, "task_") || name == "todo_write" || name == "notebook_edit":
			taskTools = append(taskTools, tool)
		case name == "web_search" || name == "web_fetch":
			webTools = append(webTools, tool)
		default:
			otherTools = append(otherTools, tool)
		}
	}

	// Display based on category filter or all
	switch category {
	case "core", "file":
		c.displayToolGroup("Core Tools", coreTools)
	case "search":
		c.displayToolGroup("Search Tools", searchTools)
	case "task":
		c.displayToolGroup("Task Management", taskTools)
	case "web":
		c.displayToolGroup("Web Tools", webTools)
	default:
		c.displayToolGroup("Core Tools", coreTools)
		c.displayToolGroup("Search Tools", searchTools)
		c.displayToolGroup("Task Management", taskTools)
		c.displayToolGroup("Web Tools", webTools)
		if len(otherTools) > 0 {
			c.displayToolGroup("Other Tools", otherTools)
		}
	}

	fmt.Println()
	fmt.Printf("Total: %d tool(s) available\n", len(allTools))
	fmt.Println("\nUse /permissions to check which tools are allowed")

	return nil
}

func (c *ToolsCommand) displayToolGroup(name string, toolsList []tools.Tool) {
	if len(toolsList) == 0 {
		return
	}

	fmt.Printf("\n%s:\n", name)
	fmt.Println(strings.Repeat("-", 40))

	for _, tool := range toolsList {
		desc := tool.Description()
		readOnly := ""
		destructive := ""

		if tool.IsReadOnly() {
			readOnly = " [RO]"
		}
		if tool.IsDestructive() {
			destructive = " [!]"
		}

		fmt.Printf("  %-20s %s%s%s\n", tool.Name(), desc, readOnly, destructive)
	}
}

func init() {
	Register(NewHelpCommand())
	Register(NewToolsCommand())
}
