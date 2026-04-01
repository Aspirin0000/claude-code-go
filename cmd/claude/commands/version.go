package commands

import (
	"context"
	"fmt"
	"runtime"
	"runtime/debug"
)

// Version information - set at build time via ldflags
var (
	// AppName is the application name
	AppName = "claude-code-go"

	// Version is the application version
	Version = "dev"

	// BuildTime is the build timestamp
	BuildTime = "unknown"

	// GitCommit is the git commit hash
	GitCommit = "unknown"
)

// VersionCommand handles the /version command
type VersionCommand struct {
	*BaseCommand
}

// NewVersionCommand creates a new version command
func NewVersionCommand() Command {
	return &VersionCommand{
		BaseCommand: NewBaseCommand(
			"version",
			"Show version information",
			CategoryGeneral,
		).WithAliases("v", "ver"),
	}
}

// Execute displays version information
func (c *VersionCommand) Execute(ctx context.Context, args []string) error {
	fmt.Printf("%s version %s\n", AppName, Version)
	fmt.Printf("  Go version: %s\n", runtime.Version())
	fmt.Printf("  Platform:   %s/%s\n", runtime.GOOS, runtime.GOARCH)

	if BuildTime != "unknown" {
		fmt.Printf("  Build time: %s\n", BuildTime)
	}

	if GitCommit != "unknown" {
		fmt.Printf("  Git commit: %s\n", GitCommit)
	}

	// Also try to read from debug.BuildInfo
	if info, ok := debug.ReadBuildInfo(); ok {
		fmt.Printf("  Module:     %s\n", info.Main.Path)
		if Version == "dev" && info.Main.Version != "" {
			fmt.Printf("  Build ver:  %s\n", info.Main.Version)
		}
	}

	return nil
}
