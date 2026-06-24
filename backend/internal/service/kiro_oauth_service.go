package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// KiroAuthEndpoint Kiro 社交登录 (Github/Google) token 刷新地址
// 参考 chaogei/Kiro-account-manager src/main/index.ts:190
const KiroAuthEndpoint = "https://prod.us-east-1.auth.desktop.kiro.dev"

// KiroOAuthService 处理 Kiro/CodeWhisperer OAuth token 刷新
//
// 支持两种刷新流程：
//   - OIDC（Builder ID / IdC）：POST https://oidc.{region}.amazonaws.com/token
//     body: {clientId, clientSecret, refreshToken, grantType:"refresh_token"}
//   - Social（Github/Google）：POST {KiroAuthEndpoint}/refreshToken
//     body: {refreshToken}，需要 KiroIDE user-agent
//
// 凭证存储约定（snake_case）：
//   - access_token / refresh_token / client_id / client_secret
//   - region / login_type / profile_arn / machine_id
//   - expires_at（Unix 秒，作为字符串存）
type KiroOAuthService struct {
	httpClient *http.Client
}

// NewKiroOAuthService 创建 Kiro OAuth 服务
func NewKiroOAuthService() *KiroOAuthService {
	return &KiroOAuthService{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// KiroTokenInfo 刷新返回的 token 信息
type KiroTokenInfo struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
	ExpiresAt    int64 // Unix 秒
}

// kiroOidcRefreshResponse OIDC /token 响应体（驼峰）
type kiroOidcRefreshResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int64  `json:"expiresIn"`
	TokenType    string `json:"tokenType"`
}

// kiroOidcRefreshRequest OIDC /token 请求体
type kiroOidcRefreshRequest struct {
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
	RefreshToken string `json:"refreshToken"`
	GrantType    string `json:"grantType"`
}

// kiroSocialRefreshRequest Social /refreshToken 请求体
type kiroSocialRefreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

// RefreshOIDC 走 IdC/BuilderID 的 OIDC 刷新流程
func (s *KiroOAuthService) RefreshOIDC(
	ctx context.Context,
	region, refreshToken, clientID, clientSecret string,
) (*KiroTokenInfo, error) {
	if strings.TrimSpace(refreshToken) == "" {
		return nil, fmt.Errorf("kiro: empty refresh token")
	}
	if strings.TrimSpace(clientID) == "" || strings.TrimSpace(clientSecret) == "" {
		return nil, fmt.Errorf("kiro: clientId/clientSecret required for OIDC refresh")
	}
	if region == "" {
		region = "us-east-1"
	}
	url := fmt.Sprintf("https://oidc.%s.amazonaws.com/token", region)

	body, err := json.Marshal(kiroOidcRefreshRequest{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RefreshToken: refreshToken,
		GrantType:    "refresh_token",
	})
	if err != nil {
		return nil, fmt.Errorf("kiro oidc: marshal: %w", err)
	}

	resp, err := s.doJSONRequest(ctx, url, body, nil)
	if err != nil {
		return nil, err
	}
	return s.parseRefreshResponse(resp, refreshToken)
}

// RefreshSocial 走 Github/Google 社交登录的刷新流程
func (s *KiroOAuthService) RefreshSocial(
	ctx context.Context,
	refreshToken, machineID string,
) (*KiroTokenInfo, error) {
	if strings.TrimSpace(refreshToken) == "" {
		return nil, fmt.Errorf("kiro: empty refresh token")
	}
	url := KiroAuthEndpoint + "/refreshToken"
	body, err := json.Marshal(kiroSocialRefreshRequest{RefreshToken: refreshToken})
	if err != nil {
		return nil, fmt.Errorf("kiro social: marshal: %w", err)
	}
	headers := map[string]string{
		"User-Agent": kiroUserAgent(machineID),
	}
	resp, err := s.doJSONRequest(ctx, url, body, headers)
	if err != nil {
		return nil, err
	}
	return s.parseRefreshResponse(resp, refreshToken)
}

// RefreshAccountToken 根据账号上的 login_type 字段路由到正确的刷新流程
func (s *KiroOAuthService) RefreshAccountToken(ctx context.Context, account *Account) (*KiroTokenInfo, error) {
	if account.Platform != PlatformKiro || account.Type != AccountTypeOAuth {
		return nil, fmt.Errorf("kiro: account is not a kiro oauth account")
	}
	refreshToken := strings.TrimSpace(account.GetCredential("refresh_token"))
	if refreshToken == "" {
		return nil, fmt.Errorf("kiro: account has no refresh_token")
	}

	loginType := strings.ToLower(strings.TrimSpace(account.GetCredential("login_type")))
	switch loginType {
	case "github", "google", "social":
		machineID := strings.TrimSpace(account.GetCredential("machine_id"))
		return s.RefreshSocial(ctx, refreshToken, machineID)
	default:
		// builder / idc / 空 → OIDC 流程
		region := strings.TrimSpace(account.GetCredential("region"))
		clientID := strings.TrimSpace(account.GetCredential("client_id"))
		clientSecret := strings.TrimSpace(account.GetCredential("client_secret"))
		return s.RefreshOIDC(ctx, region, refreshToken, clientID, clientSecret)
	}
}

// BuildAccountCredentials 把刷新结果合并成账号 credentials map
//
// 仅写入 token 与过期时间相关字段；client_id / client_secret / region / login_type
// / profile_arn / machine_id 由 MergeCredentials 在外层保留。
func (s *KiroOAuthService) BuildAccountCredentials(tokenInfo *KiroTokenInfo) map[string]any {
	creds := map[string]any{
		"access_token": tokenInfo.AccessToken,
		"expires_at":   strconv.FormatInt(tokenInfo.ExpiresAt, 10),
		// last_refreshed_at 记录本次刷新时间（Unix 秒），供固定周期强制刷新判断使用。
		"last_refreshed_at": strconv.FormatInt(time.Now().Unix(), 10),
	}
	if tokenInfo.RefreshToken != "" {
		creds["refresh_token"] = tokenInfo.RefreshToken
	}
	return creds
}

// doJSONRequest 发送 POST application/json，处理代理与超时
func (s *KiroOAuthService) doJSONRequest(
	ctx context.Context,
	url string,
	body []byte,
	extraHeaders map[string]string,
) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("kiro: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	for k, v := range extraHeaders {
		req.Header.Set(k, v)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("kiro: do request: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("kiro: refresh failed http %d: %s", resp.StatusCode, string(respBody))
	}
	return respBody, nil
}

// parseRefreshResponse 把响应体转成 KiroTokenInfo
//
// AWS OIDC 在某些情况下不返回新的 refreshToken；此时复用旧的。
// expires_in 提前 5 分钟作为 expires_at，避免边界刷新失败。
func (s *KiroOAuthService) parseRefreshResponse(body []byte, oldRefreshToken string) (*KiroTokenInfo, error) {
	var data kiroOidcRefreshResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("kiro: parse refresh response: %w (body=%s)", err, string(body))
	}
	if data.AccessToken == "" {
		return nil, fmt.Errorf("kiro: refresh response missing accessToken (body=%s)", string(body))
	}
	rt := data.RefreshToken
	if rt == "" {
		rt = oldRefreshToken
	}
	expiresIn := data.ExpiresIn
	if expiresIn <= 0 {
		expiresIn = 3600 // 兜底 1h
	}
	expiresAt := time.Now().Unix() + expiresIn - 300
	return &KiroTokenInfo{
		AccessToken:  data.AccessToken,
		RefreshToken: rt,
		ExpiresIn:    expiresIn,
		ExpiresAt:    expiresAt,
	}, nil
}

// kiroUserAgent 返回与 KiroIDE 一致的 User-Agent，社交刷新需要
//
// 参考 chaogei/Kiro-account-manager src/main/index.ts:636
const (
	kiroVersion           = "0.12.155"
	kiroAwsSdkVersionUA   = "1.0.34"
	kiroDefaultMachineHex = "kiro-default-machine-id-placeholder"
)

func kiroUserAgent(machineID string) string {
	suffix := fmt.Sprintf("KiroIDE-%s", kiroVersion)
	if machineID != "" {
		suffix = fmt.Sprintf("KiroIDE-%s-%s", kiroVersion, machineID)
	}
	return fmt.Sprintf(
		"aws-sdk-js/%s ua/2.1 os/linux lang/js md/nodejs#20.16.0 api/codewhispererstreaming#%s m/E %s",
		kiroAwsSdkVersionUA, kiroAwsSdkVersionUA, suffix,
	)
}
