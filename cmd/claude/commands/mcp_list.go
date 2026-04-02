package commands

import (
	"context"
	"fmt"

	"github.com/Aspirin0000/claude-code-go/internal/mcp"
)

func init() {
	Register(NewMCPListCommand())
}

// MCPListCommand handles listing MCP servers
type MCPListCommand struct {
	*BaseCommand
	manager *mcp.MCPManagerImpl
}

// NewMCPListCommand creates the /mcp-list command
func NewMCPListCommand() *MCPListCommand {
	return &MCPListCommand{
		BaseCommand: NewBaseCommand(
			"mcp-list",
			"List all MCP servers with status and tool counts",
			CategoryMCP,
		).WithAliases("mcps").
			WithHelp(`List all configured MCP servers

Usage: /mcp-list
       /mcps

Displays all MCP servers with:
  • Connection status (connected/disabled/auth-needed)
  • Number of available tools per server
  • Server configuration scope
  • Error messages for failed servers

Columns:
  Status      Current connection state
  Name        Server identifier
  Tools       Number of available tools (if connected)
  Type        Server type (stdio/sse)
  Scope       Configuration scope

Examples:
  /mcp-list   Show all MCP servers
  /mcps       Same as /mcp-list

See also:
  /mcp status <name>  - Detailed server information
  /mcp-add            - Add a new server
  /mcp-remove <name>  - Remove a server`),
		manager: mcp.GetGlobalMCPManager(),
	}
}

// Execute runs the MCP list command
func (c *MCPListCommand) Execute(ctx context.Context, args []string) error {
	statuses := c.manager.GetStatus()

	fmt.Println("╔════════════════════════════════════════════════════════════════╗")
	fmt.Println("║                    MCP Servers                                 ║")
	fmt.Println("╚════════════════════════════════════════════════════════════════╝")
	fmt.Println()

	if len(statuses) == 0 {
		fmt.Println("No MCP servers configured.")
		fmt.Println()
		fmt.Println("Use /mcp-add to add a new MCP server.")
		return nil
	}

	// Define status order for sorting
	statusOrder := map[mcp.MCPServerConnectionType]int{
		mcp.MCPServerConnectionTypeConnected: 0,
		mcp.MCPServerConnectionTypePending:   1,
		mcp.MCPServerConnectionTypeNeedsAuth: 2,
		mcp.MCPServerConnectionTypeFailed:    3,
		mcp.MCPServerConnectionTypeDisabled:  4,
	}

	// Simple bubble sort by status order
	for i := 0; i < len(statuses)-1; i++ {
		for j := i + 1; j < len(statuses); j++ {
			if statusOrder[statuses[i].Type] > statusOrder[statuses[j].Type] {
				statuses[i], statuses[j] = statuses[j], statuses[i]
			}
		}
	}

	// Print header
	fmt.Printf("%-10s %-20s %-8s %-10s %-10s\n", "STATUS", "NAME", "TOOLS", "TYPE", "SCOPE")

	// Print separator
	for i := 0; i < 65; i++ {
		fmt.Print("-")
	}
	fmt.Println()

	var totalTools int
	var connectedCount int

	for _, status := range statuses {
		var statusIcon, statusText string
		switch status.Type {
		case mcp.MCPServerConnectionTypeConnected:
			statusIcon = "✅"
			statusText = "connected"
			connectedCount++
		case mcp.MCPServerConnectionTypeFailed:
			statusIcon = "❌"
			statusText = "failed"
		case mcp.MCPServerConnectionTypeNeedsAuth:
			statusIcon = "🔒"
			statusText = "auth-needed"
		case mcp.MCPServerConnectionTypeDisabled:
			statusIcon = "🚫"
			statusText = "disabled"
		case mcp.MCPServerConnectionTypePending:
			statusIcon = "⏳"
			statusText = "pending"
		}

		// Get tool count for connected servers
		toolCount := "-"
		if status.Type == mcp.MCPServerConnectionTypeConnected {
			if tools, err := c.manager.GetServerTools(status.Name); err == nil {
				toolCount = fmt.Sprintf("%d", len(tools))
				totalTools += len(tools)
			}
		}

		serverType := status.Config.Type
		if serverType == "" {
			serverType = "stdio"
		}

		scope := string(status.Config.Scope)
		if scope == "" {
			scope = "-"
		}

		fmt.Printf("%-2s %-7s %-20s %-8s %-10s %-10s\n",
			statusIcon,
			statusText,
			status.Name,
			toolCount,
			serverType,
			scope,
		)

		// Show error for failed servers
		if status.Type == mcp.MCPServerConnectionTypeFailed && status.LastError != "" {
			errMsg := status.LastError
			if len(errMsg) > 50 {
				errMsg = errMsg[:47] + "..."
			}
			fmt.Printf("           Error: %s\n", errMsg)
		}
	}

	// Print summary
	fmt.Println()
	fmt.Printf("Total: %d servers", len(statuses))
	if connectedCount > 0 {
		fmt.Printf(" (%d connected, %d tools available)", connectedCount, totalTools)
	}
	fmt.Println()

	return nil
}
