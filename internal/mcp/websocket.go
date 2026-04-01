// Package mcp provides MCP WebSocket transport implementation
// Source: src/services/mcp/client.ts WebSocketClientTransport
// Batch: C-6/8 - WebSocket Transport for MCP
package mcp

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocketTransport implements ClientTransport for WebSocket connections
// Supports both ws:// and wss:// protocols with reconnection and keepalive
type WebSocketTransport struct {
	TransportBase
	serverURL   string
	headers     map[string]string
	dialer      *websocket.Dialer
	conn        *websocket.Conn
	writeMu     sync.Mutex
	reconnectMu sync.Mutex

	// Configuration
	maxReconnectAttempts int
	reconnectDelay       time.Duration
	pingInterval         time.Duration
	pongTimeout          time.Duration

	// State
	reconnectCount int
	stopPing       chan struct{}
	lastPong       time.Time
}

// WebSocketConfig configuration options for WebSocket transport
type WebSocketConfig struct {
	ServerURL            string
	Headers              map[string]string
	TLSConfig            *tls.Config
	MaxReconnectAttempts int
	ReconnectDelay       time.Duration
	PingInterval         time.Duration
	PongTimeout          time.Duration
}

// NewWebSocketTransport creates a new WebSocket transport
func NewWebSocketTransport(serverURL string, headers map[string]string) *WebSocketTransport {
	return NewWebSocketTransportWithConfig(WebSocketConfig{
		ServerURL:            serverURL,
		Headers:              headers,
		MaxReconnectAttempts: 3,
		ReconnectDelay:       2 * time.Second,
		PingInterval:         30 * time.Second,
		PongTimeout:          10 * time.Second,
	})
}

// NewWebSocketTransportWithConfig creates a new WebSocket transport with full configuration
func NewWebSocketTransportWithConfig(config WebSocketConfig) *WebSocketTransport {
	// Set defaults
	if config.MaxReconnectAttempts <= 0 {
		config.MaxReconnectAttempts = 3
	}
	if config.ReconnectDelay <= 0 {
		config.ReconnectDelay = 2 * time.Second
	}
	if config.PingInterval <= 0 {
		config.PingInterval = 30 * time.Second
	}
	if config.PongTimeout <= 0 {
		config.PongTimeout = 10 * time.Second
	}

	// Create dialer with TLS config if provided
	dialer := &websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}
	if config.TLSConfig != nil {
		dialer.TLSClientConfig = config.TLSConfig
	}

	return &WebSocketTransport{
		serverURL:            config.ServerURL,
		headers:              config.Headers,
		dialer:               dialer,
		maxReconnectAttempts: config.MaxReconnectAttempts,
		reconnectDelay:       config.ReconnectDelay,
		pingInterval:         config.PingInterval,
		pongTimeout:          config.PongTimeout,
		stopPing:             make(chan struct{}),
	}
}

// Connect establishes the WebSocket connection with reconnection support
func (t *WebSocketTransport) Connect() error {
	t.reconnectMu.Lock()
	defer t.reconnectMu.Unlock()

	if t.conn != nil {
		return fmt.Errorf("already connected")
	}

	return t.connectWithRetry()
}

// connectWithRetry attempts to connect with reconnection logic
func (t *WebSocketTransport) connectWithRetry() error {
	var lastErr error

	for attempt := 0; attempt <= t.maxReconnectAttempts; attempt++ {
		if attempt > 0 {
			t.reconnectCount = attempt
			t.triggerOnError(fmt.Errorf("reconnection attempt %d/%d", attempt, t.maxReconnectAttempts))
			time.Sleep(t.reconnectDelay)
		}

		err := t.dial()
		if err == nil {
			t.reconnectCount = 0
			return nil
		}
		lastErr = err
	}

	return fmt.Errorf("failed to connect after %d attempts: %w", t.maxReconnectAttempts, lastErr)
}

// dial performs the actual WebSocket dial
func (t *WebSocketTransport) dial() error {
	// Parse and validate URL
	u, err := url.Parse(t.serverURL)
	if err != nil {
		return fmt.Errorf("invalid server URL: %w", err)
	}

	// Ensure proper scheme
	if u.Scheme != "ws" && u.Scheme != "wss" {
		return fmt.Errorf("unsupported scheme: %s (expected ws:// or wss://)", u.Scheme)
	}

	// Prepare headers
	header := http.Header{}

	// Set User-Agent
	header.Set("User-Agent", "claude-code-go/1.0")

	// Add custom headers
	for key, value := range t.headers {
		header.Set(key, value)
	}

	// Perform WebSocket handshake
	conn, resp, err := t.dialer.Dial(u.String(), header)
	if err != nil {
		if resp != nil {
			if resp.StatusCode == http.StatusUnauthorized {
				return &McpAuthError{ServerName: t.serverURL, Message: "authentication required"}
			}
			return fmt.Errorf("websocket handshake failed: %s", resp.Status)
		}
		return fmt.Errorf("websocket dial failed: %w", err)
	}

	t.conn = conn
	t.lastPong = time.Now()

	// Set pong handler
	t.conn.SetPongHandler(func(data string) error {
		t.lastPong = time.Now()
		return nil
	})

	// Start message reading goroutine
	go t.readMessages()

	// Start ping keepalive
	go t.keepalive()

	return nil
}

// readMessages continuously reads messages from the WebSocket
func (t *WebSocketTransport) readMessages() {
	defer func() {
		// Handle disconnection
		if !t.IsClosed() {
			t.handleDisconnect()
		}
	}()

	for {
		if t.IsClosed() {
			return
		}

		// Read message
		messageType, data, err := t.conn.ReadMessage()
		if err != nil {
			if !t.IsClosed() && websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				t.triggerOnError(fmt.Errorf("websocket read error: %w", err))
			}
			return
		}

		// Handle text messages (JSON-RPC)
		if messageType == websocket.TextMessage {
			var msg JSONRPCMessage
			if err := json.Unmarshal(data, &msg); err != nil {
				t.triggerOnError(fmt.Errorf("failed to unmarshal message: %w", err))
				continue
			}
			t.triggerOnMessage(msg)
		}
	}
}

// keepalive sends periodic ping messages
func (t *WebSocketTransport) keepalive() {
	ticker := time.NewTicker(t.pingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-t.stopPing:
			return
		case <-ticker.C:
			if t.IsClosed() || t.conn == nil {
				return
			}

			// Check if we've received a pong recently
			if time.Since(t.lastPong) > t.pingInterval+t.pongTimeout {
				t.triggerOnError(fmt.Errorf("pong timeout - connection may be dead"))
				t.handleDisconnect()
				return
			}

			// Send ping
			t.writeMu.Lock()
			err := t.conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(5*time.Second))
			t.writeMu.Unlock()

			if err != nil {
				t.triggerOnError(fmt.Errorf("failed to send ping: %w", err))
				t.handleDisconnect()
				return
			}
		}
	}
}

// handleDisconnect handles unexpected disconnection with reconnection
func (t *WebSocketTransport) handleDisconnect() {
	t.reconnectMu.Lock()
	defer t.reconnectMu.Unlock()

	if t.IsClosed() || t.reconnectCount >= t.maxReconnectAttempts {
		t.triggerOnClose()
		return
	}

	// Attempt reconnection
	t.reconnectCount++
	t.triggerOnError(fmt.Errorf("connection lost, attempting reconnection %d/%d", t.reconnectCount, t.maxReconnectAttempts))

	// Close existing connection
	if t.conn != nil {
		t.conn.Close()
		t.conn = nil
	}

	// Try to reconnect
	go func() {
		time.Sleep(t.reconnectDelay)
		if err := t.connectWithRetry(); err != nil {
			t.triggerOnError(fmt.Errorf("reconnection failed: %w", err))
			t.triggerOnClose()
		}
	}()
}

// Send sends a JSON-RPC message over the WebSocket
func (t *WebSocketTransport) Send(message JSONRPCMessage) error {
	if t.IsClosed() {
		return fmt.Errorf("transport is closed")
	}

	t.reconnectMu.Lock()
	conn := t.conn
	t.reconnectMu.Unlock()

	if conn == nil {
		return fmt.Errorf("not connected")
	}

	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	t.writeMu.Lock()
	defer t.writeMu.Unlock()

	err = conn.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

// Close gracefully closes the WebSocket connection
func (t *WebSocketTransport) Close() error {
	if t.IsClosed() {
		return nil
	}

	t.MarkClosed()

	// Stop ping goroutine
	close(t.stopPing)

	t.reconnectMu.Lock()
	conn := t.conn
	t.conn = nil
	t.reconnectMu.Unlock()

	if conn != nil {
		// Send close message
		conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))

		// Wait a bit for the close handshake
		time.Sleep(100 * time.Millisecond)

		// Close the connection
		conn.Close()
	}

	t.triggerOnClose()
	return nil
}

// IsConnected returns true if the WebSocket is currently connected
func (t *WebSocketTransport) IsConnected() bool {
	t.reconnectMu.Lock()
	defer t.reconnectMu.Unlock()
	return t.conn != nil && !t.IsClosed()
}

// GetReconnectCount returns the number of reconnection attempts made
func (t *WebSocketTransport) GetReconnectCount() int {
	return t.reconnectCount
}

// SetTLSConfig updates the TLS configuration (must be called before Connect)
func (t *WebSocketTransport) SetTLSConfig(config *tls.Config) {
	t.dialer.TLSClientConfig = config
}
