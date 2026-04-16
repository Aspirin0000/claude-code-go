package commands

import (
	"context"
	"strings"
	"testing"

	"github.com/Aspirin0000/claude-code-go/internal/hooks"
	"github.com/Aspirin0000/claude-code-go/internal/types"
)

func TestHooksCommand_Empty(t *testing.T) {
	cmd := NewHooksCommand()
	out := captureOutput(func() {
		_ = cmd.Execute(context.Background(), nil)
	})

	if !strings.Contains(out, "Registered Hooks") {
		t.Errorf("expected header, got: %s", out)
	}
	if !strings.Contains(out, "No hooks currently registered") {
		t.Errorf("expected empty message, got: %s", out)
	}
}

func TestHooksCommand_WithHooks(t *testing.T) {
	mgr := hooks.GetGlobalManager()
	_ = mgr.RegisterSync(types.HookEventSessionStart, "test-hook", func(ctx context.Context, input types.HookInput) (*types.SyncHookResponse, error) {
		return &types.SyncHookResponse{}, nil
	})
	defer mgr.Unregister(types.HookEventSessionStart, "test-hook")

	cmd := NewHooksCommand()
	out := captureOutput(func() {
		_ = cmd.Execute(context.Background(), nil)
	})

	if !strings.Contains(out, "SessionStart") {
		t.Errorf("expected SessionStart event, got: %s", out)
	}
	if !strings.Contains(out, "test-hook") {
		t.Errorf("expected test-hook name, got: %s", out)
	}
}

func TestHooksCommand_Registered(t *testing.T) {
	reg := NewRegistry()
	_ = reg.Register(NewHooksCommand())
	if _, ok := reg.Get("hooks"); !ok {
		t.Error("hooks command not registered")
	}
}
