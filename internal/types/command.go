// Package types 提供核心类型定义
// 来源: src/types/command.ts (216行)
// 重构: Go 命令类型系统
package types

// ============================================================================
// 命令结果类型
// ============================================================================

// LocalCommandResult 本地命令结果
// 对应 TS: export type LocalCommandResult = ...
type LocalCommandResult struct {
	Type             string      `json:"type"`
	Value            *string     `json:"value,omitempty"`
	CompactionResult interface{} `json:"compactionResult,omitempty"`
	DisplayText      *string     `json:"displayText,omitempty"`
}

// CommandResultType 命令结果类型
const (
	CommandResultTypeText    = "text"
	CommandResultTypeCompact = "compact"
	CommandResultTypeSkip    = "skip"
)

// ============================================================================
// 提示命令类型
// ============================================================================

// PromptCommand 提示命令
// 对应 TS: export type PromptCommand = {...}
type PromptCommand struct {
	Type                  string         `json:"type"`
	ProgressMessage       string         `json:"progressMessage"`
	ContentLength         int            `json:"contentLength"`
	ArgNames              []string       `json:"argNames,omitempty"`
	AllowedTools          []string       `json:"allowedTools,omitempty"`
	Model                 *string        `json:"model,omitempty"`
	Source                string         `json:"source"` // SettingSource | 'builtin' | 'mcp' | 'plugin' | 'bundled'
	PluginInfo            *PluginInfo    `json:"pluginInfo,omitempty"`
	DisableNonInteractive *bool          `json:"disableNonInteractive,omitempty"`
	Hooks                 *HooksSettings `json:"hooks,omitempty"`
	SkillRoot             *string        `json:"skillRoot,omitempty"`
	Context               *string        `json:"context,omitempty"` // 'inline' | 'fork'
	Agent                 *string        `json:"agent,omitempty"`
	Effort                *string        `json:"effort,omitempty"`
	Paths                 []string       `json:"paths,omitempty"`
}

// PluginInfo 插件信息
type PluginInfo struct {
	PluginManifest PluginManifest `json:"pluginManifest"`
	Repository     string         `json:"repository"`
}

// HooksSettings 钩子设置
type HooksSettings struct {
	// 钩子配置
	Events map[string]interface{} `json:"events,omitempty"`
}

// ============================================================================
// 命令上下文
// ============================================================================

// LocalJSXCommandContext 本地 JSX 命令上下文
// 对应 TS: export type LocalJSXCommandContext = ...
type LocalJSXCommandContext struct {
	CanUseTool               interface{}    `json:"canUseTool,omitempty"` // CanUseToolFn
	SetMessages              interface{}    `json:"setMessages"`          // function
	Options                  CommandOptions `json:"options"`
	OnChangeAPIKey           func()         `json:"onChangeApiKey"`
	OnChangeDynamicMcpConfig interface{}    `json:"onChangeDynamicMcpConfig,omitempty"` // function
	OnInstallIDEExtension    interface{}    `json:"onInstallIdeExtension,omitempty"`    // function
	Resume                   interface{}    `json:"resume,omitempty"`                   // function
}

// CommandOptions 命令选项
type CommandOptions struct {
	DynamicMcpConfig      interface{} `json:"dynamicMcpConfig,omitempty"` // Record<string, ScopedMcpServerConfig>
	IdeInstallationStatus interface{} `json:"ideInstallationStatus"`      // IDEExtensionInstallationStatus | null
	Theme                 string      `json:"theme"`                      // ThemeName
}

// ResumeEntrypoint 恢复入口点
// 对应 TS: export type ResumeEntrypoint = ...
type ResumeEntrypoint string

const (
	ResumeEntrypointCliFlag               ResumeEntrypoint = "cli_flag"
	ResumeEntrypointSlashCommandPicker    ResumeEntrypoint = "slash_command_picker"
	ResumeEntrypointSlashCommandSessionID ResumeEntrypoint = "slash_command_session_id"
	ResumeEntrypointSlashCommandTitle     ResumeEntrypoint = "slash_command_title"
	ResumeEntrypointFork                  ResumeEntrypoint = "fork"
)

// CommandResultDisplay 命令结果显示方式
// 对应 TS: export type CommandResultDisplay = 'skip' | 'system' | 'user'
type CommandResultDisplay string

const (
	CommandResultDisplaySkip   CommandResultDisplay = "skip"
	CommandResultDisplaySystem CommandResultDisplay = "system"
	CommandResultDisplayUser   CommandResultDisplay = "user"
)

// LocalJSXCommandOnDone 命令完成回调
// 对应 TS: export type LocalJSXCommandOnDone = ...
type LocalJSXCommandOnDone func(result *string, options *OnDoneOptions)

// OnDoneOptions 完成选项
type OnDoneOptions struct {
	Display         *CommandResultDisplay `json:"display,omitempty"`
	ShouldQuery     *bool                 `json:"shouldQuery,omitempty"`
	MetaMessages    []string              `json:"metaMessages,omitempty"`
	NextInput       *string               `json:"nextInput,omitempty"`
	SubmitNextInput *bool                 `json:"submitNextInput,omitempty"`
}

// ============================================================================
// 命令可用性
// ============================================================================

// CommandAvailability 命令可用性
// 对应 TS: export type CommandAvailability = ...
type CommandAvailability string

const (
	CommandAvailabilityClaudeAi CommandAvailability = "claude-ai"
	CommandAvailabilityConsole  CommandAvailability = "console"
)

// ============================================================================
// 命令基础定义
// ============================================================================

// CommandBase 命令基础
// 对应 TS: export type CommandBase = {...}
type CommandBase struct {
	Availability                []CommandAvailability `json:"availability,omitempty"`
	Description                 string                `json:"description"`
	HasUserSpecifiedDescription *bool                 `json:"hasUserSpecifiedDescription,omitempty"`
	IsEnabled                   func() bool           `json:"-"` // 运行时检查
	IsHidden                    *bool                 `json:"isHidden,omitempty"`
	Name                        string                `json:"name"`
	Aliases                     []string              `json:"aliases,omitempty"`
	IsMcp                       *bool                 `json:"isMcp,omitempty"`
	ArgumentHint                *string               `json:"argumentHint,omitempty"`
	WhenToUse                   *string               `json:"whenToUse,omitempty"`
	Version                     *string               `json:"version,omitempty"`
	DisableModelInvocation      *bool                 `json:"disableModelInvocation,omitempty"`
	UserInvocable               *bool                 `json:"userInvocable,omitempty"`
	LoadedFrom                  *string               `json:"loadedFrom,omitempty"` // 'commands_DEPRECATED' | 'skills' | 'plugin' | 'managed' | 'bundled' | 'mcp'
	Kind                        *string               `json:"kind,omitempty"`       // 'workflow'
	Immediate                   *bool                 `json:"immediate,omitempty"`
}

// CommandType 命令类型
const (
	CommandTypePrompt   = "prompt"
	CommandTypeLocal    = "local"
	CommandTypeLocalJSX = "local-jsx"
)

// Command 命令 (联合类型)
type Command struct {
	CommandBase
	Type string `json:"type"`
	// PromptCommand 字段
	ProgressMessage *string  `json:"progressMessage,omitempty"`
	ContentLength   *int     `json:"contentLength,omitempty"`
	ArgNames        []string `json:"argNames,omitempty"`
	AllowedTools    []string `json:"allowedTools,omitempty"`
	Model           *string  `json:"model,omitempty"`
	Source          *string  `json:"source,omitempty"`
	// LocalCommand / LocalJSXCommand 字段
	SupportsNonInteractive *bool `json:"supportsNonInteractive,omitempty"`
}

// CommandSource 命令来源类型
const (
	CommandSourceUserSettings    = "userSettings"
	CommandSourceProjectSettings = "projectSettings"
	CommandSourceLocalSettings   = "localSettings"
	CommandSourceBuiltin         = "builtin"
	CommandSourceMcp             = "mcp"
	CommandSourcePlugin          = "plugin"
	CommandSourceBundled         = "bundled"
	CommandSourceManaged         = "managed"
)

// CommandRegistry 命令注册表
type CommandRegistry struct {
	Commands map[string]Command `json:"commands"`
}

// NewCommandRegistry 创建新的命令注册表
func NewCommandRegistry() *CommandRegistry {
	return &CommandRegistry{
		Commands: make(map[string]Command),
	}
}

// Register 注册命令
func (r *CommandRegistry) Register(cmd Command) {
	r.Commands[cmd.Name] = cmd
}

// Get 获取命令
func (r *CommandRegistry) Get(name string) (Command, bool) {
	cmd, ok := r.Commands[name]
	return cmd, ok
}

// List 列出所有命令
func (r *CommandRegistry) List() []Command {
	cmds := make([]Command, 0, len(r.Commands))
	for _, cmd := range r.Commands {
		cmds = append(cmds, cmd)
	}
	return cmds
}
