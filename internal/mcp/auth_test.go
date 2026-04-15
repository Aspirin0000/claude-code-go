package mcp

import (
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"
)

func TestStartOAuthCallbackServer(t *testing.T) {
	expectedState := "test_state_123"
	port := 18080

	// Start server in background
	resultCh := make(chan *OAuthCallbackResult, 1)
	errCh := make(chan error, 1)
	go func() {
		res, err := StartOAuthCallbackServer(port, expectedState)
		if err != nil {
			errCh <- err
			return
		}
		resultCh <- res
	}()

	// Give server time to start
	time.Sleep(200 * time.Millisecond)

	// Simulate browser callback with valid code and state
	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/oauth/callback?code=authcode123&state=%s", port, expectedState))
	if err != nil {
		t.Fatalf("callback request failed: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		t.Fatalf("unexpected status: %d, body: %s", resp.StatusCode, string(body))
	}

	select {
	case res := <-resultCh:
		if res.Code != "authcode123" {
			t.Errorf("expected code authcode123, got %s", res.Code)
		}
		if res.State != expectedState {
			t.Errorf("expected state %s, got %s", expectedState, res.State)
		}
	case err := <-errCh:
		t.Fatalf("server returned error: %v", err)
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for callback result")
	}
}

func TestStartOAuthCallbackServerInvalidState(t *testing.T) {
	expectedState := "test_state_123"
	port := 18081

	resultCh := make(chan *OAuthCallbackResult, 1)
	errCh := make(chan error, 1)
	go func() {
		res, err := StartOAuthCallbackServer(port, expectedState)
		if err != nil {
			errCh <- err
			return
		}
		resultCh <- res
	}()

	time.Sleep(200 * time.Millisecond)

	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/oauth/callback?code=authcode123&state=wrong_state", port))
	if err != nil {
		t.Fatalf("callback request failed: %v", err)
	}
	defer resp.Body.Close()

	select {
	case err := <-errCh:
		if err == nil {
			t.Fatal("expected error for invalid state")
		}
		if err.Error() != "oauth callback error: invalid state" {
			t.Errorf("unexpected error message: %v", err)
		}
	case res := <-resultCh:
		t.Fatalf("expected error but got result: %+v", res)
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for callback error")
	}
}

func TestStartOAuthCallbackServerErrorParam(t *testing.T) {
	port := 18082

	resultCh := make(chan *OAuthCallbackResult, 1)
	errCh := make(chan error, 1)
	go func() {
		res, err := StartOAuthCallbackServer(port, "any_state")
		if err != nil {
			errCh <- err
			return
		}
		resultCh <- res
	}()

	time.Sleep(200 * time.Millisecond)

	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/oauth/callback?error=access_denied", port))
	if err != nil {
		t.Fatalf("callback request failed: %v", err)
	}
	defer resp.Body.Close()

	select {
	case err := <-errCh:
		if err == nil {
			t.Fatal("expected error for access_denied")
		}
		if err.Error() != "oauth callback error: access_denied" {
			t.Errorf("unexpected error message: %v", err)
		}
	case res := <-resultCh:
		t.Fatalf("expected error but got result: %+v", res)
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for callback error")
	}
}

func TestExchangeCodeForTokenInvalidEndpoint(t *testing.T) {
	config := &OAuthConfig{
		Endpoints: OAuthEndpoints{},
	}
	_, err := ExchangeCodeForToken(config, "dummy_code")
	if err == nil {
		t.Fatal("expected error for missing token endpoint")
	}
}

func TestBuildAuthorizationURL(t *testing.T) {
	config := &OAuthConfig{
		ClientID: "client123",
		Endpoints: OAuthEndpoints{
			AuthorizationEndpoint: "https://auth.example.com/authorize",
		},
		CallbackPort: 8080,
	}
	url, err := BuildAuthorizationURL(config, "state123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "https://auth.example.com/authorize?client_id=client123&redirect_uri=http%3A%2F%2Flocalhost%3A8080%2Foauth%2Fcallback&response_type=code&scope=openid+profile+email&state=state123"
	if url != expected {
		t.Errorf("unexpected URL:\n got: %s\n want: %s", url, expected)
	}
}

func TestOAuthTokenValidity(t *testing.T) {
	token := &OAuthToken{
		AccessToken: "token123",
		ExpiresAt:   time.Now().Add(1 * time.Hour),
	}
	if !token.IsValid() {
		t.Error("expected token to be valid")
	}
	if token.IsExpired() {
		t.Error("expected token not to be expired")
	}

	expired := &OAuthToken{
		AccessToken: "token456",
		ExpiresAt:   time.Now().Add(-1 * time.Hour),
	}
	if expired.IsValid() {
		t.Error("expected expired token to be invalid")
	}
	if !expired.IsExpired() {
		t.Error("expected token to be expired")
	}
}
