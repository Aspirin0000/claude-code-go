// Package types 提供核心类型定义
// 来源: src/types/permissions.ts (441行)
// 重构: Go 权限类型系统
package types

// ============================================================================
// 权限模式 (Permission Modes)
// ============================================================================

// ExternalPermissionMode 外部权限模式
// 对应 TS: export type ExternalPermissionMode = ...
type ExternalPermissionMode string

const (
	ExternalPermissionModeAcceptEdits       ExternalPermissionMode = "acceptEdits"
	ExternalPermissionModeBypassPermissions ExternalPermissionMode = "bypassPermissions"
	ExternalPermissionModeDefault           ExternalPermissionMode = "default"
	ExternalPermissionModeDontAsk           ExternalPermissionMode = "dontAsk"
	ExternalPermissionModePlan              ExternalPermissionMode = "plan"
)

// ExternalPermissionModes 所有外部权限模式
// 对应 TS: export const EXTERNAL_PERMISSION_MODES = [...]
var ExternalPermissionModes = []ExternalPermissionMode{
	ExternalPermissionModeAcceptEdits,
	ExternalPermissionModeBypassPermissions,
	ExternalPermissionModeDefault,
	ExternalPermissionModeDontAsk,
	ExternalPermissionModePlan,
}

// InternalPermissionMode 内部权限模式
type InternalPermissionMode string

const (
	InternalPermissionModeAuto   InternalPermissionMode = "auto"
	InternalPermissionModeBubble InternalPermissionMode = "bubble"
)

// PermissionMode 权限模式 (外部 + 内部)
type PermissionMode = ExternalPermissionMode

// InternalPermissionModes 所有内部权限模式
var InternalPermissionModes = []PermissionMode{
	ExternalPermissionModeAcceptEdits,
	ExternalPermissionModeBypassPermissions,
	ExternalPermissionModeDefault,
	ExternalPermissionModeDontAsk,
	ExternalPermissionModePlan,
	// "auto" 需要 feature flag，这里不包含
}

// ============================================================================
// 权限行为 (Permission Behaviors)
// ============================================================================

// PermissionBehavior 权限行为
// 对应 TS: export type PermissionBehavior = 'allow' | 'deny' | 'ask'
type PermissionBehavior string

const (
	PermissionBehaviorAllow PermissionBehavior = "allow"
	PermissionBehaviorDeny  PermissionBehavior = "deny"
	PermissionBehaviorAsk   PermissionBehavior = "ask"
)

// ============================================================================
// 权限规则 (Permission Rules)
// ============================================================================

// PermissionRuleSource 权限规则来源
// 对应 TS: export type PermissionRuleSource = ...
type PermissionRuleSource string

const (
	PermissionRuleSourceUserSettings    PermissionRuleSource = "userSettings"
	PermissionRuleSourceProjectSettings PermissionRuleSource = "projectSettings"
	PermissionRuleSourceLocalSettings   PermissionRuleSource = "localSettings"
	PermissionRuleSourceFlagSettings    PermissionRuleSource = "flagSettings"
	PermissionRuleSourcePolicySettings  PermissionRuleSource = "policySettings"
	PermissionRuleSourceCliArg          PermissionRuleSource = "cliArg"
	PermissionRuleSourceCommand         PermissionRuleSource = "command"
	PermissionRuleSourceSession         PermissionRuleSource = "session"
)

// PermissionRuleValue 权限规则值
type PermissionRuleValue struct {
	ToolName    string  `json:"toolName"`
	RuleContent *string `json:"ruleContent,omitempty"`
}

// PermissionRule 权限规则
type PermissionRule struct {
	Source       PermissionRuleSource `json:"source"`
	RuleBehavior PermissionBehavior   `json:"ruleBehavior"`
	RuleValue    PermissionRuleValue  `json:"ruleValue"`
}

// ============================================================================
// 权限更新 (Permission Updates)
// ============================================================================

// PermissionUpdateDestination 权限更新目标
type PermissionUpdateDestination string

const (
	PermissionUpdateDestinationUserSettings    PermissionUpdateDestination = "userSettings"
	PermissionUpdateDestinationProjectSettings PermissionUpdateDestination = "projectSettings"
	PermissionUpdateDestinationLocalSettings   PermissionUpdateDestination = "localSettings"
	PermissionUpdateDestinationSession         PermissionUpdateDestination = "session"
	PermissionUpdateDestinationCliArg          PermissionUpdateDestination = "cliArg"
)

// PermissionUpdate 权限更新
// 对应 TS: export type PermissionUpdate = ...
type PermissionUpdate struct {
	Type        string                      `json:"type"`
	Destination PermissionUpdateDestination `json:"destination,omitempty"`
	Rules       []PermissionRuleValue       `json:"rules,omitempty"`
	Behavior    PermissionBehavior          `json:"behavior,omitempty"`
	Mode        ExternalPermissionMode      `json:"mode,omitempty"`
	Directories []string                    `json:"directories,omitempty"`
}

// ============================================================================
// 工作目录 (Working Directories)
// ============================================================================

// WorkingDirectorySource 工作目录来源
type WorkingDirectorySource = PermissionRuleSource

// AdditionalWorkingDirectory 额外工作目录
type AdditionalWorkingDirectory struct {
	Path   string                 `json:"path"`
	Source WorkingDirectorySource `json:"source"`
}

// ============================================================================
// 权限决策 (Permission Decisions)
// ============================================================================

// PermissionCommandMetadata 权限命令元数据
type PermissionCommandMetadata struct {
	Name        string                 `json:"name"`
	Description *string                `json:"description,omitempty"`
	Extra       map[string]interface{} `json:"-"`
}

// PermissionMetadata 权限元数据
type PermissionMetadata struct {
	Command *PermissionCommandMetadata `json:"command,omitempty"`
}

// PermissionAllowDecision 允许决策
type PermissionAllowDecision struct {
	Behavior       PermissionBehavior        `json:"behavior"`
	UpdatedInput   interface{}               `json:"updatedInput,omitempty"`
	UserModified   *bool                     `json:"userModified,omitempty"`
	DecisionReason *PermissionDecisionReason `json:"decisionReason,omitempty"`
	ToolUseID      *string                   `json:"toolUseId,omitempty"`
	AcceptFeedback *string                   `json:"acceptFeedback,omitempty"`
	ContentBlocks  interface{}               `json:"contentBlocks,omitempty"`
}

// PendingClassifierCheck 待处理的分类器检查
type PendingClassifierCheck struct {
	Command      string   `json:"command"`
	Cwd          string   `json:"cwd"`
	Descriptions []string `json:"descriptions"`
}

// PermissionAskDecision 询问决策
type PermissionAskDecision struct {
	Behavior               PermissionBehavior         `json:"behavior"`
	PendingClassifierCheck *PendingClassifierCheck    `json:"pendingClassifierCheck,omitempty"`
	Reasons                []PermissionDecisionReason `json:"reasons"`
}

// PermissionDenyDecision 拒绝决策
type PermissionDenyDecision struct {
	Behavior       PermissionBehavior         `json:"behavior"`
	Reasons        []PermissionDecisionReason `json:"reasons"`
	UpdatedInput   interface{}                `json:"updatedInput,omitempty"`
	DecisionReason *PermissionDecisionReason  `json:"decisionReason,omitempty"`
}

// PermissionResult 权限结果
type PermissionResult struct {
	Behavior               PermissionBehavior         `json:"behavior"`
	UpdatedInput           interface{}                `json:"updatedInput,omitempty"`
	UserModified           *bool                      `json:"userModified,omitempty"`
	DecisionReason         *PermissionDecisionReason  `json:"decisionReason,omitempty"`
	Reasons                []PermissionDecisionReason `json:"reasons,omitempty"`
	PendingClassifierCheck *PendingClassifierCheck    `json:"pendingClassifierCheck,omitempty"`
	ToolUseID              *string                    `json:"toolUseId,omitempty"`
}

// PermissionDecisionReason 权限决策原因
type PermissionDecisionReason struct {
	Type                 string  `json:"type"`
	Reason               string  `json:"reason"`
	Policy               *string `json:"policy,omitempty"`
	Classifier           *string `json:"classifier,omitempty"`
	ClassifierApprovable *bool   `json:"classifierApprovable,omitempty"`
}

// ============================================================================
// 分类器类型 (Classifier Types)
// ============================================================================

// ClassifierConfidence 分类器置信度
type ClassifierConfidence string

const (
	ClassifierConfidenceHigh   ClassifierConfidence = "high"
	ClassifierConfidenceMedium ClassifierConfidence = "medium"
	ClassifierConfidenceLow    ClassifierConfidence = "low"
)

// ClassifierResult 分类器结果
type ClassifierResult struct {
	Matches            bool                 `json:"matches"`
	MatchedDescription *string              `json:"matchedDescription,omitempty"`
	Confidence         ClassifierConfidence `json:"confidence"`
	Reason             string               `json:"reason"`
}

// ClassifierBehavior 分类器行为
type ClassifierBehavior string

const (
	ClassifierBehaviorDeny  ClassifierBehavior = "deny"
	ClassifierBehaviorAsk   ClassifierBehavior = "ask"
	ClassifierBehaviorAllow ClassifierBehavior = "allow"
)

// ClassifierUsage 分类器使用量
type ClassifierUsage struct {
	InputTokens              int `json:"inputTokens"`
	OutputTokens             int `json:"outputTokens"`
	CacheReadInputTokens     int `json:"cacheReadInputTokens"`
	CacheCreationInputTokens int `json:"cacheCreationInputTokens"`
}

// YoloClassifierResult Yolo分类器结果
type YoloClassifierResult struct {
	Thinking          *string          `json:"thinking,omitempty"`
	ShouldBlock       bool             `json:"shouldBlock"`
	Reason            string           `json:"reason"`
	Unavailable       *bool            `json:"unavailable,omitempty"`
	TranscriptTooLong *bool            `json:"transcriptTooLong,omitempty"`
	Model             string           `json:"model"`
	Usage             *ClassifierUsage `json:"usage,omitempty"`
	DurationMs        *int             `json:"durationMs,omitempty"`
	PromptLengths     *PromptLengths   `json:"promptLengths,omitempty"`
	ErrorDumpPath     *string          `json:"errorDumpPath,omitempty"`
	Stage             *string          `json:"stage,omitempty"`
	Stage1Usage       *ClassifierUsage `json:"stage1Usage,omitempty"`
	Stage1DurationMs  *int             `json:"stage1DurationMs,omitempty"`
	Stage1RequestID   *string          `json:"stage1RequestId,omitempty"`
	Stage1MsgID       *string          `json:"stage1MsgId,omitempty"`
	Stage2Usage       *ClassifierUsage `json:"stage2Usage,omitempty"`
	Stage2DurationMs  *int             `json:"stage2DurationMs,omitempty"`
	Stage2RequestID   *string          `json:"stage2RequestId,omitempty"`
	Stage2MsgID       *string          `json:"stage2MsgId,omitempty"`
}

// PromptLengths 提示长度
type PromptLengths struct {
	SystemPrompt int `json:"systemPrompt"`
	ToolCalls    int `json:"toolCalls"`
	UserPrompts  int `json:"userPrompts"`
}
