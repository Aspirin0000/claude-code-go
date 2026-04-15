package lsp

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ServerInstance represents a single LSP server lifecycle.
type ServerInstance struct {
	Client       *Client
	Command      string
	Args         []string
	Extensions   []string
	State        string // stopped, starting, running, error
	mu           sync.RWMutex
	capabilities json.RawMessage
	rootURI      string
}

// NewServerInstance creates a new server instance config (not started yet).
func NewServerInstance(command string, args []string, extensions []string, rootURI string) *ServerInstance {
	return &ServerInstance{
		Command:    command,
		Args:       args,
		Extensions: extensions,
		State:      "stopped",
		rootURI:    rootURI,
	}
}

// Start initializes the LSP server via initialize/shutdown handshake.
func (s *ServerInstance) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.State == "running" || s.State == "starting" {
		s.mu.Unlock()
		return nil
	}
	s.State = "starting"
	s.mu.Unlock()

	client, err := NewClient(s.Command, s.Args...)
	if err != nil {
		s.setState("error")
		return err
	}

	initParams := map[string]interface{}{
		"processId": nil,
		"rootUri":   s.rootURI,
		"capabilities": map[string]interface{}{
			"textDocument": map[string]interface{}{
				"hover":          map[string]interface{}{"contentFormat": []string{"plaintext", "markdown"}},
				"definition":     map[string]interface{}{"linkSupport": true},
				"references":     map[string]interface{}{},
				"documentSymbol": map[string]interface{}{"hierarchicalDocumentSymbolSupport": true},
				"callHierarchy":  map[string]interface{}{"dynamicRegistration": false},
			},
			"workspace": map[string]interface{}{
				"symbol": map[string]interface{}{},
			},
		},
		"workspaceFolders": []map[string]string{
			{"uri": s.rootURI, "name": "workspace"},
		},
	}

	result, err := client.Request(ctx, "initialize", initParams)
	if err != nil {
		_ = client.Close()
		s.setState("error")
		return fmt.Errorf("initialize failed: %w", err)
	}

	_ = client.Notify("initialized", map[string]interface{}{})

	s.mu.Lock()
	s.Client = client
	s.capabilities = result
	s.State = "running"
	s.mu.Unlock()
	return nil
}

// Stop shuts down the server gracefully.
func (s *ServerInstance) Stop() error {
	s.mu.Lock()
	client := s.Client
	s.State = "stopping"
	s.mu.Unlock()

	if client == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, _ = client.Request(ctx, "shutdown", nil)
	_ = client.Notify("exit", nil)
	_ = client.Close()

	s.setState("stopped")
	return nil
}

// Request sends a request to the underlying client if running.
func (s *ServerInstance) Request(ctx context.Context, method string, params interface{}) (json.RawMessage, error) {
	s.mu.RLock()
	client := s.Client
	state := s.State
	s.mu.RUnlock()

	if state != "running" || client == nil {
		return nil, fmt.Errorf("server not running (state: %s)", state)
	}
	return client.Request(ctx, method, params)
}

// Notify sends a notification if running.
func (s *ServerInstance) Notify(method string, params interface{}) error {
	s.mu.RLock()
	client := s.Client
	state := s.State
	s.mu.RUnlock()

	if state != "running" || client == nil {
		return fmt.Errorf("server not running")
	}
	return client.Notify(method, params)
}

func (s *ServerInstance) setState(state string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.State = state
}

// ServerConfig defines how to start an LSP server for a language.
type ServerConfig struct {
	Command    string   `json:"command"`
	Args       []string `json:"args"`
	Extensions []string `json:"extensions"`
}

// Manager routes LSP requests to the correct server instance by file extension.
type Manager struct {
	mu          sync.RWMutex
	servers     map[string]*ServerInstance // key = extension
	rootURI     string
	openFiles   map[string]bool
	openFilesMu sync.RWMutex
}

// NewManager creates a new LSP manager.
func NewManager(rootURI string) *Manager {
	return &Manager{
		servers:   make(map[string]*ServerInstance),
		rootURI:   rootURI,
		openFiles: make(map[string]bool),
	}
}

// RegisterServer registers a server configuration for given file extensions.
func (m *Manager) RegisterServer(config ServerConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()
	instance := NewServerInstance(config.Command, config.Args, config.Extensions, m.rootURI)
	for _, ext := range config.Extensions {
		m.servers[ext] = instance
	}
}

// getServerForPath returns the server instance responsible for the given file path.
func (m *Manager) getServerForPath(path string) (*ServerInstance, error) {
	ext := strings.ToLower(filepath.Ext(path))
	if ext == "" {
		return nil, fmt.Errorf("no file extension")
	}

	m.mu.RLock()
	server := m.servers[ext]
	m.mu.RUnlock()

	if server == nil {
		return nil, fmt.Errorf("no LSP server registered for extension %s", ext)
	}
	return server, nil
}

// EnsureStarted ensures the server for the given path is started.
func (m *Manager) EnsureStarted(ctx context.Context, path string) (*ServerInstance, error) {
	server, err := m.getServerForPath(path)
	if err != nil {
		return nil, err
	}
	if err := server.Start(ctx); err != nil {
		return nil, err
	}
	return server, nil
}

// OpenFile notifies the server that a file is open.
func (m *Manager) OpenFile(path string, version int, text string) error {
	server, err := m.getServerForPath(path)
	if err != nil {
		return err
	}

	m.openFilesMu.Lock()
	alreadyOpen := m.openFiles[path]
	m.openFiles[path] = true
	m.openFilesMu.Unlock()

	if alreadyOpen {
		return nil
	}

	uri := PathToURI(path)
	return server.Notify("textDocument/didOpen", map[string]interface{}{
		"textDocument": map[string]interface{}{
			"uri":        uri,
			"languageId": languageIDFromExt(filepath.Ext(path)),
			"version":    version,
			"text":       text,
		},
	})
}

// CloseFile notifies the server that a file is closed.
func (m *Manager) CloseFile(path string) error {
	server, err := m.getServerForPath(path)
	if err != nil {
		return err
	}

	m.openFilesMu.Lock()
	delete(m.openFiles, path)
	m.openFilesMu.Unlock()

	uri := PathToURI(path)
	return server.Notify("textDocument/didClose", map[string]interface{}{
		"textDocument": map[string]interface{}{"uri": uri},
	})
}

// Request sends an LSP request for the given file path.
func (m *Manager) Request(ctx context.Context, path string, method string, params interface{}) (json.RawMessage, error) {
	server, err := m.EnsureStarted(ctx, path)
	if err != nil {
		return nil, err
	}
	return server.Request(ctx, method, params)
}

// ShutdownAll stops all managed servers.
func (m *Manager) ShutdownAll() {
	m.mu.RLock()
	seen := make(map[*ServerInstance]bool)
	for _, s := range m.servers {
		seen[s] = true
	}
	m.mu.RUnlock()

	for s := range seen {
		_ = s.Stop()
	}
}

// PathToURI converts a filesystem path to a file:// URI.
func PathToURI(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}
	// Very basic URI encoding for spaces
	abs = strings.ReplaceAll(abs, " ", "%20")
	return "file://" + abs
}

// languageIDFromExt maps a file extension to a language ID.
func languageIDFromExt(ext string) string {
	switch strings.ToLower(ext) {
	case ".go":
		return "go"
	case ".js":
		return "javascript"
	case ".ts", ".tsx":
		return "typescript"
	case ".py":
		return "python"
	case ".rs":
		return "rust"
	case ".java":
		return "java"
	case ".c", ".h":
		return "c"
	case ".cpp", ".cc", ".hpp":
		return "cpp"
	default:
		return strings.TrimPrefix(ext, ".")
	}
}

// Global manager instance.
var globalManager *Manager
var globalManagerMu sync.Mutex

// GetGlobalManager returns the global LSP manager, initializing lazily with the current working directory.
func GetGlobalManager() *Manager {
	globalManagerMu.Lock()
	defer globalManagerMu.Unlock()
	if globalManager == nil {
		cwd, _ := filepath.Abs(".")
		globalManager = NewManager(PathToURI(cwd))
	}
	return globalManager
}

// SetGlobalManager sets the global LSP manager (useful for testing).
func SetGlobalManager(m *Manager) {
	globalManagerMu.Lock()
	defer globalManagerMu.Unlock()
	globalManager = m
}
