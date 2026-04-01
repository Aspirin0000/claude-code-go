// Package types 提供核心类型定义
// 来源: src/types/logs.ts (330行)
// 重构: Go 日志类型系统
package types

// SerializedMessage 序列化消息
// 对应 TS: export type SerializedMessage = Message & {...}
type SerializedMessage struct {
	Message
	Cwd        string  `json:"cwd"`
	UserType   string  `json:"userType"`
	Entrypoint *string `json:"entrypoint,omitempty"` // CLAUDE_CODE_ENTRYPOINT
	SessionID  string  `json:"sessionId"`
	Timestamp  string  `json:"timestamp"`
	Version    string  `json:"version"`
	GitBranch  *string `json:"gitBranch,omitempty"`
	Slug       *string `json:"slug,omitempty"` // Session slug
}

// LogOption 日志选项
// 对应 TS: export type LogOption = {...}
type LogOption struct {
	Date            string                    `json:"date"`
	Messages        []SerializedMessage       `json:"messages"`
	FullPath        *string                   `json:"fullPath,omitempty"`
	Value           int                       `json:"value"`
	Created         int64                     `json:"created"`  // Unix timestamp
	Modified        int64                     `json:"modified"` // Unix timestamp
	FirstPrompt     string                    `json:"firstPrompt"`
	MessageCount    int                       `json:"messageCount"`
	FileSize        *int                      `json:"fileSize,omitempty"` // bytes
	IsSidechain     bool                      `json:"isSidechain"`
	IsLite          *bool                     `json:"isLite,omitempty"`
	SessionID       *string                   `json:"sessionId,omitempty"`
	TeamName        *string                   `json:"teamName,omitempty"`
	AgentName       *string                   `json:"agentName,omitempty"`
	AgentColor      *string                   `json:"agentColor,omitempty"`
	AgentSetting    *string                   `json:"agentSetting,omitempty"`
	IsTeammate      *bool                     `json:"isTeammate,omitempty"`
	LeafUUID        *string                   `json:"leafUuid,omitempty"`
	Summary         *string                   `json:"summary,omitempty"`
	CustomTitle     *string                   `json:"customTitle,omitempty"`
	Tag             *string                   `json:"tag,omitempty"`
	GitBranch       *string                   `json:"gitBranch,omitempty"`
	ProjectPath     *string                   `json:"projectPath,omitempty"`
	PrNumber        *int                      `json:"prNumber,omitempty"`
	PrURL           *string                   `json:"prUrl,omitempty"`
	PrRepository    *string                   `json:"prRepository,omitempty"`
	Mode            *string                   `json:"mode,omitempty"` // "coordinator" | "normal"
	WorktreeSession *PersistedWorktreeSession `json:"worktreeSession,omitempty"`
}

// SummaryMessage 摘要消息
type SummaryMessage struct {
	Type     string `json:"type"`
	LeafUUID string `json:"leafUuid"`
	Summary  string `json:"summary"`
}

// CustomTitleMessage 自定义标题消息
type CustomTitleMessage struct {
	Type        string `json:"type"`
	SessionID   string `json:"sessionId"`
	CustomTitle string `json:"customTitle"`
}

// AiTitleMessage AI 生成的标题消息
type AiTitleMessage struct {
	Type      string `json:"type"`
	SessionID string `json:"sessionId"`
	AiTitle   string `json:"aiTitle"`
}

// LastPromptMessage 最后提示消息
type LastPromptMessage struct {
	Type       string `json:"type"`
	SessionID  string `json:"sessionId"`
	LastPrompt string `json:"lastPrompt"`
}

// TaskSummaryMessage 任务摘要消息
type TaskSummaryMessage struct {
	Type      string `json:"type"`
	SessionID string `json:"sessionId"`
	Summary   string `json:"summary"`
	Timestamp string `json:"timestamp"`
}

// TagMessage 标签消息
type TagMessage struct {
	Type      string `json:"type"`
	SessionID string `json:"sessionId"`
	Tag       string `json:"tag"`
}

// AgentNameMessage 代理名称消息
type AgentNameMessage struct {
	Type      string `json:"type"`
	SessionID string `json:"sessionId"`
	AgentName string `json:"agentName"`
}

// AgentColorMessage 代理颜色消息
type AgentColorMessage struct {
	Type       string `json:"type"`
	SessionID  string `json:"sessionId"`
	AgentColor string `json:"agentColor"`
}

// AgentSettingMessage 代理设置消息
type AgentSettingMessage struct {
	Type         string `json:"type"`
	SessionID    string `json:"sessionId"`
	AgentSetting string `json:"agentSetting"`
}

// PRLinkMessage PR 链接消息
type PRLinkMessage struct {
	Type         string `json:"type"`
	SessionID    string `json:"sessionId"`
	PrNumber     int    `json:"prNumber"`
	PrURL        string `json:"prUrl"`
	PrRepository string `json:"prRepository"`
	Timestamp    string `json:"timestamp"`
}

// ModeEntry 模式条目
type ModeEntry struct {
	Type      string `json:"type"`
	SessionID string `json:"sessionId"`
	Mode      string `json:"mode"` // "coordinator" | "normal"
}

// PersistedWorktreeSession 持久化的工作树会话
type PersistedWorktreeSession struct {
	OriginalCwd        string  `json:"originalCwd"`
	WorktreePath       string  `json:"worktreePath"`
	WorktreeName       string  `json:"worktreeName"`
	WorktreeBranch     *string `json:"worktreeBranch,omitempty"`
	OriginalBranch     *string `json:"originalBranch,omitempty"`
	OriginalHeadCommit *string `json:"originalHeadCommit,omitempty"`
	SessionID          string  `json:"sessionId"`
	TmuxSessionName    *string `json:"tmuxSessionName,omitempty"`
	HookBased          *bool   `json:"hookBased,omitempty"`
}

// WorktreeStateEntry 工作树状态条目
type WorktreeStateEntry struct {
	Type            string                    `json:"type"`
	SessionID       string                    `json:"sessionId"`
	WorktreeSession *PersistedWorktreeSession `json:"worktreeSession"`
}

// ContentReplacementEntry 内容替换条目
type ContentReplacementEntry struct {
	Type         string                     `json:"type"`
	SessionID    string                     `json:"sessionId"`
	AgentID      *AgentId                   `json:"agentId,omitempty"`
	Replacements []ContentReplacementRecord `json:"replacements"`
}

// ContentReplacementRecord 内容替换记录
type ContentReplacementRecord struct {
	// 简化版，实际字段根据具体实现补充
	Original interface{} `json:"original"`
	Replaced interface{} `json:"replaced"`
}

// FileHistorySnapshotMessage 文件历史快照消息
type FileHistorySnapshotMessage struct {
	Type             string      `json:"type"`
	MessageID        string      `json:"messageId"`
	Snapshot         interface{} `json:"snapshot"` // FileHistorySnapshot
	IsSnapshotUpdate bool        `json:"isSnapshotUpdate"`
}

// FileAttributionState 文件归因状态
type FileAttributionState struct {
	ContentHash        string `json:"contentHash"`
	ClaudeContribution int    `json:"claudeContribution"`
	Mtime              int64  `json:"mtime"`
}

// AttributionSnapshotMessage 归因快照消息
type AttributionSnapshotMessage struct {
	Type                              string                          `json:"type"`
	MessageID                         string                          `json:"messageId"`
	Surface                           string                          `json:"surface"`
	FileStates                        map[string]FileAttributionState `json:"fileStates"`
	PromptCount                       *int                            `json:"promptCount,omitempty"`
	PromptCountAtLastCommit           *int                            `json:"promptCountAtLastCommit,omitempty"`
	PermissionPromptCount             *int                            `json:"permissionPromptCount,omitempty"`
	PermissionPromptCountAtLastCommit *int                            `json:"permissionPromptCountAtLastCommit,omitempty"`
	EscapeCount                       *int                            `json:"escapeCount,omitempty"`
	EscapeCountAtLastCommit           *int                            `json:"escapeCountAtLastCommit,omitempty"`
}

// TranscriptMessage 转录消息
type TranscriptMessage struct {
	SerializedMessage
	ParentUUID        *string `json:"parentUuid"`
	LogicalParentUUID *string `json:"logicalParentUuid,omitempty"`
	IsSidechain       bool    `json:"isSidechain"`
	GitBranch         *string `json:"gitBranch,omitempty"`
	AgentID           *string `json:"agentId,omitempty"`
	TeamName          *string `json:"teamName,omitempty"`
	AgentName         *string `json:"agentName,omitempty"`
	AgentColor        *string `json:"agentColor,omitempty"`
	PromptID          *string `json:"promptId,omitempty"`
}

// SpeculationAcceptMessage 推测接受消息
type SpeculationAcceptMessage struct {
	Type        string `json:"type"`
	Timestamp   string `json:"timestamp"`
	TimeSavedMs int    `json:"timeSavedMs"`
}

// ContextCollapseCommitEntry 上下文压缩提交条目
type ContextCollapseCommitEntry struct {
	Type              string `json:"type"` // "marble-origami-commit"
	SessionID         string `json:"sessionId"`
	CollapseID        string `json:"collapseId"`
	SummaryUUID       string `json:"summaryUuid"`
	SummaryContent    string `json:"summaryContent"`
	Summary           string `json:"summary"`
	FirstArchivedUUID string `json:"firstArchivedUuid"`
	LastArchivedUUID  string `json:"lastArchivedUuid"`
}

// ContextCollapseSnapshotEntry 上下文压缩快照条目
type ContextCollapseSnapshotEntry struct {
	Type            string       `json:"type"` // "marble-origami-snapshot"
	SessionID       string       `json:"sessionId"`
	Staged          []StagedSpan `json:"staged"`
	Armed           bool         `json:"armed"`
	LastSpawnTokens int          `json:"lastSpawnTokens"`
}

// StagedSpan 暂存的跨度
type StagedSpan struct {
	StartUUID string  `json:"startUuid"`
	EndUUID   string  `json:"endUuid"`
	Summary   string  `json:"summary"`
	Risk      float64 `json:"risk"`
	StagedAt  int64   `json:"stagedAt"`
}

// Entry 日志条目联合类型
type Entry interface{}
