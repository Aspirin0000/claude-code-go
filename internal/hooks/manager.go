// Package hooks provides event hook management for the Claude Code application.
// It supports synchronous and asynchronous hooks for events like PreToolUse,
// UserPromptSubmit, SessionStart, etc.
package hooks

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/Aspirin0000/claude-code-go/internal/types"
)

// SyncHookFunc is a synchronous hook function.
type SyncHookFunc func(ctx context.Context, input types.HookInput) (*types.SyncHookResponse, error)

// AsyncHookFunc is an asynchronous hook function.
type AsyncHookFunc func(ctx context.Context, input types.HookInput) (*types.AsyncHookResult, error)

// registeredHook holds a registered hook callback.
type registeredHook struct {
	Name  string
	Sync  SyncHookFunc
	Async AsyncHookFunc
}

// Manager manages hook registrations and execution.
type Manager struct {
	mu    sync.RWMutex
	hooks map[types.HookEvent][]registeredHook
}

// NewManager creates a new hook manager.
func NewManager() *Manager {
	return &Manager{
		hooks: make(map[types.HookEvent][]registeredHook),
	}
}

// RegisterSync registers a synchronous hook for an event.
func (m *Manager) RegisterSync(event types.HookEvent, name string, fn SyncHookFunc) error {
	if !types.IsHookEvent(string(event)) {
		return fmt.Errorf("invalid hook event: %s", event)
	}
	if fn == nil {
		return fmt.Errorf("hook function cannot be nil")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.hooks[event] = append(m.hooks[event], registeredHook{Name: name, Sync: fn})
	return nil
}

// RegisterAsync registers an asynchronous hook for an event.
func (m *Manager) RegisterAsync(event types.HookEvent, name string, fn AsyncHookFunc) error {
	if !types.IsHookEvent(string(event)) {
		return fmt.Errorf("invalid hook event: %s", event)
	}
	if fn == nil {
		return fmt.Errorf("hook function cannot be nil")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.hooks[event] = append(m.hooks[event], registeredHook{Name: name, Async: fn})
	return nil
}

// Unregister removes all hooks with the given name for a specific event.
func (m *Manager) Unregister(event types.HookEvent, name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	list := m.hooks[event]
	filtered := make([]registeredHook, 0, len(list))
	for _, h := range list {
		if h.Name != name {
			filtered = append(filtered, h)
		}
	}
	m.hooks[event] = filtered
}

// Clear removes all hooks for an event.
func (m *Manager) Clear(event types.HookEvent) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.hooks, event)
}

// HasHooks returns true if there are any hooks registered for the event.
func (m *Manager) HasHooks(event types.HookEvent) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.hooks[event]) > 0
}

// ExecuteSync runs all synchronous hooks for an event and returns an aggregated result.
// If any hook returns Continue=false, execution stops early.
func (m *Manager) ExecuteSync(ctx context.Context, event types.HookEvent, input types.HookInput) (*types.SyncHookResponse, error) {
	m.mu.RLock()
	hooks := make([]registeredHook, len(m.hooks[event]))
	copy(hooks, m.hooks[event])
	m.mu.RUnlock()

	result := &types.SyncHookResponse{
		Continue:       boolPtr(true),
		SuppressOutput: boolPtr(false),
	}

	for _, h := range hooks {
		if h.Sync == nil {
			continue
		}
		resp, err := h.Sync(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("hook %s/%s failed: %w", event, h.Name, err)
		}
		if resp == nil {
			continue
		}
		mergeSyncResult(result, resp)
		if result.Continue != nil && !*result.Continue {
			break
		}
	}

	return result, nil
}

// ExecuteAsync runs all asynchronous hooks for an event concurrently.
func (m *Manager) ExecuteAsync(ctx context.Context, event types.HookEvent, input types.HookInput) ([]types.AsyncHookResult, error) {
	m.mu.RLock()
	hooks := make([]registeredHook, len(m.hooks[event]))
	copy(hooks, m.hooks[event])
	m.mu.RUnlock()

	var wg sync.WaitGroup
	results := make([]types.AsyncHookResult, len(hooks))
	errCh := make(chan error, len(hooks))

	for i, h := range hooks {
		if h.Async == nil {
			continue
		}
		wg.Add(1)
		go func(idx int, hook registeredHook) {
			defer wg.Done()
			resp, err := hook.Async(ctx, input)
			if err != nil {
				errCh <- fmt.Errorf("async hook %s/%s failed: %w", event, hook.Name, err)
				return
			}
			if resp != nil {
				results[idx] = *resp
			}
		}(i, h)
	}

	wg.Wait()
	close(errCh)

	var firstErr error
	for err := range errCh {
		if firstErr == nil {
			firstErr = err
		}
	}
	if firstErr != nil {
		return nil, firstErr
	}

	return results, nil
}

// mergeSyncResult merges a single hook response into the accumulator.
func mergeSyncResult(acc, resp *types.SyncHookResponse) {
	if resp.Continue != nil {
		acc.Continue = resp.Continue
	}
	if resp.SuppressOutput != nil {
		acc.SuppressOutput = resp.SuppressOutput
	}
	if resp.StopReason != nil {
		acc.StopReason = resp.StopReason
	}
	if resp.Decision != nil {
		acc.Decision = resp.Decision
	}
	if resp.Reason != nil {
		acc.Reason = resp.Reason
	}
	if resp.SystemMessage != nil {
		acc.SystemMessage = resp.SystemMessage
	}
	if resp.HookSpecificOutput != nil {
		acc.HookSpecificOutput = resp.HookSpecificOutput
	}
}

func boolPtr(b bool) *bool {
	return &b
}

// ---------------------------------------------------------------------------
// Event-specific helpers
// ---------------------------------------------------------------------------

// PreToolUseInput is the input for PreToolUse hooks.
type PreToolUseInput struct {
	ToolName string                 `json:"toolName"`
	Input    map[string]interface{} `json:"input"`
}

// PreToolUseResult aggregates the result of PreToolUse hooks.
type PreToolUseResult struct {
	Continue          bool
	Decision          types.PermissionBehavior
	DecisionReason    string
	UpdatedInput      map[string]interface{}
	AdditionalContext string
	SuppressOutput    bool
}

// ExecutePreToolUse runs PreToolUse hooks and returns a typed result.
func (m *Manager) ExecutePreToolUse(ctx context.Context, toolName string, input map[string]interface{}) (*PreToolUseResult, error) {
	hookInput := PreToolUseInput{ToolName: toolName, Input: input}
	resp, err := m.ExecuteSync(ctx, types.HookEventPreToolUse, hookInput)
	if err != nil {
		return nil, err
	}

	result := &PreToolUseResult{
		Continue:       resp.Continue == nil || *resp.Continue,
		SuppressOutput: resp.SuppressOutput != nil && *resp.SuppressOutput,
		UpdatedInput:   cloneMap(input),
	}

	if resp.Decision != nil {
		switch *resp.Decision {
		case "approve", "allow":
			result.Decision = types.PermissionBehaviorAllow
		case "block", "deny":
			result.Decision = types.PermissionBehaviorDeny
		default:
			result.Decision = types.PermissionBehaviorAsk
		}
	} else {
		result.Decision = types.PermissionBehaviorAllow
	}

	if resp.Reason != nil {
		result.DecisionReason = *resp.Reason
	}

	if resp.HookSpecificOutput != nil {
		if out, ok := resp.HookSpecificOutput.(*types.PreToolUseOutput); ok {
			if out.PermissionDecision != nil {
				result.Decision = *out.PermissionDecision
			}
			if out.PermissionDecisionReason != nil {
				result.DecisionReason = *out.PermissionDecisionReason
			}
			if len(out.UpdatedInput) > 0 {
				result.UpdatedInput = out.UpdatedInput
			}
			if out.AdditionalContext != nil {
				result.AdditionalContext = *out.AdditionalContext
			}
		} else if raw, ok := resp.HookSpecificOutput.(map[string]interface{}); ok {
			// Attempt to extract fields from raw map
			if v, ok := raw["permissionDecision"].(string); ok {
				result.Decision = types.PermissionBehavior(v)
			}
			if v, ok := raw["permissionDecisionReason"].(string); ok {
				result.DecisionReason = v
			}
			if v, ok := raw["updatedInput"].(map[string]interface{}); ok {
				result.UpdatedInput = v
			}
			if v, ok := raw["additionalContext"].(string); ok {
				result.AdditionalContext = v
			}
		}
	}

	return result, nil
}

// UserPromptSubmitInput is the input for UserPromptSubmit hooks.
type UserPromptSubmitInput struct {
	Prompt string `json:"prompt"`
}

// UserPromptSubmitResult is the result of UserPromptSubmit hooks.
type UserPromptSubmitResult struct {
	Continue          bool
	AdditionalContext string
	SuppressOutput    bool
}

// ExecuteUserPromptSubmit runs UserPromptSubmit hooks.
func (m *Manager) ExecuteUserPromptSubmit(ctx context.Context, prompt string) (*UserPromptSubmitResult, error) {
	hookInput := UserPromptSubmitInput{Prompt: prompt}
	resp, err := m.ExecuteSync(ctx, types.HookEventUserPromptSubmit, hookInput)
	if err != nil {
		return nil, err
	}

	result := &UserPromptSubmitResult{
		Continue:       resp.Continue == nil || *resp.Continue,
		SuppressOutput: resp.SuppressOutput != nil && *resp.SuppressOutput,
	}

	if resp.HookSpecificOutput != nil {
		if out, ok := resp.HookSpecificOutput.(*types.UserPromptSubmitOutput); ok {
			if out.AdditionalContext != nil {
				result.AdditionalContext = *out.AdditionalContext
			}
		} else if raw, ok := resp.HookSpecificOutput.(map[string]interface{}); ok {
			if v, ok := raw["additionalContext"].(string); ok {
				result.AdditionalContext = v
			}
		}
	}

	return result, nil
}

// SessionStartResult is the result of SessionStart hooks.
type SessionStartResult struct {
	Continue           bool
	AdditionalContext  string
	InitialUserMessage string
	WatchPaths         []string
	SuppressOutput     bool
}

// ExecuteSessionStart runs SessionStart hooks.
func (m *Manager) ExecuteSessionStart(ctx context.Context) (*SessionStartResult, error) {
	resp, err := m.ExecuteSync(ctx, types.HookEventSessionStart, nil)
	if err != nil {
		return nil, err
	}

	result := &SessionStartResult{
		Continue:       resp.Continue == nil || *resp.Continue,
		SuppressOutput: resp.SuppressOutput != nil && *resp.SuppressOutput,
	}

	if resp.HookSpecificOutput != nil {
		if out, ok := resp.HookSpecificOutput.(*types.SessionStartOutput); ok {
			if out.AdditionalContext != nil {
				result.AdditionalContext = *out.AdditionalContext
			}
			if out.InitialUserMessage != nil {
				result.InitialUserMessage = *out.InitialUserMessage
			}
			result.WatchPaths = out.WatchPaths
		} else if raw, ok := resp.HookSpecificOutput.(map[string]interface{}); ok {
			if v, ok := raw["additionalContext"].(string); ok {
				result.AdditionalContext = v
			}
			if v, ok := raw["initialUserMessage"].(string); ok {
				result.InitialUserMessage = v
			}
			if v, ok := raw["watchPaths"].([]string); ok {
				result.WatchPaths = v
			}
		}
	}

	return result, nil
}

// Global manager instance.
var globalManager = NewManager()

// GetGlobalManager returns the global hook manager.
func GetGlobalManager() *Manager {
	return globalManager
}

// SetGlobalManager replaces the global hook manager (useful for testing).
func SetGlobalManager(m *Manager) {
	globalManager = m
}

// cloneMap deep-copies a map[string]interface{} via JSON round-trip.
func cloneMap(m map[string]interface{}) map[string]interface{} {
	if m == nil {
		return nil
	}
	b, _ := json.Marshal(m)
	var out map[string]interface{}
	_ = json.Unmarshal(b, &out)
	return out
}
