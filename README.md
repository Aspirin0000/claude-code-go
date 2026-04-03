# Claude Code Go

> An unofficial Go implementation reverse-engineered and translated from Claude Code v2.1.88 Source Map

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Stars](https://img.shields.io/github/stars/Aspirin0000/claude-code-go?style=social)](https://github.com/Aspirin0000/claude-code-go/stargazers)

> ⚠️ **Disclaimer**: This repository is a technical learning project. The TypeScript source code was reverse-engineered from the Source Map (`cli.js.map`) bundled with the `@anthropic-ai/claude-code` npm package v2.1.88, and translated into Go. For educational and research purposes only. All TypeScript source code copyright belongs to Anthropic.

---

## 🚀 Quick Start

### Prerequisites

- **Go 1.21+** installed
- **Anthropic API Key** (get from [console.anthropic.com](https://console.anthropic.com))

### Install

```bash
# Clone the repository
git clone https://github.com/Aspirin0000/claude-code-go.git
cd claude-code-go

# Build
make build

# (Optional) Install to system
make install

# Or run directly without installing
make run
```

### Configure

```bash
# Set your API key
export ANTHROPIC_API_KEY="your-api-key-here"

# Or create config file
mkdir -p ~/.config/claude
cat > ~/.config/claude/config.json << 'EOF'
{
  "api_key": "your-api-key-here",
  "model": "claude-sonnet-4-20250514"
}
EOF
```

### Run

```bash
# Interactive mode
./claude

# With initial prompt
./claude -p "Help me understand this codebase"

# Use specific model
./claude --model claude-opus-4-20250514
```

---

## 📖 Project Overview

### What is Claude Code?

Claude Code is Anthropic's official AI-powered coding assistant CLI tool that helps developers:

- 🔍 Understand and navigate codebases
- ✏️ Write and edit code
- 🖥️ Execute terminal commands
- 🐛 Debug and fix issues
- 🔌 Integrate with MCP (Model Context Protocol) servers

### Purpose of This Project

This project reverse-engineers the TypeScript source code (version 2.1.88) from the Source Map bundled with the `@anthropic-ai/claude-code` npm package, and **fully translates** it into Go.

**Reverse-engineering Scope:**
- 📁 Files reverse-engineered: **2,216** (containing 4,756 TypeScript source files)
- 🔄 Translation method: File-by-file complete translation, maintaining logical consistency
- 📊 Lines of code: ~**50,000+** lines of Go code

### Why Choose the Go Version?

| Feature | Node.js/TS Version | Go Version |
|---------|-------------------|------------|
| **Distribution** | Requires Node.js runtime | Single binary, no dependencies |
| **Cross-platform** | Depends on npm packages | Static compilation, native multi-platform support |
| **Performance** | V8 interpreted execution | Go compiled to machine code, faster startup |
| **Deployment** | Complex dependency management | Single file, easy deployment |
| **Concurrency** | Callbacks/Promises | Go goroutines, better concurrency model |

---

## 🔄 Translation Progress

### Core Module Translation Status

| Module | TS Files | Go Files | Status | Description |
|--------|----------|----------|--------|-------------|
| **Type System** | ~50 | 11 | ✅ Complete | All core type definitions |
| **MCP Protocol** | ~35 | 12 | ✅ Complete | Client, config, transport layer |
| **Command System** | ~207 | 28 | 🔄 In Progress | Core slash commands implemented |
| **Tool System** | ~55 | 5 | 🔄 In Progress | Bash, Grep, Glob, File tools |
| **API Client** | ~25 | 3 | 🔄 In Progress | Retry mechanisms, authentication |
| **Config System** | ~20 | 4 | 🔄 In Progress | Viper integration |
| **OAuth Auth** | ~15 | 2 | ⏳ Planned | Token management |
| **UI System** | ~144 | 0 | ⏳ Planned | Bubble Tea framework |
| **Hooks System** | ~85 | 0 | ⏳ Planned | Event hooks |
| **Permission System** | ~24 | 1 | 🔄 In Progress | Basic permission checks |

**Overall Progress: ~25%**

### Implemented Slash Commands (28)

#### Core Commands (5)
- ✅ `/help` - Display help information
- ✅ `/status` - Display session status
- ✅ `/clear` - Clear terminal screen
- ✅ `/version` - Display version information
- ✅ `/exit` - Exit the application

#### Session Management (4)
- ✅ `/compact` - Compress conversation history
- ✅ `/resume` - Resume historical session
- ✅ `/save` - Save session to file
- ✅ `/load` - Load session from file

#### Configuration Management (3)
- ✅ `/config` - Manage configuration items
- ✅ `/model` - Switch AI model
- ✅ `/permissions` - Set permission levels

#### MCP Management (4)
- ✅ `/mcp` - MCP server management
- ✅ `/mcp-add` - Add MCP server
- ✅ `/mcp-list` - List MCP servers
- ✅ `/mcp-remove` - Remove MCP server

#### Advanced Features (8)
- ✅ `/plan` - Create execution plans
- ✅ `/review` - Review code changes
- ✅ `/tasks` - Task management
- ✅ `/todos` - Todo items
- ✅ `/memory` - Session memory
- ✅ `/cost` - Cost tracking
- ✅ `/diff` - Git diff viewing
- ✅ `/doctor` - System diagnostics

#### AI Tool Commands (4)
- ✅ `/bash` - Execute Bash commands (with safety checks)
- ✅ `/git` - Git operations assistant
- ✅ `/grep` - File content search
- ✅ `/glob` - File pattern matching

*Note: System commands (ls, ps, docker, etc.) are executed via BashTool, not as independent slash commands*

---

## ✨ Features in Go Implementation

### 🔐 Authentication System

- **API Key Support**: Directly set `ANTHROPIC_API_KEY`
- **OAuth Token**: Support OAuth bearer token
- **Multi-cloud Auth**: AWS Bedrock, Google Vertex, Azure (planned)

### 🛠️ Tool System

#### BashTool
- **Safe Execution**: Dangerous command detection (rm -rf, etc.)
- **Permission Checking**: Control execution based on permission levels
- **Timeout Control**: Default 30-second timeout
- **Read-only Detection**: Automatically identify read-only commands

#### GrepTool
- **Parallel Search**: Multi-goroutine concurrent search
- **Regex Support**: Full regular expression matching
- **Result Limiting**: Default maximum 50 results
- **Binary Detection**: Automatically skip binary files

#### GlobTool
- **Pattern Matching**: Support `*` and `**` wildcards
- **Recursive Search**: Automatically recurse subdirectories
- **Permission Checking**: Check file access permissions

#### File Operation Tools
- **FileReadTool**: Read file content, support syntax highlighting
- **FileWriteTool**: Write files, automatic backup
- **FileEditTool**: AI-assisted editing, display diff

### 🔌 MCP (Model Context Protocol)

- **Transport Layer**: Support stdio and HTTP/SSE transport
- **Server Management**: Add, remove, list MCP servers
- **Tool Discovery**: Automatically discover tools provided by MCP servers
- **Config Management**: Support environment variable expansion `${VAR:-default}`

### 📝 Command System

- **Slash Commands**: /help, /compact, /config, etc.
- **Auto-registration**: Commands automatically register with the system
- **Category Management**: Organize commands by category (core, session, config, etc.)
- **Alias Support**: Commands support multiple aliases

### 💾 Session Management

- **Save/Load**: Support JSON and Markdown formats
- **History**: Preserve session history
- **Token Tracking**: Estimate token usage
- **Compression**: Automatic compression of long conversations

---

## 🚀 Installation & Deployment

### Prerequisites

- **Go Version**: Go 1.21 or higher
- **Operating System**: Linux, macOS, Windows
- **API Access**: Anthropic API key

### Option 1: Build from Source

```bash
# 1. Clone repository
git clone https://github.com/Aspirin0000/claude-code-go.git
cd claude-code-go

# 2. Download dependencies
go mod download

# 3. Build
go build -o claude ./cmd/claude

# 4. (Optional) Install to system path
sudo mv claude /usr/local/bin/
```

### Option 2: Run Directly

```bash
# Clone repository
git clone https://github.com/Aspirin0000/claude-code-go.git
cd claude-code-go

# Run directly
go run ./cmd/claude
```

### Option 3: Cross Compile

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

### Verify Installation

```bash
# Check version
./claude version

# Run diagnostics
./claude doctor
```

---

## 📚 Usage

### Basic Usage

```bash
# Interactive conversation
./claude

# Launch with initial prompt
./claude "Help me analyze the structure of this project"

# Specify model
./claude --model claude-sonnet-4-20250514 "Optimize this code"
```

### Slash Commands

```bash
# Display help
/help

# View current status
/status

# Compress conversation history
/compact

# Save session
/save my_session.json

# Configuration management
/config get model
/config set model claude-sonnet-4-20250514

# MCP server management
/mcp list
/mcp add my-server npx my-mcp-server
```

### Environment Variables

```bash
# API Key
export ANTHROPIC_API_KEY="your-api-key"

# Or OAuth Token
export ANTHROPIC_AUTH_TOKEN="your-oauth-token"

# Config directory
export CLAUDE_CONFIG_HOME="$HOME/.config/claude"
```

---

## 🎯 Use Cases

### 1. Code Review

```bash
./claude "Review the error handling in src/services/api/client.go"
```

### 2. Code Refactoring

```bash
./claude "Refactor internal/mcp/config.go to make it more testable"
```

### 3. Debugging Issues

```bash
./claude --verbose "Why is my program experiencing deadlock?"
```

### 4. Generate Documentation

```bash
./claude "Generate complete Go doc comments for internal/types/*.go"
```

### 5. Project Analysis

```bash
./claude "Analyze the architecture of this project and explain the responsibilities of each module"
```

---

## 📂 Directory Structure

### TypeScript Original Code Structure

```
claude-code-main/src/
├── main.tsx              # CLI entry point
├── commands/             # Slash commands (~207 files)
│   ├── compact/          # Conversation compression
│   ├── config/           # Configuration management
│   ├── help/             # Help system
│   └── ...
├── tools/                # AI tools (~55 files)
│   ├── BashTool/         # Bash execution
│   ├── FileReadTool/     # File reading
│   ├── GrepTool/         # Text search
│   ├── GlobTool/         # File matching
│   └── ...
├── services/             # Service layer
│   ├── api/              # API client
│   ├── mcp/              # MCP protocol
│   └── oauth/            # OAuth authentication
├── utils/                # Utility functions (~430 files)
│   ├── bash/             # Bash parser
│   ├── permissions/      # Permission system
│   └── ...
└── types/                # Type definitions
```

### Go Implementation Code Structure

```
.
├── cmd/
│   └── claude/           # Application entry point
│       ├── main.go
│       └── commands/     # Slash command implementations
├── internal/
│   ├── mcp/              # MCP protocol implementation
│   │   ├── client.go     # MCP client
│   │   ├── config.go     # Configuration management
│   │   ├── connection.go # Connection management
│   │   └── types.go      # Type definitions
│   ├── types/            # Type system
│   │   ├── ids.go
│   │   ├── message.go
│   │   └── permissions.go
│   ├── tools/            # AI tools
│   │   ├── tools.go
│   │   ├── bash.go
│   │   └── grep.go
│   ├── settings/         # Settings management
│   ├── plugins/          # Plugin system
│   └── bootstrap/        # Bootstrap state
├── pkg/
│   └── utils/            # Common utilities
└── scripts/              # Script tools
```

---

## 🛠️ Development Guide

### Build

```bash
# Compile all packages
go build ./...

# Run tests
go test ./...

# Static analysis
go vet ./...
```

### Add New Command

```go
// cmd/claude/commands/mycommand.go
package commands

import "context"

type MyCommand struct{ *BaseCommand }

func NewMyCommand() *MyCommand {
    return &MyCommand{
        BaseCommand: NewBaseCommand(
            "mycommand",
            "Command description",
            CategoryTools,
        ),
    }
}

func (c *MyCommand) Execute(ctx context.Context, args []string) error {
    // Implementation logic
    return nil
}

func init() { Register(NewMyCommand()) }
```

---

## 🗓️ Development Roadmap

### Completed ✅

- [x] Complete type system translation (11 files)
- [x] MCP protocol core implementation (12 files, ~6,200 lines)
- [x] 28 core slash commands
- [x] Basic tool system (Bash, Grep, Glob)
- [x] Configuration system (Viper integration)
- [x] Command registry and routing

### In Progress 🔄

- [ ] Complete OAuth authentication implementation
- [ ] Complete tool system (FileEdit, Task, etc.)
- [ ] Refine permission system
- [ ] Session persistence

### Planned 📝

- [ ] UI system (Bubble Tea)
- [ ] Hooks system
- [ ] Plugin system
- [ ] Complete test suite
- [ ] CI/CD configuration

---

## 🔗 Related Links

- **Original Project**: [Anthropic Claude Code](https://www.npmjs.com/package/@anthropic-ai/claude-code)
- **Reverse-engineered Version**: [v2.1.88](https://github.com/Aspirin0000/claude-code-go/releases)

---

## ⚖️ Disclaimer

- **Source Code Copyright**: TypeScript source code copyright belongs to **Anthropic**
- **Purpose**: This repository is for educational and research purposes only
- **Non-official**: This is not an Anthropic official project
- **No Commercial Use**: Please do not use for commercial purposes

---

## 📄 License

This project uses the MIT License. Original TypeScript source code copyright belongs to Anthropic.

---

## 🌟 Star History

[![Star History Chart](https://api.star-history.com/svg?repos=Aspirin0000/claude-code-go&type=Date)](https://star-history.com/#Aspirin0000/claude-code-go&Date)

**If this project helps you, please give it a ⭐ Star!**
