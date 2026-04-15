package bootstrap

import (
	"os"
	"testing"
	"time"
)

func init() {
	os.Setenv("GO_ENV", "test")
}

func TestInitState(t *testing.T) {
	ResetStateForTests()
	InitState()

	if state == nil {
		t.Fatal("state should not be nil after InitState")
	}
	if state.SessionId == "" {
		t.Error("SessionId should be set after InitState")
	}
	if state.StartTime.IsZero() {
		t.Error("StartTime should be set after InitState")
	}
}

func TestGenerateUUID(t *testing.T) {
	u1 := generateUUID()
	u2 := generateUUID()
	if u1 == "" {
		t.Error("generateUUID should return non-empty string")
	}
	if u1 == u2 {
		t.Error("generateUUID should return unique values")
	}
}

func TestSessionId(t *testing.T) {
	ResetStateForTests()
	InitState()

	sid := GetSessionId()
	if sid == "" {
		t.Error("GetSessionId should return non-empty session id")
	}

	newSid := RegenerateSessionId(false)
	if newSid == sid {
		t.Error("RegenerateSessionId should create a new session id")
	}
	if GetParentSessionId() != nil {
		t.Error("ParentSessionId should be nil when setCurrentAsParent is false")
	}

	anotherSid := RegenerateSessionId(true)
	if anotherSid == newSid {
		t.Error("RegenerateSessionId should create another new session id")
	}
	parent := GetParentSessionId()
	if parent == nil {
		t.Fatal("ParentSessionId should be set when setCurrentAsParent is true")
	}
	if *parent != newSid {
		t.Error("ParentSessionId should be the previous session id")
	}
}

func TestSwitchSession(t *testing.T) {
	ResetStateForTests()
	InitState()

	var callbackCalled bool
	OnSessionSwitch(func(sid SessionId) {
		callbackCalled = true
	})

	projectDir := "/tmp/test-project"
	SwitchSession("sess-abc", &projectDir)

	if GetSessionId() != "sess-abc" {
		t.Errorf("expected session id sess-abc, got %s", GetSessionId())
	}
	if GetSessionProjectDir() == nil || *GetSessionProjectDir() != projectDir {
		t.Error("expected project dir to be updated")
	}

	// Callback is invoked in a goroutine - give it a moment
	time.Sleep(50 * time.Millisecond)
	if !callbackCalled {
		t.Error("session switch callback should have been called")
	}
}

func TestCwdAndProjectRoot(t *testing.T) {
	ResetStateForTests()
	SetOriginalCwd("/home/user")
	SetProjectRoot("/home/user/project")
	SetCwdState("/home/user/project/src")

	if GetOriginalCwd() != "/home/user" {
		t.Errorf("unexpected original cwd: %s", GetOriginalCwd())
	}
	if GetProjectRoot() != "/home/user/project" {
		t.Errorf("unexpected project root: %s", GetProjectRoot())
	}
	if GetCwdState() != "/home/user/project/src" {
		t.Errorf("unexpected cwd state: %s", GetCwdState())
	}
}

func TestCostTracking(t *testing.T) {
	ResetStateForTests()

	AddToTotalCostState(1.5, ModelUsage{InputTokens: 100, OutputTokens: 50}, "claude-test")
	AddToTotalCostState(0.5, ModelUsage{InputTokens: 50, OutputTokens: 25}, "claude-test")

	if GetTotalCostUSD() != 2.0 {
		t.Errorf("expected total cost 2.0, got %f", GetTotalCostUSD())
	}
	// ModelUsage is stored per-model (overwritten), so last call wins
	if GetTotalInputTokens() != 50 {
		t.Errorf("expected total input tokens 50, got %d", GetTotalInputTokens())
	}
	if GetTotalOutputTokens() != 25 {
		t.Errorf("expected total output tokens 25, got %d", GetTotalOutputTokens())
	}
}

func TestDurationTracking(t *testing.T) {
	ResetStateForTests()

	AddToTotalDurationState(5.0, 4.0)
	AddToToolDuration(3.0)

	// GetTotalDuration returns time.Since(StartTime), not accumulated API duration
	if GetTotalDuration() < 0 || GetTotalDuration() > 1*time.Second {
		t.Errorf("expected total duration near 0, got %v", GetTotalDuration())
	}
	if GetTotalAPIDuration() != 5.0 {
		t.Errorf("expected api duration 5.0, got %f", GetTotalAPIDuration())
	}
	if GetTotalAPIDurationWithoutRetries() != 4.0 {
		t.Errorf("expected api duration without retries 4.0, got %f", GetTotalAPIDurationWithoutRetries())
	}
	if GetTotalToolDuration() != 3.0 {
		t.Errorf("expected tool duration 3.0, got %f", GetTotalToolDuration())
	}
}

func TestLinesChanged(t *testing.T) {
	ResetStateForTests()

	AddToTotalLinesChanged(10, 3)
	AddToTotalLinesChanged(5, 2)

	if GetTotalLinesAdded() != 15 {
		t.Errorf("expected 15 lines added, got %d", GetTotalLinesAdded())
	}
	if GetTotalLinesRemoved() != 5 {
		t.Errorf("expected 5 lines removed, got %d", GetTotalLinesRemoved())
	}
}

func TestTurnTracking(t *testing.T) {
	ResetStateForTests()

	AddToToolDuration(1.5)
	if GetTurnToolDurationMs() != 1.5 {
		t.Errorf("expected turn tool duration 1.5, got %f", GetTurnToolDurationMs())
	}

	ResetTurnToolDuration()
	if GetTurnToolDurationMs() != 0 {
		t.Errorf("expected turn tool duration 0 after reset, got %f", GetTurnToolDurationMs())
	}

	if GetTurnToolCount() != 0 {
		t.Errorf("expected turn tool count 0, got %d", GetTurnToolCount())
	}
}

func TestTokenBudget(t *testing.T) {
	ResetStateForTests()

	budget := 1000
	SnapshotOutputTokensForTurn(&budget)
	if *GetCurrentTurnTokenBudget() != 1000 {
		t.Errorf("expected token budget 1000, got %d", *GetCurrentTurnTokenBudget())
	}

	IncrementBudgetContinuationCount()
	IncrementBudgetContinuationCount()
	if GetBudgetContinuationCount() != 2 {
		t.Errorf("expected budget continuation count 2, got %d", GetBudgetContinuationCount())
	}
}

func TestFlagsAndSettings(t *testing.T) {
	ResetStateForTests()

	SetIsInteractive(true)
	if !GetIsInteractive() {
		t.Error("expected interactive to be true")
	}

	SetClientType("test-client")
	if GetClientType() != "test-client" {
		t.Errorf("unexpected client type: %s", GetClientType())
	}

	SetSessionSource("cli")
	if GetSessionSource() != "cli" {
		t.Errorf("unexpected session source: %s", GetSessionSource())
	}

	SetSessionTrustAccepted(true)
	if !GetSessionTrustAccepted() {
		t.Error("expected session trust accepted")
	}

	SetSessionPersistenceDisabled(true)
	if !GetSessionPersistenceDisabled() {
		t.Error("expected session persistence disabled")
	}
}

func TestChromeAndCoworkFlags(t *testing.T) {
	ResetStateForTests()

	trueVal := true
	SetChromeFlagOverride(&trueVal)
	if GetChromeFlagOverride() == nil || !*GetChromeFlagOverride() {
		t.Error("expected chrome flag override to be true")
	}

	SetUseCoworkPlugins(true)
	if !GetUseCoworkPlugins() {
		t.Error("expected cowork plugins to be true")
	}
}

func TestPermissionsAndChannels(t *testing.T) {
	ResetStateForTests()

	SetSessionBypassPermissions(true)
	if !GetSessionBypassPermissions() {
		t.Error("expected session bypass permissions to be true")
	}

	channels := []ChannelEntry{{Kind: "plugin", Name: "test-plugin"}}
	SetAllowedChannels(channels)
	allowed := GetAllowedChannels()
	if len(allowed) != 1 || allowed[0].Name != "test-plugin" {
		t.Errorf("unexpected allowed channels: %+v", allowed)
	}

	SetHasDevChannels(true)
	if !GetHasDevChannels() {
		t.Error("expected has dev channels to be true")
	}
}

func TestAdditionalDirectoriesForClaudeMd(t *testing.T) {
	ResetStateForTests()

	dirs := []string{"/tmp/dir1", "/tmp/dir2"}
	SetAdditionalDirectoriesForClaudeMd(dirs)
	got := GetAdditionalDirectoriesForClaudeMd()
	if len(got) != 2 || got[0] != "/tmp/dir1" {
		t.Errorf("unexpected additional directories: %+v", got)
	}
}

func TestPostCompaction(t *testing.T) {
	ResetStateForTests()

	MarkPostCompaction()
	if !ConsumePostCompaction() {
		t.Error("expected ConsumePostCompaction to return true")
	}
	if ConsumePostCompaction() {
		t.Error("expected ConsumePostCompaction to return false on second call")
	}
}

func TestRequestIdAndTimestamp(t *testing.T) {
	ResetStateForTests()

	SetLastMainRequestId("req-123")
	if GetLastMainRequestId() != "req-123" {
		t.Errorf("unexpected request id: %s", GetLastMainRequestId())
	}

	ts := time.Now()
	SetLastApiCompletionTimestamp(&ts)
	gotTs := GetLastApiCompletionTimestamp()
	if gotTs == nil || !gotTs.Equal(ts) {
		t.Error("unexpected last api completion timestamp")
	}
}

func TestPromptId(t *testing.T) {
	ResetStateForTests()

	SetPromptId("prompt-456")
	if GetPromptId() != "prompt-456" {
		t.Errorf("unexpected prompt id: %s", GetPromptId())
	}
}

func TestStrictToolResultPairing(t *testing.T) {
	ResetStateForTests()

	SetStrictToolResultPairing(true)
	if !GetStrictToolResultPairing() {
		t.Error("expected strict tool result pairing to be true")
	}
}

func TestScheduledTasks(t *testing.T) {
	ResetStateForTests()

	SetScheduledTasksEnabled(true)
	if !GetScheduledTasksEnabled() {
		t.Error("expected scheduled tasks to be enabled")
	}
}

func TestInlinePlugins(t *testing.T) {
	ResetStateForTests()

	plugins := []string{"plugin-a", "plugin-b"}
	SetInlinePlugins(plugins)
	got := GetInlinePlugins()
	if len(got) != 2 || got[1] != "plugin-b" {
		t.Errorf("unexpected inline plugins: %+v", got)
	}
}

func TestUnknownModelCost(t *testing.T) {
	ResetStateForTests()

	SetHasUnknownModelCost()
	if !HasUnknownModelCost() {
		t.Error("expected unknown model cost to be true")
	}
}

func TestSessionStartTime(t *testing.T) {
	ResetStateForTests()
	InitState()

	if GetSessionStartTime().IsZero() {
		t.Error("expected non-zero session start time")
	}
}

func TestIsNonInteractiveSession(t *testing.T) {
	ResetStateForTests()
	SetIsInteractive(false)
	if !GetIsNonInteractiveSession() {
		t.Error("expected non-interactive session to be true")
	}
}

func TestResetStateForTestsPanicsOutsideTest(t *testing.T) {
	os.Setenv("GO_ENV", "production")
	defer os.Setenv("GO_ENV", "test")

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when calling ResetStateForTests outside test env")
		}
	}()
	ResetStateForTests()
}
