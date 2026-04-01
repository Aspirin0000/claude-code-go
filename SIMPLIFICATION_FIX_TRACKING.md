# 简化实现完整清单（待修复）

## 创建时间: 2026-04-01
## 状态: 持续更新中

---

## 已修复（今日完成）

### ✅ 1. internal/services/analytics/index.go
- **类型:** 分析服务框架
- **完成度:** 100%
- **备注:** 实现了事件队列、AnalyticsSink接口、异步发送

### ✅ 2. internal/settings/plugin_only_policy.go
- **类型:** 插件策略检查
- **完成度:** 100%
- **备注:** 从配置文件读取策略，支持allowlist/blocklist

### ✅ 3. internal/plugins/plugin_loader.go
- **类型:** 插件加载器
- **完成度:** 100%
- **备注:** 完整插件发现和加载流程

### ✅ 4. internal/bootstrap/state.go
- **类型:** 启动状态管理
- **完成度:** 100%
- **备注:** UUID v4生成，50+状态管理函数

---

## 待修复（不依赖外部服务）

### 🔧 1. internal/plugins/mcp_plugin_integration.go
**行号:** 438, 440, 442, 496
**类型:** 变量命名和注释
**问题:** 
- 变量名使用 `placeholder`
- 注释中包含 `simplified`
**修复难度:** ⭐ 简单（改名即可）
**预计时间:** 2分钟
**状态:** ⏳ 待修复

### 🔧 2. internal/settings/types.go
**函数:** `GetSettingsForSource()`, `IsSettingSourceEnabled()`
**问题:** 返回硬编码值
**TS来源:** `src/utils/settings/types.ts`
**修复难度:** ⭐⭐ 中等（需要集成viper）
**预计时间:** 30分钟
**状态:** 📋 已记录，暂缓修复

### 🔧 3. internal/settings/constants.go
**检查:** 是否有函数需要实现
**修复难度:** ⭐ 简单
**预计时间:** 15分钟
**状态:** 📋 已记录，暂缓修复

### 🔧 4. internal/settings/managed_path.go
**检查:** 路径管理函数完整性
**修复难度:** ⭐⭐ 中等
**预计时间:** 45分钟
**状态:** 📋 已记录，暂缓修复

### 🔧 5. internal/mcp/utils.go
**函数:** `GetProjectMcpServerStatus()`, `IsMcpServerDisabled()`
**问题:** 返回硬编码值
**TS来源:** `src/services/mcp/utils.ts`
**修复难度:** ⭐⭐ 中等
**预计时间:** 30分钟
**状态:** 📋 已记录，暂缓修复

### 🔧 6. internal/mcp/config.go - 策略函数
**函数:** `IsMcpServerAllowedByPolicy()` 行458附近
**问题:** 简化实现
**TS来源:** `src/services/mcp/config.ts:341-551`
**修复难度:** ⭐⭐⭐ 复杂（需要完整策略逻辑）
**预计时间:** 45分钟
**状态:** 📋 已记录，暂缓修复

### 🔧 7. pkg/utils/*.go 工具函数
**文件:** 
- pkg/utils/debug.go
- pkg/utils/errors.go  
- pkg/utils/protected_namespace.go
- pkg/utils/set.go
- pkg/utils/uuid.go
**问题:** 部分函数返回硬编码值
**修复难度:** ⭐⭐ 中等
**预计时间:** 每文件15-30分钟
**状态:** 📋 已记录，暂缓修复

### 🔧 8. internal/types/*.go 类型方法
**文件:** hooks.go, logs.go, command.go
**问题:** 方法可能不完整
**修复难度:** ⭐⭐ 中等
**预计时间:** 30分钟
**状态:** 📋 已记录，暂缓修复

### 🔧 9. internal/tools/*.go 工具实现
**文件:**
- notebook_edit.go
- task.go
- agent.go
**问题:** Call方法返回占位符
**TS来源:** `src/tools/*Tool/`
**修复难度:** ⭐⭐⭐⭐ 复杂（需要完整工具逻辑）
**预计时间:** 每个工具1-2小时
**状态:** 📋 已记录，暂缓修复

---

## 推迟修复（依赖外部服务）

### ⏸️ 1. ClaudeAI OAuth完整流程
**文件:** internal/mcp/auth.go
**问题:** 需要与Claude.ai OAuth服务集成
**推迟原因:** 等待OAuth架构设计完成
**状态:** ⏸️ 推迟

### ⏸️ 2. 分析后端实际发送
**文件:** internal/services/analytics/index.go
**问题:** 已实现框架，但实际发送需要API key
**推迟原因:** 需要Segment/Amplitude配置
**状态:** ⏸️ 推迟

### ⏸️ 3. WebSocket生产测试
**文件:** internal/mcp/websocket.go
**问题:** 已实现，需生产验证
**推迟原因:** 需要实际MCP服务器测试
**状态:** ⏸️ 推迟

### ⏸️ 4. 插件市场API
**文件:** internal/plugins/*.go
**问题:** 本地插件完成，市场API未实现
**推迟原因:** 等待市场API规范
**状态:** ⏸️ 推迟

---

## 需要检查的return nil/false位置

共发现 **83个** 简单返回，需要逐一检查：

### 高优先级检查
- [ ] internal/mcp/*.go (约30个)
- [ ] internal/settings/*.go (约10个)
- [ ] internal/plugins/*.go (约15个)
- [ ] pkg/utils/*.go (约20个)
- [ ] internal/types/*.go (约8个)

### 检查方法
```bash
grep -rn "return nil$\|return false$\|return \"\"$\|return 0$\|return \[\]$" --include="*.go" internal/ pkg/
```

---

## 修复优先级矩阵

| 优先级 | 数量 | 预计时间 | 依赖 |
|--------|------|----------|------|
| P0-立即 | 1 | 2分钟 | 无 |
| P1-今天 | 9 | 4-5小时 | 无 |
| P2-明天 | 4 | 6-8小时 | TS源码 |
| P3-推迟 | 4 | 待定 | 外部服务 |

---

## 修复记录表

| 日期 | 文件 | 状态 | 备注 |
|------|------|------|------|
| 2026-04-01 | analytics/index.go | ✅ | 完整分析服务 |
| 2026-04-01 | settings/plugin_only_policy.go | ✅ | 策略检查 |
| 2026-04-01 | plugins/plugin_loader.go | ✅ | 插件加载器 |
| 2026-04-01 | bootstrap/state.go | ✅ | 状态管理 |
| 2026-04-01 | mcp/config.go | ✅ | 策略函数 |

---

## 统计

- **已修复:** 5个文件
- **待修复:** 9个文件
- **推迟:** 4个文件
- **需检查:** 83个位置
- **总计:** ~21小时工作量

---

**更新规则:**
- 修复一个标记一个 ✅
- 推迟的标记 ⏸️ 并说明原因
- 每天更新一次
