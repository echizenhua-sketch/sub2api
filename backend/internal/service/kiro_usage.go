package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Kiro 用量查询 REST API
//
// 参考 chaogei/Kiro-account-manager src/main/index.ts:973-1200 (getUsageLimitsRest)。
// GET https://q.{region}.amazonaws.com/getUsageLimits?origin=AI_EDITOR&resourceType=AGENTIC_REQUEST&isEmailRequired=true&profileArn={arn}
// 主端点 403 时 fallback 到另一区域端点。

var kiroUsageRestEndpoints = map[string]string{
	"us-east-1":    "https://q.us-east-1.amazonaws.com",
	"eu-central-1": "https://q.eu-central-1.amazonaws.com",
}

func kiroUsageRestBase(region string) string {
	r := strings.TrimSpace(region)
	if r == "" {
		return kiroUsageRestEndpoints["us-east-1"]
	}
	if v, ok := kiroUsageRestEndpoints[r]; ok {
		return v
	}
	if strings.HasPrefix(r, "eu-") {
		return kiroUsageRestEndpoints["eu-central-1"]
	}
	return kiroUsageRestEndpoints["us-east-1"]
}

func kiroUsageRestFallbackBase(region string) string {
	if kiroUsageRestBase(region) == kiroUsageRestEndpoints["eu-central-1"] {
		return kiroUsageRestEndpoints["us-east-1"]
	}
	return kiroUsageRestEndpoints["eu-central-1"]
}

// KiroUsageBreakdown 单个资源类型的用量明细。
type KiroUsageBreakdown struct {
	Type         string  `json:"type"`
	ResourceType string  `json:"resourceType"`
	DisplayName  string  `json:"displayName"`
	CurrentUsage float64 `json:"currentUsage"`
	UsageLimit   float64 `json:"usageLimit"`
	Unit         string  `json:"unit"`
	OverageCap   float64 `json:"overageCap"`
}

// KiroUsageLimitsResponse 是 getUsageLimits 的响应（仅取需要的字段）。
type KiroUsageLimitsResponse struct {
	UsageBreakdownList []KiroUsageBreakdown `json:"usageBreakdownList"`
	NextDateReset      json.RawMessage      `json:"nextDateReset"`
	SubscriptionInfo   *struct {
		SubscriptionName  string `json:"subscriptionName"`
		SubscriptionTitle string `json:"subscriptionTitle"`
		SubscriptionType  string `json:"subscriptionType"`
		Status            string `json:"status"`
	} `json:"subscriptionInfo"`
	UserInfo *struct {
		Email  string `json:"email"`
		UserID string `json:"userId"`
	} `json:"userInfo"`
}

// KiroUsageInfo 是归一化后的 kiro 用量信息。
type KiroUsageInfo struct {
	CurrentUsage     float64
	UsageLimit       float64
	Utilization      float64
	ResetsAt         *time.Time
	SubscriptionType string
	Email            string
}

// FetchUsageLimits 调用 getUsageLimits 获取 kiro 账号用量。
func (s *KiroOAuthService) FetchUsageLimits(ctx context.Context, account *Account) (*KiroUsageInfo, error) {
	if account == nil || !account.IsKiro() {
		return nil, fmt.Errorf("kiro usage: account is not kiro")
	}
	accessToken := strings.TrimSpace(account.GetCredential("access_token"))
	if accessToken == "" {
		return nil, fmt.Errorf("kiro usage: account has no access_token")
	}

	region := strings.TrimSpace(account.GetCredential("region"))
	machineID := resolveKiroMachineID(account)
	profileArn := resolveKiroProfileArn(account)

	params := url.Values{}
	params.Set("origin", "AI_EDITOR")
	params.Set("resourceType", "AGENTIC_REQUEST")
	params.Set("isEmailRequired", "true")
	if profileArn != "" {
		params.Set("profileArn", profileArn)
	}
	path := "/getUsageLimits?" + params.Encode()

	resp, err := s.doKiroUsageGet(ctx, kiroUsageRestBase(region)+path, accessToken, machineID)
	if err == nil && resp.statusCode == http.StatusForbidden {
		// 主端点 403，尝试 fallback 区域。
		resp, err = s.doKiroUsageGet(ctx, kiroUsageRestFallbackBase(region)+path, accessToken, machineID)
	}
	if err != nil {
		return nil, err
	}
	if resp.statusCode < 200 || resp.statusCode >= 300 {
		return nil, fmt.Errorf("kiro usage: http %d: %s", resp.statusCode, string(resp.body))
	}

	var parsed KiroUsageLimitsResponse
	if err := json.Unmarshal(resp.body, &parsed); err != nil {
		return nil, fmt.Errorf("kiro usage: parse response: %w (body=%s)", err, string(resp.body))
	}

	return buildKiroUsageInfo(&parsed), nil
}

type kiroUsageHTTPResult struct {
	statusCode int
	body       []byte
}

func (s *KiroOAuthService) doKiroUsageGet(ctx context.Context, fullURL, accessToken, machineID string) (*kiroUsageHTTPResult, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("kiro usage: build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("User-Agent", kiroUserAgent(machineID))
	req.Header.Set("x-amz-user-agent", fmt.Sprintf("aws-sdk-js/%s KiroIDE %s %s", kiroAwsSdkVersionUA, kiroVersion, machineID))

	httpResp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("kiro usage: do request: %w", err)
	}
	defer httpResp.Body.Close()
	body, _ := io.ReadAll(httpResp.Body)
	return &kiroUsageHTTPResult{statusCode: httpResp.StatusCode, body: body}, nil
}

// buildKiroUsageInfo 从响应归一化用量。优先取 AGENTIC_REQUEST 资源；缺则取第一个。
func buildKiroUsageInfo(resp *KiroUsageLimitsResponse) *KiroUsageInfo {
	info := &KiroUsageInfo{}
	if resp.SubscriptionInfo != nil {
		info.SubscriptionType = strings.TrimSpace(resp.SubscriptionInfo.SubscriptionType)
		if info.SubscriptionType == "" {
			info.SubscriptionType = strings.TrimSpace(resp.SubscriptionInfo.SubscriptionTitle)
		}
	}
	if resp.UserInfo != nil {
		info.Email = strings.TrimSpace(resp.UserInfo.Email)
	}

	var chosen *KiroUsageBreakdown
	for i := range resp.UsageBreakdownList {
		b := &resp.UsageBreakdownList[i]
		rt := strings.ToUpper(strings.TrimSpace(b.ResourceType))
		if rt == "" {
			rt = strings.ToUpper(strings.TrimSpace(b.Type))
		}
		if rt == "AGENTIC_REQUEST" {
			chosen = b
			break
		}
		if chosen == nil {
			chosen = b
		}
	}
	if chosen != nil {
		info.CurrentUsage = chosen.CurrentUsage
		info.UsageLimit = chosen.UsageLimit
		if chosen.UsageLimit > 0 {
			info.Utilization = chosen.CurrentUsage / chosen.UsageLimit * 100
		}
	}

	info.ResetsAt = parseKiroResetDate(resp.NextDateReset)
	return info
}

// parseKiroResetDate 解析 nextDateReset，支持 Unix 秒时间戳（数字）或 ISO 字符串。
func parseKiroResetDate(raw json.RawMessage) *time.Time {
	if len(raw) == 0 {
		return nil
	}
	// 尝试数字（Unix 秒）。
	var num float64
	if err := json.Unmarshal(raw, &num); err == nil && num > 0 {
		t := time.Unix(int64(num), 0)
		return &t
	}
	// 尝试字符串。
	var str string
	if err := json.Unmarshal(raw, &str); err == nil && strings.TrimSpace(str) != "" {
		if t, perr := time.Parse(time.RFC3339, str); perr == nil {
			return &t
		}
	}
	return nil
}
