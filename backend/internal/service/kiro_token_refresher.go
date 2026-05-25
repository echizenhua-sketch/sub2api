package service

import (
	"context"
	"fmt"
	"time"
)

const (
	// kiroRefreshWindow Kiro token 提前刷新窗口：10 分钟
	// AWS OIDC accessToken 默认 1 小时有效，提前 10 分钟刷新留足余量
	kiroRefreshWindow = 10 * time.Minute
)

// KiroTokenRefresher 实现 TokenRefresher 接口
type KiroTokenRefresher struct {
	kiroOAuthService *KiroOAuthService
}

// NewKiroTokenRefresher 创建 Kiro token 刷新器
func NewKiroTokenRefresher(kiroOAuthService *KiroOAuthService) *KiroTokenRefresher {
	return &KiroTokenRefresher{
		kiroOAuthService: kiroOAuthService,
	}
}

// CacheKey 返回用于分布式锁的缓存键
func (r *KiroTokenRefresher) CacheKey(account *Account) string {
	return KiroTokenCacheKey(account)
}

// CanRefresh 检查是否能处理此账号
func (r *KiroTokenRefresher) CanRefresh(account *Account) bool {
	return account.Platform == PlatformKiro && account.Type == AccountTypeOAuth
}

// NeedsRefresh 检查 token 是否需要刷新
//
// 与 Antigravity 一致：固定 10 分钟窗口，忽略全局配置。
// AWS OIDC 默认 1h 过期，全局窗口（默认 5 分钟）对它太紧。
func (r *KiroTokenRefresher) NeedsRefresh(account *Account, _ time.Duration) bool {
	if !r.CanRefresh(account) {
		return false
	}
	expiresAt := account.GetCredentialAsTime("expires_at")
	if expiresAt == nil {
		return false
	}
	timeUntilExpiry := time.Until(*expiresAt)
	needs := timeUntilExpiry < kiroRefreshWindow
	if needs {
		fmt.Printf("[KiroTokenRefresher] Account %d needs refresh: expires_at=%s, time_until_expiry=%v, window=%v\n",
			account.ID, expiresAt.Format("2006-01-02 15:04:05"), timeUntilExpiry, kiroRefreshWindow)
	}
	return needs
}

// Refresh 执行 token 刷新；保留原有 credentials 中的字段
func (r *KiroTokenRefresher) Refresh(ctx context.Context, account *Account) (map[string]any, error) {
	tokenInfo, err := r.kiroOAuthService.RefreshAccountToken(ctx, account)
	if err != nil {
		return nil, err
	}
	newCreds := r.kiroOAuthService.BuildAccountCredentials(tokenInfo)
	return MergeCredentials(account.Credentials, newCreds), nil
}
