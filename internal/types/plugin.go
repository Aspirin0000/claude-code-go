// Package types 提供核心类型定义
// 来源: src/types/plugin.ts (363行)
// 重构: Go 插件类型系统
package types

// ============================================================================
// 插件清单 (Plugin Manifest)
// ============================================================================

// PluginManifest 插件清单
// 对应 TS: export type PluginManifest = ...
type PluginManifest struct {
	Name        string         `json:"name"`
	Version     string         `json:"version"`
	Description string         `json:"description"`
	Author      *PluginAuthor  `json:"author,omitempty"`
	Repository  *string        `json:"repository,omitempty"`
	License     *string        `json:"license,omitempty"`
	Commands    interface{}    `json:"commands,omitempty"` // CommandMetadata[] | Record<string, CommandMetadata>
	Agents      interface{}    `json:"agents,omitempty"`   // string[] | AgentDefinition
	Skills      interface{}    `json:"skills,omitempty"`   // BundledSkillDefinition[]
	Hooks       *HooksSettings `json:"hooks,omitempty"`
	McpServers  interface{}    `json:"mcpServers,omitempty"` // Record<string, McpServerConfig>
	LspServers  interface{}    `json:"lspServers,omitempty"` // Record<string, LspServerConfig>
}

// PluginAuthor 插件作者
type PluginAuthor struct {
	Name  string  `json:"name"`
	Email *string `json:"email,omitempty"`
}

// CommandMetadata 命令元数据
type CommandMetadata struct {
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Aliases     []string           `json:"aliases,omitempty"`
	Arguments   []ArgumentMetadata `json:"arguments,omitempty"`
	IsHidden    *bool              `json:"isHidden,omitempty"`
}

// ArgumentMetadata 参数元数据
type ArgumentMetadata struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Required    *bool   `json:"required,omitempty"`
	Type        *string `json:"type,omitempty"`
}

// ============================================================================
// 内置插件定义
// ============================================================================

// BuiltinPluginDefinition 内置插件定义
// 对应 TS: export type BuiltinPluginDefinition = {...}
type BuiltinPluginDefinition struct {
	Name           string         `json:"name"`
	Description    string         `json:"description"`
	Version        *string        `json:"version,omitempty"`
	Skills         interface{}    `json:"skills,omitempty"`
	Hooks          *HooksSettings `json:"hooks,omitempty"`
	McpServers     interface{}    `json:"mcpServers,omitempty"`
	IsAvailable    func() bool    `json:"-"` // 运行时检查
	DefaultEnabled *bool          `json:"defaultEnabled,omitempty"`
}

// ============================================================================
// 插件仓库和配置
// ============================================================================

// PluginRepository 插件仓库
// 对应 TS: export type PluginRepository = {...}
type PluginRepository struct {
	URL         string  `json:"url"`
	Branch      string  `json:"branch"`
	LastUpdated *string `json:"lastUpdated,omitempty"`
	CommitSha   *string `json:"commitSha,omitempty"`
}

// PluginConfig 插件配置
type PluginConfig struct {
	Repositories map[string]PluginRepository `json:"repositories"`
}

// ============================================================================
// 已加载的插件
// ============================================================================

// LoadedPlugin 已加载的插件
// 对应 TS: export type LoadedPlugin = {...}
type LoadedPlugin struct {
	Name              string                     `json:"name"`
	Manifest          PluginManifest             `json:"manifest"`
	Path              string                     `json:"path"`
	Source            string                     `json:"source"`
	Repository        string                     `json:"repository"`
	Enabled           *bool                      `json:"enabled,omitempty"`
	IsBuiltin         *bool                      `json:"isBuiltin,omitempty"`
	Sha               *string                    `json:"sha,omitempty"`
	CommandsPath      *string                    `json:"commandsPath,omitempty"`
	CommandsPaths     []string                   `json:"commandsPaths,omitempty"`
	CommandsMetadata  map[string]CommandMetadata `json:"commandsMetadata,omitempty"`
	AgentsPath        *string                    `json:"agentsPath,omitempty"`
	AgentsPaths       []string                   `json:"agentsPaths,omitempty"`
	SkillsPath        *string                    `json:"skillsPath,omitempty"`
	SkillsPaths       []string                   `json:"skillsPaths,omitempty"`
	OutputStylesPath  *string                    `json:"outputStylesPath,omitempty"`
	OutputStylesPaths []string                   `json:"outputStylesPaths,omitempty"`
	HooksConfig       *HooksSettings             `json:"hooksConfig,omitempty"`
	McpServers        interface{}                `json:"mcpServers,omitempty"`
	LspServers        interface{}                `json:"lspServers,omitempty"`
	Settings          map[string]interface{}     `json:"settings,omitempty"`
}

// ============================================================================
// 插件组件和错误
// ============================================================================

// PluginComponent 插件组件类型
// 对应 TS: export type PluginComponent = ...
type PluginComponent string

const (
	PluginComponentCommands     PluginComponent = "commands"
	PluginComponentAgents       PluginComponent = "agents"
	PluginComponentSkills       PluginComponent = "skills"
	PluginComponentHooks        PluginComponent = "hooks"
	PluginComponentOutputStyles PluginComponent = "output-styles"
)

// PluginError 插件错误
// 对应 TS: export type PluginError = ...
type PluginError struct {
	Type                  string   `json:"type"`
	Source                string   `json:"source"`
	Plugin                *string  `json:"plugin,omitempty"`
	Path                  *string  `json:"path,omitempty"`
	Component             *string  `json:"component,omitempty"`
	GitURL                *string  `json:"gitUrl,omitempty"`
	AuthType              *string  `json:"authType,omitempty"`
	Operation             *string  `json:"operation,omitempty"`
	URL                   *string  `json:"url,omitempty"`
	Details               *string  `json:"details,omitempty"`
	ManifestPath          *string  `json:"manifestPath,omitempty"`
	ParseError            *string  `json:"parseError,omitempty"`
	ValidationErrors      []string `json:"validationErrors,omitempty"`
	PluginID              *string  `json:"pluginId,omitempty"`
	Marketplace           *string  `json:"marketplace,omitempty"`
	AvailableMarketplaces []string `json:"availableMarketplaces,omitempty"`
	Reason                *string  `json:"reason,omitempty"`
	ServerName            *string  `json:"serverName,omitempty"`
	ValidationError       *string  `json:"validationError,omitempty"`
	DuplicateOf           *string  `json:"duplicateOf,omitempty"`
	HookPath              *string  `json:"hookPath,omitempty"`
	McpbPath              *string  `json:"mcpbPath,omitempty"`
	ExitCode              *int     `json:"exitCode,omitempty"`
	Signal                *string  `json:"signal,omitempty"`
	Method                *string  `json:"method,omitempty"`
	TimeoutMs             *int     `json:"timeoutMs,omitempty"`
	Error                 *string  `json:"error,omitempty"`
	BlockedByBlocklist    *bool    `json:"blockedByBlocklist,omitempty"`
	AllowedSources        []string `json:"allowedSources,omitempty"`
	Dependency            *string  `json:"dependency,omitempty"`
	InstallPath           *string  `json:"installPath,omitempty"`
}

// PluginErrorType 插件错误类型
const (
	PluginErrorTypeGeneric                      = "generic-error"
	PluginErrorTypePathNotFound                 = "path-not-found"
	PluginErrorTypeGitAuthFailed                = "git-auth-failed"
	PluginErrorTypeGitTimeout                   = "git-timeout"
	PluginErrorTypeNetworkError                 = "network-error"
	PluginErrorTypeManifestParseError           = "manifest-parse-error"
	PluginErrorTypeManifestValidationError      = "manifest-validation-error"
	PluginErrorTypePluginNotFound               = "plugin-not-found"
	PluginErrorTypeMarketplaceNotFound          = "marketplace-not-found"
	PluginErrorTypeMarketplaceLoadFailed        = "marketplace-load-failed"
	PluginErrorTypeMcpConfigInvalid             = "mcp-config-invalid"
	PluginErrorTypeMcpServerSuppressedDuplicate = "mcp-server-suppressed-duplicate"
	PluginErrorTypeMcpbDownloadFailed           = "mcpb-download-failed"
	PluginErrorTypeMcpbExtractFailed            = "mcpb-extract-failed"
	PluginErrorTypeMcpbInvalidManifest          = "mcpb-invalid-manifest"
	PluginErrorTypeLspConfigInvalid             = "lsp-config-invalid"
	PluginErrorTypeLspServerStartFailed         = "lsp-server-start-failed"
	PluginErrorTypeLspServerCrashed             = "lsp-server-crashed"
	PluginErrorTypeLspRequestTimeout            = "lsp-request-timeout"
	PluginErrorTypeLspRequestFailed             = "lsp-request-failed"
	PluginErrorTypeMarketplaceBlockedByPolicy   = "marketplace-blocked-by-policy"
	PluginErrorTypeDependencyUnsatisfied        = "dependency-unsatisfied"
	PluginErrorTypePluginCacheMiss              = "plugin-cache-miss"
	PluginErrorTypeHookLoadFailed               = "hook-load-failed"
	PluginErrorTypeComponentLoadFailed          = "component-load-failed"
)

// PluginLoadResult 插件加载结果
type PluginLoadResult struct {
	Enabled  []LoadedPlugin `json:"enabled"`
	Disabled []LoadedPlugin `json:"disabled"`
	Errors   []PluginError  `json:"errors"`
}

// ============================================================================
// 插件管理
// ============================================================================

// PluginManager 插件管理器接口
type PluginManager interface {
	LoadPlugin(source string, name string) (*LoadedPlugin, error)
	UnloadPlugin(name string) error
	EnablePlugin(name string) error
	DisablePlugin(name string) error
	ListPlugins() []LoadedPlugin
	GetPlugin(name string) (*LoadedPlugin, bool)
}

// GetPluginErrorMessage 获取插件错误消息
// 对应 TS: export function getPluginErrorMessage(error: PluginError): string
func GetPluginErrorMessage(error PluginError) string {
	switch error.Type {
	case PluginErrorTypeGeneric:
		if error.Error != nil {
			return *error.Error
		}
		return "Unknown error"
	case PluginErrorTypePathNotFound:
		if error.Path != nil && error.Component != nil {
			return "Path not found: " + *error.Path + " (" + *error.Component + ")"
		}
	case PluginErrorTypePluginNotFound:
		if error.PluginID != nil && error.Marketplace != nil {
			return "Plugin " + *error.PluginID + " not found in marketplace " + *error.Marketplace
		}
	case PluginErrorTypeNetworkError:
		if error.URL != nil {
			msg := "Network error: " + *error.URL
			if error.Details != nil {
				msg += " - " + *error.Details
			}
			return msg
		}
	}
	return "Plugin error: " + error.Type
}
