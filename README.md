# Claude Code Go

使用 Go 语言重构的 Claude Code CLI 工具。

## 项目简介

本项目是 Anthropic Claude Code CLI 的 Go 语言实现，提供交互式 AI 编程助手功能。

## 功能特性

- 🤖 **Claude API 集成**：支持多提供商（Anthropic、AWS Bedrock、Google Vertex、Azure）
- 🛠️ **工具系统**：
  - Bash 命令执行
  - 文件读写编辑
  - Grep 搜索
  - Glob 文件匹配
  - Web 搜索和获取
  - 任务管理
  - Agent 子代理
- 💬 **对话管理**：支持多轮对话、上下文压缩
- 🖥️ **终端 UI**：使用 Charm 框架构建交互式界面
- 🔌 **MCP 支持**：Model Context Protocol 协议支持
- 🔐 **权限管理**：细粒度的工具权限控制

## 项目进度

### 整体进度: ~20% 完成

| 模块 | 进度 | 状态 |
|------|------|------|
| **MCP服务** | 95% | 12个文件, ~6,200行 ✅ |
| **命令系统** | 20% | 41/207 命令 ✅ |
| **类型系统** | 100% | 11个文件完成 ✅ |
| **工具系统** | 16% | 9/55 工具 ✅ |

### 已实现的命令 (41个)

**核心命令**: /help, /status, /clear, /version, /exit  
**会话管理**: /compact, /resume, /save, /load  
**配置管理**: /config, /model, /permissions  
**MCP管理**: /mcp, /mcp-add, /mcp-list, /mcp-remove  
**工具命令**: /bash, /git, /grep, /glob  
**文件操作**: /ls, /read, /edit, /write, /rm, /mkdir, /cp, /mv, /cd, /pwd, /touch  
**高级功能**: /plan, /review, /tasks, /todos, /memory  
**系统命令**: /whoami, /hostname, /env, /date, /cal  

*最后更新: 2026-04-01*

---

```
.
├── cmd/
│   └── claude/          # 主程序入口
├── internal/
│   ├── api/             # Claude API 客户端
│   ├── config/          # 配置管理
│   ├── tools/           # 工具系统
│   ├── ui/              # 终端 UI
│   ├── mcp/             # MCP 协议实现
│   └── state/           # 状态管理
├── pkg/                 # 公共包
└── docs/                # 文档
```

## 安装

```bash
go install github.com/Aspirin0000/claude-code-go/cmd/claude@latest
```

## 使用方法

```bash
# 启动交互式会话
claude

# 管道模式
echo "分析这个文件" | claude -p

# 指定模型
claude --model claude-sonnet-4-20250514
```

## 配置

配置文件位于 `~/.config/claude/config.json`：

```json
{
  "api_key": "your-api-key",
  "model": "claude-sonnet-4-20250514",
  "theme": "dark"
}
```

## 开发

```bash
# 克隆仓库
git clone https://github.com/Aspirin0000/claude-code-go.git

# 安装依赖
go mod tidy

# 运行测试
go test ./...

# 构建
go build -o claude ./cmd/claude
```

## 许可证

MIT
