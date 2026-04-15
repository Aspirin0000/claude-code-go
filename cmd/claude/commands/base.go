// Package commands provides all CLI command implementations.
// Source: src/commands/ (207 TS files)
// Refactor: Go command system in batch mode.
package commands

import (
	"context"
	"fmt"
)

// Command interface.
// Maps to TS: Command type definition.
type Command interface {
	// Name command name (e.g. "help", "bash")
	Name() string

	// Aliases command aliases
	Aliases() []string

	// Description command description
	Description() string

	// Category command category
	Category() CommandCategory

	// Execute runs the command
	Execute(ctx context.Context, args []string) error

	// Help returns help text
	Help() string
}

// CommandCategory command category.
type CommandCategory string

const (
	CategoryGeneral  CommandCategory = "general"  // General commands
	CategorySession  CommandCategory = "session"  // Session management
	CategoryConfig   CommandCategory = "config"   // Configuration
	CategoryMCP      CommandCategory = "mcp"      // MCP management
	CategoryTools    CommandCategory = "tools"    // Tool commands
	CategoryFiles    CommandCategory = "files"    // File operations
	CategoryAdvanced CommandCategory = "advanced" // Advanced features
	CategoryPlugins  CommandCategory = "plugins"  // Plugin management
)

// String returns the display name for the category.
func (c CommandCategory) String() string {
	switch c {
	case CategoryGeneral:
		return "General"
	case CategorySession:
		return "Session"
	case CategoryConfig:
		return "Config"
	case CategoryMCP:
		return "MCP"
	case CategoryTools:
		return "Tools"
	case CategoryFiles:
		return "Files"
	case CategoryAdvanced:
		return "Advanced"
	case CategoryPlugins:
		return "Plugins"
	default:
		return string(c)
	}
}

// BaseCommand base command implementation.
type BaseCommand struct {
	name        string
	aliases     []string
	description string
	category    CommandCategory
	helpText    string
}

// Name returns the command name.
func (c *BaseCommand) Name() string {
	return c.name
}

// Aliases returns command aliases.
func (c *BaseCommand) Aliases() []string {
	return c.aliases
}

// Description returns the command description.
func (c *BaseCommand) Description() string {
	return c.description
}

// Category returns the command category.
func (c *BaseCommand) Category() CommandCategory {
	return c.category
}

// Help returns the help text.
func (c *BaseCommand) Help() string {
	return c.helpText
}

// NewBaseCommand creates a base command.
func NewBaseCommand(name, description string, category CommandCategory) *BaseCommand {
	return &BaseCommand{
		name:        name,
		description: description,
		category:    category,
		helpText: fmt.Sprintf("/%s - %s\n\nUsage: /%s [args]",
			name, description, name),
	}
}

// WithAliases sets aliases.
func (c *BaseCommand) WithAliases(aliases ...string) *BaseCommand {
	c.aliases = aliases
	return c
}

// WithHelp sets help text.
func (c *BaseCommand) WithHelp(help string) *BaseCommand {
	c.helpText = help
	return c
}
