package mcp

import (
	"context"
	"fmt"
	"time"
)

const (
	MaxRetryAttempts        = 2
	DefaultExecutionTimeout = 100_000_000 * time.Millisecond
)

type ExecutionState int

const (
	ExecutionStateStarted ExecutionState = iota
	ExecutionStateInProgress
	ExecutionStateCompleted
	ExecutionStateFailed
)

func (s ExecutionState) String() string {
	switch s {
	case ExecutionStateStarted:
		return "started"
	case ExecutionStateInProgress:
		return "in_progress"
	case ExecutionStateCompleted:
		return "completed"
	case ExecutionStateFailed:
		return "failed"
	default:
		return "unknown"
	}
}

type ExecutionProgress struct {
	State       ExecutionState
	Message     string
	StartTime   time.Time
	ElapsedMs   int64
	Attempt     int
	MaxAttempts int
}

type ProgressCallback func(progress ExecutionProgress)

type ExecutionContext struct {
	Ctx              context.Context
	Timeout          time.Duration
	ProgressCallback ProgressCallback
	ServerName       string
	ToolName         string
	startTime        time.Time
}

func NewExecutionContext(ctx context.Context, serverName, toolName string) *ExecutionContext {
	return &ExecutionContext{
		Ctx:        ctx,
		Timeout:    DefaultExecutionTimeout,
		ServerName: serverName,
		ToolName:   toolName,
	}
}

func (ec *ExecutionContext) WithTimeout(timeout time.Duration) *ExecutionContext {
	ec.Timeout = timeout
	return ec
}

func (ec *ExecutionContext) WithProgressCallback(callback ProgressCallback) *ExecutionContext {
	ec.ProgressCallback = callback
	return ec
}

func (ec *ExecutionContext) reportProgress(state ExecutionState, message string, attempt, maxAttempts int) {
	if ec.ProgressCallback == nil {
		return
	}
	elapsed := int64(0)
	if !ec.startTime.IsZero() {
		elapsed = time.Since(ec.startTime).Milliseconds()
	}
	ec.ProgressCallback(ExecutionProgress{
		State:       state,
		Message:     message,
		StartTime:   ec.startTime,
		ElapsedMs:   elapsed,
		Attempt:     attempt,
		MaxAttempts: maxAttempts,
	})
}

type ToolExecutorWithRetry struct {
	connectionManager *ConnectionManager
}

func NewToolExecutorWithRetry(cm *ConnectionManager) *ToolExecutorWithRetry {
	return &ToolExecutorWithRetry{connectionManager: cm}
}

func (te *ToolExecutorWithRetry) Execute(ctx *ExecutionContext, args map[string]interface{}) (*CallToolResult, error) {
	return te.executeInternal(ctx, args, 1, 1)
}

func (te *ToolExecutorWithRetry) ExecuteWithRetry(ctx *ExecutionContext, args map[string]interface{}) (*CallToolResult, error) {
	return te.executeInternal(ctx, args, 1, MaxRetryAttempts)
}

func (te *ToolExecutorWithRetry) executeInternal(ctx *ExecutionContext, args map[string]interface{}, attempt, maxAttempts int) (*CallToolResult, error) {
	if attempt == 1 {
		ctx.startTime = time.Now()
	}
	execCtx, cancel := context.WithTimeout(ctx.Ctx, ctx.Timeout)
	defer cancel()

	ctx.reportProgress(ExecutionStateStarted, fmt.Sprintf("Starting tool '%s' on server '%s'", ctx.ToolName, ctx.ServerName), attempt, maxAttempts)

	client, exists := te.connectionManager.GetClient(ctx.ServerName)
	if !exists || client == nil {
		ctx.reportProgress(ExecutionStateFailed, "Server not connected", attempt, maxAttempts)
		return nil, &McpToolCallError{
			Message:          fmt.Sprintf("MCP server '%s' not connected", ctx.ServerName),
			TelemetryMessage: "server_not_connected",
			McpMeta:          map[string]interface{}{"serverName": ctx.ServerName, "toolName": ctx.ToolName},
		}
	}

	ctx.reportProgress(ExecutionStateInProgress, fmt.Sprintf("Executing tool '%s'", ctx.ToolName), attempt, maxAttempts)

	resultChan := make(chan *CallToolResult, 1)
	errChan := make(chan error, 1)

	go func() {
		result, err := client.CallTool(ctx.ToolName, args)
		if err != nil {
			errChan <- err
			return
		}
		resultChan <- result
	}()

	select {
	case result := <-resultChan:
		ctx.reportProgress(ExecutionStateCompleted, "Tool execution completed", attempt, maxAttempts)
		return result, nil

	case err := <-errChan:
		wrappedErr := te.wrapError(err, ctx.ServerName, ctx.ToolName)
		if IsMcpSessionExpiredError(err) && attempt < maxAttempts {
			ctx.reportProgress(ExecutionStateInProgress, fmt.Sprintf("Session expired, retrying (attempt %d/%d)", attempt+1, maxAttempts), attempt, maxAttempts)
			ClearClientCache(ctx.ServerName)
			if reconnectErr := te.reconnect(ctx); reconnectErr != nil {
				ctx.reportProgress(ExecutionStateFailed, fmt.Sprintf("Reconnection failed: %v", reconnectErr), attempt, maxAttempts)
				return nil, te.wrapError(reconnectErr, ctx.ServerName, ctx.ToolName)
			}
			return te.executeInternal(ctx, args, attempt+1, maxAttempts)
		}
		ctx.reportProgress(ExecutionStateFailed, fmt.Sprintf("Tool execution failed: %v", wrappedErr), attempt, maxAttempts)
		return nil, wrappedErr

	case <-execCtx.Done():
		elapsed := time.Since(ctx.startTime).Milliseconds()
		ctx.reportProgress(ExecutionStateFailed, fmt.Sprintf("Tool execution timeout after %dms", elapsed), attempt, maxAttempts)
		return nil, &McpToolCallError{
			Message:          fmt.Sprintf("MCP tool '%s' execution timeout after %dms", ctx.ToolName, elapsed),
			TelemetryMessage: "execution_timeout",
			McpMeta:          map[string]interface{}{"serverName": ctx.ServerName, "toolName": ctx.ToolName, "timeoutMs": ctx.Timeout.Milliseconds(), "elapsedMs": elapsed},
		}
	}
}

func (te *ToolExecutorWithRetry) reconnect(ctx *ExecutionContext) error {
	_ = te.connectionManager.DisconnectServer(ctx.ServerName)
	config, exists := te.connectionManager.configs[ctx.ServerName]
	if !exists {
		return fmt.Errorf("server configuration not found for '%s'", ctx.ServerName)
	}
	conn, err := te.connectionManager.ConnectServer(ctx.Ctx, ctx.ServerName, config, nil)
	if err != nil {
		return fmt.Errorf("failed to reconnect: %w", err)
	}
	if _, ok := conn.(*ConnectedMCPServer); !ok {
		return fmt.Errorf("server '%s' failed to connect properly", ctx.ServerName)
	}
	return nil
}

func (te *ToolExecutorWithRetry) wrapError(err error, serverName, toolName string) error {
	if err == nil {
		return nil
	}
	if _, ok := err.(*McpToolCallError); ok {
		return err
	}
	if IsMcpSessionExpiredError(err) {
		return &McpSessionExpiredError{ServerName: serverName}
	}
	if authErr, ok := err.(*McpAuthError); ok {
		return &McpToolCallError{
			Message:          authErr.Error(),
			TelemetryMessage: "authentication_error",
			McpMeta:          map[string]interface{}{"serverName": serverName, "toolName": toolName},
		}
	}
	if rpcErr, ok := err.(*JSONRPCError); ok {
		return &McpToolCallError{
			Message:          fmt.Sprintf("JSON-RPC error %d: %s", rpcErr.Code, rpcErr.Message),
			TelemetryMessage: "jsonrpc_error",
			McpMeta:          map[string]interface{}{"serverName": serverName, "toolName": toolName, "code": rpcErr.Code},
		}
	}
	return &McpToolCallError{
		Message:          fmt.Sprintf("MCP tool '%s' call failed: %v", toolName, err),
		TelemetryMessage: "tool_call_failed",
		McpMeta:          map[string]interface{}{"serverName": serverName, "toolName": toolName, "error": err.Error()},
	}
}

type ExecutionResult struct {
	Result   *CallToolResult
	Error    error
	Attempts int
	Duration time.Duration
	State    ExecutionState
}

func (er *ExecutionResult) IsSuccess() bool {
	return er.Error == nil && er.State == ExecutionStateCompleted
}

func (er *ExecutionResult) IsRetried() bool {
	return er.Attempts > 1
}
