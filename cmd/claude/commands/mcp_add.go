package commands

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/Aspirin0000/claude-code-go/internal/config"
	"github.com/Aspirin0000/claude-code-go/internal/mcp"
	"github.com/Aspirin0000/claude-code-go/pkg/utils"
)

func init() {
	Register(NewMCPAddCommand())
}

// MCPAddCommand handles adding new MCP servers
type MCPAddCommand struct {
	*BaseCommand
	manager *mcp.MCPManagerImpl
}

// NewMCPAddCommand creates the /mcp-add command
func NewMCPAddCommand() *MCPAddCommand {
	return &MCPAddCommand{
		BaseCommand: NewBaseCommand(
			"mcp-add",
			"Add a new MCP server",
			CategoryMCP,
		).WithHelp(`Add a new MCP server configuration

Usage: /mcp-add <name> <command> [args...]
       /mcp-add <name> --type=sse <url>

Arguments:
  name              Server name (unique identifier)
  command           Command to execute (for stdio type)
  args              Optional command arguments

Options:
  --type=TYPE       Server type: stdio (default) or sse
  --url=URL         Server URL (required for sse type)
  --scope=SCOPE     Config scope: local, user, project (default: project)

Examples:
  /mcp-add filesystem npx -y @modelcontextprotocol/server-filesystem /path
  /mcp-add myserver --type=sse https://example.com/mcp
  /mcp-add toolserver node /path/to/server.js --scope=user

Interactive Mode:
  Run /mcp-add without arguments to enter interactive mode.`),
		manager: mcp.GetGlobalMCPManager(),
	}
}

// Execute runs the MCP add command
func (c *MCPAddCommand) Execute(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return c.interactiveMode()
	}

	name := args[0]
	if name == "" {
		return fmt.Errorf("server name is required")
	}

	// Parse flags and arguments
	serverType := "stdio"
	var serverURL string
	scope := mcp.ConfigScopeProject
	var cmdArgs []string

	i := 1
	for i < len(args) {
		arg := args[i]
		if strings.HasPrefix(arg, "--") {
			// Parse flags
			if strings.HasPrefix(arg, "--type=") {
				serverType = strings.TrimPrefix(arg, "--type=")
			} else if strings.HasPrefix(arg, "--url=") {
				serverURL = strings.TrimPrefix(arg, "--url=")
			} else if strings.HasPrefix(arg, "--scope=") {
				scopeStr := strings.TrimPrefix(arg, "--scope=")
				scope = mcp.ConfigScope(scopeStr)
			}
		} else {
			cmdArgs = append(cmdArgs, arg)
		}
		i++
	}

	// Validate based on type
	var serverConfig mcp.McpServerConfig
	switch serverType {
	case "stdio":
		if len(cmdArgs) == 0 {
			return fmt.Errorf("command is required for stdio type servers")
		}
		serverConfig = mcp.McpServerConfig{
			Type:    "stdio",
			Command: cmdArgs[0],
		}
		if len(cmdArgs) > 1 {
			serverConfig.Args = cmdArgs[1:]
		}
	case "sse":
		if serverURL == "" && len(cmdArgs) > 0 {
			serverURL = cmdArgs[0]
		}
		if serverURL == "" {
			return fmt.Errorf("URL is required for sse type servers")
		}
		serverConfig = mcp.McpServerConfig{
			Type: "sse",
			URL:  serverURL,
		}
	default:
		return fmt.Errorf("unsupported server type: %s", serverType)
	}

	return c.addServer(name, serverConfig, scope)
}

// interactiveMode prompts user for server details
func (c *MCPAddCommand) interactiveMode() error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("╔════════════════════════════════════════════════════════╗")
	fmt.Println("║              Add New MCP Server                        ║")
	fmt.Println("╚════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Get server name
	fmt.Print("Server name: ")
	name, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("server name is required")
	}

	// Get server type
	fmt.Print("Server type (stdio/sse) [stdio]: ")
	serverType, _ := reader.ReadString('\n')
	serverType = strings.TrimSpace(serverType)
	if serverType == "" {
		serverType = "stdio"
	}

	var serverConfig mcp.McpServerConfig

	switch serverType {
	case "stdio":
		fmt.Print("Command: ")
		command, _ := reader.ReadString('\n')
		command = strings.TrimSpace(command)
		if command == "" {
			return fmt.Errorf("command is required")
		}

		fmt.Print("Arguments (space-separated): ")
		argsStr, _ := reader.ReadString('\n')
		argsStr = strings.TrimSpace(argsStr)
		var args []string
		if argsStr != "" {
			args = strings.Fields(argsStr)
		}

		serverConfig = mcp.McpServerConfig{
			Type:    "stdio",
			Command: command,
			Args:    args,
		}

	case "sse":
		fmt.Print("Server URL: ")
		url, _ := reader.ReadString('\n')
		url = strings.TrimSpace(url)
		if url == "" {
			return fmt.Errorf("URL is required")
		}
		serverConfig = mcp.McpServerConfig{
			Type: "sse",
			URL:  url,
		}

	default:
		return fmt.Errorf("unsupported server type: %s", serverType)
	}

	// Get scope
	fmt.Print("Config scope (local/user/project) [project]: ")
	scopeStr, _ := reader.ReadString('\n')
	scopeStr = strings.TrimSpace(scopeStr)
	if scopeStr == "" {
		scopeStr = "project"
	}
	scope := mcp.ConfigScope(scopeStr)

	return c.addServer(name, serverConfig, scope)
}

// addServer adds a new MCP server
func (c *MCPAddCommand) addServer(name string, config mcp.McpServerConfig, scope mcp.ConfigScope) error {
	// Check if server already exists
	if _, exists := c.manager.GetServerStatus(name); exists {
		return fmt.Errorf("server '%s' already exists. Use /mcp-remove first if you want to replace it", name)
	}

	fmt.Printf("Adding MCP server '%s'...\n", name)

	// Create scoped config
	scopedConfig := mcp.ScopedMcpServerConfig{
		McpServerConfig: config,
		Scope:           scope,
	}

	// Add to manager
	if err := c.manager.AddServer(name, scopedConfig); err != nil {
		return fmt.Errorf("failed to add server: %w", err)
	}

	// Save to configuration file
	if err := c.saveToConfig(name, config, scope); err != nil {
		// Don't fail if save fails, just log it
		utils.LogForDebugging(fmt.Sprintf("Failed to save config: %v", err))
	}

	// Get updated status
	status, _ := c.manager.GetServerStatus(name)

	fmt.Println()
	if status != nil && status.Connected {
		fmt.Printf("✅ Server '%s' added and connected successfully!\n", name)

		// Show tool count
		if tools, err := c.manager.GetServerTools(name); err == nil {
			fmt.Printf("   Available tools: %d\n", len(tools))
		}
	} else {
		fmt.Printf("⚠️  Server '%s' added but not connected.\n", name)
		if status != nil && status.LastError != "" {
			fmt.Printf("   Error: %s\n", status.LastError)
		}
	}

	fmt.Println()
	fmt.Println("Commands:")
	fmt.Printf("  /mcp status %s  - Check server status\n", name)
	fmt.Println("  /mcp-list       - List all servers")

	return nil
}

// saveToConfig saves the MCP server to the configuration file
func (c *MCPAddCommand) saveToConfig(name string, serverConfig mcp.McpServerConfig, scope mcp.ConfigScope) error {
	// Load existing config
	cfgPath := config.GetConfigPath()
	cfg, err := config.Load(cfgPath)
	if err != nil {
		cfg = config.DefaultConfig()
	}

	// Get project path for project scope
	var projectPath string
	if scope == mcp.ConfigScopeProject {
		projectPath = getCurrentProjectPath()
	}

	// Get or create project config
	projectCfg, exists := cfg.Projects[projectPath]
	if !exists {
		projectCfg = config.ProjectConfig{
			AllowedTools: []string{},
			MCPServers:   make(map[string]config.MCPServerConfig),
		}
	}

	// Add MCP server
	projectCfg.MCPServers[name] = config.MCPServerConfig{
		Command: serverConfig.Command,
		Args:    serverConfig.Args,
		Env:     serverConfig.Env,
	}

	// Update projects
	cfg.Projects[projectPath] = projectCfg

	// Save config
	return cfg.Save(cfgPath)
}

// getCurrentProjectPath returns the current working directory as project path
func getCurrentProjectPath() string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return cwd
}
