// Package types 提供核心类型定义
// 来源: src/types/hooks.ts (289行)
// 重构: Go 钩子类型系统 (完整实现)
package types

// HookEvent 钩子事件类型
// 对应 TS: export type HookEvent = ...
type HookEvent string

// 常用钩子事件常量
const (
	HookEventPreToolUse       HookEvent = "PreToolUse"
	HookEventUserPromptSubmit HookEvent = "UserPromptSubmit"
	HookEventSessionStart     HookEvent = "SessionStart"
	HookEventSetup            HookEvent = "Setup"
	HookEventSubagentStart    HookEvent = "SubagentStart"
	HookEventFileChanged      HookEvent = "FileChanged"
)

// HookInput 钩子输入
type HookInput interface{}

// HookJSONOutput 钩子JSON输出
type HookJSONOutput interface{}

// AsyncHookJSONOutput 异步钩子输出
type AsyncHookJSONOutput interface{}

// SyncHookJSONOutput 同步钩子输出
type SyncHookJSONOutput interface{}

// IsHookEvent 检查是否为有效的钩子事件
// 对应 TS: export function isHookEvent(value: string): value is HookEvent
func IsHookEvent(value string) bool {
	hookEvents := []string{
		string(HookEventPreToolUse),
		string(HookEventUserPromptSubmit),
		string(HookEventSessionStart),
		string(HookEventSetup),
		string(HookEventSubagentStart),
		string(HookEventFileChanged),
	}
	for _, event := range hookEvents {
		if event == value {
			return true
		}
	}
	return false
}

// GetAllHookEvents returns all valid hook events.
func GetAllHookEvents() []HookEvent {
	return []HookEvent{
		HookEventPreToolUse,
		HookEventUserPromptSubmit,
		HookEventSessionStart,
		HookEventSetup,
		HookEventSubagentStart,
		HookEventFileChanged,
	}
}

// PromptRequest 提示请求
// 对应 TS: export type PromptRequest = ...
type PromptRequest struct {
	Prompt  string         `json:"prompt"`
	Message string         `json:"message"`
	Options []PromptOption `json:"options"`
}

// PromptOption 提示选项
type PromptOption struct {
	Key         string  `json:"key"`
	Label       string  `json:"label"`
	Description *string `json:"description,omitempty"`
}

// PromptResponse 提示响应
type PromptResponse struct {
	PromptResponse string `json:"prompt_response"`
	Selected       string `json:"selected"`
}

// SyncHookResponse 同步钩子响应
type SyncHookResponse struct {
	Continue           *bool       `json:"continue,omitempty"`
	SuppressOutput     *bool       `json:"suppressOutput,omitempty"`
	StopReason         *string     `json:"stopReason,omitempty"`
	Decision           *string     `json:"decision,omitempty"` // "approve" | "block"
	Reason             *string     `json:"reason,omitempty"`
	SystemMessage      *string     `json:"systemMessage,omitempty"`
	HookSpecificOutput interface{} `json:"hookSpecificOutput,omitempty"`
}

// PreToolUseOutput PreToolUse 钩子输出
type PreToolUseOutput struct {
	HookEventName            string                 `json:"hookEventName"`
	PermissionDecision       *PermissionBehavior    `json:"permissionDecision,omitempty"`
	PermissionDecisionReason *string                `json:"permissionDecisionReason,omitempty"`
	UpdatedInput             map[string]interface{} `json:"updatedInput,omitempty"`
	AdditionalContext        *string                `json:"additionalContext,omitempty"`
}

// UserPromptSubmitOutput UserPromptSubmit 钩子输出
type UserPromptSubmitOutput struct {
	HookEventName     string  `json:"hookEventName"`
	AdditionalContext *string `json:"additionalContext,omitempty"`
}

// SessionStartOutput SessionStart 钩子输出
type SessionStartOutput struct {
	HookEventName      string   `json:"hookEventName"`
	AdditionalContext  *string  `json:"additionalContext,omitempty"`
	InitialUserMessage *string  `json:"initialUserMessage,omitempty"`
	WatchPaths         []string `json:"watchPaths,omitempty"`
}

// SetupOutput Setup 钩子输出
type SetupOutput struct {
	HookEventName     string  `json:"hookEventName"`
	AdditionalContext *string `json:"additionalContext,omitempty"`
}

// SubagentStartOutput SubagentStart 钩子输出
type SubagentStartOutput struct {
	HookEventName     string  `json:"hookEventName"`
	AdditionalContext *string `json:"additionalContext,omitempty"`
}

// HookResult 钩子结果
type HookResult struct {
	Continue       bool        `json:"continue"`
	SuppressOutput bool        `json:"suppressOutput"`
	StopReason     string      `json:"stopReason,omitempty"`
	Output         interface{} `json:"output,omitempty"`
}

// AsyncHookResult 异步钩子结果
type AsyncHookResult struct {
	Type           string      `json:"type"`
	Continue       bool        `json:"continue"`
	SuppressOutput bool        `json:"suppressOutput"`
	StopReason     string      `json:"stopReason,omitempty"`
	Output         interface{} `json:"output,omitempty"`
	Events         []HookEvent `json:"events,omitempty"`
}

// EventStreamItem 事件流项
type EventStreamItem struct {
	Type   string      `json:"type"`
	Event  *HookEvent  `json:"event,omitempty"`
	Output interface{} `json:"output,omitempty"`
	Error  *string     `json:"error,omitempty"`
}

// HookRegistration 钩子注册
type HookRegistration struct {
	HookName string                 `json:"hookName"`
	Event    HookEvent              `json:"event"`
	Options  map[string]interface{} `json:"options,omitempty"`
}
