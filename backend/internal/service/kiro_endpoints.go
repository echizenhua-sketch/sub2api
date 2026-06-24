package service

// Kiro 上游 endpoints
//
// 参考 chaogei/Kiro-account-manager src/main/proxy/kiroApi.ts:72-93。
// 三个候选 endpoint 用同一个请求体格式（CLI 端点稍有差异），
// 通过 :amz-target 头切换上游服务路由。
//
// Phase 3b MVP：只使用 codewhisperer 单端点；多 endpoint 轮转留待 Phase 5/6 完善。

const (
	// KiroEndpointCodeWhisperer KiroIDE 默认走的端点。最稳，配额最大。
	KiroEndpointCodeWhisperer = "https://codewhisperer.us-east-1.amazonaws.com/generateAssistantResponse"

	// KiroEndpointAmazonQ codewhisperer 限流时的 fallback 端点。
	KiroEndpointAmazonQ = "https://q.us-east-1.amazonaws.com/generateAssistantResponse"

	// KiroEndpointAmazonQCLI Amazon Q CLI 专用端点。请求体需删 agentContinuationId / agentTaskType，
	// origin 改 "AmazonQ"。Phase 5+ 再启用。
	KiroEndpointAmazonQCLI = "https://q.us-east-1.amazonaws.com/SendMessageStreaming"

	// KiroAmzTargetGenerate AmazonCodeWhisperer/AmazonQ 的 :amz-target 头
	KiroAmzTargetGenerate = "AmazonCodeWhispererStreamingService.GenerateAssistantResponse"

	// KiroAmzTargetSendMessage AmazonQ-CLI 的 :amz-target 头
	KiroAmzTargetSendMessage = "AmazonQDeveloperStreamingService.SendMessage"
)

// KiroEndpoint 描述一个 Kiro 上游 endpoint
type KiroEndpoint struct {
	URL       string // 完整 URL
	AmzTarget string // :amz-target 头
	CLIMode   bool   // true 表示需要删 agentContinuationId / agentTaskType + origin=AmazonQ
}

// kiroDefaultEndpoints 用作端点轮转列表（Phase 5/6 时用，当前 forwardKiro 只用第一个）
var kiroDefaultEndpoints = []KiroEndpoint{
	{URL: KiroEndpointCodeWhisperer, AmzTarget: KiroAmzTargetGenerate, CLIMode: false},
	{URL: KiroEndpointAmazonQ, AmzTarget: KiroAmzTargetGenerate, CLIMode: false},
}

// kiroPrimaryEndpoint Phase 3b MVP 单 endpoint 入口
func kiroPrimaryEndpoint() KiroEndpoint {
	return kiroDefaultEndpoints[0]
}
