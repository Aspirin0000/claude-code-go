package hooks

import (
	"context"
	"errors"
	"testing"

	"github.com/Aspirin0000/claude-code-go/internal/types"
)

func TestManagerRegisterAndExecuteSync(t *testing.T) {
	m := NewManager()
	called := false

	err := m.RegisterSync(types.HookEventPreToolUse, "test", func(ctx context.Context, input types.HookInput) (*types.SyncHookResponse, error) {
		called = true
		return &types.SyncHookResponse{
			Continue: boolPtr(true),
		}, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	resp, err := m.ExecuteSync(context.Background(), types.HookEventPreToolUse, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Errorf("expected hook to be called")
	}
	if resp.Continue == nil || !*resp.Continue {
		t.Errorf("expected continue=true")
	}
}

func TestManagerInvalidEvent(t *testing.T) {
	m := NewManager()
	err := m.RegisterSync("InvalidEvent", "test", func(ctx context.Context, input types.HookInput) (*types.SyncHookResponse, error) {
		return nil, nil
	})
	if err == nil {
		t.Errorf("expected error for invalid event")
	}
}

func TestManagerStopOnContinueFalse(t *testing.T) {
	m := NewManager()
	callCount := 0

	m.RegisterSync(types.HookEventPreToolUse, "first", func(ctx context.Context, input types.HookInput) (*types.SyncHookResponse, error) {
		callCount++
		return &types.SyncHookResponse{Continue: boolPtr(false), Decision: strPtr("deny")}, nil
	})
	m.RegisterSync(types.HookEventPreToolUse, "second", func(ctx context.Context, input types.HookInput) (*types.SyncHookResponse, error) {
		callCount++
		return &types.SyncHookResponse{Continue: boolPtr(true)}, nil
	})

	resp, err := m.ExecuteSync(context.Background(), types.HookEventPreToolUse, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if callCount != 1 {
		t.Errorf("expected 1 hook call, got %d", callCount)
	}
	if resp.Continue == nil || *resp.Continue {
		t.Errorf("expected continue=false")
	}
}

func TestManagerExecuteAsync(t *testing.T) {
	m := NewManager()
	called := false

	m.RegisterAsync(types.HookEventFileChanged, "test", func(ctx context.Context, input types.HookInput) (*types.AsyncHookResult, error) {
		called = true
		return &types.AsyncHookResult{Continue: true}, nil
	})

	results, err := m.ExecuteAsync(context.Background(), types.HookEventFileChanged, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Errorf("expected async hook to be called")
	}
	if len(results) != 1 || !results[0].Continue {
		t.Errorf("expected one positive result")
	}
}

func TestManagerAsyncError(t *testing.T) {
	m := NewManager()
	m.RegisterAsync(types.HookEventFileChanged, "fail", func(ctx context.Context, input types.HookInput) (*types.AsyncHookResult, error) {
		return nil, errors.New("boom")
	})

	_, err := m.ExecuteAsync(context.Background(), types.HookEventFileChanged, nil)
	if err == nil {
		t.Errorf("expected error from failing async hook")
	}
}

func TestManagerUnregister(t *testing.T) {
	m := NewManager()
	m.RegisterSync(types.HookEventPreToolUse, "a", func(ctx context.Context, input types.HookInput) (*types.SyncHookResponse, error) {
		return nil, nil
	})
	m.RegisterSync(types.HookEventPreToolUse, "b", func(ctx context.Context, input types.HookInput) (*types.SyncHookResponse, error) {
		return nil, nil
	})

	m.Unregister(types.HookEventPreToolUse, "a")

	m.mu.RLock()
	count := len(m.hooks[types.HookEventPreToolUse])
	m.mu.RUnlock()
	if count != 1 {
		t.Errorf("expected 1 hook after unregister, got %d", count)
	}
}

func TestExecutePreToolUse(t *testing.T) {
	m := NewManager()
	m.RegisterSync(types.HookEventPreToolUse, "test", func(ctx context.Context, input types.HookInput) (*types.SyncHookResponse, error) {
		return &types.SyncHookResponse{
			Continue: boolPtr(false),
			Decision: strPtr("deny"),
			Reason:   strPtr("unsafe"),
			HookSpecificOutput: &types.PreToolUseOutput{
				PermissionDecision:       pbPtr(types.PermissionBehaviorDeny),
				PermissionDecisionReason: strPtr("classified unsafe"),
				UpdatedInput:             map[string]interface{}{"cmd": "echo safe"},
				AdditionalContext:        strPtr("be careful"),
			},
		}, nil
	})

	result, err := m.ExecutePreToolUse(context.Background(), "bash", map[string]interface{}{"cmd": "rm -rf /"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Continue {
		t.Errorf("expected continue=false")
	}
	if result.Decision != types.PermissionBehaviorDeny {
		t.Errorf("expected deny decision, got %s", result.Decision)
	}
	if result.DecisionReason != "classified unsafe" {
		t.Errorf("unexpected reason: %s", result.DecisionReason)
	}
	if result.UpdatedInput["cmd"] != "echo safe" {
		t.Errorf("expected updated input, got %v", result.UpdatedInput)
	}
	if result.AdditionalContext != "be careful" {
		t.Errorf("unexpected context: %s", result.AdditionalContext)
	}
}

func TestExecuteUserPromptSubmit(t *testing.T) {
	m := NewManager()
	m.RegisterSync(types.HookEventUserPromptSubmit, "test", func(ctx context.Context, input types.HookInput) (*types.SyncHookResponse, error) {
		return &types.SyncHookResponse{
			Continue: boolPtr(true),
			HookSpecificOutput: &types.UserPromptSubmitOutput{
				AdditionalContext: strPtr("context from hook"),
			},
		}, nil
	})

	result, err := m.ExecuteUserPromptSubmit(context.Background(), "hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Continue {
		t.Errorf("expected continue=true")
	}
	if result.AdditionalContext != "context from hook" {
		t.Errorf("unexpected context: %s", result.AdditionalContext)
	}
}

func TestExecuteSessionStart(t *testing.T) {
	m := NewManager()
	m.RegisterSync(types.HookEventSessionStart, "test", func(ctx context.Context, input types.HookInput) (*types.SyncHookResponse, error) {
		return &types.SyncHookResponse{
			Continue: boolPtr(true),
			HookSpecificOutput: &types.SessionStartOutput{
				AdditionalContext:  strPtr("started"),
				InitialUserMessage: strPtr("help me"),
				WatchPaths:         []string{"/tmp"},
			},
		}, nil
	})

	result, err := m.ExecuteSessionStart(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Continue {
		t.Errorf("expected continue=true")
	}
	if result.AdditionalContext != "started" {
		t.Errorf("unexpected context")
	}
	if result.InitialUserMessage != "help me" {
		t.Errorf("unexpected initial message")
	}
	if len(result.WatchPaths) != 1 || result.WatchPaths[0] != "/tmp" {
		t.Errorf("unexpected watch paths: %v", result.WatchPaths)
	}
}

func TestGlobalManager(t *testing.T) {
	orig := GetGlobalManager()
	newM := NewManager()
	SetGlobalManager(newM)
	if GetGlobalManager() != newM {
		t.Errorf("expected global manager to be updated")
	}
	SetGlobalManager(orig)
}

func strPtr(s string) *string {
	return &s
}

func pbPtr(p types.PermissionBehavior) *types.PermissionBehavior {
	return &p
}
