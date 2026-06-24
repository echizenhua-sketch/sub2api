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
	// forceInterval 固定周期强制刷新间隔；0 表示关闭固定周期，沿用过期前窗口逻辑。
	forceInterval time.Duration
}

// NewKiroTokenRefresher 创建 Kiro token 刷新器
//
// forceInterval > 0 时启用固定周期强制刷新：每隔该时间无条件刷新一次 token，
// 不看过期时间（但仍保留过期前 kiroRefreshWindow 兜底刷新）。
// forceInterval == 0 时维持原"过期前 10 分钟提前刷新"逻辑。
func NewKiroTokenRefresher(kiroOAuthService *KiroOAuthService, forceInterval time.Duration) *KiroTokenRefresher {
	return &KiroTokenRefresher{
		kiroOAuthService: kiroOAuthService,
		forceInterval:    forceInterval,
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
// 两种模式：
//   - forceInterval == 0（默认）：过期前 kiroRefreshWindow（10 分钟）刷新。
//   - forceInterval > 0：固定周期强制刷新——距上次刷新（last_refreshed_at）满 forceInterval
//     即刷新；同时保留过期前 kiroRefreshWindow 兜底，避免周期未到但 token 已临近过期。
//
// 第二个参数（全局 refreshWindow）对 kiro 不适用，忽略。
func (r *KiroTokenRefresher) NeedsRefresh(account *Account, _ time.Duration) bool {
	if !r.CanRefresh(account) {
		return false
	}

	// 过期前兜底：无论哪种模式，token 临近过期都要刷。
	expiresAt := account.GetCredentialAsTime("expires_at")
	if expiresAt != nil && time.Until(*expiresAt) < kiroRefreshWindow {
		fmt.Printf("[KiroTokenRefresher] Account %d needs refresh (expiry window): expires_at=%s, time_until_expiry=%v\n",
			account.ID, expiresAt.Format("2006-01-02 15:04:05"), time.Until(*expiresAt))
		return true
	}

	// 固定周期强制刷新模式。
	if r.forceInterval > 0 {
		lastRefreshed := account.GetCredentialAsTime("last_refreshed_at")
		if lastRefreshed == nil {
			// 从未记录刷新时间（老账号或刚导入）：立即刷新一次以建立基线。
			fmt.Printf("[KiroTokenRefresher] Account %d needs refresh (force interval, no baseline)\n", account.ID)
			return true
		}
		sinceLast := time.Since(*lastRefreshed)
		if sinceLast >= r.forceInterval {
			fmt.Printf("[KiroTokenRefresher] Account %d needs refresh (force interval): since_last=%v interval=%v\n",
				account.ID, sinceLast, r.forceInterval)
			return true
		}
		return false
	}

	// 默认模式：仅过期前窗口刷新（上面已判过期，这里返回 false）。
	if expiresAt == nil {
		return false
	}
	return false
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
