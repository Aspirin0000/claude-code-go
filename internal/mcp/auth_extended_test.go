package mcp

import (
	"testing"
	"time"
)

// ============================================================================
// OAuth Token Tests
// ============================================================================

func TestOAuthTokenIsExpired(t *testing.T) {
	// Token with zero expiry time (never expires)
	token := &OAuthToken{
		AccessToken: "test-token",
		ExpiresAt:   time.Time{},
	}
	if token.IsExpired() {
		t.Error("token with zero expiry should not be expired")
	}

	// Token that expires in the future
	token2 := &OAuthToken{
		AccessToken: "test-token-2",
		ExpiresAt:   time.Now().Add(1 * time.Hour),
	}
	if token2.IsExpired() {
		t.Error("token that expires in future should not be expired")
	}

	// Token that expired in the past
	token3 := &OAuthToken{
		AccessToken: "test-token-3",
		ExpiresAt:   time.Now().Add(-1 * time.Hour),
	}
	if !token3.IsExpired() {
		t.Error("token that expired in past should be expired")
	}
}

func TestOAuthTokenIsValid(t *testing.T) {
	// Nil token
	var nilToken *OAuthToken
	if nilToken != nil && nilToken.IsValid() {
		t.Error("nil token should not be valid")
	}

	// Empty access token
	token := &OAuthToken{
		AccessToken: "",
		ExpiresAt:   time.Now().Add(1 * time.Hour),
	}
	if token.IsValid() {
		t.Error("token with empty access token should not be valid")
	}

	// Valid token
	token2 := &OAuthToken{
		AccessToken: "valid-token",
		ExpiresAt:   time.Now().Add(1 * time.Hour),
	}
	if !token2.IsValid() {
		t.Error("valid token should be valid")
	}

	// Expired token
	token3 := &OAuthToken{
		AccessToken: "expired-token",
		ExpiresAt:   time.Now().Add(-1 * time.Hour),
	}
	if token3.IsValid() {
		t.Error("expired token should not be valid")
	}
}

// ============================================================================
// OAuth Config Tests
// ============================================================================

func TestOAuthConfigGetCallbackURL(t *testing.T) {
	config := &OAuthConfig{
		CallbackPort: 8080,
	}
	url := config.GetCallbackURL()
	expected := "http://localhost:8080/oauth/callback"
	if url != expected {
		t.Errorf("expected %s, got %s", expected, url)
	}

	// Default port
	config2 := &OAuthConfig{}
	url2 := config2.GetCallbackURL()
	expected2 := "http://localhost:8080/oauth/callback"
	if url2 != expected2 {
		t.Errorf("expected %s, got %s", expected2, url2)
	}
}

func TestOAuthConfigGetDefaultScopes(t *testing.T) {
	config := &OAuthConfig{}
	scopes := config.GetDefaultScopes()
	if len(scopes) != 3 {
		t.Errorf("expected 3 default scopes, got %d", len(scopes))
	}
	expected := []string{"openid", "profile", "email"}
	for i, scope := range scopes {
		if scope != expected[i] {
			t.Errorf("expected scope %s, got %s", expected[i], scope)
		}
	}
}

// ============================================================================
// FileTokenStorage Tests
// ============================================================================

func TestFileTokenStorage(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewFileTokenStorage(tmpDir)

	// Test saving token
	token := &OAuthToken{
		AccessToken:  "access123",
		RefreshToken: "refresh456",
		ExpiresAt:    time.Now().Add(1 * time.Hour),
		TokenType:    "Bearer",
	}

	err := storage.Save("test-server", token)
	if err != nil {
		t.Fatalf("failed to save token: %v", err)
	}

	// Test loading token
	loaded, err := storage.Load("test-server")
	if err != nil {
		t.Fatalf("failed to load token: %v", err)
	}
	if loaded == nil {
		t.Fatal("expected loaded token to not be nil")
	}
	if loaded.AccessToken != token.AccessToken {
		t.Errorf("expected access token %s, got %s", token.AccessToken, loaded.AccessToken)
	}
	if loaded.RefreshToken != token.RefreshToken {
		t.Errorf("expected refresh token %s, got %s", token.RefreshToken, loaded.RefreshToken)
	}

	// Test exists
	if !storage.Exists("test-server") {
		t.Error("expected token to exist")
	}

	// Test delete
	err = storage.Delete("test-server")
	if err != nil {
		t.Fatalf("failed to delete token: %v", err)
	}

	if storage.Exists("test-server") {
		t.Error("expected token to not exist after deletion")
	}

	// Test loading non-existent token
	loaded2, err := storage.Load("non-existent")
	if err != nil {
		t.Fatalf("failed to load non-existent token: %v", err)
	}
	if loaded2 != nil {
		t.Error("expected nil for non-existent token")
	}
}

func TestFileTokenStorageSanitizeServerID(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"server-name", "server-name"},
		{"server_name", "server_name"},
		{"server.name", "server_name"},
		{"server/name", "server_name"},
		{"server:name", "server_name"},
		{"server@name", "server_name"},
	}

	for _, test := range tests {
		result := sanitizeServerID(test.input)
		if result != test.expected {
			t.Errorf("sanitizeServerID(%q) = %q, expected %q", test.input, result, test.expected)
		}
	}
}

// ============================================================================
// ClaudeAuthProvider Tests
// ============================================================================

func TestClaudeAuthProviderTokens(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewFileTokenStorage(tmpDir)
	config := &OAuthConfig{
		ClientID:     "test-client",
		CallbackPort: 8080,
	}

	provider := NewClaudeAuthProvider("test-server", config, storage)

	// Test with no token
	token, err := provider.Tokens()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != nil {
		t.Error("expected nil token when no token exists")
	}

	// Test IsAuthenticated
	if provider.IsAuthenticated() {
		t.Error("expected not authenticated when no token")
	}

	// Set token
	testToken := &OAuthToken{
		AccessToken:  "test-access",
		RefreshToken: "test-refresh",
		ExpiresAt:    time.Now().Add(1 * time.Hour),
	}
	err = provider.SetToken(testToken)
	if err != nil {
		t.Fatalf("failed to set token: %v", err)
	}

	// Test IsAuthenticated after setting token
	if !provider.IsAuthenticated() {
		t.Error("expected authenticated after setting valid token")
	}

	// Test Tokens after setting
	token, err = provider.Tokens()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token == nil {
		t.Fatal("expected non-nil token")
	}
	if token.AccessToken != testToken.AccessToken {
		t.Errorf("expected access token %s, got %s", testToken.AccessToken, token.AccessToken)
	}
}

func TestClaudeAuthProviderRevoke(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewFileTokenStorage(tmpDir)
	config := &OAuthConfig{
		ClientID:     "test-client",
		CallbackPort: 8080,
	}

	provider := NewClaudeAuthProvider("test-server", config, storage)

	// Revoke with no token should not error
	err := provider.Revoke()
	if err != nil {
		t.Fatalf("unexpected error revoking empty token: %v", err)
	}

	// Set and revoke token
	testToken := &OAuthToken{
		AccessToken:  "test-access",
		RefreshToken: "test-refresh",
		ExpiresAt:    time.Now().Add(1 * time.Hour),
	}
	err = provider.SetToken(testToken)
	if err != nil {
		t.Fatalf("failed to set token: %v", err)
	}

	if !provider.IsAuthenticated() {
		t.Error("expected authenticated before revoke")
	}

	err = provider.Revoke()
	if err != nil {
		t.Fatalf("unexpected error revoking token: %v", err)
	}

	if provider.IsAuthenticated() {
		t.Error("expected not authenticated after revoke")
	}
}

// ============================================================================
// Helper Functions Tests
// ============================================================================

func TestGenerateState(t *testing.T) {
	state1 := GenerateState()
	state2 := GenerateState()

	if state1 == "" {
		t.Error("expected non-empty state")
	}

	if state1 == state2 {
		t.Error("expected different states")
	}

	if !startsWith(state1, "state_") {
		t.Errorf("expected state to start with 'state_', got %s", state1)
	}
}

func TestGetAuthHeader(t *testing.T) {
	// Nil token
	header := GetAuthHeader(nil)
	if header != "" {
		t.Errorf("expected empty header for nil token, got %s", header)
	}

	// Empty access token
	token := &OAuthToken{AccessToken: ""}
	header = GetAuthHeader(token)
	if header != "" {
		t.Errorf("expected empty header for empty token, got %s", header)
	}

	// Token with default type
	token2 := &OAuthToken{
		AccessToken: "test-token",
		TokenType:   "",
	}
	header = GetAuthHeader(token2)
	expected := "Bearer test-token"
	if header != expected {
		t.Errorf("expected %s, got %s", expected, header)
	}

	// Token with custom type
	token3 := &OAuthToken{
		AccessToken: "test-token",
		TokenType:   "Custom",
	}
	header = GetAuthHeader(token3)
	expected = "Custom test-token"
	if header != expected {
		t.Errorf("expected %s, got %s", expected, header)
	}
}

func TestGetTokenStoragePath(t *testing.T) {
	path := GetTokenStoragePath()
	if path == "" {
		t.Error("expected non-empty token storage path")
	}
}

func TestCreateAuthProviderFromConfig(t *testing.T) {
	clientID := "test-client"
	callbackPort := 9090
	oauthConfig := &McpOAuthConfig{
		ClientID:     &clientID,
		CallbackPort: &callbackPort,
	}

	provider, err := CreateAuthProviderFromConfig("test-server", oauthConfig)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if provider == nil {
		t.Fatal("expected non-nil provider")
	}

	// Test with nil config
	_, err = CreateAuthProviderFromConfig("test-server", nil)
	if err == nil {
		t.Error("expected error for nil config")
	}
}

func TestJoinScopes(t *testing.T) {
	tests := []struct {
		scopes   []string
		expected string
	}{
		{[]string{}, ""},
		{[]string{"openid"}, "openid"},
		{[]string{"openid", "profile"}, "openid profile"},
		{[]string{"openid", "profile", "email"}, "openid profile email"},
	}

	for _, test := range tests {
		result := joinScopes(test.scopes)
		if result != test.expected {
			t.Errorf("joinScopes(%v) = %q, expected %q", test.scopes, result, test.expected)
		}
	}
}

// ============================================================================
// BuildAuthorizationURL Tests
// ============================================================================

func TestBuildAuthorizationURLError(t *testing.T) {
	config := &OAuthConfig{
		ClientID: "client123",
		Endpoints: OAuthEndpoints{
			AuthorizationEndpoint: "",
		},
	}
	_, err := BuildAuthorizationURL(config, "state123")
	if err == nil {
		t.Error("expected error for missing authorization endpoint")
	}
}

func TestBuildAuthorizationURLWithAudience(t *testing.T) {
	config := &OAuthConfig{
		ClientID: "client123",
		Endpoints: OAuthEndpoints{
			AuthorizationEndpoint: "https://auth.example.com/authorize",
		},
		CallbackPort: 8080,
		Audience:     "test-audience",
	}
	url, err := BuildAuthorizationURL(config, "state123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !contains(url, "audience=test-audience") {
		t.Errorf("expected URL to contain audience parameter, got %s", url)
	}
}

// ============================================================================
// ExchangeCodeForToken Tests
// ============================================================================

func TestExchangeCodeForTokenMissingEndpoint(t *testing.T) {
	config := &OAuthConfig{
		ClientID: "client123",
		Endpoints: OAuthEndpoints{
			TokenEndpoint: "",
		},
	}
	_, err := ExchangeCodeForToken(config, "code123")
	if err == nil {
		t.Error("expected error for missing token endpoint")
	}
}

// ============================================================================
// Helper function
// ============================================================================

func startsWith(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
