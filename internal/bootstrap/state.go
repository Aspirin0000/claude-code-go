// Package bootstrap provides startup state management
// Source: src/bootstrap/state.ts
// Refactor: Go bootstrap state with full session management and UUID generation
package bootstrap

import (
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
)

// SessionId type for session identifiers
type SessionId string

// ChannelEntry represents a channel entry for allowlist
type ChannelEntry struct {
	Kind        string `json:"kind"` // "plugin" or "server"
	Name        string `json:"name"`
	Marketplace string `json:"marketplace,omitempty"`
	Dev         bool   `json:"dev,omitempty"`
}

// AttributedCounter interface for metrics counters
type AttributedCounter interface {
	Add(value float64, attributes map[string]interface{})
}

// ModelUsage tracks usage for a specific model
type ModelUsage struct {
	InputTokens              int `json:"inputTokens"`
	OutputTokens             int `json:"outputTokens"`
	CacheReadInputTokens     int `json:"cacheReadInputTokens"`
	CacheCreationInputTokens int `json:"cacheCreationInputTokens"`
	WebSearchRequests        int `json:"webSearchRequests"`
}

// State holds all the global state
// DO NOT ADD MORE STATE HERE - BE JUDICIOUS WITH GLOBAL STATE
type State struct {
	// Working directories
	OriginalCwd string
	ProjectRoot string
	Cwd         string

	// Cost tracking
	TotalCostUSD                   float64
	TotalAPIDuration               float64
	TotalAPIDurationWithoutRetries float64
	TotalToolDuration              float64
	HasUnknownModelCost            bool

	// Turn tracking
	TurnHookDurationMs       float64
	TurnToolDurationMs       float64
	TurnClassifierDurationMs float64
	TurnToolCount            int
	TurnHookCount            int
	TurnClassifierCount      int

	// Session timing
	StartTime           time.Time
	LastInteractionTime time.Time

	// Lines changed
	TotalLinesAdded   int
	TotalLinesRemoved int

	// Model usage
	ModelUsage map[string]ModelUsage

	// Session identifiers
	SessionId       SessionId
	ParentSessionId *SessionId

	// Session state
	IsInteractive              bool
	StrictToolResultPairing    bool
	SessionTrustAccepted       bool
	SessionPersistenceDisabled bool

	// Client info
	ClientType    string
	SessionSource string

	// Flags
	ChromeFlagOverride       *bool
	UseCoworkPlugins         bool
	SessionBypassPermissions bool
	ScheduledTasksEnabled    bool

	// Token budget tracking
	OutputTokensAtTurnStart int
	CurrentTurnTokenBudget  *int
	BudgetContinuationCount int

	// Prompt tracking
	PromptId                   string
	LastMainRequestId          string
	LastApiCompletionTimestamp *time.Time
	PendingPostCompaction      bool

	// Channel allowlist
	AllowedChannels []ChannelEntry
	HasDevChannels  bool

	// Plugin state
	InlinePlugins []string

	// Additional directories for CLAUDE.md loading
	AdditionalDirectoriesForClaudeMd []string

	// Cache for plan slugs
	PlanSlugCache map[string]string

	// Invoked skills for preservation across compaction
	InvokedSkills map[string]InvokedSkillInfo

	// Slow operations tracking
	SlowOperations []SlowOperation

	// Beta header latches
	AfkModeHeaderLatched      *bool
	FastModeHeaderLatched     *bool
	CacheEditingHeaderLatched *bool
	ThinkingClearLatched      *bool

	// Prompt cache
	PromptCache1hAllowlist []string
	PromptCache1hEligible  *bool

	// Teleported session tracking
	TeleportedSessionInfo *TeleportedSessionInfo

	// Plan mode tracking
	HasExitedPlanMode           bool
	NeedsPlanModeExitAttachment bool
	NeedsAutoModeExitAttachment bool

	// LSP recommendation
	LspRecommendationShownThisSession bool

	// System prompt cache
	SystemPromptSectionCache map[string]*string

	// Last emitted date
	LastEmittedDate *string

	// Session project directory (null = derive from originalCwd)
	SessionProjectDir *string

	// Session cron tasks
	SessionCronTasks []SessionCronTask

	// Session created teams
	SessionCreatedTeams map[string]bool
}

// InvokedSkillInfo tracks an invoked skill
type InvokedSkillInfo struct {
	SkillName string
	SkillPath string
	Content   string
	InvokedAt time.Time
	AgentId   *string
}

// SlowOperation tracks a slow operation
type SlowOperation struct {
	Operation  string
	DurationMs int
	Timestamp  time.Time
}

// TeleportedSessionInfo tracks teleported session state
type TeleportedSessionInfo struct {
	IsTeleported          bool
	HasLoggedFirstMessage bool
	SessionId             string
}

// SessionCronTask represents a session-only cron task
type SessionCronTask struct {
	ID        string
	Cron      string
	Prompt    string
	CreatedAt time.Time
	Recurring bool
	AgentId   *string
}

var (
	// Global state - initialized once
	state *State

	// Mutex for thread-safe operations
	stateMu sync.RWMutex

	// Session switch signal callbacks
	sessionSwitchCallbacks []func(SessionId)
)

// InitState initializes the global state
func InitState() {
	stateMu.Lock()
	defer stateMu.Unlock()

	if state != nil {
		return // Already initialized
	}

	// Resolve symlinks in cwd
	resolvedCwd, err := os.Getwd()
	if err != nil {
		resolvedCwd = "."
	}

	// Try to get real path
	if realpath, err := filepath.EvalSymlinks(resolvedCwd); err == nil {
		resolvedCwd = realpath
	}

	state = &State{
		OriginalCwd:                      resolvedCwd,
		ProjectRoot:                      resolvedCwd,
		Cwd:                              resolvedCwd,
		SessionId:                        SessionId(generateUUID()),
		StartTime:                        time.Now(),
		LastInteractionTime:              time.Now(),
		ModelUsage:                       make(map[string]ModelUsage),
		ClientType:                       "cli",
		PlanSlugCache:                    make(map[string]string),
		InvokedSkills:                    make(map[string]InvokedSkillInfo),
		SystemPromptSectionCache:         make(map[string]*string),
		SessionCreatedTeams:              make(map[string]bool),
		AdditionalDirectoriesForClaudeMd: []string{},
		AllowedChannels:                  []ChannelEntry{},
		InlinePlugins:                    []string{},
		SlowOperations:                   []SlowOperation{},
		SessionCronTasks:                 []SessionCronTask{},
	}
}

// generateUUID generates a proper UUID v4
func generateUUID() string {
	u, err := uuid.NewRandom()
	if err != nil {
		// Fallback to custom UUID generation if crypto/rand fails
		return generateFallbackUUID()
	}
	return u.String()
}

// generateFallbackUUID generates a UUID-like string without crypto/rand
func generateFallbackUUID() string {
	// Use time + random bytes as fallback
	now := time.Now().UnixNano()
	b := make([]byte, 16)

	// Fill with timestamp-based data
	for i := 0; i < 8; i++ {
		b[i] = byte(now >> (i * 8))
	}

	// Fill remaining with process ID and random data
	pid := os.Getpid()
	b[8] = byte(pid)
	b[9] = byte(pid >> 8)

	// Add some randomness from crypto/rand
	rand.Read(b[10:])

	// Set version (4) and variant bits per RFC 4122
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80

	return fmt.Sprintf("%x-%x-%x-%x-%x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// GetSessionId returns the current session ID
func GetSessionId() SessionId {
	stateMu.RLock()
	defer stateMu.RUnlock()

	if state == nil {
		InitState()
	}
	return state.SessionId
}

// RegenerateSessionId regenerates the session ID
// If setCurrentAsParent is true, sets current session as parent
func RegenerateSessionId(setCurrentAsParent bool) SessionId {
	stateMu.Lock()
	defer stateMu.Unlock()

	if state == nil {
		InitState()
	}

	if setCurrentAsParent {
		parentId := state.SessionId
		state.ParentSessionId = &parentId
	}

	// Clear plan slug cache entry for old session
	delete(state.PlanSlugCache, string(state.SessionId))

	// Generate new session ID
	state.SessionId = SessionId(generateUUID())
	state.SessionProjectDir = nil

	return state.SessionId
}

// GetParentSessionId returns the parent session ID
func GetParentSessionId() *SessionId {
	stateMu.RLock()
	defer stateMu.RUnlock()

	if state == nil {
		return nil
	}
	return state.ParentSessionId
}

// SwitchSession atomically switches the active session
func SwitchSession(sessionId SessionId, projectDir *string) {
	stateMu.Lock()
	defer stateMu.Unlock()

	if state == nil {
		InitState()
	}

	// Clear plan slug cache entry for old session
	delete(state.PlanSlugCache, string(state.SessionId))

	state.SessionId = sessionId
	state.SessionProjectDir = projectDir

	// Notify callbacks
	for _, callback := range sessionSwitchCallbacks {
		go callback(sessionId)
	}
}

// OnSessionSwitch registers a callback for session switches
func OnSessionSwitch(callback func(SessionId)) {
	stateMu.Lock()
	defer stateMu.Unlock()

	sessionSwitchCallbacks = append(sessionSwitchCallbacks, callback)
}

// GetSessionProjectDir returns the session project directory
func GetSessionProjectDir() *string {
	stateMu.RLock()
	defer stateMu.RUnlock()

	if state == nil {
		return nil
	}
	return state.SessionProjectDir
}

// GetOriginalCwd returns the original working directory
func GetOriginalCwd() string {
	stateMu.RLock()
	defer stateMu.RUnlock()

	if state == nil {
		InitState()
	}
	return state.OriginalCwd
}

// GetProjectRoot returns the stable project root directory
func GetProjectRoot() string {
	stateMu.RLock()
	defer stateMu.RUnlock()

	if state == nil {
		InitState()
	}
	return state.ProjectRoot
}

// SetOriginalCwd sets the original working directory
func SetOriginalCwd(cwd string) {
	stateMu.Lock()
	defer stateMu.Unlock()

	if state == nil {
		InitState()
	}
	state.OriginalCwd = cwd
}

// SetProjectRoot sets the project root directory
// Only for --worktree startup flag - mid-session tools should NOT call this
func SetProjectRoot(cwd string) {
	stateMu.Lock()
	defer stateMu.Unlock()

	if state == nil {
		InitState()
	}
	state.ProjectRoot = cwd
}

// GetCwdState returns the current working directory state
func GetCwdState() string {
	stateMu.RLock()
	defer stateMu.RUnlock()

	if state == nil {
		InitState()
	}
	return state.Cwd
}

// SetCwdState sets the current working directory state
func SetCwdState(cwd string) {
	stateMu.Lock()
	defer stateMu.Unlock()

	if state == nil {
		InitState()
	}
	state.Cwd = cwd
}

// AddToTotalDurationState adds to total duration tracking
func AddToTotalDurationState(duration, durationWithoutRetries float64) {
	stateMu.Lock()
	defer stateMu.Unlock()

	if state == nil {
		InitState()
	}
	state.TotalAPIDuration += duration
	state.TotalAPIDurationWithoutRetries += durationWithoutRetries
}

// AddToTotalCostState adds cost and model usage
func AddToTotalCostState(cost float64, modelUsage ModelUsage, model string) {
	stateMu.Lock()
	defer stateMu.Unlock()

	if state == nil {
		InitState()
	}
	state.TotalCostUSD += cost
	state.ModelUsage[model] = modelUsage
}

// GetTotalCostUSD returns total cost
func GetTotalCostUSD() float64 {
	stateMu.RLock()
	defer stateMu.RUnlock()

	if state == nil {
		return 0
	}
	return state.TotalCostUSD
}

// GetTotalAPIDuration returns total API duration
func GetTotalAPIDuration() float64 {
	stateMu.RLock()
	defer stateMu.RUnlock()

	if state == nil {
		return 0
	}
	return state.TotalAPIDuration
}

// GetTotalDuration returns total session duration
func GetTotalDuration() time.Duration {
	stateMu.RLock()
	defer stateMu.RUnlock()

	if state == nil {
		return 0
	}
	return time.Since(state.StartTime)
}

// GetTotalAPIDurationWithoutRetries returns API duration without retries
func GetTotalAPIDurationWithoutRetries() float64 {
	stateMu.RLock()
	defer stateMu.RUnlock()

	if state == nil {
		return 0
	}
	return state.TotalAPIDurationWithoutRetries
}

// GetTotalToolDuration returns total tool duration
func GetTotalToolDuration() float64 {
	stateMu.RLock()
	defer stateMu.RUnlock()

	if state == nil {
		return 0
	}
	return state.TotalToolDuration
}

// AddToToolDuration adds to tool duration
func AddToToolDuration(duration float64) {
	stateMu.Lock()
	defer stateMu.Unlock()

	if state == nil {
		InitState()
	}
	state.TotalToolDuration += duration
	state.TurnToolDurationMs += duration
	state.TurnToolCount++
}

// GetTurnToolDurationMs returns turn tool duration
func GetTurnToolDurationMs() float64 {
	stateMu.RLock()
	defer stateMu.RUnlock()

	if state == nil {
		return 0
	}
	return state.TurnToolDurationMs
}

// ResetTurnToolDuration resets turn tool duration
func ResetTurnToolDuration() {
	stateMu.Lock()
	defer stateMu.Unlock()

	if state == nil {
		return
	}
	state.TurnToolDurationMs = 0
	state.TurnToolCount = 0
}

// GetTurnToolCount returns turn tool count
func GetTurnToolCount() int {
	stateMu.RLock()
	defer stateMu.RUnlock()

	if state == nil {
		return 0
	}
	return state.TurnToolCount
}

// AddToTotalLinesChanged adds lines changed
func AddToTotalLinesChanged(added, removed int) {
	stateMu.Lock()
	defer stateMu.Unlock()

	if state == nil {
		InitState()
	}
	state.TotalLinesAdded += added
	state.TotalLinesRemoved += removed
}

// GetTotalLinesAdded returns total lines added
func GetTotalLinesAdded() int {
	stateMu.RLock()
	defer stateMu.RUnlock()

	if state == nil {
		return 0
	}
	return state.TotalLinesAdded
}

// GetTotalLinesRemoved returns total lines removed
func GetTotalLinesRemoved() int {
	stateMu.RLock()
	defer stateMu.RUnlock()

	if state == nil {
		return 0
	}
	return state.TotalLinesRemoved
}

// GetTotalInputTokens returns total input tokens
func GetTotalInputTokens() int {
	stateMu.RLock()
	defer stateMu.RUnlock()

	if state == nil {
		return 0
	}

	total := 0
	for _, usage := range state.ModelUsage {
		total += usage.InputTokens
	}
	return total
}

// GetTotalOutputTokens returns total output tokens
func GetTotalOutputTokens() int {
	stateMu.RLock()
	defer stateMu.RUnlock()

	if state == nil {
		return 0
	}

	total := 0
	for _, usage := range state.ModelUsage {
		total += usage.OutputTokens
	}
	return total
}

// GetTotalCacheReadInputTokens returns total cache read input tokens
func GetTotalCacheReadInputTokens() int {
	stateMu.RLock()
	defer stateMu.RUnlock()

	if state == nil {
		return 0
	}

	total := 0
	for _, usage := range state.ModelUsage {
		total += usage.CacheReadInputTokens
	}
	return total
}

// GetTotalCacheCreationInputTokens returns total cache creation input tokens
func GetTotalCacheCreationInputTokens() int {
	stateMu.RLock()
	defer stateMu.RUnlock()

	if state == nil {
		return 0
	}

	total := 0
	for _, usage := range state.ModelUsage {
		total += usage.CacheCreationInputTokens
	}
	return total
}

// SnapshotOutputTokensForTurn snapshots output tokens for turn budget tracking
func SnapshotOutputTokensForTurn(budget *int) {
	stateMu.Lock()
	defer stateMu.Unlock()

	if state == nil {
		InitState()
	}

	state.OutputTokensAtTurnStart = GetTotalOutputTokens()
	state.CurrentTurnTokenBudget = budget
	state.BudgetContinuationCount = 0
}

// GetTurnOutputTokens returns output tokens for current turn
func GetTurnOutputTokens() int {
	return GetTotalOutputTokens() - getOutputTokensAtTurnStart()
}

// getOutputTokensAtTurnStart returns the snapshot value (internal)
func getOutputTokensAtTurnStart() int {
	stateMu.RLock()
	defer stateMu.RUnlock()

	if state == nil {
		return 0
	}
	return state.OutputTokensAtTurnStart
}

// GetCurrentTurnTokenBudget returns the current turn token budget
func GetCurrentTurnTokenBudget() *int {
	stateMu.RLock()
	defer stateMu.RUnlock()

	if state == nil {
		return nil
	}
	return state.CurrentTurnTokenBudget
}

// GetBudgetContinuationCount returns budget continuation count
func GetBudgetContinuationCount() int {
	stateMu.RLock()
	defer stateMu.RUnlock()

	if state == nil {
		return 0
	}
	return state.BudgetContinuationCount
}

// IncrementBudgetContinuationCount increments budget continuation count
func IncrementBudgetContinuationCount() {
	stateMu.Lock()
	defer stateMu.Unlock()

	if state == nil {
		return
	}
	state.BudgetContinuationCount++
}

// SetHasUnknownModelCost marks that an unknown model cost was encountered
func SetHasUnknownModelCost() {
	stateMu.Lock()
	defer stateMu.Unlock()

	if state == nil {
		InitState()
	}
	state.HasUnknownModelCost = true
}

// HasUnknownModelCost returns whether unknown model cost was encountered
func HasUnknownModelCost() bool {
	stateMu.RLock()
	defer stateMu.RUnlock()

	if state == nil {
		return false
	}
	return state.HasUnknownModelCost
}

// UpdateLastInteractionTime marks that an interaction occurred
func UpdateLastInteractionTime() {
	stateMu.Lock()
	defer stateMu.Unlock()

	if state == nil {
		InitState()
	}
	state.LastInteractionTime = time.Now()
}

// GetLastInteractionTime returns the last interaction time
func GetLastInteractionTime() time.Time {
	stateMu.RLock()
	defer stateMu.RUnlock()

	if state == nil {
		return time.Now()
	}
	return state.LastInteractionTime
}

// GetIsInteractive returns whether session is interactive
func GetIsInteractive() bool {
	stateMu.RLock()
	defer stateMu.RUnlock()

	if state == nil {
		return false
	}
	return state.IsInteractive
}

// SetIsInteractive sets whether session is interactive
func SetIsInteractive(value bool) {
	stateMu.Lock()
	defer stateMu.Unlock()

	if state == nil {
		InitState()
	}
	state.IsInteractive = value
}

// GetClientType returns the client type
func GetClientType() string {
	stateMu.RLock()
	defer stateMu.RUnlock()

	if state == nil {
		return "cli"
	}
	return state.ClientType
}

// SetClientType sets the client type
func SetClientType(clientType string) {
	stateMu.Lock()
	defer stateMu.Unlock()

	if state == nil {
		InitState()
	}
	state.ClientType = clientType
}

// GetSessionSource returns the session source
func GetSessionSource() string {
	stateMu.RLock()
	defer stateMu.RUnlock()

	if state == nil {
		return ""
	}
	return state.SessionSource
}

// SetSessionSource sets the session source
func SetSessionSource(source string) {
	stateMu.Lock()
	defer stateMu.Unlock()

	if state == nil {
		InitState()
	}
	state.SessionSource = source
}

// GetStrictToolResultPairing returns strict tool result pairing setting
func GetStrictToolResultPairing() bool {
	stateMu.RLock()
	defer stateMu.RUnlock()

	if state == nil {
		return false
	}
	return state.StrictToolResultPairing
}

// SetStrictToolResultPairing sets strict tool result pairing
func SetStrictToolResultPairing(value bool) {
	stateMu.Lock()
	defer stateMu.Unlock()

	if state == nil {
		InitState()
	}
	state.StrictToolResultPairing = value
}

// GetSessionTrustAccepted returns session trust accepted flag
func GetSessionTrustAccepted() bool {
	stateMu.RLock()
	defer stateMu.RUnlock()

	if state == nil {
		return false
	}
	return state.SessionTrustAccepted
}

// SetSessionTrustAccepted sets session trust accepted flag
func SetSessionTrustAccepted(accepted bool) {
	stateMu.Lock()
	defer stateMu.Unlock()

	if state == nil {
		InitState()
	}
	state.SessionTrustAccepted = accepted
}

// GetSessionPersistenceDisabled returns session persistence disabled flag
func GetSessionPersistenceDisabled() bool {
	stateMu.RLock()
	defer stateMu.RUnlock()

	if state == nil {
		return false
	}
	return state.SessionPersistenceDisabled
}

// SetSessionPersistenceDisabled sets session persistence disabled flag
func SetSessionPersistenceDisabled(disabled bool) {
	stateMu.Lock()
	defer stateMu.Unlock()

	if state == nil {
		InitState()
	}
	state.SessionPersistenceDisabled = disabled
}

// GetChromeFlagOverride returns chrome flag override
func GetChromeFlagOverride() *bool {
	stateMu.RLock()
	defer stateMu.RUnlock()

	if state == nil {
		return nil
	}
	return state.ChromeFlagOverride
}

// SetChromeFlagOverride sets chrome flag override
func SetChromeFlagOverride(value *bool) {
	stateMu.Lock()
	defer stateMu.Unlock()

	if state == nil {
		InitState()
	}
	state.ChromeFlagOverride = value
}

// GetUseCoworkPlugins returns whether to use cowork plugins directory
func GetUseCoworkPlugins() bool {
	stateMu.RLock()
	defer stateMu.RUnlock()

	if state == nil {
		return false
	}
	return state.UseCoworkPlugins
}

// SetUseCoworkPlugins sets whether to use cowork plugins directory
func SetUseCoworkPlugins(value bool) {
	stateMu.Lock()
	defer stateMu.Unlock()

	if state == nil {
		InitState()
	}
	state.UseCoworkPlugins = value
}

// GetSessionBypassPermissions returns session bypass permissions mode
func GetSessionBypassPermissions() bool {
	stateMu.RLock()
	defer stateMu.RUnlock()

	if state == nil {
		return false
	}
	return state.SessionBypassPermissions
}

// SetSessionBypassPermissions sets session bypass permissions mode
func SetSessionBypassPermissions(enabled bool) {
	stateMu.Lock()
	defer stateMu.Unlock()

	if state == nil {
		InitState()
	}
	state.SessionBypassPermissions = enabled
}

// GetScheduledTasksEnabled returns scheduled tasks enabled flag
func GetScheduledTasksEnabled() bool {
	stateMu.RLock()
	defer stateMu.RUnlock()

	if state == nil {
		return false
	}
	return state.ScheduledTasksEnabled
}

// SetScheduledTasksEnabled sets scheduled tasks enabled flag
func SetScheduledTasksEnabled(enabled bool) {
	stateMu.Lock()
	defer stateMu.Unlock()

	if state == nil {
		InitState()
	}
	state.ScheduledTasksEnabled = enabled
}

// GetInlinePlugins returns inline plugins list
func GetInlinePlugins() []string {
	stateMu.RLock()
	defer stateMu.RUnlock()

	if state == nil {
		return []string{}
	}
	return state.InlinePlugins
}

// SetInlinePlugins sets inline plugins list
func SetInlinePlugins(plugins []string) {
	stateMu.Lock()
	defer stateMu.Unlock()

	if state == nil {
		InitState()
	}
	state.InlinePlugins = plugins
}

// GetAllowedChannels returns allowed channels
func GetAllowedChannels() []ChannelEntry {
	stateMu.RLock()
	defer stateMu.RUnlock()

	if state == nil {
		return []ChannelEntry{}
	}
	return state.AllowedChannels
}

// SetAllowedChannels sets allowed channels
func SetAllowedChannels(channels []ChannelEntry) {
	stateMu.Lock()
	defer stateMu.Unlock()

	if state == nil {
		InitState()
	}
	state.AllowedChannels = channels
}

// GetHasDevChannels returns whether any dev channels are loaded
func GetHasDevChannels() bool {
	stateMu.RLock()
	defer stateMu.RUnlock()

	if state == nil {
		return false
	}
	return state.HasDevChannels
}

// SetHasDevChannels sets dev channels flag
func SetHasDevChannels(value bool) {
	stateMu.Lock()
	defer stateMu.Unlock()

	if state == nil {
		InitState()
	}
	state.HasDevChannels = value
}

// GetAdditionalDirectoriesForClaudeMd returns additional directories for CLAUDE.md
func GetAdditionalDirectoriesForClaudeMd() []string {
	stateMu.RLock()
	defer stateMu.RUnlock()

	if state == nil {
		return []string{}
	}
	return state.AdditionalDirectoriesForClaudeMd
}

// SetAdditionalDirectoriesForClaudeMd sets additional directories for CLAUDE.md
func SetAdditionalDirectoriesForClaudeMd(dirs []string) {
	stateMu.Lock()
	defer stateMu.Unlock()

	if state == nil {
		InitState()
	}
	state.AdditionalDirectoriesForClaudeMd = dirs
}

// MarkPostCompaction marks that a compaction just occurred
func MarkPostCompaction() {
	stateMu.Lock()
	defer stateMu.Unlock()

	if state == nil {
		InitState()
	}
	state.PendingPostCompaction = true
}

// ConsumePostCompaction consumes and returns the post-compaction flag
func ConsumePostCompaction() bool {
	stateMu.Lock()
	defer stateMu.Unlock()

	if state == nil {
		return false
	}

	was := state.PendingPostCompaction
	state.PendingPostCompaction = false
	return was
}

// GetLastMainRequestId returns the last main request ID
func GetLastMainRequestId() string {
	stateMu.RLock()
	defer stateMu.RUnlock()

	if state == nil {
		return ""
	}
	return state.LastMainRequestId
}

// SetLastMainRequestId sets the last main request ID
func SetLastMainRequestId(requestId string) {
	stateMu.Lock()
	defer stateMu.Unlock()

	if state == nil {
		InitState()
	}
	state.LastMainRequestId = requestId
}

// GetLastApiCompletionTimestamp returns the last API completion timestamp
func GetLastApiCompletionTimestamp() *time.Time {
	stateMu.RLock()
	defer stateMu.RUnlock()

	if state == nil {
		return nil
	}
	return state.LastApiCompletionTimestamp
}

// SetLastApiCompletionTimestamp sets the last API completion timestamp
func SetLastApiCompletionTimestamp(timestamp *time.Time) {
	stateMu.Lock()
	defer stateMu.Unlock()

	if state == nil {
		InitState()
	}
	state.LastApiCompletionTimestamp = timestamp
}

// GetPromptId returns the current prompt ID
func GetPromptId() string {
	stateMu.RLock()
	defer stateMu.RUnlock()

	if state == nil {
		return ""
	}
	return state.PromptId
}

// SetPromptId sets the current prompt ID
func SetPromptId(id string) {
	stateMu.Lock()
	defer stateMu.Unlock()

	if state == nil {
		InitState()
	}
	state.PromptId = id
}

// ResetStateForTests resets state for testing (only works in test mode)
func ResetStateForTests() {
	// Only allow in test environment
	if os.Getenv("GO_ENV") != "test" && os.Getenv("NODE_ENV") != "test" {
		panic("ResetStateForTests can only be called in tests")
	}

	stateMu.Lock()
	defer stateMu.Unlock()

	state = nil
	sessionSwitchCallbacks = nil
}

// GetSessionStartTime returns the session start time
// Corresponds to TS: export function getSessionStartTime(): Date
func GetSessionStartTime() time.Time {
	stateMu.RLock()
	defer stateMu.RUnlock()

	if state == nil {
		InitState()
	}
	return state.StartTime
}

// GetIsNonInteractiveSession returns whether the session is non-interactive
// Corresponds to TS: export function getIsNonInteractiveSession(): boolean
func GetIsNonInteractiveSession() bool {
	return !GetIsInteractive()
}
