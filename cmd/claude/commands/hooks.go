package commands

import (
	"context"
	"fmt"

	"github.com/Aspirin0000/claude-code-go/internal/hooks"
	"github.com/Aspirin0000/claude-code-go/internal/types"
)

// HooksCommand shows registered event hooks
type HooksCommand struct {
	*BaseCommand
}

// NewHooksCommand creates the /hooks command
func NewHooksCommand() *HooksCommand {
	return &HooksCommand{
		BaseCommand: NewBaseCommand(
			"hooks",
			"Show registered event hooks",
			CategoryAdvanced,
		).
			WithHelp(`Usage: /hooks

Display all registered hooks grouped by event type.
This is useful for debugging hook behavior and verifying plugin integrations.
`),
	}
}

// Execute runs the hooks command
func (c *HooksCommand) Execute(ctx context.Context, args []string) error {
	mgr := hooks.GetGlobalManager()
	events := types.GetAllHookEvents()

	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║                  Registered Hooks                       ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Println()

	hasAny := false
	for _, event := range events {
		info := mgr.ListHooks(event)
		if len(info) == 0 {
			continue
		}
		hasAny = true
		fmt.Printf("📌 %s (%d)\n", event, len(info))
		for _, h := range info {
			kind := "async"
			if h.IsSync {
				kind = "sync"
			}
			fmt.Printf("   • %s [%s]\n", h.Name, kind)
		}
		fmt.Println()
	}

	if !hasAny {
		fmt.Println("ℹ️  No hooks currently registered.")
		fmt.Println("   Hooks are typically registered by plugins or extensions.")
		fmt.Println()
	}

	return nil
}

func init() {
	Register(NewHooksCommand())
}
