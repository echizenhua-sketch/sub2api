package service

import (
	"strings"
)

// Kiro 错误码分级
//
// 参考 chaogei/Kiro-account-manager src/main/proxy/accountPool.ts:12-30
// 与 src/main/proxy/proxyServer.ts 的错误处理。
//
// sub2api 的 failover 由 GatewayService.shouldFailoverUpstreamError + handleErrorResponse
// 实现，本文件只负责打分类标签用于日志与未来更细粒度的退避策略。

// KiroErrorClass 错误分类
type KiroErrorClass int

const (
	// KiroErrUnknown 未知错误，不走 failover
	KiroErrUnknown KiroErrorClass = iota
	// KiroErrRecoverable 可恢复（402/403/429）→ 切号；429 触发 quotaExhaustedAt 标记
	KiroErrRecoverable
	// KiroErrAuthNeedsRefresh 401 → 刷 token；当前 sub2api 直接走标准 failover，
	// 由后台 TokenRefreshService 异步刷新
	KiroErrAuthNeedsRefresh
	// KiroErrFatal 致命错误（4xx 客户端错误，5xx 服务端错误）→ 不切号
	KiroErrFatal
)

// String 返回分类名称
func (c KiroErrorClass) String() string {
	switch c {
	case KiroErrRecoverable:
		return "recoverable"
	case KiroErrAuthNeedsRefresh:
		return "auth_needs_refresh"
	case KiroErrFatal:
		return "fatal"
	default:
		return "unknown"
	}
}

// classifyKiroError 根据 HTTP 状态码和响应体判断错误等级
//
// 决策表：
//   - 401             → AuthNeedsRefresh（先尝试刷 token）
//   - 402, 403, 429   → Recoverable（切号；429 标 quotaExhaustedAt）
//   - 400 (CONTENT_LENGTH_EXCEEDS_THRESHOLD) → Fatal（请求过大，切号无意义）
//   - 422             → Fatal（请求格式非法）
//   - 5xx             → Fatal（不切号；走 retry）
//   - 其他            → Unknown
func classifyKiroError(httpStatus int, body []byte) KiroErrorClass {
	switch httpStatus {
	case 401:
		return KiroErrAuthNeedsRefresh
	case 402, 403, 429:
		return KiroErrRecoverable
	case 422:
		return KiroErrFatal
	}

	if httpStatus == 400 && bytesContainsAny(body, "CONTENT_LENGTH_EXCEEDS_THRESHOLD", "PayloadTooLarge") {
		return KiroErrFatal
	}

	if httpStatus >= 500 && httpStatus < 600 {
		return KiroErrFatal
	}
	if httpStatus >= 400 && httpStatus < 500 {
		return KiroErrFatal
	}
	return KiroErrUnknown
}

// IsKiroQuotaExhausted 是否是配额耗尽信号（用于设置 quotaExhaustedAt）
func IsKiroQuotaExhausted(httpStatus int) bool {
	return httpStatus == 402 || httpStatus == 429
}

// bytesContainsAny 任一子串命中即返回 true
func bytesContainsAny(b []byte, candidates ...string) bool {
	s := string(b)
	for _, c := range candidates {
		if strings.Contains(s, c) {
			return true
		}
	}
	return false
}
