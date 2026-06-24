package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

// KiroImportRequest 是 kiro 账号批量导入请求。
// Data 支持两种格式（自动识别）：
//   - JSON 数组：[{"email","refresh_token","client_id","client_secret","region","login_type","machine_id"}...]
//   - Kiro manage 导出包：{"accounts":[{"email","machineId","credentials":{"refreshToken","clientId","clientSecret"...}}...]}
//   - 卡密文本：每行 邮箱----密码----RefreshToken----ClientId----ClientSecret，分隔符支持 ----、Tab、连续空格
type KiroImportRequest struct {
	Data                    string   `json:"data"`
	DefaultLoginType        string   `json:"default_login_type"`
	Region                  string   `json:"region"`
	GroupIDs                []int64  `json:"group_ids"`
	ProxyID                 *int64   `json:"proxy_id"`
	Concurrency             *int     `json:"concurrency"`
	Priority                *int     `json:"priority"`
	RateMultiplier          *float64 `json:"rate_multiplier"`
	ExpiresAt               *int64   `json:"expires_at"`
	AutoPauseOnExpired      *bool    `json:"auto_pause_on_expired"`
	SkipDefaultGroupBind    *bool    `json:"skip_default_group_bind"`
	ConfirmMixedChannelRisk *bool    `json:"confirm_mixed_channel_risk"`
}

// KiroImportItem 是单条导入凭据解析后的结构。
type KiroImportItem struct {
	Email        string `json:"email"`
	Password     string `json:"password"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	Region       string `json:"region"`
	LoginType    string `json:"login_type"`
	MachineID    string `json:"machine_id"`
	ProfileArn   string `json:"profile_arn"`
	Extra        map[string]any
}

type kiroImportEntry struct {
	Index int
	Item  KiroImportItem
}

// KiroImportResult 是批量导入的汇总结果。
type KiroImportResult struct {
	Total   int                      `json:"total"`
	Created int                      `json:"created"`
	Skipped int                      `json:"skipped"`
	Failed  int                      `json:"failed"`
	Errors  []KiroImportErrorMessage `json:"errors,omitempty"`
}

// KiroImportErrorMessage 描述单条导入失败/跳过的原因。
type KiroImportErrorMessage struct {
	Index   int    `json:"index"`
	Name    string `json:"name,omitempty"`
	Message string `json:"message"`
}

// ImportKiro 处理 kiro 账号批量导入。
// POST /api/v1/admin/accounts/import-kiro
func (h *AccountHandler) ImportKiro(c *gin.Context) {
	var req KiroImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	if req.Concurrency != nil && *req.Concurrency < 0 {
		response.BadRequest(c, "concurrency must be >= 0")
		return
	}
	if req.Priority != nil && *req.Priority < 0 {
		response.BadRequest(c, "priority must be >= 0")
		return
	}
	if req.RateMultiplier != nil && *req.RateMultiplier < 0 {
		response.BadRequest(c, "rate_multiplier must be >= 0")
		return
	}

	entries, err := parseKiroImportEntries(req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if len(entries) == 0 {
		response.BadRequest(c, "请输入 kiro 凭据（JSON 数组或卡密格式：邮箱----密码----RefreshToken----ClientId----ClientSecret）")
		return
	}

	executeAdminIdempotentJSON(c, "admin.accounts.import_kiro", req, service.DefaultWriteIdempotencyTTL(), func(ctx context.Context) (any, error) {
		return h.importKiroAccounts(ctx, req, entries)
	})
}

// parseKiroImportEntries 解析请求体为待导入条目，自动识别 JSON 与卡密格式。
func parseKiroImportEntries(req KiroImportRequest) ([]kiroImportEntry, error) {
	trimmed := strings.TrimSpace(req.Data)
	if trimmed == "" {
		return nil, nil
	}

	defaultLoginType := strings.ToLower(strings.TrimSpace(req.DefaultLoginType))
	if defaultLoginType == "" {
		defaultLoginType = "builder"
	}
	defaultRegion := strings.TrimSpace(req.Region)
	if defaultRegion == "" {
		defaultRegion = "us-east-1"
	}

	entries := make([]kiroImportEntry, 0)

	// 优先尝试 JSON（Sub2API 数组/单对象，或 Kiro manage 导出包）。
	var rawList []map[string]any
	if strings.HasPrefix(trimmed, "[") {
		if err := json.Unmarshal([]byte(trimmed), &rawList); err != nil {
			return nil, fmt.Errorf("JSON 解析失败: %v", err)
		}
	} else if strings.HasPrefix(trimmed, "{") {
		var single map[string]any
		if err := json.Unmarshal([]byte(trimmed), &single); err != nil {
			return nil, fmt.Errorf("JSON 解析失败: %v", err)
		}
		rawList = expandKiroImportJSONObjects(single)
	}

	if rawList != nil {
		for i, raw := range rawList {
			item := kiroItemFromJSON(raw, defaultLoginType, defaultRegion)
			if strings.TrimSpace(item.RefreshToken) == "" {
				continue
			}
			entries = append(entries, kiroImportEntry{Index: i, Item: item})
		}
		return entries, nil
	}

	// 卡密格式：逐行解析，跳过空行和 # 注释。
	lines := strings.Split(trimmed, "\n")
	idx := 0
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := splitKiroKamiLine(line)
		item := KiroImportItem{
			Email:        strings.TrimSpace(getPart(parts, 0)),
			RefreshToken: strings.TrimSpace(getPart(parts, 2)),
			ClientID:     strings.TrimSpace(getPart(parts, 3)),
			ClientSecret: strings.TrimSpace(getPart(parts, 4)),
			LoginType:    defaultLoginType,
			Region:       defaultRegion,
		}
		rawPwd := strings.TrimSpace(getPart(parts, 1))
		if rawPwd != "" && rawPwd != "no_password" {
			item.Password = rawPwd
		}
		if strings.TrimSpace(item.RefreshToken) == "" {
			continue
		}
		entries = append(entries, kiroImportEntry{Index: idx, Item: item})
		idx++
	}
	return entries, nil
}

// splitKiroKamiLine 按 ---- / Tab / 连续空格切分卡密行。
func splitKiroKamiLine(line string) []string {
	if strings.Contains(line, "----") {
		return strings.Split(line, "----")
	}
	if strings.Contains(line, "\t") {
		return strings.Split(line, "\t")
	}
	return strings.Fields(line)
}

func getPart(parts []string, i int) string {
	if i < len(parts) {
		return parts[i]
	}
	return ""
}

func expandKiroImportJSONObjects(raw map[string]any) []map[string]any {
	if accounts, ok := raw["accounts"]; ok {
		if list := mapSliceFromAny(accounts); len(list) > 0 {
			return list
		}
	}
	return []map[string]any{raw}
}

func mapSliceFromAny(value any) []map[string]any {
	switch v := value.(type) {
	case []any:
		list := make([]map[string]any, 0, len(v))
		for _, item := range v {
			if m, ok := item.(map[string]any); ok {
				list = append(list, m)
			}
		}
		return list
	case []map[string]any:
		return v
	case map[string]any:
		list := make([]map[string]any, 0, len(v))
		for _, item := range v {
			if m, ok := item.(map[string]any); ok {
				list = append(list, m)
			}
		}
		return list
	default:
		return nil
	}
}

func kiroItemFromJSON(raw map[string]any, defaultLoginType, defaultRegion string) KiroImportItem {
	credentials := mapFromAny(raw["credentials"])
	item := KiroImportItem{
		Email:        firstStringKey(raw, "email", "_email"),
		Password:     firstStringKey(raw, "password"),
		AccessToken:  firstStringKey(raw, "access_token", "accessToken"),
		RefreshToken: firstStringKey(raw, "refresh_token", "refreshToken"),
		ClientID:     firstStringKey(raw, "client_id", "clientId"),
		ClientSecret: firstStringKey(raw, "client_secret", "clientSecret"),
		Region:       firstStringKey(raw, "region"),
		LoginType:    firstStringKey(raw, "login_type", "loginType", "authMethod", "provider"),
		MachineID:    firstStringKey(raw, "machine_id", "machineId"),
		ProfileArn:   firstStringKey(raw, "profile_arn", "profileArn"),
	}
	if credentials != nil {
		item.AccessToken = firstKiroImportNonEmpty(item.AccessToken, firstStringKey(credentials, "access_token", "accessToken"))
		item.RefreshToken = firstKiroImportNonEmpty(item.RefreshToken, firstStringKey(credentials, "refresh_token", "refreshToken"))
		item.ClientID = firstKiroImportNonEmpty(item.ClientID, firstStringKey(credentials, "client_id", "clientId"))
		item.ClientSecret = firstKiroImportNonEmpty(item.ClientSecret, firstStringKey(credentials, "client_secret", "clientSecret"))
		item.Region = firstKiroImportNonEmpty(item.Region, firstStringKey(credentials, "region"))
		item.LoginType = firstKiroImportNonEmpty(item.LoginType, firstStringKey(credentials, "login_type", "loginType", "authMethod", "provider"))
		item.MachineID = firstKiroImportNonEmpty(item.MachineID, firstStringKey(credentials, "machine_id", "machineId"))
		item.ProfileArn = firstKiroImportNonEmpty(item.ProfileArn, firstStringKey(credentials, "profile_arn", "profileArn"))
	}
	item.Extra = buildKiroManageExtra(raw, credentials)
	item.LoginType = normalizeKiroLoginType(item.LoginType, defaultLoginType)
	if item.Region == "" {
		item.Region = defaultRegion
	}
	return item
}

func mapFromAny(value any) map[string]any {
	if m, ok := value.(map[string]any); ok {
		return m
	}
	return nil
}

func firstKiroImportNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

// normalizeKiroLoginType 把 kiro-account-manager 的 provider/authMethod 归一到 sub2api 的 login_type。
//   - github/google → 同名（social 刷新）
//   - social → github（按社交处理）
//   - builderid/enterprise/idc/builder → builder（OIDC 刷新）
func normalizeKiroLoginType(value, fallback string) string {
	v := strings.ToLower(strings.TrimSpace(value))
	switch v {
	case "github":
		return "github"
	case "google":
		return "google"
	case "social":
		return "github"
	case "idc":
		return "idc"
	case "builder", "builderid", "enterprise":
		return "builder"
	case "":
		if fallback == "" {
			return "builder"
		}
		return fallback
	default:
		return v
	}
}

func firstStringKey(raw map[string]any, keys ...string) string {
	for _, k := range keys {
		if v, ok := raw[k]; ok {
			if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
				return strings.TrimSpace(s)
			}
		}
	}
	return ""
}

func buildKiroManageExtra(raw, credentials map[string]any) map[string]any {
	if raw == nil {
		return nil
	}
	manage := make(map[string]any)
	copyStringExtra(manage, raw, "id", "id")
	copyStringExtra(manage, raw, "user_id", "userId")
	copyStringExtra(manage, raw, "idp", "idp")
	copyStringExtra(manage, raw, "status", "status")
	copyStringExtra(manage, raw, "tags", "tags")
	copyStringExtra(manage, raw, "created_at", "createdAt")
	copyStringExtra(manage, raw, "last_used_at", "lastUsedAt")
	copyStringExtra(manage, raw, "last_checked_at", "lastCheckedAt")
	copyStringExtra(manage, raw, "machine_id", "machineId")
	copyStringExtra(manage, credentials, "auth_method", "authMethod")
	copyStringExtra(manage, credentials, "provider", "provider")
	copyStringExtra(manage, credentials, "csrf_token", "csrfToken")
	copyStringExtra(manage, credentials, "credentials_expires_at", "expiresAt")
	copyObjectExtra(manage, raw, "subscription", "subscription")
	copyObjectExtra(manage, raw, "usage", "usage")
	if len(manage) == 0 {
		return nil
	}
	return map[string]any{
		"kiro_manage": manage,
	}
}

func copyStringExtra(dst map[string]any, src map[string]any, dstKey, srcKey string) {
	if src == nil {
		return
	}
	if value := firstStringKey(src, srcKey); value != "" {
		dst[dstKey] = value
	}
}

func copyObjectExtra(dst map[string]any, src map[string]any, dstKey, srcKey string) {
	if src == nil {
		return
	}
	if value, ok := src[srcKey]; ok && value != nil {
		dst[dstKey] = value
	}
}

func (h *AccountHandler) importKiroAccounts(ctx context.Context, req KiroImportRequest, entries []kiroImportEntry) (KiroImportResult, error) {
	result := KiroImportResult{
		Total:  len(entries),
		Errors: make([]KiroImportErrorMessage, 0),
	}

	skipCheck := req.ConfirmMixedChannelRisk != nil && *req.ConfirmMixedChannelRisk
	skipDefaultGroupBind := req.SkipDefaultGroupBind != nil && *req.SkipDefaultGroupBind

	// 预加载已存在的 kiro 账号用于去重（按 refresh_token）。
	existing, err := h.listAccountsFiltered(ctx, service.PlatformKiro, service.AccountTypeOAuth, "", "", 0, "", "created_at", "desc")
	if err != nil {
		return result, err
	}
	existingByRefresh := make(map[string]bool, len(existing))
	for i := range existing {
		rt := strings.TrimSpace(existing[i].GetCredential("refresh_token"))
		if rt != "" {
			existingByRefresh[rt] = true
		}
	}

	concurrency := 0
	if req.Concurrency != nil {
		concurrency = *req.Concurrency
	}
	priority := 0
	if req.Priority != nil {
		priority = *req.Priority
	}

	for _, entry := range entries {
		item := entry.Item
		displayName := item.Email
		if displayName == "" {
			displayName = fmt.Sprintf("kiro-%d", entry.Index+1)
		}

		if strings.TrimSpace(item.RefreshToken) == "" {
			result.Failed++
			result.Errors = append(result.Errors, KiroImportErrorMessage{Index: entry.Index, Name: displayName, Message: "缺少 refresh_token"})
			continue
		}

		// 去重：相同 refresh_token 视为已存在。
		if existingByRefresh[strings.TrimSpace(item.RefreshToken)] {
			result.Skipped++
			result.Errors = append(result.Errors, KiroImportErrorMessage{Index: entry.Index, Name: displayName, Message: "账号已存在（refresh_token 重复）"})
			continue
		}

		// 构造凭据并做条件校验（social 仅需 refresh_token；OIDC 需要 client_id/secret）。
		credentials := buildKiroImportCredentials(item)
		probe := &service.Account{
			Platform:    service.PlatformKiro,
			Type:        service.AccountTypeOAuth,
			Credentials: credentials,
		}
		if verr := probe.ValidateKiroOAuthCredentials(); verr != nil {
			result.Failed++
			result.Errors = append(result.Errors, KiroImportErrorMessage{Index: entry.Index, Name: displayName, Message: verr.Error()})
			continue
		}

		// 实时验证：用 refresh_token 刷新换取 access_token。
		tokenInfo, rerr := h.kiroOAuthService.RefreshAccountToken(ctx, probe)
		if rerr != nil {
			result.Failed++
			result.Errors = append(result.Errors, KiroImportErrorMessage{Index: entry.Index, Name: displayName, Message: "验证失败: " + rerr.Error()})
			continue
		}

		// 用刷新结果更新 token 字段，保留 client_id/secret/login_type/region/machine_id/profile_arn。
		credentials["access_token"] = tokenInfo.AccessToken
		credentials["expires_at"] = strconv.FormatInt(tokenInfo.ExpiresAt, 10)
		if strings.TrimSpace(tokenInfo.RefreshToken) != "" {
			credentials["refresh_token"] = tokenInfo.RefreshToken
		}

		account, cerr := h.adminService.CreateAccount(ctx, &service.CreateAccountInput{
			Name:                  displayName,
			Platform:              service.PlatformKiro,
			Type:                  service.AccountTypeOAuth,
			Credentials:           credentials,
			Extra:                 item.Extra,
			ProxyID:               req.ProxyID,
			Concurrency:           concurrency,
			Priority:              priority,
			RateMultiplier:        req.RateMultiplier,
			GroupIDs:              req.GroupIDs,
			ExpiresAt:             req.ExpiresAt,
			AutoPauseOnExpired:    req.AutoPauseOnExpired,
			SkipDefaultGroupBind:  skipDefaultGroupBind,
			SkipMixedChannelCheck: skipCheck,
		})
		if cerr != nil {
			result.Failed++
			result.Errors = append(result.Errors, KiroImportErrorMessage{Index: entry.Index, Name: displayName, Message: cerr.Error()})
			continue
		}

		existingByRefresh[strings.TrimSpace(item.RefreshToken)] = true
		_ = account
		result.Created++
	}

	return result, nil
}

// buildKiroImportCredentials 把导入条目转成账号 credentials map（仅写入非空字段）。
func buildKiroImportCredentials(item KiroImportItem) map[string]any {
	creds := map[string]any{
		"refresh_token": strings.TrimSpace(item.RefreshToken),
		"login_type":    strings.TrimSpace(item.LoginType),
	}
	if v := strings.TrimSpace(item.AccessToken); v != "" {
		creds["access_token"] = v
	}
	if v := strings.TrimSpace(item.ClientID); v != "" {
		creds["client_id"] = v
	}
	if v := strings.TrimSpace(item.ClientSecret); v != "" {
		creds["client_secret"] = v
	}
	if v := strings.TrimSpace(item.Region); v != "" {
		creds["region"] = v
	}
	if v := strings.TrimSpace(item.MachineID); v != "" {
		creds["machine_id"] = v
	}
	if v := strings.TrimSpace(item.ProfileArn); v != "" {
		creds["profile_arn"] = v
	}
	if v := strings.TrimSpace(item.Email); v != "" {
		creds["email"] = v
	}
	return creds
}
