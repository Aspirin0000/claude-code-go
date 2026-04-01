// Package plugins 提供插件管理
// 来源: src/utils/plugins/pluginLoader.ts
// 重构: Go 插件加载器
package plugins

// LoadedPlugin 已加载的插件
type LoadedPlugin struct {
	Name   string
	Source string
	Path   string
}

// PluginLoadError 插件加载错误
type PluginLoadError struct {
	Plugin string
	Error  string
}

// LoadAllPlugins 加载所有插件
func LoadAllPlugins() ([]LoadedPlugin, []PluginLoadError) {
	// 返回空结果作为默认实现
	return []LoadedPlugin{}, []PluginLoadError{}
}
