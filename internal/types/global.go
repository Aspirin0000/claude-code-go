// Package types 提供核心类型定义
// 来源: src/types/global.d.ts (91行)
// 重构: Go 常量和全局声明
package types

// ============================================================================
// 构建时常量
// 对应 TS: declare const BUILD_TARGET, BUILD_ENV, INTERFACE_TYPE
// 注意: Go 中使用构建标签或 ldflags 注入
// ============================================================================

// BuildInfo 构建信息
// 对应 TS 中的 MACRO 常量
type BuildInfo struct {
	Version          string // 对应 MACRO.VERSION
	BuildTime        string // 对应 MACRO.BUILD_TIME
	FeedbackChannel  string // 对应 MACRO.FEEDBACK_CHANNEL
	IssuesExplainer  string // 对应 MACRO.ISSUES_EXPLAINER
	NativePackageURL string // 对应 MACRO.NATIVE_PACKAGE_URL
	PackageURL       string // 对应 MACRO.PACKAGE_URL
	VersionChangelog string // 对应 MACRO.VERSION_CHANGELOG
}

// BuildTarget 构建目标平台
// 对应 TS: declare const BUILD_TARGET: string
type BuildTarget string

const (
	BuildTargetDarwinAMD64 BuildTarget = "darwin-amd64"
	BuildTargetDarwinARM64 BuildTarget = "darwin-arm64"
	BuildTargetLinuxAMD64  BuildTarget = "linux-amd64"
	BuildTargetLinuxARM64  BuildTarget = "linux-arm64"
	BuildTargetWindows     BuildTarget = "windows"
)

// BuildEnv 构建环境
// 对应 TS: declare const BUILD_ENV: string
type BuildEnv string

const (
	BuildEnvDevelopment BuildEnv = "development"
	BuildEnvProduction  BuildEnv = "production"
	BuildEnvStaging     BuildEnv = "staging"
)

// InterfaceType 接口类型
// 对应 TS: declare const INTERFACE_TYPE: string
type InterfaceType string

const (
	InterfaceTypeCLI InterfaceType = "cli"
	InterfaceTypeTUI InterfaceType = "tui"
	InterfaceTypeGUI InterfaceType = "gui"
	InterfaceTypeWeb InterfaceType = "web"
)

// ============================================================================
// 文件加载器类型
// 对应 TS: declare module '*.md' / '*.txt' / '*.html' / '*.css'
// 注意: Go 中通过 embed 包实现，此处仅为类型声明
// ============================================================================

// FileContent 文件内容类型
type FileContent string

// EmbeddedFile 嵌入式文件接口
type EmbeddedFile interface {
	Content() string
}

// ============================================================================
// API 指标类型
// 对应 TS: type ApiMetricEntry = {...}
// ============================================================================

// ApiMetricEntry API 性能指标条目
type ApiMetricEntry struct {
	TtftMs                 int64 // Time To First Token (毫秒)
	FirstTokenTime         int64
	LastTokenTime          int64
	ResponseLengthBaseline int
	EndResponseLength      int
}

// ============================================================================
// 钩子时序阈值
// 对应 TS: declare const HOOK_TIMING_DISPLAY_THRESHOLD_MS: number
// ============================================================================

// HookTimingDisplayThresholdMs 钩子时序显示阈值（毫秒）
const HookTimingDisplayThresholdMs = 100
