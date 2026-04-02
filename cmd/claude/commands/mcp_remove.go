package commands

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/Aspirin0000/claude-code-go/internal/config"
	"github.com/Aspirin0000/claude-code-go/internal/mcp"
)

func init() {
	Register(NewMCPRemoveCommand())
}

// MCPRemoveCommand handles removing MCP servers
type MCPRemoveCommand struct {
	*BaseCommand
	manager *mcp.MCPManagerImpl
}

// NewMCPRemoveCommand creates the /mcp-remove command
func NewMCPRemoveCommand() *MCPRemoveCommand {
	return &MCPRemoveCommand{
		BaseCommand: NewBaseCommand(
			"mcp-remove",
			"Remove an MCP server configuration",
			CategoryMCP,
		).WithHelp(`Remove an MCP server from configuration

Usage: /mcp-remove <name>

Arguments:
  name    Name of the MCP server to remove

Description:
  This command will:
  1. Disconnect from the server if currently connected
  2. Remove the server from the configuration
  3. Clean up any associated resources

You will be asked to confirm before the server is removed.

Examples:
  /mcp-remove filesystem   Remove the 'filesystem' server
  /mcp-remove myserver     Remove the 'myserver' server

Note: This action cannot be undone. The server configuration
will be permanently deleted.`),
		manager: mcp.GetGlobalMCPManager(),
	}
}

// Execute runs the MCP remove command
func (c *MCPRemoveCommand) Execute(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: /mcp-remove <server-name>")
	}

	name := args[0]

	// Check if server exists
	status, exists := c.manager.GetServerStatus(name)
	if !exists {
		return fmt.Errorf("server '%s' not found", name)
	}

	// Show server details
	fmt.Println("╔════════════════════════════════════════════════════════╗")
	fmt.Printf("║  Remove MCP Server: %-30s    ║\n", name)
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

	fmt.Printf("Status:  %s\n", statusText)
	fmt.Printf("Type:    %s\n", status.Config.Type)
	fmt.Printf("Scope:   %s\n", status.Config.Scope)

	if status.Config.Command != "" {
		fmt.Printf("Command: %s\n", status.Config.Command)
	}
	if status.Config.URL != "" {
		fmt.Printf("URL:     %s\n", status.Config.URL)
	}

	// Get tool count if connected
	if status.Type == mcp.MCPServerConnectionTypeConnected {
		if tools, err := c.manager.GetServerTools(name); err == nil {
			fmt.Printf("Tools:   %d available\n", len(tools))
		}
	}

	fmt.Println()

	// Ask for confirmation
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Are you sure you want to remove this server? [y/N]: ")
	response, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	response = strings.TrimSpace(strings.ToLower(response))
	if response != "y" && response != "yes" {
		fmt.Println("Removal cancelled.")
		return nil
	}

	// Remove the server
	fmt.Printf("\nRemoving server '%s'...\n", name)

	if err := c.manager.RemoveServer(name); err != nil {
		return fmt.Errorf("failed to remove server: %w", err)
	}

	// Remove from config file
	if err := c.removeFromConfig(name); err != nil {
		// Log but don't fail
		fmt.Printf("Warning: Could not remove from config file: %v\n", err)
	}

	fmt.Printf("✅ Server '%s' has been removed successfully.\n", name)

	return nil
}

// removeFromConfig removes the MCP server from the configuration file
func (c *MCPRemoveCommand) removeFromConfig(name string) error {
	// Load existing config
	cfgPath := config.GetConfigPath()
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return err
	}

	// Check all projects for this server
	removed := false
	for projectPath, projectCfg := range cfg.Projects {
		if _, exists := projectCfg.MCPServers[name]; exists {
			delete(projectCfg.MCPServers, name)
			cfg.Projects[projectPath] = projectCfg
			removed = true
		}
	}

	if !removed {
		return fmt.Errorf("server not found in any project configuration")
	}

	// Save config
	return cfg.Save(cfgPath)
}
