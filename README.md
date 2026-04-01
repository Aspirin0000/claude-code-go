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

## 项目结构

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
