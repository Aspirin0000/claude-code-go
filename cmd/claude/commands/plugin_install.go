package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Aspirin0000/claude-code-go/internal/plugins"
)

// PluginInstallCommand installs a plugin from a source
type PluginInstallCommand struct {
	*BaseCommand
}

// NewPluginInstallCommand creates the /plugin-install command
func NewPluginInstallCommand() *PluginInstallCommand {
	return &PluginInstallCommand{
		BaseCommand: NewBaseCommand(
			"plugin-install",
			"Install a plugin from npm, GitHub, URL, or local path",
			CategoryPlugins,
		).WithAliases("install-plugin").
			WithHelp(`Usage: /plugin-install <source>

Install a plugin from various sources.

Sources:
  local:/path/to/plugin     - Install from local directory
  npm:package-name          - Install from npm registry
  github:user/repo          - Install from GitHub repository
  url:https://git-url       - Install from git URL

Examples:
  /plugin-install local:./my-plugin
  /plugin-install npm:claude-code-go-plugin
  /plugin-install github:username/claude-plugin
  /plugin-install url:https://gitlab.com/user/plugin.git

Aliases: /install-plugin`),
	}
}

// Execute runs the plugin-install command
func (c *PluginInstallCommand) Execute(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: /plugin-install <source>")
	}

	sourceStr := args[0]
	source, err := c.parseSource(sourceStr)
	if err != nil {
		return err
	}

	// Generate target directory
	pluginName := c.extractPluginName(source)
	targetDir := filepath.Join(plugins.GetPluginsDirectory(), "installed", pluginName)

	// Check if already exists
	if _, err := os.Stat(targetDir); err == nil {
		return fmt.Errorf("plugin %q already installed at %s", pluginName, targetDir)
	}

	fmt.Printf("Installing plugin %q from %s...\n", pluginName, source.Type)

	plugin, err := plugins.InstallPlugin(source, targetDir)
	if err != nil {
		return fmt.Errorf("installation failed: %w", err)
	}

	fmt.Println()
	fmt.Println("✅ Plugin installed successfully!")
	fmt.Printf("   Name: %s\n", plugin.Manifest.Name)
	fmt.Printf("   Path: %s\n", targetDir)
	if plugin.Manifest.Description != "" {
		fmt.Printf("   Description: %s\n", plugin.Manifest.Description)
	}
	fmt.Println()
	fmt.Println("Use /plugins to see all installed plugins.")
	fmt.Println()

	return nil
}

func (c *PluginInstallCommand) parseSource(sourceStr string) (plugins.PluginSource, error) {
	parts := strings.SplitN(sourceStr, ":", 2)
	if len(parts) != 2 {
		return plugins.PluginSource{}, fmt.Errorf("invalid source format; expected type:value, got: %s", sourceStr)
	}

	sourceType := strings.ToLower(parts[0])
	value := parts[1]

	switch sourceType {
	case "local":
		return plugins.PluginSource{Type: "local", Path: value}, nil
	case "npm":
		return plugins.PluginSource{Type: "npm", Package: value}, nil
	case "github":
		repoParts := strings.SplitN(value, "@", 2)
		repo := repoParts[0]
		ref := ""
		if len(repoParts) > 1 {
			ref = repoParts[1]
		}
		return plugins.PluginSource{Type: "github", Repo: repo, Ref: ref}, nil
	case "url":
		return plugins.PluginSource{Type: "url", URL: value}, nil
	default:
		return plugins.PluginSource{}, fmt.Errorf("unsupported source type: %s", sourceType)
	}
}

func (c *PluginInstallCommand) extractPluginName(source plugins.PluginSource) string {
	switch source.Type {
	case "local":
		return filepath.Base(source.Path)
	case "npm":
		return source.Package
	case "github":
		parts := strings.Split(source.Repo, "/")
		if len(parts) > 0 {
			return parts[len(parts)-1]
		}
		return source.Repo
	case "url":
		// Extract last path component without .git
		base := filepath.Base(source.URL)
		return strings.TrimSuffix(base, ".git")
	default:
		return "unknown"
	}
}

func init() {
	Register(NewPluginInstallCommand())
}
