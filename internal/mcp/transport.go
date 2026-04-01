// Package mcp 提供 MCP 传输层实现
// 来源: src/services/mcp/client.ts (800-1200行等效)
// 批次: C-3/8 - 传输层实现 (SSE, WebSocket, HTTP, stdio)
package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"
)

// ============================================================================
// 基础传输接口和实现
// ============================================================================

// TransportBase 传输层基础结构
type TransportBase struct {
	onMessage func(message JSONRPCMessage)
	onClose   func()
	onError   func(err error)
	mu        sync.RWMutex
	closed    bool
}

// SetOnMessage 设置消息处理器
func (t *TransportBase) SetOnMessage(handler func(message JSONRPCMessage)) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.onMessage = handler
}

// SetOnClose 设置关闭处理器
func (t *TransportBase) SetOnClose(handler func()) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.onClose = handler
}

// SetOnError 设置错误处理器
func (t *TransportBase) SetOnError(handler func(err error)) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.onError = handler
}

// triggerOnMessage 触发消息回调
func (t *TransportBase) triggerOnMessage(message JSONRPCMessage) {
	t.mu.RLock()
	handler := t.onMessage
	t.mu.RUnlock()
	if handler != nil {
		handler(message)
	}
}

// triggerOnClose 触发关闭回调
func (t *TransportBase) triggerOnClose() {
	t.mu.RLock()
	handler := t.onClose
	t.mu.RUnlock()
	if handler != nil {
		handler()
	}
}

// triggerOnError 触发错误回调
func (t *TransportBase) triggerOnError(err error) {
	t.mu.RLock()
	handler := t.onError
	t.mu.RUnlock()
	if handler != nil {
		handler(err)
	}
}

// IsClosed 检查传输是否已关闭
func (t *TransportBase) IsClosed() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.closed
}

// MarkClosed 标记为已关闭
func (t *TransportBase) MarkClosed() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.closed = true
}

// ============================================================================
// HTTP 传输实现
// ============================================================================

// HTTPTransport HTTP 传输实现
// 对应 TS: StreamableHTTPClientTransport
type HTTPTransport struct {
	TransportBase
	baseURL    *url.URL
	httpClient *http.Client
	headers    map[string]string
	authToken  string
}

// NewHTTPTransport 创建新的 HTTP 传输
func NewHTTPTransport(serverURL string, headers map[string]string, authToken ...string) (*HTTPTransport, error) {
	parsedURL, err := url.Parse(serverURL)
	if err != nil {
		return nil, fmt.Errorf("invalid server URL: %w", err)
	}

	transport := &HTTPTransport{
		baseURL:    parsedURL,
		httpClient: &http.Client{Timeout: MCPRequestTimeoutMs * time.Millisecond},
		headers:    headers,
	}

	if len(authToken) > 0 && authToken[0] != "" {
		transport.authToken = authToken[0]
	}

	return transport, nil
}

// Connect 建立 HTTP 连接
func (t *HTTPTransport) Connect() error {
	// HTTP 是无状态的，不需要持久连接
	// 只需验证服务器可达
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", t.baseURL.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	for key, value := range t.headers {
		req.Header.Set(key, value)
	}
	req.Header.Set("Accept", MCPStreamableHttpAccept)
	req.Header.Set("User-Agent", "claude-code-go/1.0")

	if t.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+t.authToken)
	}

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return &McpAuthError{ServerName: "http-server", Message: "authentication required"}
	}

	// 开始 SSE 事件流监听
	go t.listenForEvents()

	return nil
}

// listenForEvents 监听 SSE 事件流
func (t *HTTPTransport) listenForEvents() {
	req, err := http.NewRequest("GET", t.baseURL.String(), nil)
	if err != nil {
		t.triggerOnError(err)
		return
	}

	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")

	for key, value := range t.headers {
		req.Header.Set(key, value)
	}

	resp, err := t.httpClient.Do(req)
	if err != nil {
		t.triggerOnError(err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.triggerOnError(fmt.Errorf("SSE connection failed: %s", resp.Status))
		return
	}

	reader := bufio.NewReader(resp.Body)
	for {
		if t.IsClosed() {
			return
		}

		line, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF && !t.IsClosed() {
				t.triggerOnError(err)
			}
			return
		}

		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			var message JSONRPCMessage
			if err := json.Unmarshal([]byte(data), &message); err == nil {
				t.triggerOnMessage(message)
			}
		}
	}
}

// Send 发送消息
func (t *HTTPTransport) Send(message JSONRPCMessage) error {
	if t.IsClosed() {
		return fmt.Errorf("transport is closed")
	}

	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), MCPRequestTimeoutMs*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", t.baseURL.String(), strings.NewReader(string(data)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", MCPStreamableHttpAccept)
	req.Header.Set("User-Agent", "claude-code-go/1.0")

	for key, value := range t.headers {
		req.Header.Set(key, value)
	}

	if t.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+t.authToken)
	}

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return &McpAuthError{ServerName: "http-server", Message: "authentication required"}
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	return nil
}

// Close 关闭传输
func (t *HTTPTransport) Close() error {
	if t.IsClosed() {
		return nil
	}
	t.MarkClosed()
	t.httpClient.CloseIdleConnections()
	t.triggerOnClose()
	return nil
}

// ============================================================================
// Stdio 传输实现
// ============================================================================

// StdioTransport Stdio 传输实现
// 对应 TS: StdioClientTransport
type StdioTransport struct {
	TransportBase
	command string
	args    []string
	env     map[string]string
	cmd     *exec.Cmd
	stdin   io.WriteCloser
	stdout  io.ReadCloser
	stderr  io.ReadCloser
	pid     int
	writeMu sync.Mutex
}

// NewStdioTransport 创建新的 Stdio 传输
func NewStdioTransport(command string, args []string, env map[string]string) *StdioTransport {
	return &StdioTransport{
		command: command,
		args:    args,
		env:     env,
	}
}

// Connect 启动子进程并建立连接
func (t *StdioTransport) Connect() error {
	cmd := exec.Command(t.command, t.args...)

	// 设置环境变量
	cmd.Env = os.Environ()
	for key, value := range t.env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}

	// 获取管道
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// 启动进程
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start process: %w", err)
	}

	t.cmd = cmd
	t.stdin = stdin
	t.stdout = stdout
	t.stderr = stderr
	t.pid = cmd.Process.Pid

	// 启动读取 goroutine
	go t.readStdout()
	go t.readStderr()

	// 监控进程退出
	go t.waitForExit()

	return nil
}

// readStdout 读取标准输出
func (t *StdioTransport) readStdout() {
	scanner := bufio.NewScanner(t.stdout)
	for scanner.Scan() {
		if t.IsClosed() {
			return
		}

		line := scanner.Text()
		var message JSONRPCMessage
		if err := json.Unmarshal([]byte(line), &message); err == nil {
			t.triggerOnMessage(message)
		}
	}

	if err := scanner.Err(); err != nil && !t.IsClosed() {
		t.triggerOnError(fmt.Errorf("stdout read error: %w", err))
	}
}

// readStderr 读取标准错误
func (t *StdioTransport) readStderr() {
	scanner := bufio.NewScanner(t.stderr)
	var stderrOutput strings.Builder

	for scanner.Scan() {
		if t.IsClosed() {
			return
		}

		line := scanner.Text()
		if stderrOutput.Len() < 64*1024 {
			stderrOutput.WriteString(line)
			stderrOutput.WriteString("\n")
		}
	}

	if stderrOutput.Len() > 0 {
		// 记录 stderr 输出用于调试
		fmt.Fprintf(os.Stderr, "[MCP StdioServer] %s", stderrOutput.String())
	}
}

// waitForExit 等待进程退出
func (t *StdioTransport) waitForExit() {
	if t.cmd == nil {
		return
	}

	err := t.cmd.Wait()
	if err != nil && !t.IsClosed() {
		t.triggerOnError(fmt.Errorf("process exited: %w", err))
	}

	t.triggerOnClose()
}

// Send 发送消息到子进程
func (t *StdioTransport) Send(message JSONRPCMessage) error {
	if t.IsClosed() {
		return fmt.Errorf("transport is closed")
	}

	t.writeMu.Lock()
	defer t.writeMu.Unlock()

	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// 写入 JSON 行
	_, err = fmt.Fprintf(t.stdin, "%s\n", data)
	if err != nil {
		return fmt.Errorf("failed to write to stdin: %w", err)
	}

	return nil
}

// Close 关闭传输并终止子进程
func (t *StdioTransport) Close() error {
	if t.IsClosed() {
		return nil
	}
	t.MarkClosed()

	// 发送终止信号序列
	if t.cmd != nil && t.cmd.Process != nil {
		// 首先尝试 SIGINT
		t.cmd.Process.Signal(syscall.SIGINT)

		// 等待 100ms
		time.Sleep(100 * time.Millisecond)

		// 如果还在运行，尝试 SIGTERM
		if t.cmd.ProcessState == nil || !t.cmd.ProcessState.Exited() {
			t.cmd.Process.Signal(syscall.SIGTERM)
			time.Sleep(400 * time.Millisecond)
		}

		// 最后强制终止
		if t.cmd.ProcessState == nil || !t.cmd.ProcessState.Exited() {
			t.cmd.Process.Kill()
		}
	}

	// 关闭管道
	if t.stdin != nil {
		t.stdin.Close()
	}
	if t.stdout != nil {
		t.stdout.Close()
	}
	if t.stderr != nil {
		t.stderr.Close()
	}

	t.triggerOnClose()
	return nil
}

// GetPID 获取子进程 PID
func (t *StdioTransport) GetPID() int {
	return t.pid
}

// ============================================================================
// SSE 传输实现
// ============================================================================

// SSETransport SSE 传输实现
// 对应 TS: SSEClientTransport
type SSETransport struct {
	TransportBase
	serverURL   string
	headers     map[string]string
	client      *http.Client
	eventSource io.ReadCloser
}

// NewSSETransport 创建新的 SSE 传输
func NewSSETransport(serverURL string, headers map[string]string) *SSETransport {
	return &SSETransport{
		serverURL: serverURL,
		headers:   headers,
		client:    &http.Client{Timeout: 0}, // SSE 连接是长期的
	}
}

// Connect 建立 SSE 连接
func (t *SSETransport) Connect() error {
	req, err := http.NewRequest("GET", t.serverURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")

	for key, value := range t.headers {
		req.Header.Set(key, value)
	}

	resp, err := t.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	t.eventSource = resp.Body

	// 启动事件读取 goroutine
	go t.readEvents()

	return nil
}

// readEvents 读取 SSE 事件
func (t *SSETransport) readEvents() {
	reader := bufio.NewReader(t.eventSource)
	var eventData strings.Builder

	for {
		if t.IsClosed() {
			return
		}

		line, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF && !t.IsClosed() {
				t.triggerOnError(fmt.Errorf("SSE read error: %w", err))
			}
			return
		}

		line = strings.TrimSpace(line)

		if line == "" {
			// 空行表示事件结束
			if eventData.Len() > 0 {
				data := eventData.String()
				var message JSONRPCMessage
				if err := json.Unmarshal([]byte(data), &message); err == nil {
					t.triggerOnMessage(message)
				}
				eventData.Reset()
			}
		} else if strings.HasPrefix(line, "data: ") {
			// 累积事件数据
			if eventData.Len() > 0 {
				eventData.WriteString("\n")
			}
			eventData.WriteString(strings.TrimPrefix(line, "data: "))
		}
	}
}

// Send 发送消息 (SSE 传输使用 POST 请求)
func (t *SSETransport) Send(message JSONRPCMessage) error {
	if t.IsClosed() {
		return fmt.Errorf("transport is closed")
	}

	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), MCPRequestTimeoutMs*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", t.serverURL, strings.NewReader(string(data)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	for key, value := range t.headers {
		req.Header.Set(key, value)
	}

	resp, err := t.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	return nil
}

// Close 关闭 SSE 连接
func (t *SSETransport) Close() error {
	if t.IsClosed() {
		return nil
	}
	t.MarkClosed()

	if t.eventSource != nil {
		t.eventSource.Close()
	}

	t.triggerOnClose()
	return nil
}

// ============================================================================
// 传输工厂函数
// ============================================================================

// TransportType 传输类型
// 对应 TS: transport type strings
type TransportType string

const (
	TransportTypeStdio         TransportType = "stdio"
	TransportTypeSSE           TransportType = "sse"
	TransportTypeSSEIDE        TransportType = "sse-ide"
	TransportTypeHTTP          TransportType = "http"
	TransportTypeWebSocket     TransportType = "ws"
	TransportTypeWebSocketIDE  TransportType = "ws-ide"
	TransportTypeClaudeAIProxy TransportType = "claudeai-proxy"
	TransportTypeSDK           TransportType = "sdk"
)

// CreateTransport 根据配置创建传输
// 对应 TS: transport creation in connectToServer
func CreateTransport(serverType TransportType, serverURL string, headers map[string]string, authToken ...string) (ClientTransport, error) {
	switch serverType {
	case TransportTypeStdio, "":
		// stdio 传输需要特殊处理，由调用方提供命令参数
		return nil, fmt.Errorf("stdio transport requires command configuration, use NewStdioTransport directly")

	case TransportTypeSSE, TransportTypeSSEIDE:
		return NewSSETransport(serverURL, headers), nil

	case TransportTypeHTTP:
		return NewHTTPTransport(serverURL, headers, authToken...)

	case TransportTypeClaudeAIProxy:
		// Claude AI 代理使用 HTTP 传输
		return NewHTTPTransport(serverURL, headers, authToken...)

	case TransportTypeWebSocket, TransportTypeWebSocketIDE:
		return NewWebSocketTransport(serverURL, headers), nil

	default:
		return nil, fmt.Errorf("unsupported transport type: %s", serverType)
	}
}

// CreateStdioTransportFromConfig 从配置创建 stdio 传输
// 对应 TS: StdioClientTransport creation
func CreateStdioTransportFromConfig(config McpStdioServerConfig) *StdioTransport {
	command := config.Command
	args := config.Args

	// 支持 CLAUDE_CODE_SHELL_PREFIX 环境变量
	if shellPrefix := os.Getenv("CLAUDE_CODE_SHELL_PREFIX"); shellPrefix != "" {
		command = shellPrefix
		args = []string{config.Command}
		if len(config.Args) > 0 {
			args = append(args, config.Args...)
		}
	}

	// 合并环境变量
	env := make(map[string]string)
	for k, v := range config.Env {
		env[k] = v
	}

	return NewStdioTransport(command, args, env)
}
