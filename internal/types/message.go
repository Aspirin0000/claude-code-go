// Package types 提供核心类型定义
// 来源: src/types/message.ts (167行)
// 重构: Go 消息类型系统
package types

// MessageType 消息类型枚举
// 对应 TS: export type MessageType = 'user' | 'assistant' | 'system' | ...
type MessageType string

const (
	MessageTypeUser                MessageType = "user"
	MessageTypeAssistant           MessageType = "assistant"
	MessageTypeSystem              MessageType = "system"
	MessageTypeAttachment          MessageType = "attachment"
	MessageTypeProgress            MessageType = "progress"
	MessageTypeGroupedToolUse      MessageType = "grouped_tool_use"
	MessageTypeCollapsedReadSearch MessageType = "collapsed_read_search"
)

// ContentItem 内容块元素
// 对应 TS: export type ContentItem = ContentBlockParam | ContentBlock
type ContentItem interface{}

// MessageContent 消息内容
// 对应 TS: export type MessageContent = string | ContentBlockParam[] | ContentBlock[]
type MessageContent interface{}

// TypedMessageContent 类型化的内容数组
// 对应 TS: export type TypedMessageContent = ContentItem[]
type TypedMessageContent []ContentItem

// Message 基础消息类型
// 对应 TS: export type Message = {...}
type Message struct {
	Type                      MessageType  `json:"type"`
	UUID                      string       `json:"uuid"`
	IsMeta                    *bool        `json:"isMeta,omitempty"`
	IsCompactSummary          *bool        `json:"isCompactSummary,omitempty"`
	ToolUseResult             interface{}  `json:"toolUseResult,omitempty"`
	IsVisibleInTranscriptOnly *bool        `json:"isVisibleInTranscriptOnly,omitempty"`
	Attachment                *Attachment  `json:"attachment,omitempty"`
	Message                   *MessageData `json:"message,omitempty"`
}

// Attachment 附件数据
type Attachment struct {
	Type      string                 `json:"type"`
	ToolUseID *string                `json:"toolUseID,omitempty"`
	Extra     map[string]interface{} `json:"-"` // [key: string]: unknown
}

// MessageData 消息内部数据
type MessageData struct {
	Role    *string                `json:"role,omitempty"`
	ID      *string                `json:"id,omitempty"`
	Content MessageContent         `json:"content,omitempty"`
	Usage   interface{}            `json:"usage,omitempty"`
	Extra   map[string]interface{} `json:"-"`
}

// CompactMetadata 紧凑元数据
type CompactMetadata map[string]interface{}

// PreservedSegment 保留的会话段
type PreservedSegment struct {
	HeadUUID   string                 `json:"headUuid"`
	TailUUID   string                 `json:"tailUuid"`
	AnchorUUID string                 `json:"anchorUuid"`
	Extra      map[string]interface{} `json:"-"`
}

// StopHookInfo 停止钩子信息
type StopHookInfo struct {
	Command    *string                `json:"command,omitempty"`
	DurationMs *int                   `json:"durationMs,omitempty"`
	Extra      map[string]interface{} `json:"-"`
}

// CollapsedReadSearchGroup 折叠的读取搜索组
type CollapsedReadSearchGroup struct {
	Type                  MessageType            `json:"type"`
	UUID                  string                 `json:"uuid"`
	Timestamp             interface{}            `json:"timestamp,omitempty"`
	SearchCount           int                    `json:"searchCount"`
	ReadCount             int                    `json:"readCount"`
	ListCount             int                    `json:"listCount"`
	ReplCount             int                    `json:"replCount"`
	MemorySearchCount     int                    `json:"memorySearchCount"`
	MemoryReadCount       int                    `json:"memoryReadCount"`
	MemoryWriteCount      int                    `json:"memoryWriteCount"`
	ReadFilePaths         []string               `json:"readFilePaths"`
	SearchArgs            []string               `json:"searchArgs"`
	LatestDisplayHint     *string                `json:"latestDisplayHint,omitempty"`
	Messages              []CollapsibleMessage   `json:"messages"`
	DisplayMessage        CollapsibleMessage     `json:"displayMessage"`
	McpCallCount          *int                   `json:"mcpCallCount,omitempty"`
	McpServerNames        []string               `json:"mcpServerNames,omitempty"`
	BashCount             *int                   `json:"bashCount,omitempty"`
	GitOpBashCount        *int                   `json:"gitOpBashCount,omitempty"`
	Commits               []Commit               `json:"commits,omitempty"`
	Pushes                []Push                 `json:"pushes,omitempty"`
	Branches              []Branch               `json:"branches,omitempty"`
	Prs                   []Pr                   `json:"prs,omitempty"`
	HookTotalMs           *int                   `json:"hookTotalMs,omitempty"`
	HookCount             *int                   `json:"hookCount,omitempty"`
	HookInfos             []StopHookInfo         `json:"hookInfos,omitempty"`
	RelevantMemories      []Memory               `json:"relevantMemories,omitempty"`
	TeamMemorySearchCount *int                   `json:"teamMemorySearchCount,omitempty"`
	TeamMemoryReadCount   *int                   `json:"teamMemoryReadCount,omitempty"`
	TeamMemoryWriteCount  *int                   `json:"teamMemoryWriteCount,omitempty"`
	Extra                 map[string]interface{} `json:"-"`
}

// Commit Git 提交信息
type Commit struct {
	Sha  string `json:"sha"`
	Kind string `json:"kind"` // CommitKind
}

// Push Git 推送信息
type Push struct {
	Branch string `json:"branch"`
}

// Branch Git 分支信息
type Branch struct {
	Ref    string `json:"ref"`
	Action string `json:"action"` // BranchAction
}

// Pr Pull Request 信息
type Pr struct {
	Number int     `json:"number"`
	URL    *string `json:"url,omitempty"`
	Action string  `json:"action"` // PrAction
}

// Memory 记忆信息
type Memory struct {
	Path    string `json:"path"`
	Content string `json:"content"`
	MtimeMs int64  `json:"mtimeMs"`
}

// =====================================================
// 具体消息类型（结构体嵌入）
// =====================================================

// AssistantMessage 助手消息
type AssistantMessage struct {
	Message
}

// NewAssistantMessage 创建助手消息
func NewAssistantMessage() AssistantMessage {
	return AssistantMessage{
		Message: Message{Type: MessageTypeAssistant},
	}
}

// AttachmentMessage 附件消息
type AttachmentMessage struct {
	Message
	Attachment Attachment `json:"attachment"`
}

// ProgressMessage 进度消息
type ProgressMessage struct {
	Message
	Data interface{} `json:"data"`
}

// SystemMessage 系统消息
type SystemMessage struct {
	Message
}

// NewSystemMessage 创建系统消息
func NewSystemMessage() SystemMessage {
	return SystemMessage{
		Message: Message{Type: MessageTypeSystem},
	}
}

// UserMessage 用户消息
type UserMessage struct {
	Message
}

// NewUserMessage 创建用户消息
func NewUserMessage() UserMessage {
	return UserMessage{
		Message: Message{Type: MessageTypeUser},
	}
}

// NormalizedUserMessage 标准化用户消息
type NormalizedUserMessage = UserMessage

// NormalizedAssistantMessage 标准化助手消息
type NormalizedAssistantMessage = AssistantMessage

// NormalizedMessage 标准化消息
type NormalizedMessage = Message

// =====================================================
// 系统消息子类型
// =====================================================

// SystemCompactBoundaryMessage 系统紧凑边界消息
type SystemCompactBoundaryMessage struct {
	Message
	CompactMetadata struct {
		PreservedSegment *PreservedSegment      `json:"preservedSegment,omitempty"`
		Extra            map[string]interface{} `json:"-"`
	} `json:"compactMetadata"`
}

// SystemStopHookSummaryMessage 系统停止钩子摘要消息
type SystemStopHookSummaryMessage struct {
	Message
	Subtype         string         `json:"subtype"`
	HookLabel       string         `json:"hookLabel"`
	HookCount       int            `json:"hookCount"`
	TotalDurationMs *int           `json:"totalDurationMs,omitempty"`
	HookInfos       []StopHookInfo `json:"hookInfos"`
}

// GroupedToolUseMessage 分组工具使用消息
type GroupedToolUseMessage struct {
	Message
	ToolName       string                       `json:"toolName"`
	Messages       []NormalizedAssistantMessage `json:"messages"`
	Results        []NormalizedUserMessage      `json:"results"`
	DisplayMessage interface{}                  `json:"displayMessage"`
}

// =====================================================
// 可渲染/可折叠消息类型
// =====================================================

// RenderableMessage 可渲染消息接口
type RenderableMessage interface{}

// CollapsibleMessage 可折叠消息接口
type CollapsibleMessage interface{}

// =====================================================
// 事件类型
// =====================================================

// RequestStartEvent 请求开始事件
type RequestStartEvent struct {
	Type  string                 `json:"type"`
	Extra map[string]interface{} `json:"-"`
}

// StreamEvent 流式事件
type StreamEvent struct {
	Type  string                 `json:"type"`
	Extra map[string]interface{} `json:"-"`
}

// MessageOrigin 消息来源
type MessageOrigin string

// PartialCompactDirection 部分紧凑方向
type PartialCompactDirection string

// SystemMessageLevel 系统消息级别
type SystemMessageLevel string
