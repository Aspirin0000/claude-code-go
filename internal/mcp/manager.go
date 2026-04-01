// Package mcp provides MCP (Model Context Protocol) service management
// Source: src/services/mcp/manager.ts
// Batch: C-8/8 - Final Integration and Manager
package mcp

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Aspirin0000/claude-code-go/internal/api"
	"github.com/Aspirin0000/claude-code-go/internal/config"
	"github.com/Aspirin0000/claude-code-go/pkg/utils"
)

// ============================================================================
// ToolExecutor - Tool execution coordinator
// ============================================================================

// ToolExecutor coordinates tool execution across multiple MCP servers
type ToolExecutor struct {
	mu      sync.RWMutex
	servers map[string]*MCPClient
}

// NewToolExecutor creates a new tool executor
func NewToolExecutor() *ToolExecutor {
	return &ToolExecutor{
		servers: make(map[string]*MCPClient),
	}
}

// RegisterServer registers a server for tool execution
func (te *ToolExecutor) RegisterServer(name string, client *MCPClient) {
	te.mu.Lock()
	defer te.mu.Unlock()
	te.servers[name] = client
}

// UnregisterServer unregisters a server
func (te *ToolExecutor) UnregisterServer(name string) {
	te.mu.Lock()
	defer te.mu.Unlock()
	delete(te.servers, name)
}

// ExecuteTool executes a tool on the appropriate server
// Tool name format: mcp__<server>__<tool>
func (te *ToolExecutor) ExecuteTool(
	ctx context.Context,
	toolName string,
	arguments map[string]interface{},
) (*CallToolResult, error) {
	// Parse tool name to extract server and actual tool name
	serverName, actualToolName, err := te.parseToolName(toolName)
	if err != nil {
		return nil, err
	}

	te.mu.RLock()
	client, exists := te.servers[serverName]
	te.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("server '%s' not found for tool '%s'", serverName, toolName)
	}

	if !client.IsConnected() {
		return nil, fmt.Errorf("server '%s' is not connected", serverName)
	}

	// Execute with timeout from context or default
	result, err := client.CallTool(actualToolName, arguments)
	if err != nil {
		return nil, fmt.Errorf("tool execution failed on server '%s': %w", serverName, err)
	}

	return result, nil
}

// parseToolName parses a tool name into server and actual tool name
// Format: mcp__<server>__<tool> or just <tool>
func (te *ToolExecutor) parseToolName(toolName string) (serverName string, actualTool string, err error) {
	if !strings.HasPrefix(toolName, "mcp__") {
		// Not an MCP tool
		return "", toolName, nil
	}

	parts := strings.SplitN(toolName, "__", 3)
	if len(parts) < 3 {
		return "", "", fmt.Errorf("invalid MCP tool name format: %s", toolName)
	}

	return parts[1], parts[2], nil
}

// FindServerForTool finds which server provides a specific tool
func (te *ToolExecutor) FindServerForTool(toolName string) (string, bool) {
	serverName, _, err := te.parseToolName(toolName)
	if err != nil {
		return "", false
	}

	te.mu.RLock()
	_, exists := te.servers[serverName]
	te.mu.RUnlock()

	return serverName, exists
}

// ============================================================================
// MCPManagerImpl - Main MCP manager
// ============================================================================

// ServerStatus represents the status of an MCP server
type ServerStatus struct {
	Name        string                  `json:"name"`
	Type        MCPServerConnectionType `json:"type"`
	Connected   bool                    `json:"connected"`
	LastError   string                  `json:"lastError,omitempty"`
	ToolCount   int                     `json:"toolCount,omitempty"`
	Config      ScopedMcpServerConfig   `json:"config"`
	LastChecked time.Time               `json:"lastChecked"`
}

// MCPManagerImpl ties all MCP functionality together
type MCPManagerImpl struct {
	connectionManager *ConnectionManager
	cache             *MCPCache
	executor          *ToolExecutor
	authProviders     map[string]*ClaudeAuthProvider
	configs           map[string]ScopedMcpServerConfig
	status            map[string]*ServerStatus
	mu                sync.RWMutex
	ctx               context.Context
	cancel            context.CancelFunc
}

// NewMCPManager creates a new MCP manager
func NewMCPManager() *MCPManagerImpl {
	ctx, cancel := context.WithCancel(context.Background())
	return &MCPManagerImpl{
		connectionManager: NewConnectionManager(),
		cache:             NewMCPCache(),
		executor:          NewToolExecutor(),
		authProviders:     make(map[string]*ClaudeAuthProvider),
		configs:           make(map[string]ScopedMcpServerConfig),
		status:            make(map[string]*ServerStatus),
		ctx:               ctx,
		cancel:            cancel,
	}
}

// Initialize sets up the MCP manager from configuration
func (m *MCPManagerImpl) Initialize(cfg *config.Config) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Load all MCP server configurations
	allConfigs := m.loadAllConfigs(cfg)

	// Apply policy filtering
	allowedConfigs, blockedServers := m.applyPolicyFilter(allConfigs)
	if len(blockedServers) > 0 {
		utils.LogForDebugging(fmt.Sprintf("Blocked %d servers by policy: %v", len(blockedServers), blockedServers))
	}

	// Filter out disabled servers
	for name, serverCfg := range allowedConfigs {
		if IsMcpServerDisabled(name) {
			m.status[name] = &ServerStatus{
				Name:        name,
				Type:        MCPServerConnectionTypeDisabled,
				Connected:   false,
				Config:      serverCfg,
				LastChecked: time.Now(),
			}
			continue
		}

		m.configs[name] = serverCfg
	}

	utils.LogEvent("tengu_mcp_manager_initialized", map[string]interface{}{
		"serverCount": len(m.configs),
		"disabled":    len(allowedConfigs) - len(m.configs),
	})

	return nil
}

// loadAllConfigs loads all MCP configurations from various sources
func (m *MCPManagerImpl) loadAllConfigs(cfg *config.Config) map[string]ScopedMcpServerConfig {
	allConfigs := make(map[string]ScopedMcpServerConfig)

	// 1. Load from project config
	if cfg != nil {
		for projectPath, projectCfg := range cfg.Projects {
			for name, serverCfg := range projectCfg.MCPServers {
				fullName := name
				if projectPath != "" {
					fullName = fmt.Sprintf("%s/%s", projectPath, name)
				}
				allConfigs[fullName] = ScopedMcpServerConfig{
					McpServerConfig: McpServerConfig{
						Type:    "stdio",
						Command: serverCfg.Command,
						Args:    serverCfg.Args,
						Env:     serverCfg.Env,
					},
					Scope: ConfigScopeProject,
				}
			}
		}
	}

	// 2. Load from environment or other sources
	// This can be extended to load from:
	// - Local config files
	// - User config
	// - Global config
	// - ClaudeAI managed configs
	// - Plugin configs

	return allConfigs
}

// applyPolicyFilter applies policy filtering to server configurations
func (m *MCPManagerImpl) applyPolicyFilter(
	configs map[string]ScopedMcpServerConfig,
) (map[string]ScopedMcpServerConfig, []string) {
	allowed := make(map[string]ScopedMcpServerConfig)
	blocked := []string{}

	for name, cfg := range configs {
		if IsMcpServerAllowedByPolicy(name, &cfg.McpServerConfig) {
			allowed[name] = cfg
		} else {
			blocked = append(blocked, name)
		}
	}

	return allowed, blocked
}

// ConnectAll connects to all configured servers
func (m *MCPManagerImpl) ConnectAll() []BatchConnectResult {
	m.mu.RLock()
	configs := make(map[string]ScopedMcpServerConfig, len(m.configs))
	for name, cfg := range m.configs {
		configs[name] = cfg
	}
	m.mu.RUnlock()

	if len(configs) == 0 {
		return nil
	}

	results := m.connectionManager.ConnectBatch(m.ctx, configs, 0)

	// Process results and update executor
	for _, result := range results {
		m.processConnectResult(result)
	}

	return results
}

// processConnectResult processes a connection result
func (m *MCPManagerImpl) processConnectResult(result BatchConnectResult) {
	m.mu.Lock()
	defer m.mu.Unlock()

	status := &ServerStatus{
		Name:        result.Name,
		LastChecked: time.Now(),
	}

	if result.Error != nil {
		status.Type = MCPServerConnectionTypeFailed
		status.Connected = false
		status.LastError = result.Error.Error()
	} else {
		switch conn := result.Connection.(type) {
		case *ConnectedMCPServer:
			status.Type = MCPServerConnectionTypeConnected
			status.Connected = true
			if client, ok := conn.Client.(*MCPClient); ok {
				m.executor.RegisterServer(result.Name, client)
			}
		case *NeedsAuthMCPServer:
			status.Type = MCPServerConnectionTypeNeedsAuth
			status.Connected = false
		case *FailedMCPServer:
			status.Type = MCPServerConnectionTypeFailed
			status.Connected = false
			if conn.Error != nil {
				status.LastError = *conn.Error
			}
		case *PendingMCPServer:
			status.Type = MCPServerConnectionTypePending
			status.Connected = false
		}
	}

	if cfg, exists := m.configs[result.Name]; exists {
		status.Config = cfg
	}

	m.status[result.Name] = status
}

// GetAllTools aggregates tools from all connected servers
func (m *MCPManagerImpl) GetAllTools() ([]api.Tool, error) {
	connected := m.connectionManager.ListConnected()

	var allTools []api.Tool
	var mu sync.Mutex
	var errs []error

	for _, serverName := range connected {
		client, exists := m.connectionManager.GetClient(serverName)
		if !exists {
			continue
		}

		tools, err := FetchToolsForClient(client)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to fetch tools from '%s': %w", serverName, err))
			continue
		}

		mu.Lock()
		allTools = append(allTools, tools...)
		mu.Unlock()
	}

	if len(errs) > 0 {
		// Log errors but return tools we could get
		for _, err := range errs {
			LogMCPDebug("manager", err.Error())
		}
	}

	return allTools, nil
}

// ExecuteTool finds the appropriate server and executes a tool
func (m *MCPManagerImpl) ExecuteTool(
	ctx context.Context,
	toolName string,
	arguments map[string]interface{},
) (*CallToolResult, error) {
	// First try the executor
	result, err := m.executor.ExecuteTool(ctx, toolName, arguments)
	if err == nil {
		return result, nil
	}

	// If executor fails, try to find server and execute directly
	serverName, actualToolName, err := m.parseToolName(toolName)
	if err != nil {
		return nil, err
	}

	client, exists := m.connectionManager.GetClient(serverName)
	if !exists {
		return nil, fmt.Errorf("server '%s' not found", serverName)
	}

	if !client.IsConnected() {
		// Try to reconnect
		m.mu.RLock()
		config, hasConfig := m.configs[serverName]
		m.mu.RUnlock()

		if !hasConfig {
			return nil, fmt.Errorf("server '%s' configuration not found", serverName)
		}

		_, err := m.connectionManager.ConnectServer(ctx, serverName, config, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to reconnect to server '%s': %w", serverName, err)
		}

		// Re-register with executor
		m.executor.RegisterServer(serverName, client)
	}

	return client.CallTool(actualToolName, arguments)
}

// parseToolName parses a tool name into server and actual tool name
func (m *MCPManagerImpl) parseToolName(toolName string) (serverName string, actualTool string, err error) {
	if !strings.HasPrefix(toolName, "mcp__") {
		return "", toolName, nil
	}

	parts := strings.SplitN(toolName, "__", 3)
	if len(parts) < 3 {
		return "", "", fmt.Errorf("invalid MCP tool name format: %s", toolName)
	}

	return parts[1], parts[2], nil
}

// AddServer adds and connects a new server
func (m *MCPManagerImpl) AddServer(name string, config ScopedMcpServerConfig) error {
	m.mu.Lock()
	m.configs[name] = config
	m.mu.Unlock()

	// Check policy
	if !IsMcpServerAllowedByPolicy(name, &config.McpServerConfig) {
		return fmt.Errorf("server '%s' is not allowed by policy", name)
	}

	// Check if disabled
	if IsMcpServerDisabled(name) {
		m.mu.Lock()
		m.status[name] = &ServerStatus{
			Name:        name,
			Type:        MCPServerConnectionTypeDisabled,
			Connected:   false,
			Config:      config,
			LastChecked: time.Now(),
		}
		m.mu.Unlock()
		return fmt.Errorf("server '%s' is disabled", name)
	}

	// Connect to the server
	conn, err := m.connectionManager.ConnectServer(m.ctx, name, config, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to server '%s': %w", name, err)
	}

	// Process result
	m.processConnectResult(BatchConnectResult{
		Name:       name,
		Connection: conn,
		Error:      nil,
	})

	return nil
}

// RemoveServer disconnects and removes a server
func (m *MCPManagerImpl) RemoveServer(name string) error {
	// Disconnect from the server
	if err := m.connectionManager.DisconnectServer(name); err != nil {
		LogMCPDebug(name, fmt.Sprintf("Error disconnecting: %v", err))
	}

	// Unregister from executor
	m.executor.UnregisterServer(name)

	// Clear cache for this server
	ClearClientCache(name)

	m.mu.Lock()
	defer m.mu.Unlock()

	// Remove from configs
	delete(m.configs, name)

	// Update status
	delete(m.status, name)

	// Remove auth provider if exists
	if provider, exists := m.authProviders[name]; exists {
		_ = provider.Revoke()
		delete(m.authProviders, name)
	}

	return nil
}

// RestartServer disconnects and reconnects a server
func (m *MCPManagerImpl) RestartServer(name string) error {
	m.mu.RLock()
	config, exists := m.configs[name]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("server '%s' not found", name)
	}

	// Disconnect
	if err := m.connectionManager.DisconnectServer(name); err != nil {
		LogMCPDebug(name, fmt.Sprintf("Error disconnecting during restart: %v", err))
	}

	m.executor.UnregisterServer(name)

	// Reconnect
	conn, err := m.connectionManager.ConnectServer(m.ctx, name, config, nil)
	if err != nil {
		return fmt.Errorf("failed to reconnect to server '%s': %w", name, err)
	}

	m.processConnectResult(BatchConnectResult{
		Name:       name,
		Connection: conn,
		Error:      nil,
	})

	return nil
}

// GetStatus returns the status of all servers
func (m *MCPManagerImpl) GetStatus() []ServerStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	statuses := make([]ServerStatus, 0, len(m.status))
	for _, status := range m.status {
		statuses = append(statuses, *status)
	}

	return statuses
}

// GetServerStatus returns the status of a specific server
func (m *MCPManagerImpl) GetServerStatus(name string) (*ServerStatus, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status, exists := m.status[name]
	if !exists {
		return nil, false
	}

	// Make a copy
	statusCopy := *status
	return &statusCopy, true
}

// HealthCheck verifies all connections and updates status
func (m *MCPManagerImpl) HealthCheck() []ServerStatus {
	m.mu.Lock()
	defer m.mu.Unlock()

	results := make([]ServerStatus, 0, len(m.configs))

	for name, config := range m.configs {
		status := &ServerStatus{
			Name:        name,
			LastChecked: time.Now(),
			Config:      config,
		}

		// Check if connected
		if client, exists := m.connectionManager.GetClient(name); exists && client.IsConnected() {
			status.Connected = true
			status.Type = MCPServerConnectionTypeConnected

			// Try to get tool count
			tools, err := client.GetTools()
			if err == nil {
				status.ToolCount = len(tools)
			}
		} else {
			status.Connected = false
			if existingStatus, exists := m.status[name]; exists {
				status.Type = existingStatus.Type
				status.LastError = existingStatus.LastError
			} else {
				status.Type = MCPServerConnectionTypeFailed
			}
		}

		m.status[name] = status
		results = append(results, *status)
	}

	return results
}

// Shutdown cleans up all resources
func (m *MCPManagerImpl) Shutdown() error {
	m.cancel()

	// Disconnect all servers
	m.connectionManager.DisconnectAll()

	// Clear executor
	m.executor = NewToolExecutor()

	// Clear all cache
	ClearAllCache()

	m.mu.Lock()
	defer m.mu.Unlock()

	// Revoke all auth tokens
	for name, provider := range m.authProviders {
		_ = provider.Revoke()
		delete(m.authProviders, name)
	}

	// Clear configs and status
	m.configs = make(map[string]ScopedMcpServerConfig)
	m.status = make(map[string]*ServerStatus)

	return nil
}

// GetConnectedServers returns a list of connected server names
func (m *MCPManagerImpl) GetConnectedServers() []string {
	return m.connectionManager.ListConnected()
}

// IsServerConnected checks if a server is connected
func (m *MCPManagerImpl) IsServerConnected(name string) bool {
	client, exists := m.connectionManager.GetClient(name)
	if !exists {
		return false
	}
	return client.IsConnected()
}

// GetServerTools gets tools for a specific server
func (m *MCPManagerImpl) GetServerTools(serverName string) ([]api.Tool, error) {
	client, exists := m.connectionManager.GetClient(serverName)
	if !exists {
		return nil, fmt.Errorf("server '%s' not found", serverName)
	}

	return FetchToolsForClient(client)
}

// RegisterAuthProvider registers an auth provider for a server
func (m *MCPManagerImpl) RegisterAuthProvider(serverName string, provider *ClaudeAuthProvider) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.authProviders[serverName] = provider
}

// GetAuthProvider gets the auth provider for a server
func (m *MCPManagerImpl) GetAuthProvider(serverName string) (*ClaudeAuthProvider, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	provider, exists := m.authProviders[serverName]
	return provider, exists
}

// GetState returns the complete MCP state for serialization
func (m *MCPManagerImpl) GetState() MCPCliState {
	m.mu.RLock()
	defer m.mu.RUnlock()

	state := MCPCliState{
		Clients:         make([]SerializedClient, 0),
		Configs:         make(map[string]ScopedMcpServerConfig),
		Tools:           make([]SerializedTool, 0),
		Resources:       make(map[string][]ServerResource),
		NormalizedNames: make(map[string]string),
	}

	// Add clients
	for name, status := range m.status {
		state.Clients = append(state.Clients, SerializedClient{
			Name: name,
			Type: status.Type,
		})
	}

	// Add configs
	for name, cfg := range m.configs {
		state.Configs[name] = cfg
	}

	// Add tools from connected servers
	for _, serverName := range m.connectionManager.ListConnected() {
		if client, exists := m.connectionManager.GetClient(serverName); exists {
			tools, err := FetchToolsForClient(client)
			if err == nil {
				for _, tool := range tools {
					state.Tools = append(state.Tools, SerializedTool{
						Name:        tool.Name,
						Description: tool.Description,
						InputJSONSchema: map[string]interface{}{
							"type":       "object",
							"properties": tool.InputSchema,
						},
						IsMcp: func() *bool { b := true; return &b }(),
					})
				}
			}
		}
	}

	return state
}

// GetMCPCache returns the MCP cache instance
func (m *MCPManagerImpl) GetMCPCache() *MCPCache {
	return m.cache
}

// GetConnectionManager returns the connection manager
func (m *MCPManagerImpl) GetConnectionManager() *ConnectionManager {
	return m.connectionManager
}

// GetExecutor returns the tool executor
func (m *MCPManagerImpl) GetExecutor() *ToolExecutor {
	return m.executor
}

// Global MCP Manager instance
var (
	globalMCPManager     *MCPManagerImpl
	globalMCPManagerOnce sync.Once
)

// GetGlobalMCPManager returns the global MCP manager instance
func GetGlobalMCPManager() *MCPManagerImpl {
	globalMCPManagerOnce.Do(func() {
		globalMCPManager = NewMCPManager()
	})
	return globalMCPManager
}

// SetGlobalMCPManager sets the global MCP manager instance (for testing)
func SetGlobalMCPManager(manager *MCPManagerImpl) {
	globalMCPManager = manager
}
