// Package types 提供核心类型定义
// 来源: src/types/messageQueueTypes.ts (10行)
// 重构: Go 消息队列类型
package types

// QueueOperationMessage 队列操作消息
// 对应 TS: export type QueueOperationMessage = {...}
type QueueOperationMessage struct {
	Type      string                 `json:"type"`
	Operation QueueOperation         `json:"operation"`
	Timestamp string                 `json:"timestamp"`
	SessionID string                 `json:"sessionId"`
	Content   *string                `json:"content,omitempty"`
	Extra     map[string]interface{} `json:"-"`
}

// QueueOperation 队列操作类型
// 对应 TS: export type QueueOperation = 'enqueue' | 'dequeue' | 'remove' | string
type QueueOperation string

const (
	QueueOperationEnqueue QueueOperation = "enqueue"
	QueueOperationDequeue QueueOperation = "dequeue"
	QueueOperationRemove  QueueOperation = "remove"
)
