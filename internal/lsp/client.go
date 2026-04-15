// Package lsp provides a lightweight Language Server Protocol client and manager.
package lsp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
)

// Client is a lightweight JSON-RPC client over stdio.
type Client struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout *bufio.Reader
	stderr io.ReadCloser

	mu       sync.Mutex
	closed   bool
	pending  map[string]chan *jsonRPCResponse
	handlers map[string]func(params json.RawMessage)

	idCounter int64
}

// jsonRPCMessage is the envelope for JSON-RPC messages.
type jsonRPCMessage struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *jsonRPCError   `json:"error,omitempty"`
}

// jsonRPCResponse is used for pending request resolution.
type jsonRPCResponse struct {
	Result json.RawMessage
	Error  *jsonRPCError
}

// jsonRPCError represents a JSON-RPC error object.
type jsonRPCError struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// Error implements the error interface.
func (e *jsonRPCError) Error() string {
	return fmt.Sprintf("jsonrpc error %d: %s", e.Code, e.Message)
}

// NewClient starts a new LSP client by spawning the given command.
func NewClient(command string, args ...string) (*Client, error) {
	cmd := exec.Command(command, args...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdin pipe: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start LSP server: %w", err)
	}

	client := &Client{
		cmd:      cmd,
		stdin:    stdin,
		stdout:   bufio.NewReader(stdout),
		stderr:   stderr,
		pending:  make(map[string]chan *jsonRPCResponse),
		handlers: make(map[string]func(params json.RawMessage)),
	}

	go client.readLoop()
	return client, nil
}

// readLoop continuously reads JSON-RPC messages from stdout.
func (c *Client) readLoop() {
	for {
		msg, err := c.readMessage()
		if err != nil {
			c.closePending(fmt.Errorf("read loop ended: %w", err))
			return
		}

		if msg.Method != "" {
			// Server-to-client notification or request
			if len(msg.ID) == 0 {
				// Notification
				c.mu.Lock()
				handler := c.handlers[msg.Method]
				c.mu.Unlock()
				if handler != nil {
					handler(msg.Params)
				}
			}
			continue
		}

		if len(msg.ID) > 0 {
			var idStr string
			if err := json.Unmarshal(msg.ID, &idStr); err == nil {
				idStr = strconv.Quote(idStr)
			}
			if idStr == "" {
				var idNum int64
				if err := json.Unmarshal(msg.ID, &idNum); err == nil {
					idStr = strconv.FormatInt(idNum, 10)
				}
			}

			c.mu.Lock()
			ch, ok := c.pending[idStr]
			delete(c.pending, idStr)
			c.mu.Unlock()

			if ok {
				ch <- &jsonRPCResponse{Result: msg.Result, Error: msg.Error}
			}
		}
	}
}

// readMessage reads a single JSON-RPC message using Content-Length headers.
func (c *Client) readMessage() (*jsonRPCMessage, error) {
	var contentLength int
	for {
		line, err := c.stdout.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line = strings.TrimSpace(line)
		if line == "" {
			break
		}
		if strings.HasPrefix(line, "Content-Length: ") {
			val := strings.TrimPrefix(line, "Content-Length: ")
			if n, err := strconv.Atoi(val); err == nil {
				contentLength = n
			}
		}
	}

	if contentLength == 0 {
		return nil, fmt.Errorf("missing Content-Length header")
	}

	body := make([]byte, contentLength)
	_, err := io.ReadFull(c.stdout, body)
	if err != nil {
		return nil, err
	}

	var msg jsonRPCMessage
	if err := json.Unmarshal(body, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

// writeMessage sends a JSON-RPC message to the server's stdin.
func (c *Client) writeMessage(msg *jsonRPCMessage) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return fmt.Errorf("client closed")
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	header := fmt.Sprintf("Content-Length: %d\r\n\r\n", len(body))
	if _, err := c.stdin.Write([]byte(header)); err != nil {
		return err
	}
	if _, err := c.stdin.Write(body); err != nil {
		return err
	}
	return nil
}

// nextID generates the next request ID.
func (c *Client) nextID() int64 {
	return atomic.AddInt64(&c.idCounter, 1)
}

// Request sends a JSON-RPC request and waits for a response.
func (c *Client) Request(ctx context.Context, method string, params interface{}) (json.RawMessage, error) {
	paramsBytes, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	id := c.nextID()
	idStr := strconv.FormatInt(id, 10)
	msg := &jsonRPCMessage{
		JSONRPC: "2.0",
		ID:      json.RawMessage(idStr),
		Method:  method,
		Params:  paramsBytes,
	}

	ch := make(chan *jsonRPCResponse, 1)
	c.mu.Lock()
	c.pending[idStr] = ch
	c.mu.Unlock()

	if err := c.writeMessage(msg); err != nil {
		c.mu.Lock()
		delete(c.pending, idStr)
		c.mu.Unlock()
		return nil, err
	}

	select {
	case resp := <-ch:
		if resp.Error != nil {
			return nil, resp.Error
		}
		return resp.Result, nil
	case <-ctx.Done():
		c.mu.Lock()
		delete(c.pending, idStr)
		c.mu.Unlock()
		return nil, ctx.Err()
	}
}

// Notify sends a JSON-RPC notification (no response expected).
func (c *Client) Notify(method string, params interface{}) error {
	paramsBytes, err := json.Marshal(params)
	if err != nil {
		return err
	}

	msg := &jsonRPCMessage{
		JSONRPC: "2.0",
		Method:  method,
		Params:  paramsBytes,
	}
	return c.writeMessage(msg)
}

// OnNotification registers a handler for server-to-client notifications.
func (c *Client) OnNotification(method string, handler func(params json.RawMessage)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.handlers[method] = handler
}

// Close shuts down the client and kills the server process.
func (c *Client) Close() error {
	c.mu.Lock()
	c.closed = true
	for id, ch := range c.pending {
		close(ch)
		delete(c.pending, id)
	}
	c.mu.Unlock()

	if c.stdin != nil {
		_ = c.stdin.Close()
	}
	if c.cmd != nil && c.cmd.Process != nil {
		_ = c.cmd.Process.Kill()
		_, _ = c.cmd.Process.Wait()
	}
	return nil
}

// closePending closes all pending channels with an error.
func (c *Client) closePending(err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for id, ch := range c.pending {
		ch <- &jsonRPCResponse{Error: &jsonRPCError{Code: -1, Message: err.Error()}}
		delete(c.pending, id)
	}
}
