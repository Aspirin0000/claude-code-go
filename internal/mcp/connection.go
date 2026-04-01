// Package mcp 提供 MCP 服务器连接管理
// 来源: src/services/mcp/client.ts (800-1200行等效)
// 批次: C-3/8 - 服务器连接管理、批量连接、错误处理
package mcp

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Aspirin0000/claude-code-go/pkg/utils"
)

// ConnectionManager MCP 连接管理器
// 管理多个 MCP 服务器的连接生命周期
type ConnectionManager struct {
	mu        sync.RWMutex
	clients   map[string]*MCPClient
	configs   map[string]ScopedMcpServerConfig
	status    map[string]ConnectionStatus
	cancelFns map[string]context.CancelFunc
}

// ConnectionStatus 连接状态
type ConnectionStatus int

const (
	StatusDisconnected ConnectionStatus = iota
	StatusConnecting
	StatusConnected
	StatusFailed
	StatusNeedsAuth
)

func (s ConnectionStatus) String() string {
	switch s {
	case StatusDisconnected:
		return "disconnected"
	case StatusConnecting:
		return "connecting"
	case StatusConnected:
		return "connected"
	case StatusFailed:
		return "failed"
	case StatusNeedsAuth:
		return "needs-auth"
	default:
		return "unknown"
	}
}

// NewConnectionManager 创建新的连接管理器
func NewConnectionManager() *ConnectionManager {
	return &ConnectionManager{
		clients:   make(map[string]*MCPClient),
		configs:   make(map[string]ScopedMcpServerConfig),
		status:    make(map[string]ConnectionStatus),
		cancelFns: make(map[string]context.CancelFunc),
	}
}

// ConnectServer 连接到单个 MCP 服务器
// 对应 TS: connectToServer 函数
func (cm *ConnectionManager) ConnectServer(
	ctx context.Context,
	name string,
	config ScopedMcpServerConfig,
	stats *ConnectionStats,
) (MCPServerConnection, error) {
	cm.mu.Lock()

	// 检查是否已在连接中
	if status, exists := cm.status[name]; exists && status == StatusConnecting {
		cm.mu.Unlock()
		return nil, fmt.Errorf("server '%s' is already connecting", name)
	}

	cm.status[name] = StatusConnecting
	cm.configs[name] = config
	cm.mu.Unlock()

	connectStartTime := time.Now()

	// 创建带超时的上下文
	connectCtx, cancel := context.WithTimeout(ctx, time.Duration(GetConnectionTimeoutMs())*time.Millisecond)
	defer cancel()

	// 根据配置类型创建传输
	transport, err := cm.createTransport(config)
	if err != nil {
		cm.setStatus(name, StatusFailed)
		return nil, fmt.Errorf("failed to create transport: %w", err)
	}

	// 创建客户端
	client := NewMCPClient(name, config)

	// 设置回调
	client.SetOnClose(func() {
		cm.setStatus(name, StatusDisconnected)
		cm.removeClient(name)
	})

	client.SetOnError(func(err error) {
		LogMCPDebug(name, fmt.Sprintf("Client error: %v", err))
		// 检查是否为会话过期错误
		if IsMcpSessionExpiredError(err) {
			cm.setStatus(name, StatusFailed)
			cm.removeClient(name)
		}
	})

	// 建立连接
	errChan := make(chan error, 1)
	go func() {
		errChan <- client.Connect(transport)
	}()

	select {
	case err := <-errChan:
		if err != nil {
			cm.setStatus(name, StatusFailed)

			// 检查是否为认证错误
			if _, ok := err.(*McpAuthError); ok {
				cm.setStatus(name, StatusNeedsAuth)
				return &NeedsAuthMCPServer{
					Name:   name,
					Type:   MCPServerConnectionTypeNeedsAuth,
					Config: config,
				}, nil
			}

			return nil, err
		}
	case <-connectCtx.Done():
		cm.setStatus(name, StatusFailed)
		transport.Close()
		return nil, fmt.Errorf("connection timeout after %dms", GetConnectionTimeoutMs())
	}

	// 连接成功
	cm.mu.Lock()
	cm.clients[name] = client
	cm.status[name] = StatusConnected
	cm.mu.Unlock()

	// 记录成功事件
	connectionDurationMs := time.Since(connectStartTime).Milliseconds()
	utils.LogEvent("tengu_mcp_server_connection_succeeded", map[string]interface{}{
		"connectionDurationMs": connectionDurationMs,
		"transportType":        config.Type,
		"serverName":           name,
	})

	return &ConnectedMCPServer{
		Name:         name,
		Type:         MCPServerConnectionTypeConnected,
		Client:       client,
		Capabilities: client.capabilities,
		ServerInfo:   &client.serverInfo,
		Instructions: &client.instructions,
		Config:       config,
	}, nil
}

// ConnectBatch 批量连接多个服务器
// 对应 TS: p-map 批处理逻辑
func (cm *ConnectionManager) ConnectBatch(
	ctx context.Context,
	configs map[string]ScopedMcpServerConfig,
	batchSize int,
) []BatchConnectResult {
	if batchSize <= 0 {
		batchSize = GetMcpServerConnectionBatchSize()
	}

	results := make([]BatchConnectResult, 0, len(configs))
	var mu sync.Mutex

	// 使用信号量控制并发
	sem := make(chan struct{}, batchSize)
	var wg sync.WaitGroup

	for name, config := range configs {
		wg.Add(1)
		go func(name string, config ScopedMcpServerConfig) {
			defer wg.Done()

			sem <- struct{}{}        // 获取信号量
			defer func() { <-sem }() // 释放信号量

			conn, err := cm.ConnectServer(ctx, name, config, nil)

			mu.Lock()
			results = append(results, BatchConnectResult{
				Name:       name,
				Connection: conn,
				Error:      err,
			})
			mu.Unlock()
		}(name, config)
	}

	wg.Wait()
	return results
}

// BatchConnectResult 批量连接结果
type BatchConnectResult struct {
	Name       string
	Connection MCPServerConnection
	Error      error
}

// createTransport 根据配置创建传输
func (cm *ConnectionManager) createTransport(config ScopedMcpServerConfig) (ClientTransport, error) {
	serverType := TransportType(config.Type)
	if config.Type == "" {
		serverType = TransportTypeStdio
	}

	switch serverType {
	case TransportTypeStdio:
		if config.Command == "" {
			return nil, fmt.Errorf("stdio command is empty")
		}
		stdioConfig := McpStdioServerConfig{
			Command: config.Command,
			Args:    config.Args,
			Env:     config.Env,
		}
		return CreateStdioTransportFromConfig(stdioConfig), nil

	case TransportTypeSSE, TransportTypeSSEIDE:
		url := GetServerUrl(config.McpServerConfig)
		headers := make(map[string]string)
		return NewSSETransport(url, headers), nil

	case TransportTypeHTTP, TransportTypeClaudeAIProxy:
		url := GetServerUrl(config.McpServerConfig)
		headers := make(map[string]string)
		return NewHTTPTransport(url, headers)

	default:
		return nil, fmt.Errorf("unsupported transport type: %s", config.Type)
	}
}

// DisconnectServer 断开单个服务器连接
func (cm *ConnectionManager) DisconnectServer(name string) error {
	cm.mu.Lock()
	client, exists := cm.clients[name]
	cancelFn := cm.cancelFns[name]
	cm.mu.Unlock()

	if !exists {
		return nil
	}

	// 取消正在进行的操作
	if cancelFn != nil {
		cancelFn()
	}

	// 关闭客户端
	if client != nil {
		if err := client.Close(); err != nil {
			LogMCPDebug(name, fmt.Sprintf("Error closing client: %v", err))
		}
	}

	cm.removeClient(name)
	cm.setStatus(name, StatusDisconnected)

	return nil
}

// DisconnectAll 断开所有连接
func (cm *ConnectionManager) DisconnectAll() {
	cm.mu.RLock()
	names := make([]string, 0, len(cm.clients))
	for name := range cm.clients {
		names = append(names, name)
	}
	cm.mu.RUnlock()

	for _, name := range names {
		cm.DisconnectServer(name)
	}
}

// GetClient 获取客户端
func (cm *ConnectionManager) GetClient(name string) (*MCPClient, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	client, exists := cm.clients[name]
	return client, exists
}

// GetStatus 获取连接状态
func (cm *ConnectionManager) GetStatus(name string) ConnectionStatus {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	if status, exists := cm.status[name]; exists {
		return status
	}
	return StatusDisconnected
}

// ListConnected 获取所有已连接的服务器
func (cm *ConnectionManager) ListConnected() []string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	connected := make([]string, 0)
	for name, status := range cm.status {
		if status == StatusConnected {
			connected = append(connected, name)
		}
	}
	return connected
}

// setStatus 设置状态（内部使用，需在外部加锁后调用或小心使用）
func (cm *ConnectionManager) setStatus(name string, status ConnectionStatus) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.status[name] = status
}

// removeClient 移除客户端（内部使用）
func (cm *ConnectionManager) removeClient(name string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	delete(cm.clients, name)
}

// EnsureConnected 确保服务器已连接
// 对应 TS: ensureConnectedClient
func (cm *ConnectionManager) EnsureConnected(name string, config ScopedMcpServerConfig) (*MCPClient, error) {
	// 检查现有连接
	if client, exists := cm.GetClient(name); exists && client.IsConnected() {
		return client, nil
	}

	// 重新连接
	conn, err := cm.ConnectServer(context.Background(), name, config, nil)
	if err != nil {
		return nil, err
	}

	connected, ok := conn.(*ConnectedMCPServer)
	if !ok {
		return nil, fmt.Errorf("server '%s' is not connected", name)
	}

	return connected.Client.(*MCPClient), nil
}

// CallTool 调用 MCP 工具
// 对应 TS: callMCPTool
func (cm *ConnectionManager) CallTool(
	ctx context.Context,
	serverName string,
	toolName string,
	args map[string]interface{},
) (*CallToolResult, error) {
	client, exists := cm.GetClient(serverName)
	if !exists || client == nil {
		return nil, fmt.Errorf("server '%s' not connected", serverName)
	}

	if !client.IsConnected() {
		return nil, fmt.Errorf("server '%s' connection lost", serverName)
	}

	return client.CallTool(toolName, args)
}
