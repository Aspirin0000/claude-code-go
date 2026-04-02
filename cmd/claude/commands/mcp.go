// Package commands provides MCP server management commands
// Source: src/commands/mcp/
// Refactor: Go MCP command system
package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/Aspirin0000/claude-code-go/internal/mcp"
)

func init() {
	Register(NewMCPCommand())
}

// MCPCommand handles MCP server management
type MCPCommand struct {
	*BaseCommand
	manager *mcp.MCPManagerImpl
}

// NewMCPCommand creates the /mcp command
func NewMCPCommand() *MCPCommand {
	return &MCPCommand{
		BaseCommand: NewBaseCommand(
			"mcp",
			"Manage MCP servers - show status, list servers, or get specific server details",
			CategoryMCP,
		).WithHelp(`MCP Server Management

Usage: /mcp                          Show MCP status overview
       /mcp list                     List all configured servers
       /mcp status <name>            Show specific server status

Examples:
  /mcp                    Display overall MCP status summary
  /mcp list              Show all MCP servers with their status
  /mcp status filesystem Check status of the 'filesystem' server`),
		manager: mcp.GetGlobalMCPManager(),
	}
}

// Execute runs the MCP command
func (c *MCPCommand) Execute(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return c.showOverview()
	}

	switch args[0] {
	case "list":
		return c.listServers()
	case "status":
		if len(args) < 2 {
			return fmt.Errorf("usage: /mcp status <server-name>")
		}
		return c.showStatus(args[1])
	default:
		return fmt.Errorf("unknown subcommand: %s. Use 'list' or 'status <name>'", args[0])
	}
}

// showOverview displays overall MCP status
func (c *MCPCommand) showOverview() error {
	statuses := c.manager.GetStatus()

	fmt.Println("╔════════════════════════════════════════════════════════╗")
	fmt.Println("║              MCP Server Status Overview                ║")
	fmt.Println("╚════════════════════════════════════════════════════════╝")
	fmt.Println()

	if len(statuses) == 0 {
		fmt.Println("No MCP servers configured.")
		fmt.Println()
		fmt.Println("Use /mcp-add to add a new MCP server.")
		return nil
	}

	var connected, failed, needsAuth, disabled, pending int

	for _, status := range statuses {
		switch status.Type {
		case mcp.MCPServerConnectionTypeConnected:
			connected++
		case mcp.MCPServerConnectionTypeFailed:
			failed++
		case mcp.MCPServerConnectionTypeNeedsAuth:
			needsAuth++
		case mcp.MCPServerConnectionTypeDisabled:
			disabled++
		case mcp.MCPServerConnectionTypePending:
			pending++
		}
	}

	fmt.Printf("Total Servers: %d\n", len(statuses))
	fmt.Println()
	fmt.Printf("  ✅ Connected:    %d\n", connected)
	fmt.Printf("  ❌ Failed:       %d\n", failed)
	fmt.Printf("  🔒 Needs Auth:   %d\n", needsAuth)
	fmt.Printf("  🚫 Disabled:     %d\n", disabled)
	fmt.Printf("  ⏳ Pending:      %d\n", pending)
	fmt.Println()

	// Show connected servers with tool counts
	if connected > 0 {
		fmt.Println("Connected Servers:")
		for _, status := range statuses {
			if status.Type == mcp.MCPServerConnectionTypeConnected {
				toolCount := 0
				if tools, err := c.manager.GetServerTools(status.Name); err == nil {
					toolCount = len(tools)
				}
				fmt.Printf("  • %s (%d tools)\n", status.Name, toolCount)
			}
		}
		fmt.Println()
	}

	// Show servers needing attention
	if needsAuth > 0 {
		fmt.Println("Servers Needing Authentication:")
		for _, status := range statuses {
			if status.Type == mcp.MCPServerConnectionTypeNeedsAuth {
				fmt.Printf("  • %s\n", status.Name)
			}
		}
		fmt.Println()
	}

	if failed > 0 {
		fmt.Println("Failed Servers:")
		for _, status := range statuses {
			if status.Type == mcp.MCPServerConnectionTypeFailed {
				fmt.Printf("  • %s", status.Name)
				if status.LastError != "" {
					fmt.Printf(" - %s", status.LastError)
				}
				fmt.Println()
			}
		}
		fmt.Println()
	}

	fmt.Println("Commands:")
	fmt.Println("  /mcp list           - List all servers")
	fmt.Println("  /mcp status <name>  - Show server details")
	fmt.Println("  /mcp-add            - Add a new server")
	fmt.Println("  /mcp-remove <name>  - Remove a server")

	return nil
}

// listServers lists all configured MCP servers
func (c *MCPCommand) listServers() error {
	statuses := c.manager.GetStatus()

	fmt.Println("╔════════════════════════════════════════════════════════╗")
	fmt.Println("║              MCP Server List                           ║")
	fmt.Println("╚════════════════════════════════════════════════════════╝")
	fmt.Println()

	if len(statuses) == 0 {
		fmt.Println("No MCP servers configured.")
		return nil
	}

	// Sort by status type
	order := map[mcp.MCPServerConnectionType]int{
		mcp.MCPServerConnectionTypeConnected: 0,
		mcp.MCPServerConnectionTypePending:   1,
		mcp.MCPServerConnectionTypeNeedsAuth: 2,
		mcp.MCPServerConnectionTypeFailed:    3,
		mcp.MCPServerConnectionTypeDisabled:  4,
	}

	for i := 0; i < len(statuses)-1; i++ {
		for j := i + 1; j < len(statuses); j++ {
			if order[statuses[i].Type] > order[statuses[j].Type] {
				statuses[i], statuses[j] = statuses[j], statuses[i]
			}
		}
	}

	for _, status := range statuses {
		var icon, statusText string
		switch status.Type {
		case mcp.MCPServerConnectionTypeConnected:
			icon = "✅"
			statusText = "Connected"
		case mcp.MCPServerConnectionTypeFailed:
			icon = "❌"
			statusText = "Failed"
		case mcp.MCPServerConnectionTypeNeedsAuth:
			icon = "🔒"
			statusText = "Needs Auth"
		case mcp.MCPServerConnectionTypeDisabled:
			icon = "🚫"
			statusText = "Disabled"
		case mcp.MCPServerConnectionTypePending:
			icon = "⏳"
			statusText = "Pending"
		}

		toolCount := 0
		if status.Type == mcp.MCPServerConnectionTypeConnected {
			if tools, err := c.manager.GetServerTools(status.Name); err == nil {
				toolCount = len(tools)
			}
		}

		fmt.Printf("%s %-20s %-12s", icon, status.Name, statusText)

		if toolCount > 0 {
			fmt.Printf(" (%d tools)", toolCount)
		}

		if status.Config.Scope != "" {
			fmt.Printf(" [%s]", status.Config.Scope)
		}

		fmt.Println()

		if status.LastError != "" && status.Type == mcp.MCPServerConnectionTypeFailed {
			fmt.Printf("   Error: %s\n", status.LastError)
		}
	}

	fmt.Println()
	fmt.Printf("Total: %d servers\n", len(statuses))

	return nil
}

// showStatus displays detailed status for a specific server
func (c *MCPCommand) showStatus(name string) error {
	status, exists := c.manager.GetServerStatus(name)
	if !exists {
		return fmt.Errorf("server '%s' not found", name)
	}

	fmt.Println("╔════════════════════════════════════════════════════════╗")
	fmt.Printf("║  MCP Server: %-36s  ║\n", name)
	fmt.Println("╚════════════════════════════════════════════════════════╝")
	fmt.Println()

	var statusText string
	switch status.Type {
	case mcp.MCPServerConnectionTypeConnected:
		statusText = "✅ Connected"
	case mcp.MCPServerConnectionTypeFailed:
		statusText = "❌ Failed"
	case mcp.MCPServerConnectionTypeNeedsAuth:
		statusText = "🔒 Needs Authentication"
	case mcp.MCPServerConnectionTypeDisabled:
		statusText = "🚫 Disabled"
	case mcp.MCPServerConnectionTypePending:
		statusText = "⏳ Pending"
	}

	fmt.Printf("Status:      %s\n", statusText)
	fmt.Printf("Type:        %s\n", status.Config.Type)
	fmt.Printf("Scope:       %s\n", status.Config.Scope)
	fmt.Printf("Last Check:  %s\n", status.LastChecked.Format("2006-01-02 15:04:05"))

	if status.Config.Command != "" {
		fmt.Printf("Command:     %s\n", status.Config.Command)
		if len(status.Config.Args) > 0 {
			fmt.Printf("Arguments:   %s\n", strings.Join(status.Config.Args, " "))
		}
	}

	if status.Config.URL != "" {
		fmt.Printf("URL:         %s\n", status.Config.URL)
	}

	if status.LastError != "" {
		fmt.Printf("\nError: %s\n", status.LastError)
	}

	// Show tools if connected
	if status.Type == mcp.MCPServerConnectionTypeConnected {
		tools, err := c.manager.GetServerTools(name)
		if err == nil && len(tools) > 0 {
			fmt.Printf("\nAvailable Tools (%d):\n", len(tools))
			for _, tool := range tools {
				desc := tool.Description
				if len(desc) > 60 {
					desc = desc[:57] + "..."
				}
				fmt.Printf("  • %-25s %s\n", tool.Name, desc)
			}
		}
	}

	fmt.Println()
	fmt.Println("Commands:")
	fmt.Printf("  /mcp-remove %s  - Remove this server\n", name)

	return nil
}
