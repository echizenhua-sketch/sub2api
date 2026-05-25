package service

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// Anthropic Messages API → Kiro CodeWhisperer payload 翻译
//
// 上游契约参考 chaogei/Kiro-account-manager src/main/proxy/kiroApi.ts:743-918
// 与 src/main/proxy/translator.ts:713-942。
//
// Phase 3b MVP 范围：
//   - 完整 payload 字段树（conversationState / currentMessage / history /
//     userInputMessage / assistantResponseMessage / userInputMessageContext）
//   - tools 仅放 currentMessage
//   - tool_use / tool_result 配对
//   - sanitizeConversation 实装 user 起 / user 收 / 严格交替
//
// 留待 Phase 5 完善：
//   - 7 步完整 sanitize（孤儿 toolResult、缺失 toolResult 占位）
//   - 工具名 hash 截断（>64 字符）
//   - prompt cache fingerprint 与 cachePoint 注入
//   - thinking / reasoningContent 块在 history 中的丢弃逻辑

// KiroPayload 是发往 generateAssistantResponse / SendMessageStreaming 的请求体
type KiroPayload struct {
	ConversationState              KiroConversationState  `json:"conversationState"`
	ProfileArn                     string                 `json:"profileArn,omitempty"`
	InferenceConfig                *KiroInferenceConfig   `json:"inferenceConfig,omitempty"`
	AdditionalModelRequestFields   map[string]any         `json:"additionalModelRequestFields,omitempty"`
}

type KiroConversationState struct {
	AgentContinuationId string             `json:"agentContinuationId,omitempty"`
	AgentTaskType       string             `json:"agentTaskType,omitempty"`
	ChatTriggerType     string             `json:"chatTriggerType"`
	ConversationID      string             `json:"conversationId"`
	CurrentMessage      KiroCurrentMessage `json:"currentMessage"`
	History             []KiroHistoryItem  `json:"history,omitempty"`
}

type KiroCurrentMessage struct {
	UserInputMessage KiroUserInputMessage `json:"userInputMessage"`
}

type KiroUserInputMessage struct {
	Content                  string                          `json:"content"`
	ModelID                  string                          `json:"modelId"`
	Origin                   string                          `json:"origin"`
	Images                   []KiroImage                     `json:"images,omitempty"`
	UserInputMessageContext  *KiroUserInputMessageContext    `json:"userInputMessageContext,omitempty"`
}

type KiroImage struct {
	Format string         `json:"format"`
	Source KiroImageSource `json:"source"`
}

type KiroImageSource struct {
	Bytes string `json:"bytes"`
}

type KiroUserInputMessageContext struct {
	Tools       []KiroToolWrapper `json:"tools,omitempty"`
	ToolResults []KiroToolResult  `json:"toolResults,omitempty"`
}

// KiroToolWrapper 包一层 toolSpecification 是 Kiro 的硬要求
type KiroToolWrapper struct {
	ToolSpecification KiroToolSpec `json:"toolSpecification"`
}

type KiroToolSpec struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	InputSchema map[string]any `json:"inputSchema"`
}

type KiroToolResult struct {
	ToolUseID string                 `json:"toolUseId"`
	Content   []KiroToolResultBlock  `json:"content"`
	Status    string                 `json:"status,omitempty"` // success / error
}

type KiroToolResultBlock struct {
	Text string `json:"text"`
}

// KiroHistoryItem history 数组里的元素：要么 user 要么 assistant
type KiroHistoryItem struct {
	UserInputMessage          *KiroUserInputMessage          `json:"userInputMessage,omitempty"`
	AssistantResponseMessage  *KiroAssistantResponseMessage  `json:"assistantResponseMessage,omitempty"`
}

type KiroAssistantResponseMessage struct {
	Content  string         `json:"content"`
	ToolUses []KiroToolUse  `json:"toolUses,omitempty"`
}

type KiroToolUse struct {
	ToolUseID string         `json:"toolUseId"`
	Name      string         `json:"name"`
	Input     map[string]any `json:"input"`
}

type KiroInferenceConfig struct {
	MaxTokens   int     `json:"maxTokens,omitempty"`
	Temperature float64 `json:"temperature,omitempty"`
	TopP        float64 `json:"topP,omitempty"`
}

// kiroAnthropicRequest Anthropic Messages API 请求体（只解析需要的字段）
type kiroAnthropicRequest struct {
	Model       string             `json:"model"`
	System      json.RawMessage    `json:"system,omitempty"`
	Messages    []json.RawMessage  `json:"messages"`
	Tools       []json.RawMessage  `json:"tools,omitempty"`
	MaxTokens   int                `json:"max_tokens,omitempty"`
	Temperature *float64           `json:"temperature,omitempty"`
	TopP        *float64           `json:"top_p,omitempty"`
	Thinking    json.RawMessage    `json:"thinking,omitempty"`
	Stream      bool               `json:"stream,omitempty"`
}

// kiroAnthropicMessage Anthropic message 结构
type kiroAnthropicMessage struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
}

// kiroAnthropicContentBlock content 数组里的单个 block
type kiroAnthropicContentBlock struct {
	Type      string          `json:"type"`
	Text      string          `json:"text,omitempty"`
	ID        string          `json:"id,omitempty"`
	Name      string          `json:"name,omitempty"`
	Input     json.RawMessage `json:"input,omitempty"`
	ToolUseID string          `json:"tool_use_id,omitempty"`
	Content   json.RawMessage `json:"content,omitempty"`
	Source    *struct {
		Type      string `json:"type"`
		MediaType string `json:"media_type"`
		Data      string `json:"data"`
	} `json:"source,omitempty"`
}

// AnthropicToKiro 把 Anthropic Messages API 请求体翻译成 Kiro payload。
//
// account 用于：
//   - resolveProfileArn (Builder/Github/Google ARN 选择)
//   - 模型 ID 透传（kiro 上游不映射模型，模型名直接用 Anthropic 标准名）
//
// 流程：
//  1. 解析 Anthropic body
//  2. system 拼成 history 头部 user 消息 + 固定 assistant 回复
//  3. 遍历 messages：tool_use → assistantResponseMessage.toolUses；
//     tool_result → user 消息的 toolResults；text → content
//  4. 最后一条 user 消息变成 currentMessage（带 tools）
//  5. sanitizeConversation 整理 history
//
// 注：为了让 Kiro 上游接受工具名，所有 tool_use.name / tools[].name 都通过同一个
// KiroToolNameRegistry 编码；客户端可在 forwardKiro 内部用 registry.Decode 还原。
func AnthropicToKiro(body []byte, account *Account, ep KiroEndpoint) (*KiroPayload, error) {
	payload, _, err := AnthropicToKiroWithRegistry(body, account, ep)
	return payload, err
}

// AnthropicToKiroWithRegistry 同 AnthropicToKiro，但同时返回工具名注册表，
// 供响应解析阶段还原 toolUse.name 给客户端。
func AnthropicToKiroWithRegistry(body []byte, account *Account, ep KiroEndpoint) (*KiroPayload, *KiroToolNameRegistry, error) {
	registry := NewKiroToolNameRegistry()

	var req kiroAnthropicRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, nil, fmt.Errorf("kiro translator: parse body: %w", err)
	}

	// 1. 把 system 提到 history 头部
	systemText := flattenSystemField(req.System)
	historyItems := []KiroHistoryItem{}
	if systemText != "" {
		historyItems = append(historyItems,
			KiroHistoryItem{UserInputMessage: &KiroUserInputMessage{
				Content: systemText,
				ModelID: req.Model,
				Origin:  kiroOriginForEndpoint(ep),
			}},
			KiroHistoryItem{AssistantResponseMessage: &KiroAssistantResponseMessage{
				Content: "I will follow these instructions.",
			}},
		)
	}

	// 2. 转换 messages
	convertedMsgs := make([]KiroHistoryItem, 0, len(req.Messages))
	for _, raw := range req.Messages {
		var msg kiroAnthropicMessage
		if err := json.Unmarshal(raw, &msg); err != nil {
			return nil, nil, fmt.Errorf("kiro translator: parse message: %w", err)
		}
		blocks, err := parseContentBlocks(msg.Content)
		if err != nil {
			return nil, nil, err
		}

		switch msg.Role {
		case "user":
			user, err := userBlocksToKiro(blocks, req.Model, ep)
			if err != nil {
				return nil, nil, err
			}
			convertedMsgs = append(convertedMsgs, KiroHistoryItem{UserInputMessage: user})
		case "assistant":
			assistant := assistantBlocksToKiro(blocks, registry)
			convertedMsgs = append(convertedMsgs, KiroHistoryItem{AssistantResponseMessage: assistant})
		}
	}

	// 3. sanitize 合并后的 history
	allMsgs := append(historyItems, convertedMsgs...)
	allMsgs = sanitizeKiroConversation(allMsgs, req.Model, ep)

	if len(allMsgs) == 0 {
		return nil, nil, fmt.Errorf("kiro translator: no messages")
	}

	// 4. 最后一条 user 拆出来做 currentMessage（必为 user）
	last := allMsgs[len(allMsgs)-1]
	if last.UserInputMessage == nil {
		// sanitize 应该保证最后一条是 user。兜底：追加一条 Continue
		allMsgs = append(allMsgs, KiroHistoryItem{UserInputMessage: &KiroUserInputMessage{
			Content: "Continue",
			ModelID: req.Model,
			Origin:  kiroOriginForEndpoint(ep),
		}})
		last = allMsgs[len(allMsgs)-1]
	}
	current := *last.UserInputMessage
	history := allMsgs[:len(allMsgs)-1]

	// 5. tools 只放 currentMessage
	if len(req.Tools) > 0 {
		tools, err := toolsToKiro(req.Tools, registry)
		if err != nil {
			return nil, nil, err
		}
		if current.UserInputMessageContext == nil {
			current.UserInputMessageContext = &KiroUserInputMessageContext{}
		}
		current.UserInputMessageContext.Tools = tools
	}

	state := KiroConversationState{
		ChatTriggerType: "MANUAL",
		ConversationID:  kiroConversationID(history),
		CurrentMessage:  KiroCurrentMessage{UserInputMessage: current},
		History:         history,
	}
	if !ep.CLIMode {
		state.AgentContinuationId = uuid.NewString()
		state.AgentTaskType = "vibe"
	}

	payload := &KiroPayload{
		ConversationState: state,
		ProfileArn:        resolveKiroProfileArn(account),
	}

	if req.MaxTokens > 0 || req.Temperature != nil || req.TopP != nil {
		payload.InferenceConfig = &KiroInferenceConfig{MaxTokens: req.MaxTokens}
		if req.Temperature != nil {
			payload.InferenceConfig.Temperature = *req.Temperature
		}
		if req.TopP != nil {
			payload.InferenceConfig.TopP = *req.TopP
		}
	}

	// thinking 透传：仅 Claude 4+
	if len(req.Thinking) > 0 && string(req.Thinking) != "null" {
		payload.AdditionalModelRequestFields = map[string]any{
			"thinking": map[string]any{"type": "adaptive"},
		}
	}

	return payload, registry, nil
}

// flattenSystemField Anthropic 的 system 字段可以是 string 或 [{type:"text",text:"..."}]
func flattenSystemField(raw json.RawMessage) string {
	if len(raw) == 0 || string(raw) == "null" {
		return ""
	}
	var asString string
	if err := json.Unmarshal(raw, &asString); err == nil {
		return asString
	}
	var blocks []kiroAnthropicContentBlock
	if err := json.Unmarshal(raw, &blocks); err == nil {
		var sb strings.Builder
		for _, b := range blocks {
			if b.Type == "text" {
				if sb.Len() > 0 {
					sb.WriteString("\n\n")
				}
				sb.WriteString(b.Text)
			}
		}
		return sb.String()
	}
	return ""
}

// parseContentBlocks Anthropic 的 message.content 可以是 string 或 block 数组
func parseContentBlocks(raw json.RawMessage) ([]kiroAnthropicContentBlock, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}
	var asString string
	if err := json.Unmarshal(raw, &asString); err == nil {
		return []kiroAnthropicContentBlock{{Type: "text", Text: asString}}, nil
	}
	var blocks []kiroAnthropicContentBlock
	if err := json.Unmarshal(raw, &blocks); err != nil {
		return nil, fmt.Errorf("parse content blocks: %w", err)
	}
	return blocks, nil
}

// userBlocksToKiro user 角色的 blocks → Kiro user 消息
//
// text → content；image → images；tool_result → toolResults
func userBlocksToKiro(blocks []kiroAnthropicContentBlock, model string, ep KiroEndpoint) (*KiroUserInputMessage, error) {
	msg := &KiroUserInputMessage{
		ModelID: model,
		Origin:  kiroOriginForEndpoint(ep),
	}
	var textParts []string
	var images []KiroImage
	var toolResults []KiroToolResult

	for _, b := range blocks {
		switch b.Type {
		case "text":
			if b.Text != "" {
				textParts = append(textParts, b.Text)
			}
		case "image":
			if b.Source != nil && b.Source.Type == "base64" {
				images = append(images, KiroImage{
					Format: extractImageFormat(b.Source.MediaType),
					Source: KiroImageSource{Bytes: b.Source.Data},
				})
			}
		case "tool_result":
			tr := KiroToolResult{ToolUseID: b.ToolUseID, Status: "success"}
			// content 可以是 string 或 block 数组
			if len(b.Content) > 0 {
				var asString string
				if err := json.Unmarshal(b.Content, &asString); err == nil {
					tr.Content = []KiroToolResultBlock{{Text: asString}}
				} else {
					var inner []kiroAnthropicContentBlock
					if err := json.Unmarshal(b.Content, &inner); err == nil {
						for _, ib := range inner {
							if ib.Type == "text" {
								tr.Content = append(tr.Content, KiroToolResultBlock{Text: ib.Text})
							}
						}
					}
				}
			}
			if len(tr.Content) == 0 {
				tr.Content = []KiroToolResultBlock{{Text: ""}}
			}
			toolResults = append(toolResults, tr)
		}
	}

	msg.Content = strings.Join(textParts, "\n\n")
	if len(images) > 0 {
		msg.Images = images
	}
	if len(toolResults) > 0 {
		msg.UserInputMessageContext = &KiroUserInputMessageContext{ToolResults: toolResults}
	}
	return msg, nil
}

// assistantBlocksToKiro assistant 角色的 blocks → Kiro assistant 消息
//
// text → content；tool_use → toolUses；thinking/redacted_thinking 在 history 中丢弃。
// tool_use.name 经 registry 编码以保证 ≤64 字符 + 合法字符集。
func assistantBlocksToKiro(blocks []kiroAnthropicContentBlock, registry *KiroToolNameRegistry) *KiroAssistantResponseMessage {
	var textParts []string
	var toolUses []KiroToolUse

	for _, b := range blocks {
		switch b.Type {
		case "text":
			if b.Text != "" {
				textParts = append(textParts, b.Text)
			}
		case "tool_use":
			tu := KiroToolUse{ToolUseID: b.ID, Name: registry.Encode(b.Name)}
			if len(b.Input) > 0 {
				_ = json.Unmarshal(b.Input, &tu.Input)
			}
			if tu.Input == nil {
				tu.Input = map[string]any{}
			}
			toolUses = append(toolUses, tu)
		// thinking / redacted_thinking 在 history 中丢弃（kiro 上游禁止传入）
		}
	}

	msg := &KiroAssistantResponseMessage{Content: strings.Join(textParts, "\n\n")}
	if msg.Content == "" && len(toolUses) > 0 {
		// 纯 tool_use 时 content 不能为空字符串
		msg.Content = " "
	}
	if len(toolUses) > 0 {
		msg.ToolUses = toolUses
	}
	return msg
}

// toolsToKiro Anthropic tools[] → Kiro userInputMessageContext.tools
//
// 工具名 >64 字符或带非法字符时，通过 registry 编码为可逆缩写；description >10237 截断。
// registry 也用于把 history 中既有的 tool_use.name 编码（保证一致性）。
func toolsToKiro(rawTools []json.RawMessage, registry *KiroToolNameRegistry) ([]KiroToolWrapper, error) {
	out := make([]KiroToolWrapper, 0, len(rawTools))
	for _, raw := range rawTools {
		var tool struct {
			Name        string         `json:"name"`
			Description string         `json:"description"`
			InputSchema map[string]any `json:"input_schema"`
		}
		if err := json.Unmarshal(raw, &tool); err != nil {
			return nil, fmt.Errorf("kiro translator: parse tool: %w", err)
		}
		spec := KiroToolSpec{
			Name:        registry.Encode(tool.Name),
			Description: TruncateKiroToolDescription(tool.Description),
			InputSchema: tool.InputSchema,
		}
		if spec.InputSchema == nil {
			spec.InputSchema = map[string]any{"type": "object"}
		}
		out = append(out, KiroToolWrapper{ToolSpecification: spec})
	}
	return out, nil
}

// sanitizeKiroConversation 实装完整的 7 步 sanitize 规则（参考 kiroApi.ts:725-739）：
//
//  1. 必须以 user 起（缺则插 Hello 占位）
//  2. 移除空 user 消息（content 为空且无 toolResults），保留首条
//  3. toolResult 紧跟对应的 toolUse assistant 消息
//  4. 移除孤儿 toolResult（前一条不是含 toolUse 的 assistant 或 toolUseId 不匹配）
//  5. 给每个含 toolUse 的 assistant 消息后补缺失的 toolResult 占位
//  6. 严格交替 user/assistant（连续同 role 间插对方占位）
//  7. 必须以 user 收（缺则追加 Continue 占位）
func sanitizeKiroConversation(msgs []KiroHistoryItem, model string, ep KiroEndpoint) []KiroHistoryItem {
	if len(msgs) == 0 {
		return msgs
	}

	// Step 1: 必须以 user 起
	if msgs[0].UserInputMessage == nil {
		hello := KiroHistoryItem{UserInputMessage: &KiroUserInputMessage{
			Content: "Hello",
			ModelID: model,
			Origin:  kiroOriginForEndpoint(ep),
		}}
		msgs = append([]KiroHistoryItem{hello}, msgs...)
	}

	// Step 2: 移除空 user 消息（保留首条）
	cleaned := make([]KiroHistoryItem, 0, len(msgs))
	for i, m := range msgs {
		if m.UserInputMessage != nil {
			isEmpty := strings.TrimSpace(m.UserInputMessage.Content) == "" &&
				(m.UserInputMessage.UserInputMessageContext == nil ||
					(len(m.UserInputMessage.UserInputMessageContext.ToolResults) == 0 &&
						len(m.UserInputMessage.UserInputMessageContext.Tools) == 0)) &&
				len(m.UserInputMessage.Images) == 0
			if isEmpty && i != 0 {
				continue
			}
		}
		cleaned = append(cleaned, m)
	}
	msgs = cleaned

	// Step 3 + 4 + 5: toolResult 配对
	msgs = repairKiroToolPairing(msgs, model, ep)

	// Step 6: 严格交替
	out := make([]KiroHistoryItem, 0, len(msgs))
	for i, m := range msgs {
		out = append(out, m)
		if i+1 >= len(msgs) {
			break
		}
		next := msgs[i+1]
		curIsUser := m.UserInputMessage != nil
		nextIsUser := next.UserInputMessage != nil
		if curIsUser && nextIsUser {
			out = append(out, KiroHistoryItem{AssistantResponseMessage: &KiroAssistantResponseMessage{
				Content: "Understood.",
			}})
		} else if !curIsUser && !nextIsUser {
			out = append(out, KiroHistoryItem{UserInputMessage: &KiroUserInputMessage{
				Content: "Continue",
				ModelID: model,
				Origin:  kiroOriginForEndpoint(ep),
			}})
		}
	}

	// Step 7: 必须以 user 收
	if out[len(out)-1].UserInputMessage == nil {
		out = append(out, KiroHistoryItem{UserInputMessage: &KiroUserInputMessage{
			Content: "Continue",
			ModelID: model,
			Origin:  kiroOriginForEndpoint(ep),
		}})
	}

	return out
}

// repairKiroToolPairing 完成 sanitize 的 step 3/4/5
//
// 算法：从前往后扫，记录上一条 assistant 消息的 toolUse 集合（pendingToolIDs）。
// 遇到 user 消息时：
//   - 把它的 toolResults 中 toolUseId 在 pendingToolIDs 里的那条留下
//   - 不在的视为孤儿，丢弃
//   - 处理完后，pendingToolIDs 中没收到对应 toolResult 的 ID，统一插一条
//     status=error 的占位 toolResult
//
// assistant 消息时：刷新 pendingToolIDs 为本消息的 toolUse 集合。
func repairKiroToolPairing(msgs []KiroHistoryItem, model string, ep KiroEndpoint) []KiroHistoryItem {
	out := make([]KiroHistoryItem, 0, len(msgs))
	var pending []string // 上一条 assistant 留下的待处理 toolUseId

	flushMissingToolResults := func() {
		if len(pending) == 0 {
			return
		}
		// 给每个未配对的 toolUse 加一个占位 user 消息
		ctx := &KiroUserInputMessageContext{}
		for _, id := range pending {
			ctx.ToolResults = append(ctx.ToolResults, KiroToolResult{
				ToolUseID: id,
				Content:   []KiroToolResultBlock{{Text: "Tool execution failed"}},
				Status:    "error",
			})
		}
		out = append(out, KiroHistoryItem{UserInputMessage: &KiroUserInputMessage{
			ModelID:                 model,
			Origin:                  kiroOriginForEndpoint(ep),
			UserInputMessageContext: ctx,
		}})
		pending = nil
	}

	for _, m := range msgs {
		switch {
		case m.AssistantResponseMessage != nil:
			// 上一条 assistant 留下的 pending 未处理 → 兜底占位
			flushMissingToolResults()
			out = append(out, m)
			pending = nil
			for _, tu := range m.AssistantResponseMessage.ToolUses {
				pending = append(pending, tu.ToolUseID)
			}
		case m.UserInputMessage != nil:
			// 过滤掉孤儿 toolResult；保留 toolUseId 在 pending 里的
			ctx := m.UserInputMessage.UserInputMessageContext
			if ctx != nil && len(ctx.ToolResults) > 0 {
				kept := make([]KiroToolResult, 0, len(ctx.ToolResults))
				for _, tr := range ctx.ToolResults {
					if containsString(pending, tr.ToolUseID) {
						kept = append(kept, tr)
						pending = removeString(pending, tr.ToolUseID)
					}
				}
				ctx.ToolResults = kept
				if len(ctx.ToolResults) == 0 && len(ctx.Tools) == 0 &&
					strings.TrimSpace(m.UserInputMessage.Content) == "" &&
					len(m.UserInputMessage.Images) == 0 {
					// 整条消息只有孤儿 toolResult，被全部丢弃 → 这条 user 也丢
					continue
				}
			}
			// 仍有 pending 未处理：先在这条 user 之前插占位
			if len(pending) > 0 && (ctx == nil || len(ctx.ToolResults) == 0) {
				flushMissingToolResults()
			}
			out = append(out, m)
			// pending 仍可能不为空（这条 user 不带 toolResult）
		}
	}
	// 收尾
	flushMissingToolResults()
	return out
}

func containsString(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

func removeString(slice []string, s string) []string {
	out := make([]string, 0, len(slice))
	for _, v := range slice {
		if v != s {
			out = append(out, v)
		}
	}
	return out
}

// kiroConversationID 用 history 前 2 条消息内容 sha256 前 32 字符派生稳定 ID
//
// 同会话同 history → 同 ID，让 Kiro 后端的 prompt cache 命中。
// history 为空时返回随机 UUID。
func kiroConversationID(history []KiroHistoryItem) string {
	if len(history) == 0 {
		return uuid.NewString()
	}
	var sb strings.Builder
	for i := 0; i < 2 && i < len(history); i++ {
		h := history[i]
		if h.UserInputMessage != nil {
			sb.WriteString("U:")
			sb.WriteString(h.UserInputMessage.Content)
		} else if h.AssistantResponseMessage != nil {
			sb.WriteString("A:")
			sb.WriteString(h.AssistantResponseMessage.Content)
		}
		sb.WriteString("\n")
	}
	sum := sha256.Sum256([]byte(sb.String()))
	return hex.EncodeToString(sum[:])[:32]
}

// kiroOriginForEndpoint cli 端点要求 origin=AmazonQ，其他 AI_EDITOR
func kiroOriginForEndpoint(ep KiroEndpoint) string {
	if ep.CLIMode {
		return "AmazonQ"
	}
	return "AI_EDITOR"
}

// resolveKiroProfileArn 按 login_type 返回正确的 profileArn
//
// 账号 credentials.profile_arn 显式存储优先；否则按 login_type 选预设：
//   - github / google → EHGA3GRVQMUK
//   - 其他（builder / idc / 空）→ AAAACCCCXXXX
func resolveKiroProfileArn(account *Account) string {
	if account == nil {
		return ""
	}
	if v := strings.TrimSpace(account.GetCredential("profile_arn")); v != "" {
		return v
	}
	loginType := strings.ToLower(strings.TrimSpace(account.GetCredential("login_type")))
	switch loginType {
	case "github", "google":
		return "arn:aws:codewhisperer:us-east-1:699475941385:profile/EHGA3GRVQMUK"
	default:
		return "arn:aws:codewhisperer:us-east-1:638616132270:profile/AAAACCCCXXXX"
	}
}

// extractImageFormat "image/png" → "png"
func extractImageFormat(mediaType string) string {
	if i := strings.IndexByte(mediaType, '/'); i >= 0 {
		return mediaType[i+1:]
	}
	return mediaType
}
