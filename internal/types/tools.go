// Package types 提供核心类型定义
// 来源: src/types/tools.ts (12行)
// 重构: Go 工具进度类型
package types

// 工具进度类型 (原 TS 中都是 any，这里定义为 interface{})

// AgentToolProgress 代理工具进度
type AgentToolProgress interface{}

// BashProgress Bash 进度
type BashProgress interface{}

// MCPProgress MCP 进度
type MCPProgress interface{}

// REPLToolProgress REPL 工具进度
type REPLToolProgress interface{}

// SkillToolProgress 技能工具进度
type SkillToolProgress interface{}

// TaskOutputProgress 任务输出进度
type TaskOutputProgress interface{}

// ToolProgressData 工具进度数据
type ToolProgressData interface{}

// WebSearchProgress Web 搜索进度
type WebSearchProgress interface{}

// ShellProgress Shell 进度
type ShellProgress interface{}

// PowerShellProgress PowerShell 进度
type PowerShellProgress interface{}

// SdkWorkflowProgress SDK 工作流进度
type SdkWorkflowProgress interface{}
