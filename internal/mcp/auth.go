// Package mcp 提供 MCP OAuth 认证支持
// 来源: src/services/mcp/auth.ts
// 重构: Go MCP OAuth 认证
package mcp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/Aspirin0000/claude-code-go/pkg/utils"
)

// ============================================================================
// OAuth 令牌类型
// ============================================================================

// OAuthToken OAuth 令牌
// 对应 TS: OAuthToken
type OAuthToken struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	ExpiresAt    time.Time `json:"expires_at,omitempty"`
	TokenType    string    `json:"token_type,omitempty"`
}

// IsExpired 检查令牌是否已过期
func (t *OAuthToken) IsExpired() bool {
	if t.ExpiresAt.IsZero() {
		return false
	}
	return time.Now().After(t.ExpiresAt)
}

// IsValid 检查令牌是否有效（非空且未过期）
func (t *OAuthToken) IsValid() bool {
	if t == nil || t.AccessToken == "" {
		return false
	}
	return !t.IsExpired()
}

// ============================================================================
// OAuth 配置
// ============================================================================

// OAuthEndpoints OAuth 端点配置
type OAuthEndpoints struct {
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
	RevocationEndpoint    string `json:"revocation_endpoint,omitempty"`
	UserInfoEndpoint      string `json:"userinfo_endpoint,omitempty"`
}

// OAuthConfig OAuth 配置
type OAuthConfig struct {
	ClientID              string         `json:"client_id"`
	ClientSecret          string         `json:"client_secret,omitempty"`
	CallbackPort          int            `json:"callback_port,omitempty"`
	AuthServerMetadataURL string         `json:"auth_server_metadata_url,omitempty"`
	Endpoints             OAuthEndpoints `json:"endpoints"`
	Scopes                []string       `json:"scopes,omitempty"`
	Audience              string         `json:"audience,omitempty"`
}

// GetDefaultScopes 获取默认 OAuth 范围
func (c *OAuthConfig) GetDefaultScopes() []string {
	if len(c.Scopes) > 0 {
		return c.Scopes
	}
	return []string{"openid", "profile", "email"}
}

// GetCallbackURL 获取回调 URL
func (c *OAuthConfig) GetCallbackURL() string {
	port := c.CallbackPort
	if port == 0 {
		port = 8080
	}
	return fmt.Sprintf("http://localhost:%d/oauth/callback", port)
}

// ============================================================================
// 令牌存储接口
// ============================================================================

// TokenStorage 令牌存储接口
type TokenStorage interface {
	Save(serverID string, token *OAuthToken) error
	Load(serverID string) (*OAuthToken, error)
	Delete(serverID string) error
	Exists(serverID string) bool
}

// FileTokenStorage 基于文件的令牌存储
type FileTokenStorage struct {
	basePath string
	mu       sync.RWMutex
}

// NewFileTokenStorage 创建文件令牌存储
func NewFileTokenStorage(basePath string) *FileTokenStorage {
	return &FileTokenStorage{
		basePath: basePath,
	}
}

// getTokenPath 获取令牌文件路径
func (s *FileTokenStorage) getTokenPath(serverID string) string {
	// 清理服务器 ID 以用于文件名
	safeID := sanitizeServerID(serverID)
	return filepath.Join(s.basePath, fmt.Sprintf("%s_token.json", safeID))
}

// sanitizeServerID 清理服务器 ID
func sanitizeServerID(serverID string) string {
	// 替换不安全的字符
	result := ""
	for _, c := range serverID {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_' {
			result += string(c)
		} else {
			result += "_"
		}
	}
	return result
}

// Save 保存令牌
func (s *FileTokenStorage) Save(serverID string, token *OAuthToken) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := os.MkdirAll(s.basePath, 0700); err != nil {
		return fmt.Errorf("创建令牌目录失败: %w", err)
	}

	data, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化令牌失败: %w", err)
	}

	path := s.getTokenPath(serverID)
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("写入令牌文件失败: %w", err)
	}

	return nil
}

// Load 加载令牌
func (s *FileTokenStorage) Load(serverID string) (*OAuthToken, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	path := s.getTokenPath(serverID)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("读取令牌文件失败: %w", err)
	}

	var token OAuthToken
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, fmt.Errorf("解析令牌失败: %w", err)
	}

	return &token, nil
}

// Delete 删除令牌
func (s *FileTokenStorage) Delete(serverID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := s.getTokenPath(serverID)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("删除令牌文件失败: %w", err)
	}

	return nil
}

// Exists 检查令牌是否存在
func (s *FileTokenStorage) Exists(serverID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	path := s.getTokenPath(serverID)
	_, err := os.Stat(path)
	return err == nil
}

// ============================================================================
// ClaudeAuthProvider OAuth 认证提供者
// ============================================================================

// ClaudeAuthProvider Claude OAuth 认证提供者
type ClaudeAuthProvider struct {
	serverID   string
	config     *OAuthConfig
	storage    TokenStorage
	token      *OAuthToken
	mu         sync.RWMutex
	httpClient *http.Client
}

// NewClaudeAuthProvider 创建新的 Claude Auth Provider
func NewClaudeAuthProvider(serverID string, config *OAuthConfig, storage TokenStorage) *ClaudeAuthProvider {
	if storage == nil {
		// 使用默认文件存储
		tokenPath := filepath.Join(utils.GetClaudeConfigHomeDir(), "mcp-tokens")
		storage = NewFileTokenStorage(tokenPath)
	}

	return &ClaudeAuthProvider{
		serverID: serverID,
		config:   config,
		storage:  storage,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Tokens 获取当前令牌
// 如果令牌不存在，会尝试从存储加载
func (p *ClaudeAuthProvider) Tokens() (*OAuthToken, error) {
	p.mu.RLock()
	token := p.token
	p.mu.RUnlock()

	if token != nil && token.IsValid() {
		return token, nil
	}

	// 尝试从存储加载
	storedToken, err := p.storage.Load(p.serverID)
	if err != nil {
		return nil, err
	}

	if storedToken != nil {
		p.mu.Lock()
		p.token = storedToken
		p.mu.Unlock()
	}

	return storedToken, nil
}

// Refresh 刷新访问令牌
func (p *ClaudeAuthProvider) Refresh() (*OAuthToken, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// 检查是否有刷新令牌
	currentToken := p.token
	if currentToken == nil {
		var err error
		currentToken, err = p.storage.Load(p.serverID)
		if err != nil {
			return nil, fmt.Errorf("加载现有令牌失败: %w", err)
		}
	}

	if currentToken == nil || currentToken.RefreshToken == "" {
		return nil, fmt.Errorf("没有可用的刷新令牌")
	}

	// 准备刷新请求
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", currentToken.RefreshToken)
	data.Set("client_id", p.config.ClientID)
	if p.config.ClientSecret != "" {
		data.Set("client_secret", p.config.ClientSecret)
	}

	// 发送刷新请求
	tokenURL := p.config.Endpoints.TokenEndpoint
	if tokenURL == "" {
		return nil, fmt.Errorf("未配置令牌端点")
	}

	resp, err := p.httpClient.PostForm(tokenURL, data)
	if err != nil {
		return nil, fmt.Errorf("刷新令牌请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("刷新令牌失败: HTTP %d", resp.StatusCode)
	}

	// 解析响应
	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		TokenType    string `json:"token_type"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("解析令牌响应失败: %w", err)
	}

	// 创建新令牌
	newToken := &OAuthToken{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		TokenType:    tokenResp.TokenType,
	}

	if tokenResp.ExpiresIn > 0 {
		newToken.ExpiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	}

	// 如果响应中没有新的刷新令牌，保留旧的
	if newToken.RefreshToken == "" {
		newToken.RefreshToken = currentToken.RefreshToken
	}

	// 更新存储
	p.token = newToken
	if err := p.storage.Save(p.serverID, newToken); err != nil {
		return nil, fmt.Errorf("保存刷新后的令牌失败: %w", err)
	}

	return newToken, nil
}

// Revoke 撤销令牌
func (p *ClaudeAuthProvider) Revoke() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	token := p.token
	if token == nil {
		var err error
		token, err = p.storage.Load(p.serverID)
		if err != nil {
			return fmt.Errorf("加载令牌失败: %w", err)
		}
	}

	if token == nil {
		return nil // 没有令牌需要撤销
	}

	// 如果配置了撤销端点，发送撤销请求
	revokeURL := p.config.Endpoints.RevocationEndpoint
	if revokeURL != "" && token.AccessToken != "" {
		data := url.Values{}
		data.Set("token", token.AccessToken)
		data.Set("client_id", p.config.ClientID)
		if p.config.ClientSecret != "" {
			data.Set("client_secret", p.config.ClientSecret)
		}

		resp, err := p.httpClient.PostForm(revokeURL, data)
		if err == nil {
			resp.Body.Close()
		}
		// 即使撤销请求失败，也继续清理本地令牌
	}

	// 清除内存中的令牌
	p.token = nil

	// 删除存储的令牌
	if err := p.storage.Delete(p.serverID); err != nil {
		return fmt.Errorf("删除存储的令牌失败: %w", err)
	}

	return nil
}

// IsAuthenticated 检查是否已认证
func (p *ClaudeAuthProvider) IsAuthenticated() bool {
	token, _ := p.Tokens()
	return token != nil && token.IsValid()
}

// SetToken 手动设置令牌（用于测试或从外部获取令牌）
func (p *ClaudeAuthProvider) SetToken(token *OAuthToken) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.token = token
	return p.storage.Save(p.serverID, token)
}

// ============================================================================
// 辅助函数
// ============================================================================

// GetTokenStoragePath 获取默认令牌存储路径
func GetTokenStoragePath() string {
	return filepath.Join(utils.GetClaudeConfigHomeDir(), "mcp-tokens")
}

// CreateAuthProviderFromConfig 从配置创建认证提供者
func CreateAuthProviderFromConfig(serverID string, oauthConfig *McpOAuthConfig) (*ClaudeAuthProvider, error) {
	if oauthConfig == nil {
		return nil, fmt.Errorf("OAuth 配置为空")
	}

	config := &OAuthConfig{
		CallbackPort: 8080,
	}

	if oauthConfig.ClientID != nil {
		config.ClientID = *oauthConfig.ClientID
	}

	if oauthConfig.CallbackPort != nil {
		config.CallbackPort = *oauthConfig.CallbackPort
	}

	// 这里可以从 metadata URL 获取端点配置
	// Endpoint configuration can be retrieved from metadata URL or assumed to be known

	storage := NewFileTokenStorage(GetTokenStoragePath())
	return NewClaudeAuthProvider(serverID, config, storage), nil
}

// BuildAuthorizationURL 构建授权 URL
func BuildAuthorizationURL(config *OAuthConfig, state string) (string, error) {
	authURL := config.Endpoints.AuthorizationEndpoint
	if authURL == "" {
		return "", fmt.Errorf("未配置授权端点")
	}

	params := url.Values{}
	params.Set("response_type", "code")
	params.Set("client_id", config.ClientID)
	params.Set("redirect_uri", config.GetCallbackURL())
	params.Set("state", state)

	scopes := config.GetDefaultScopes()
	if len(scopes) > 0 {
		params.Set("scope", joinScopes(scopes))
	}

	if config.Audience != "" {
		params.Set("audience", config.Audience)
	}

	return authURL + "?" + params.Encode(), nil
}

// ExchangeCodeForToken 用授权码交换令牌
func ExchangeCodeForToken(config *OAuthConfig, code string) (*OAuthToken, error) {
	tokenURL := config.Endpoints.TokenEndpoint
	if tokenURL == "" {
		return nil, fmt.Errorf("未配置令牌端点")
	}

	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", config.GetCallbackURL())
	data.Set("client_id", config.ClientID)
	if config.ClientSecret != "" {
		data.Set("client_secret", config.ClientSecret)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.PostForm(tokenURL, data)
	if err != nil {
		return nil, fmt.Errorf("令牌请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("令牌请求失败: HTTP %d", resp.StatusCode)
	}

	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		TokenType    string `json:"token_type"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("解析令牌响应失败: %w", err)
	}

	token := &OAuthToken{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		TokenType:    tokenResp.TokenType,
	}

	if tokenResp.ExpiresIn > 0 {
		token.ExpiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	}

	return token, nil
}

// PerformOAuthFlow 执行完整的 OAuth 授权流程
func PerformOAuthFlow(config *OAuthConfig) (*OAuthToken, error) {
	state := GenerateState()
	authURL, err := BuildAuthorizationURL(config, state)
	if err != nil {
		return nil, err
	}

	// Open browser
	openBrowser(authURL)

	// Start callback server
	result, err := StartOAuthCallbackServer(config.CallbackPort, state)
	if err != nil {
		return nil, err
	}

	// Exchange code for token
	token, err := ExchangeCodeForToken(config, result.Code)
	if err != nil {
		return nil, err
	}

	return token, nil
}

func openBrowser(url string) {
	var cmd string
	var args []string
	switch os := os.Getenv("GOOS"); os {
	case "darwin":
		cmd = "open"
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	default:
		cmd = "xdg-open"
	}
	args = append(args, url)
	exec.Command(cmd, args...).Start()
}

// joinScopes 将范围数组连接成空格分隔的字符串
func joinScopes(scopes []string) string {
	result := ""
	for i, scope := range scopes {
		if i > 0 {
			result += " "
		}
		result += scope
	}
	return result
}

// OAuthCallbackResult 回调结果
type OAuthCallbackResult struct {
	Code  string
	State string
	Error string
}

// StartOAuthCallbackServer 启动本地 OAuth 回调服务器
func StartOAuthCallbackServer(port int, expectedState string) (*OAuthCallbackResult, error) {
	result := make(chan *OAuthCallbackResult, 1)
	mux := http.NewServeMux()
	server := &http.Server{Addr: fmt.Sprintf(":%d", port), Handler: mux}

	mux.HandleFunc("/oauth/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		state := r.URL.Query().Get("state")
		errMsg := r.URL.Query().Get("error")

		if errMsg != "" {
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte("<html><body><h2>Authentication Failed</h2><p>You can close this window.</p></body></html>"))
			result <- &OAuthCallbackResult{Error: errMsg}
			return
		}

		if code == "" {
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte("<html><body><h2>Authentication Failed</h2><p>No authorization code received. You can close this window.</p></body></html>"))
			result <- &OAuthCallbackResult{Error: "missing authorization code"}
			return
		}

		if expectedState != "" && state != expectedState {
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte("<html><body><h2>Authentication Failed</h2><p>Invalid state parameter. You can close this window.</p></body></html>"))
			result <- &OAuthCallbackResult{Error: "invalid state"}
			return
		}

		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte("<html><body><h2>Authentication Successful</h2><p>You can close this window and return to Claude Code.</p></body></html>"))
		result <- &OAuthCallbackResult{Code: code, State: state}
	})

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			result <- &OAuthCallbackResult{Error: err.Error()}
		}
	}()

	select {
	case res := <-result:
		server.Close()
		if res.Error != "" {
			return nil, fmt.Errorf("oauth callback error: %s", res.Error)
		}
		return res, nil
	case <-time.After(5 * time.Minute):
		server.Close()
		return nil, fmt.Errorf("oauth callback timeout")
	}
}

// GenerateState 生成随机的 state 参数
func GenerateState() string {
	// Generate state using timestamp and random number
	return fmt.Sprintf("state_%d", time.Now().UnixNano())
}

// GetAuthHeader 获取认证头
func GetAuthHeader(token *OAuthToken) string {
	if token == nil || token.AccessToken == "" {
		return ""
	}

	tokenType := token.TokenType
	if tokenType == "" {
		tokenType = "Bearer"
	}

	return fmt.Sprintf("%s %s", tokenType, token.AccessToken)
}
