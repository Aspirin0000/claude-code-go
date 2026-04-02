# Claude Code Go

> 基于 Claude Code v2.1.88 Source Map 还原并翻译为 Go 语言的非官方实现

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Stars](https://img.shields.io/github/stars/Aspirin0000/claude-code-go?style=social)](https://github.com/Aspirin0000/claude-code-go/stargazers)

> ⚠️ **声明**：本仓库为技术学习项目，基于 Anthropic Claude Code npm 包 v2.1.88 的 Source Map 还原 TypeScript 源码，并翻译为 Go 语言实现。仅供学习研究使用，源码版权归 Anthropic 所有。

---

## 📖 项目说明

### 什么是 Claude Code？

Claude Code 是 Anthropic 官方推出的 AI 编程助手命令行工具，可以帮助开发者：

- 🔍 理解和导航代码库
- ✏️ 编写和编辑代码
- 🖥️ 执行终端命令
- 🐛 调试和修复问题
- 🔌 与 MCP (Model Context Protocol) 服务器集成

### 本项目目的

本项目通过 `@anthropic-ai/claude-code` npm 发布包内附带的 Source Map（`cli.js.map`）还原了 TypeScript 源码（版本 2.1.88），并将其**完整翻译**为 Go 语言实现。

**还原规模：**
- 📁 还原文件数：**2,216** 个（含 4756 个 TypeScript 源文件）
- 🔄 翻译方式：逐文件完整翻译，保持逻辑一致
- 📊 代码行数：约 **50,000+** 行 Go 代码

### 为什么选择 Go 版本？

| 特性 | Node.js/TS 版本 | Go 版本 |
|------|----------------|---------|
| **分发方式** | 需要 Node.js 运行时 | 单二进制文件，无需依赖 |
| **跨平台** | 依赖 npm 包 | 静态编译，原生支持多平台 |
| **性能** | V8 解释执行 | Go 编译为机器码，更快启动 |
| **部署** | 依赖管理复杂 | 单文件，易于部署 |
| **并发** | 回调/Promise | Go 协程，更好的并发模型 |

---

## 🔄 翻译进度

### 核心模块翻译状态

| 模块 | TS 文件数 | Go 文件数 | 状态 | 说明 |
|------|-----------|-----------|------|------|
| **类型系统** | ~50 | 11 | ✅ 完成 | 所有核心类型定义 |
| **MCP 协议** | ~35 | 12 | ✅ 完成 | 客户端、配置、传输层 |
| **命令系统** | ~207 | 28 | 🔄 进行中 | 核心斜杠命令已实现 |
| **工具系统** | ~55 | 5 | 🔄 进行中 | Bash、Grep、Glob、File 工具 |
| **API 客户端** | ~25 | 3 | 🔄 进行中 | 重试机制、认证 |
| **配置系统** | ~20 | 4 | 🔄 进行中 | Viper 集成 |
| **OAuth 认证** | ~15 | 2 | ⏳ 计划中 | Token 管理 |
| **UI 系统** | ~144 | 0 | ⏳ 计划中 | Bubble Tea 框架 |
| **Hooks 系统** | ~85 | 0 | ⏳ 计划中 | 事件钩子 |
| **权限系统** | ~24 | 1 | 🔄 进行中 | 基础权限检查 |

**整体进度：约 25%**

### 已实现的斜杠命令 (28个)

#### 核心命令 (5)
- ✅ `/help` - 显示帮助信息
- ✅ `/status` - 显示会话状态
- ✅ `/clear` - 清除终端屏幕
- ✅ `/version` - 显示版本信息
- ✅ `/exit` - 退出应用

#### 会话管理 (4)
- ✅ `/compact` - 压缩对话历史
- ✅ `/resume` - 恢复历史会话
- ✅ `/save` - 保存会话到文件
- ✅ `/load` - 从文件加载会话

#### 配置管理 (3)
- ✅ `/config` - 管理配置项
- ✅ `/model` - 切换 AI 模型
- ✅ `/permissions` - 设置权限级别

#### MCP 管理 (4)
- ✅ `/mcp` - MCP 服务器管理
- ✅ `/mcp-add` - 添加 MCP 服务器
- ✅ `/mcp-list` - 列出 MCP 服务器
- ✅ `/mcp-remove` - 移除 MCP 服务器

#### 高级功能 (8)
- ✅ `/plan` - 创建执行计划
- ✅ `/review` - 审查代码变更
- ✅ `/tasks` - 任务管理
- ✅ `/todos` - 待办事项
- ✅ `/memory` - 会话记忆
- ✅ `/cost` - 成本追踪
- ✅ `/diff` - Git 差异查看
- ✅ `/doctor` - 系统诊断

#### AI 工具命令 (4)
- ✅ `/bash` - 执行 Bash 命令（带安全检查）
- ✅ `/git` - Git 操作助手
- ✅ `/grep` - 文件内容搜索
- ✅ `/glob` - 文件名模式匹配

*注：系统命令（ls, ps, docker 等）通过 BashTool 执行，不作为独立斜杠命令*

---

## ✨ Go 版本实现功能

### 🔐 认证系统

- **API Key 支持**：直接设置 `ANTHROPIC_API_KEY`
- **OAuth Token**：支持 OAuth bearer token
- **多云认证**：AWS Bedrock、Google Vertex、Azure（计划中）

### 🛠️ 工具系统

#### BashTool
- **安全执行**：危险命令检测（rm -rf 等）
- **权限检查**：根据权限级别控制执行
- **超时控制**：默认 30 秒超时
- **只读检测**：自动识别只读命令

#### GrepTool
- **并行搜索**：多 goroutine 并发搜索
- **正则支持**：完整的正则表达式匹配
- **结果限制**：默认最多 50 个结果
- **二进制检测**：自动跳过二进制文件

#### GlobTool
- **模式匹配**：支持 `*` 和 `**` 通配符
- **递归搜索**：自动递归子目录
- **权限检查**：检查文件访问权限

#### 文件操作工具
- **FileReadTool**：读取文件内容，支持语法高亮
- **FileWriteTool**：写入文件，自动备份
- **FileEditTool**：AI 辅助编辑，显示 diff

### 🔌 MCP (Model Context Protocol)

- **传输层**：支持 stdio 和 HTTP/SSE 传输
- **服务器管理**：添加、删除、列出 MCP 服务器
- **工具发现**：自动发现 MCP 服务器提供的工具
- **配置管理**：支持环境变量扩展 `${VAR:-default}`

### 📝 命令系统

- **斜杠命令**：/help, /compact, /config 等
- **自动注册**：命令自动注册到系统
- **分类管理**：按类别组织命令（核心、会话、配置等）
- **别名支持**：命令支持多个别名

### 💾 会话管理

- **保存/加载**：支持 JSON 和 Markdown 格式
- **历史记录**：保留会话历史
- **Token 追踪**：估算 token 使用量
- **压缩功能**：自动压缩长对话

---

## 🚀 安装部署

### 前置要求

- **Go 版本**：Go 1.21 或更高版本
- **操作系统**：Linux、macOS、Windows
- **API 访问**：Anthropic API 密钥

### 方式一：从源码构建

```bash
# 1. 克隆仓库
git clone https://github.com/Aspirin0000/claude-code-go.git
cd claude-code-go

# 2. 下载依赖
go mod download

# 3. 构建
go build -o claude ./cmd/claude

# 4. (可选) 安装到系统路径
sudo mv claude /usr/local/bin/
```

### 方式二：直接运行

```bash
# 克隆仓库
git clone https://github.com/Aspirin0000/claude-code-go.git
cd claude-code-go

# 直接运行
go run ./cmd/claude
```

### 方式三：交叉编译

```bash
# Linux AMD64
GOOS=linux GOARCH=amd64 go build -o claude-linux-amd64 ./cmd/claude

# macOS ARM64 (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o claude-darwin-arm64 ./cmd/claude

# macOS AMD64 (Intel)
GOOS=darwin GOARCH=amd64 go build -o claude-darwin-amd64 ./cmd/claude

# Windows AMD64
GOOS=windows GOARCH=amd64 go build -o claude-windows-amd64.exe ./cmd/claude
```

### 验证安装

```bash
# 检查版本
./claude version

# 运行诊断
./claude doctor
```

---

## 📚 使用方法

### 基本用法

```bash
# 交互式对话
./claude

# 带初始提示启动
./claude "帮我分析这个项目的结构"

# 指定模型
./claude --model claude-sonnet-4-20250514 "优化这段代码"
```

### 斜杠命令

```bash
# 显示帮助
/help

# 查看当前状态
/status

# 压缩对话历史
/compact

# 保存会话
/save my_session.json

# 配置管理
/config get model
/config set model claude-sonnet-4-20250514

# MCP 服务器管理
/mcp list
/mcp add my-server npx my-mcp-server
```

### 环境变量

```bash
# API Key
export ANTHROPIC_API_KEY="your-api-key"

# 或 OAuth Token
export ANTHROPIC_AUTH_TOKEN="your-oauth-token"

# 配置目录
export CLAUDE_CONFIG_HOME="$HOME/.config/claude"
```

---

## 🎯 使用场景

### 1. 代码审查

```bash
./claude "审查 src/services/api/client.go 的错误处理"
```

### 2. 代码重构

```bash
./claude "重构 internal/mcp/config.go，使其更易测试"
```

### 3. 调试问题

```bash
./claude --verbose "为什么我的程序出现死锁？"
```

### 4. 生成文档

```bash
./claude "为 internal/types/*.go 生成完整的 Go 文档注释"
```

### 5. 项目分析

```bash
./claude "分析这个项目的架构，解释各模块的职责"
```

---

## 📂 目录结构

### TypeScript 原始代码结构

```
claude-code-main/src/
├── main.tsx              # CLI 入口
├── commands/             # 斜杠命令（~207个文件）
│   ├── compact/          # 对话压缩
│   ├── config/           # 配置管理
│   ├── help/             # 帮助系统
│   └── ...
├── tools/                # AI 工具（~55个文件）
│   ├── BashTool/         # Bash 执行
│   ├── FileReadTool/     # 文件读取
│   ├── GrepTool/         # 文本搜索
│   ├── GlobTool/         # 文件匹配
│   └── ...
├── services/             # 服务层
│   ├── api/              # API 客户端
│   ├── mcp/              # MCP 协议
│   └── oauth/            # OAuth 认证
├── utils/                # 工具函数（~430个文件）
│   ├── bash/             # Bash 解析器
│   ├── permissions/      # 权限系统
│   └── ...
└── types/                # 类型定义
```

### Go 实现代码结构

```
.
├── cmd/
│   └── claude/           # 程序入口
│       ├── main.go
│       └── commands/     # 斜杠命令实现
├── internal/
│   ├── mcp/              # MCP 协议实现
│   │   ├── client.go     # MCP 客户端
│   │   ├── config.go     # 配置管理
│   │   ├── connection.go # 连接管理
│   │   └── types.go      # 类型定义
│   ├── types/            # 类型系统
│   │   ├── ids.go
│   │   ├── message.go
│   │   └── permissions.go
│   ├── tools/            # AI 工具
│   │   ├── tools.go
│   │   ├── bash.go
│   │   └── grep.go
│   ├── settings/         # 设置管理
│   ├── plugins/          # 插件系统
│   └── bootstrap/        # 启动状态
├── pkg/
│   └── utils/            # 公共工具
└── scripts/              # 脚本工具
```

---

## 🛠️ 开发指南

### 构建

```bash
# 编译所有包
go build ./...

# 运行测试
go test ./...

# 静态分析
go vet ./...
```

### 添加新命令

```go
// cmd/claude/commands/mycommand.go
package commands

import "context"

type MyCommand struct{ *BaseCommand }

func NewMyCommand() *MyCommand {
    return &MyCommand{
        BaseCommand: NewBaseCommand(
            "mycommand",
            "命令描述",
            CategoryTools,
        ),
    }
}

func (c *MyCommand) Execute(ctx context.Context, args []string) error {
    // 实现逻辑
    return nil
}

func init() { Register(NewMyCommand()) }
```

---

## 🗓️ 开发路线

### 已完成 ✅

- [x] 类型系统完整翻译（11个文件）
- [x] MCP 协议核心实现（12个文件，~6,200行）
- [x] 28个核心斜杠命令
- [x] 基础工具系统（Bash、Grep、Glob）
- [x] 配置系统（Viper 集成）
- [x] 命令注册表和路由

### 进行中 🔄

- [ ] OAuth 认证完整实现
- [ ] 工具系统完善（FileEdit、Task 等）
- [ ] 权限系统细化
- [ ] 会话持久化

### 计划中 📝

- [ ] UI 系统（Bubble Tea）
- [ ] Hooks 系统
- [ ] 插件系统
- [ ] 完整的测试套件
- [ ] CI/CD 配置

---

## 🔗 相关链接

- **原始项目**: [Anthropic Claude Code](https://www.npmjs.com/package/@anthropic-ai/claude-code)
- **还原版本**: [v2.1.88](https://github.com/Aspirin0000/claude-code-go/releases)

---

## ⚖️ 声明

- **源码版权**：TypeScript 源码版权归 **Anthropic** 所有
- **使用目的**：本仓库仅供学习和研究使用
- **非官方**：这不是 Anthropic 官方项目
- **禁止商用**：请勿用于商业用途

---

## 📄 License

本项目采用 MIT 许可证。原始 TypeScript 源码版权归 Anthropic 所有。

---

## 🌟 Star History

[![Star History Chart](https://api.star-history.com/svg?repos=Aspirin0000/claude-code-go&type=Date)](https://star-history.com/#Aspirin0000/claude-code-go&Date)

**如果这个项目对你有帮助，请给个 ⭐ Star！**
