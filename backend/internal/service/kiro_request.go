package service

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// 构建发往 Kiro 上游的请求头与 machineId
//
// 参考 chaogei/Kiro-account-manager src/main/proxy/kiroApi.ts:973-1014。

// resolveKiroMachineID 优先用 credentials 里写明的；缺则 sha256("kiro-device-{accountId}") 做兜底
func resolveKiroMachineID(account *Account) string {
	if account == nil {
		return ""
	}
	if v := strings.TrimSpace(account.GetCredential("machine_id")); v != "" {
		return v
	}
	sum := sha256.Sum256([]byte(fmt.Sprintf("kiro-device-%d", account.ID)))
	return hex.EncodeToString(sum[:])
}

// buildKiroHeaders 构造一个仿 KiroIDE 0.12.155 的请求头集合
//
// account.login_type = idc 时使用 CLI/IDC 风格 user-agent，否则用 KiroIDE 风格。
//
// 出于 Phase 3b 的可控范围，这里走非 IDC 路径。Phase 6 端到端验证通过后再考虑 IDC。
func buildKiroHeaders(account *Account, accessToken string, ep KiroEndpoint) map[string]string {
	machineID := resolveKiroMachineID(account)
	mode := "spec"
	if strings.ToLower(strings.TrimSpace(account.GetCredential("login_type"))) == "idc" {
		mode = "vibe"
	}
	return map[string]string{
		"Content-Type":           "application/json",
		"Accept":                 "application/vnd.amazon.eventstream",
		"Authorization":          "Bearer " + accessToken,
		"x-amzn-kiro-agent-mode": mode,
		"x-amz-target":           ep.AmzTarget,
		"x-amz-user-agent":       fmt.Sprintf("aws-sdk-js/%s KiroIDE %s %s", kiroAwsSdkVersionUA, kiroVersion, machineID),
		"User-Agent":             kiroUserAgent(machineID),
		"amz-sdk-invocation-id":  uuid.NewString(),
		"amz-sdk-request":        "attempt=1; max=3",
	}
}
