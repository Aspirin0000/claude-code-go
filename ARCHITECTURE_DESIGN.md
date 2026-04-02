# Go 命令系统架构设计
## 学习 TS 架构后的正确实现

## 架构层次

### 1. 斜杠命令 (Slash Commands)
**用户主动输入**: `/help`, `/compact`, `/config`

**TS 对应**: `src/commands/` 目录
- 每个命令独立目录（如 `commands/compact/`）
- 包含 `index.ts` 和实际逻辑文件
- 延迟加载（lazy require）

**Go 实现**: `cmd/claude/commands/`
- 保留命令：help, status, compact, resume, save, load, config, model, permissions, mcp, plan, review
- 删除：ls, ps, docker, npm 等系统命令包装器

### 2. 工具系统 (Tools)
**AI 调用**: `BashTool`, `GrepTool`, `FileReadTool`

**TS 对应**: `src/tools/<ToolName>/`
- 每个工具一个目录
- 包含 `call()` 方法执行逻辑
- 权限检查、输入验证、结果渲染

**Go 实现**: `internal/tools/`
- 已实现：BashTool, GrepTool, GlobTool, FileReadTool, FileWriteTool, FileEditTool
- 待实现：完整的 BashTool 权限系统

### 3. Bash 执行
**TS 流程**:
1. 用户输入命令
2. AI 决定调用 BashTool
3. BashTool 调用 `runShellCommand()`
4. 权限检查 → 沙箱执行 → 流式输出

**Go 应该**:
- 不要创建独立的 /ls /ps 命令
- 强化 BashTool 的权限系统
- 所有系统命令通过 BashTool 执行

## 需要修复的

### 🔴 立即删除（88个系统命令包装器）
ls, read, edit, write, rm, mkdir, cp, mv, cd, pwd, touch, chmod, chown
docker, kubectl, npm, pip, yarn, make, cmake, gradle, mvn
nano, vim, code, python, ruby, node, rustc
ssh, scp, ftp, sftp, wget, ping, curl
ps, kill, df, du, top, htop
以及所有其他系统工具包装器

### ✅ 保留（真正的斜杠命令）
help, status, compact, resume, save, load
config, model, permissions
mcp, mcp-add, mcp-list, mcp-remove
plan, review, tasks, memory
clear, version, exit

### 🔴 需要重写
compact.go - 集成 AI 服务（不是本地假摘要）

## 统计修正
- **真正需要的斜杠命令**: ~20-30 个
- **已正确实现**: ~15 个
- **需要修复**: compact, 删除多余的 88 个
- **完成度**: 不是 48%，而是约 15-20 个命令已完成（真正需要的）

## 下一步
1. 删除 88 个系统命令包装器
2. 强化 BashTool（权限、沙箱）
3. 修复 compact.go（集成 AI）
4. 实现其他真正的斜杠命令（如 TS 中的 diff, cost, commit 等）
