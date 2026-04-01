// Package state 提供应用状态管理
package state

import (
	"sync"
)

// Message 消息类型
type Message struct {
	UUID    string
	Type    string // user, assistant, system
	Role    string
	Content string
}

// AppState 应用状态
type AppState struct {
	mu sync.RWMutex

	Messages    []Message
	Tools       []string
	SessionID   string
	CWD         string
	ProjectRoot string

	// 对话相关
	ConversationID string
	TurnCount      int
}

// NewAppState 创建新的应用状态
func NewAppState() *AppState {
	return &AppState{
		Messages: make([]Message, 0),
		Tools:    make([]string, 0),
	}
}

// AddMessage 添加消息
func (s *AppState) AddMessage(msg Message) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Messages = append(s.Messages, msg)
}

// GetMessages 获取所有消息
func (s *AppState) GetMessages() []Message {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 返回副本
	result := make([]Message, len(s.Messages))
	copy(result, s.Messages)
	return result
}

// ClearMessages 清空消息
func (s *AppState) ClearMessages() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Messages = make([]Message, 0)
}

// SetSessionID 设置会话 ID
func (s *AppState) SetSessionID(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.SessionID = id
}

// SetCWD 设置当前工作目录
func (s *AppState) SetCWD(cwd string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.CWD = cwd
}

// IncrementTurn 增加轮次计数
func (s *AppState) IncrementTurn() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.TurnCount++
}

// GlobalState 全局状态实例
var GlobalState = NewAppState()
