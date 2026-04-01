# 简化实现修复清单

## 严重问题文件（需要完整实现）

### 1. internal/services/analytics/index.go
**问题:** LogEvent 是空函数
**应该:** 实际发送分析事件到后端（Segment/Amplitude/自定义）
**依赖:** 需要配置分析后端URL和API密钥
**操作:** 暂不修复，依赖外部服务配置

### 2. internal/settings/plugin_only_policy.go  
**问题:** 返回固定 false
**应该:** 从配置文件读取插件策略设置
**操作:** 可以修复 - 从 viper/config 读取

### 3. internal/plugins/mcp_plugin_integration.go
**问题:** 返回空映射，LoadAllPluginsCacheOnly 无操作
**应该:** 实际扫描插件目录，加载插件配置
**依赖:** 需要完整的插件系统
**操作:** 暂不修复，等待插件系统架构确定

### 3. internal/plugins/plugin_loader.go
**问题:** 返回空结果
**应该:** 实际加载插件
**操作:** 暂不修复

### 4. internal/settings/types.go
**问题:** GetSettingsForSource 返回空设置
**应该:** 从各来源读取实际设置
**操作:** 可以修复 - 集成 viper 配置

### 5. internal/bootstrap/state.go
**问题:** generateSessionID 简单实现
**应该:** 生成真正的唯一会话ID
**操作:** 可以修复 - 使用 UUID

### 6. pkg/utils/ 多个文件
**检查中:** uuid.go, set.go, sleep.go 等
**状态:** 需要逐一验证

## 修复优先级

P0（立即修复）:
- [ ] internal/settings/plugin_only_policy.go
- [ ] internal/settings/types.go  
- [ ] internal/bootstrap/state.go

P1（可以稍后）:
- [ ] 插件相关（需要架构设计）
- [ ] 分析服务（需要外部配置）

## 修复原则

不是删除"简化"二字，而是：
1. 分析原 TS 代码逻辑
2. 完整实现相同功能
3. 集成到现有系统
4. 添加适当的错误处理
5. 确保线程安全
