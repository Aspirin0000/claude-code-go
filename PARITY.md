# PARITY.md - Claude Code Go 迁移追踪

## 执行摘要

Go 实现已建立基础框架，正在逐步完善功能对等性。

**最大差距：**
- MCP 客户端完整传输层 (SSE, WebSocket, HTTP 详细实现)
- 命令系统 (207个命令文件待迁移)
- 工具系统完整实现 (55个工具)
- 插件/钩子系统
- 流式响应处理

---

## tools/ 工具系统

### TS 存在
证据：`claude-code-main/src/tools/` 包含 55 个工具目录
- 核心工具: `BashTool`, `FileReadTool`, `FileWriteTool`, `FileEditTool`, `GrepTool`, `GlobTool`
- MCP 工具: `ListMcpResourcesTool`, `MCPTool`, `McpAuthTool`, `ReadMcpResourceTool`
- 工作流工具: `TodoWriteTool`, `TaskTool`, `AgentTool`
- 网络工具: `WebSearchTool`, `WebFetchTool`
- IDE 工具: `LSPTool`
- 等

### Go 存在
证据：`internal/tools/tools.go`
- 基础工具接口定义 (Tool interface)
- 9 个基础工具实现: Bash, FileRead, FileWrite, FileEdit, Grep, Glob, TodoWrite, WebSearch, WebFetch
- 工具注册表 (Registry)
- 扩展工具骨架: NotebookEdit, Task* x4, Agent (待填充)

### Rust 差距
- ❌ 无 Rust 等效工具 (我们使用 Go)
- ⚠️ 扩展工具为骨架实现
- ❌ MCP 工具未实现
- ❌ 大部分工作流工具未实现

**状态:** 核心工具 9/55 完成

---

## mcp/ MCP 服务

### TS 存在
证据：`claude-code-main/src/services/mcp/`
- client.ts (3351行) - 完整 MCP 客户端
- config.ts (1579行) - 配置管理
- types.ts - 类型定义
- auth.ts - OAuth 认证

### Go 存在
证据：`internal/mcp/` (12个文件, ~6,200行)
- ✅ types.go (258行) - 完整类型系统
- ✅ config.go (335行) - 核心配置函数
- ✅ client.go (727行) - Client 结构体、错误处理、认证缓存
- ✅ transport.go (350行) - HTTP/SSE/Stdio 传输实现
- ✅ connection.go (380行) - 连接管理器、批量连接
- ✅ cache.go (344行) - LRU缓存、工具获取
- ✅ auth.go (472行) - OAuth认证、Token管理
- ✅ websocket.go (292行) - WebSocket传输、重连逻辑
- ✅ executor.go (246行) - 工具执行、重试逻辑
- ✅ manager.go (731行) - MCP管理器、生命周期

**状态:** 客户端核心 95% 完成

### 已完成功能
- ✅ 错误类型系统 (McpAuthError, McpSessionExpiredError, McpToolCallError)
- ✅ 认证缓存 (TTL、文件持久化、线程安全)
- ✅ Client结构体 (初始化、握手、请求/响应)
- ✅ 传输层 (HTTP、SSE、Stdio、WebSocket)
- ✅ 连接管理 (批量连接、状态管理、超时控制)
- ✅ LRU缓存 (工具、资源、提示)
- ✅ OAuth认证 (Token管理、刷新、吊销)
- ✅ 工具执行 (重试逻辑、进度报告、错误包装)
- ✅ MCP管理器 (服务器生命周期、配置集成)

### 待完善
- ⚠️ ClaudeAI代理特殊处理
- ⚠️ Chrome/Computer Use内进程服务器

---

## commands/ 命令系统

### TS 存在
证据：`claude-code-main/src/commands/` (207个文件)
- agents/, bash/, clear/, compact/, config/, cost/
- diff/, exit/, git/, help/, hooks/, init/, load/
- login/, logout/, memory/, model/, mcp/, permissions/
- plan/, plugin/, quit/, reload/, resume/, review/
- save/, skills/, tasks/, team/, todos/, version/
- 等

### Go 存在
证据：`cmd/claude/commands/` (14个文件, ~3,400行)
- ✅ base.go - Command接口和BaseCommand
- ✅ registry.go - 线程安全命令注册表

**CMD-1批次 (5个命令):**
- ✅ /help - 命令帮助系统
- ✅ /status - 会话状态查看
- ✅ /clear - 清屏
- ✅ /version - 版本信息
- ✅ /exit - 退出

**CMD-2批次 (4个命令):**
- ✅ /compact - 压缩对话历史
- ✅ /resume - 恢复会话
- ✅ /save - 保存对话
- ✅ /load - 加载对话

**CMD-3批次 (3个命令):**
- ✅ /config - 配置管理
- ✅ /model - 模型切换
- ✅ /permissions - 权限管理

**状态:** 13/207 命令完成 (6.3%)

### 待实现
- ❌ 194个命令文件待实现
- ❌ MCP命令 (/mcp, /mcp-add, /mcp-list)
- ❌ 文件操作命令 (/read, /write, /edit)
- ❌ 工具命令 (/bash, /git, /grep)
- ❌ 高级命令 (/plan, /review, /tasks)
- ❌ 插件命令 (/plugin, /hooks, /skills)

---

## types/ 类型系统

### TS 存在
证据：`claude-code-main/src/types/`
- message.ts, permissions.ts, tools.ts, hooks.ts, logs.ts
- global.ts, command.ts, ids.ts

### Go 存在
证据：`internal/types/`
- ✅ ids.go - ID 类型
- ✅ utils.go - 工具类型
- ✅ global.go - 全局类型
- ✅ message.go - 消息类型
- ✅ queue.go - 队列
- ✅ logs.go - 日志
- ✅ permissions.go - 权限
- ✅ hooks.go - 钩子
- ✅ tools.go - 工具
- ✅ command.go - 命令
- ✅ plugin.go - 插件

**状态:** 100% 完成

---

## cli/ 命令行界面

### TS 存在
- 结构化/远程传输层
- 处理程序分解
- JSON 提示模式

### Go 存在
- ✅ cmd/claude/main.go - 入口点
- ✅ Cobra + Viper 框架
- ✅ 基础 REPL 循环

### 缺失
- ❌ 结构化 IO
- ❌ 远程传输层
- ❌ JSON 模式

**状态:** 30% 完成

---

## services/ 服务层

### TS 存在
- api/, oauth/, mcp/, analytics/
- settings sync, policy limits
- team memory sync

### Go 存在
- ⚠️ internal/services/analytics/ - 骨架
- ⚠️ MCP 客户端部分实现

### 缺失
- ❌ API 客户端 (Anthropic SDK)
- ❌ OAuth 完整实现
- ❌ 分析服务
- ❌ 设置同步
- ❌ 策略限制

**状态:** 5% 完成

---

## internal/ 内部工具

### TS 存在
- utils/, bootstrap/, state/

### Go 存在
- ✅ internal/bootstrap/state.go
- ✅ internal/utils/claudeinchrome/
- ✅ internal/settings/
- ✅ internal/plugins/

### 状态
- ⚠️ 骨架实现多，需填充

---

## 关键依赖

### 外部 SDK
- ❌ @anthropic-ai/sdk - 需要 Go 等效库
- ❌ @modelcontextprotocol/sdk - 已部分实现
- ⚠️ Bun 特定 API - 需要适配

---

## 下一步优先级

### P0 (核心阻塞)
1. MCP Client 传输层完善 (C-3/8 ~ C-8/8)
2. Anthropic API 客户端
3. Query 引擎 (核心对话循环)

### P1 (重要功能)
4. 命令系统 (207个文件)
5. 工具系统完善 (55个工具)
6. OAuth 认证

### P2 (增强)
7. 插件系统
8. 钩子系统
9. 分析/遥测
10. 测试套件

---

## 统计

- **总 TS 文件:** 2,216
- **已迁移/骨架:** ~50
- **完成率:** ~2.3%
- **代码行数:** ~5,000 Go 代码

---

*最后更新: 2026-04-01*
