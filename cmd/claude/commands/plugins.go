package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Aspirin0000/claude-code-go/internal/plugins"
)

// PluginsCommand manages plugins
type PluginsCommand struct {
	*BaseCommand
}

// NewPluginsCommand creates the /plugins command
func NewPluginsCommand() *PluginsCommand {
	return &PluginsCommand{
		BaseCommand: NewBaseCommand(
			"plugins",
			"List installed plugins and show plugin directory",
			CategoryAdvanced,
		).WithAliases("plugin").
			WithHelp(`Usage: /plugins

List installed plugins from the plugin cache and directories.
Shows plugin directory path and basic plugin information.

Aliases: /plugin`),
	}
}

// Execute runs the plugins command
func (c *PluginsCommand) Execute(ctx context.Context, args []string) error {
	pluginDir := plugins.GetPluginsDirectory()
	cacheDir := plugins.GetPluginCachePath()

	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║                     Plugin Manager                      ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Printf("Plugin directory: %s\n", pluginDir)
	fmt.Printf("Cache directory:  %s\n", cacheDir)
	fmt.Println()

	installed := c.listInstalledPlugins(cacheDir)
	if len(installed) == 0 {
		fmt.Println("ℹ️  No plugins currently installed.")
		fmt.Println("   Plugins are loaded from the cache directory above.")
	} else {
		fmt.Printf("Installed plugins (%d):\n", len(installed))
		for _, p := range installed {
			fmt.Printf("  • %s\n", p)
		}
	}
	fmt.Println()
	return nil
}

func (c *PluginsCommand) listInstalledPlugins(cacheDir string) []string {
	var result []string
	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		return result
	}
	for _, entry := range entries {
		if entry.IsDir() {
			name := entry.Name()
			// For versioned cache, look one level deeper
			subPath := filepath.Join(cacheDir, name)
			subEntries, _ := os.ReadDir(subPath)
			for _, sub := range subEntries {
				if sub.IsDir() {
					result = append(result, fmt.Sprintf("%s@%s", sub.Name(), name))
				}
			}
		}
	}
	return result
}

func init() {
	Register(NewPluginsCommand())
}
