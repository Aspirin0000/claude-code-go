// Package state provides auto-save functionality for sessions
package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Aspirin0000/claude-code-go/internal/config"
)

// SessionStorage handles auto-saving and loading of conversation sessions
type SessionStorage struct {
	autoSaveDir string
	enabled     bool
}

// NewSessionStorage creates a new session storage manager
func NewSessionStorage(cfg *config.Config) *SessionStorage {
	return &SessionStorage{
		autoSaveDir: cfg.GetAutoSaveDir(),
		enabled:     cfg.AutoSave,
	}
}

// SaveSession saves the current session to disk
func (ss *SessionStorage) SaveSession(state *AppState, name string) error {
	if !ss.enabled {
		return nil
	}

	// Ensure directory exists
	if err := os.MkdirAll(ss.autoSaveDir, 0755); err != nil {
		return fmt.Errorf("failed to create sessions directory: %w", err)
	}

	// Generate filename
	if name == "" {
		name = ss.generateSessionName(state)
	}
	filename := ss.sanitizeFilename(name) + ".json"
	filepath := filepath.Join(ss.autoSaveDir, filename)

	// Prepare session data
	sessionData := struct {
		SessionID    string    `json:"session_id"`
		CWD          string    `json:"cwd"`
		ProjectRoot  string    `json:"project_root"`
		SavedAt      time.Time `json:"saved_at"`
		Messages     []Message `json:"messages"`
		MessageCount int       `json:"message_count"`
	}{
		SessionID:    state.SessionID,
		CWD:          state.CWD,
		ProjectRoot:  state.ProjectRoot,
		SavedAt:      time.Now(),
		Messages:     state.GetMessages(),
		MessageCount: len(state.GetMessages()),
	}

	// Serialize and save
	data, err := json.MarshalIndent(sessionData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize session: %w", err)
	}

	if err := os.WriteFile(filepath, data, 0644); err != nil {
		return fmt.Errorf("failed to write session file: %w", err)
	}

	return nil
}

// LoadSession loads a session from disk
func (ss *SessionStorage) LoadSession(filename string) (*AppState, error) {
	if !strings.HasSuffix(filename, ".json") {
		filename += ".json"
	}

	filepath := filepath.Join(ss.autoSaveDir, filename)
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read session file: %w", err)
	}

	var sessionData struct {
		SessionID   string    `json:"session_id"`
		CWD         string    `json:"cwd"`
		ProjectRoot string    `json:"project_root"`
		SavedAt     time.Time `json:"saved_at"`
		Messages    []Message `json:"messages"`
	}

	if err := json.Unmarshal(data, &sessionData); err != nil {
		return nil, fmt.Errorf("failed to parse session file: %w", err)
	}

	// Restore state
	state := NewAppState()
	state.SetSessionID(sessionData.SessionID)
	state.SetCWD(sessionData.CWD)
	state.ProjectRoot = sessionData.ProjectRoot
	state.SetMessages(sessionData.Messages)

	return state, nil
}

// ListSessions returns a list of available saved sessions
func (ss *SessionStorage) ListSessions() ([]SessionInfo, error) {
	entries, err := os.ReadDir(ss.autoSaveDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []SessionInfo{}, nil
		}
		return nil, fmt.Errorf("failed to read sessions directory: %w", err)
	}

	var sessions []SessionInfo
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		sessionName := strings.TrimSuffix(entry.Name(), ".json")
		sessions = append(sessions, SessionInfo{
			Name:     sessionName,
			Modified: info.ModTime(),
			Size:     info.Size(),
		})
	}

	// Sort by modification time (newest first)
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].Modified.After(sessions[j].Modified)
	})

	return sessions, nil
}

// DeleteSession removes a saved session
func (ss *SessionStorage) DeleteSession(filename string) error {
	if !strings.HasSuffix(filename, ".json") {
		filename += ".json"
	}
	filepath := filepath.Join(ss.autoSaveDir, filename)
	return os.Remove(filepath)
}

// AutoSave performs an automatic save if enabled
func (ss *SessionStorage) AutoSave(state *AppState) error {
	if !ss.enabled {
		return nil
	}
	return ss.SaveSession(state, "")
}

// generateSessionName creates a default session name based on context
func (ss *SessionStorage) generateSessionName(state *AppState) string {
	// Try to use directory name
	if state.CWD != "" {
		dir := filepath.Base(state.CWD)
		if dir != "." && dir != "/" {
			return fmt.Sprintf("%s_%s", dir, time.Now().Format("2006-01-02_15-04"))
		}
	}

	// Fallback to timestamp
	return fmt.Sprintf("session_%s", time.Now().Format("2006-01-02_15-04-05"))
}

// sanitizeFilename removes unsafe characters from filename
func (ss *SessionStorage) sanitizeFilename(name string) string {
	// Replace unsafe characters
	unsafe := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	result := name
	for _, char := range unsafe {
		result = strings.ReplaceAll(result, char, "_")
	}
	return result
}

// SessionInfo contains metadata about a saved session
type SessionInfo struct {
	Name     string
	Modified time.Time
	Size     int64
}

// FormatSize returns a human-readable file size
func (si SessionInfo) FormatSize() string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)

	switch {
	case si.Size >= GB:
		return fmt.Sprintf("%.2f GB", float64(si.Size)/GB)
	case si.Size >= MB:
		return fmt.Sprintf("%.2f MB", float64(si.Size)/MB)
	case si.Size >= KB:
		return fmt.Sprintf("%.2f KB", float64(si.Size)/KB)
	default:
		return fmt.Sprintf("%d B", si.Size)
	}
}

// GlobalSessionStorage is the global session storage instance
var GlobalSessionStorage *SessionStorage

// InitSessionStorage initializes the global session storage
func InitSessionStorage(cfg *config.Config) {
	GlobalSessionStorage = NewSessionStorage(cfg)
}
