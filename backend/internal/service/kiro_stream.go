package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/Wei-Shaw/sub2api/internal/pkg/logger"
)

// forwardKiro 是 Kiro/CodeWhisperer 上游的网关入口
//
// 调用链镜像 forwardBedrock：解析模型 → 构建 payload → 选 endpoint → 发请求 →
// 解析 AWS Event Stream → 转 Anthropic SSE → 返回 ForwardResult。
//
// Phase 3b 的范围：
//   - 单 endpoint（codewhisperer），固定不轮转
//   - 401/403/429 不做被动刷新与切号（Phase 4 实装）
//   - 重试 / TLS 走标准 httpUpstream
func (s *GatewayService) forwardKiro(
	ctx context.Context,
	c *gin.Context,
	account *Account,
	parsed *ParsedRequest,
	startTime time.Time,
) (*ForwardResult, error) {
	reqStream := parsed.Stream
	body := parsed.Body

	if account == nil || !account.IsKiro() {
		return nil, fmt.Errorf("forwardKiro: account is not kiro")
	}
	accessToken := strings.TrimSpace(account.GetCredential("access_token"))
	if accessToken == "" {
		return nil, fmt.Errorf("forwardKiro: account has no access_token")
	}

	endpoint := kiroPrimaryEndpoint()

	payload, registry, err := AnthropicToKiroWithRegistry(body, account, endpoint)
	if err != nil {
		return nil, fmt.Errorf("forwardKiro: translate request: %w", err)
	}

	upstreamBody, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("forwardKiro: marshal payload: %w", err)
	}

	headers := buildKiroHeaders(account, accessToken, endpoint)

	upstreamReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.URL, bytes.NewReader(upstreamBody))
	if err != nil {
		return nil, fmt.Errorf("forwardKiro: build upstream request: %w", err)
	}
	for k, v := range headers {
		upstreamReq.Header.Set(k, v)
	}

	proxyURL := ""
	if account.ProxyID != nil && account.Proxy != nil {
		proxyURL = account.Proxy.URL()
	}

	logger.LegacyPrintf("service.gateway", "[Kiro] dispatch: account=%d name=%s model=%s stream=%v endpoint=%s",
		account.ID, account.Name, parsed.Model, reqStream, endpoint.URL)

	resp, err := s.httpUpstream.DoWithTLS(upstreamReq, proxyURL, account.ID, account.Concurrency, nil)
	if err != nil {
		if resp != nil && resp.Body != nil {
			_ = resp.Body.Close()
		}
		c.JSON(http.StatusBadGateway, gin.H{
			"type": "error",
			"error": gin.H{
				"type":    "upstream_error",
				"message": "Kiro upstream request failed",
			},
		})
		return nil, fmt.Errorf("forwardKiro: upstream request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if awsReqID := resp.Header.Get("x-amzn-requestid"); awsReqID != "" && resp.Header.Get("x-request-id") == "" {
		resp.Header.Set("x-request-id", awsReqID)
	}

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
		_ = resp.Body.Close()
		// 重新装回 body 给 handleErrorResponse 用
		resp.Body = io.NopCloser(bytes.NewReader(respBody))

		class := classifyKiroError(resp.StatusCode, respBody)
		logger.LegacyPrintf("service.gateway",
			"[Kiro] account=%d upstream %d (%s): %s",
			account.ID, resp.StatusCode, class, truncateString(string(respBody), 500))

		// 走标准 failover；shouldFailoverUpstreamError 已覆盖 401/403/429/5xx
		return s.handleErrorResponse(ctx, resp, c, account)
	}

	usage := &ClaudeUsage{}
	var firstTokenMs *int
	var clientDisconnect bool

	if reqStream {
		streamResult, err := s.handleKiroStreamingResponse(ctx, resp, c, account, startTime, parsed.Model, registry)
		if err != nil {
			return nil, err
		}
		usage = streamResult.usage
		firstTokenMs = streamResult.firstTokenMs
		clientDisconnect = streamResult.clientDisconnect
	} else {
		usage, err = s.handleKiroNonStreamingResponse(ctx, resp, c, parsed.Model, registry)
		if err != nil {
			return nil, err
		}
	}

	return &ForwardResult{
		RequestID:        resp.Header.Get("x-amzn-requestid"),
		Usage:            *usage,
		Model:            parsed.Model,
		UpstreamModel:    parsed.Model, // Kiro 不映射模型
		Stream:           reqStream,
		Duration:         time.Since(startTime),
		FirstTokenMs:     firstTokenMs,
		ClientDisconnect: clientDisconnect,
	}, nil
}

// handleKiroStreamingResponse 把 Kiro 的 AWS Event Stream 翻译成 Anthropic SSE 写回客户端
//
// Kiro 事件类型（参考 kiroApi.ts:1300-1610）：
//   - assistantResponseEvent: {content: string} → text_delta
//   - toolUseEvent: 多帧累积 input 直到 stop=true
//   - messageMetadataEvent: {tokenUsage:{...}} → message_delta usage
//   - reasoningContentEvent: {text/signature/redactedContent} → thinking_delta
//   - meteringEvent: 内部计费，不下发 SSE
//
// Phase 3b MVP：实现 assistantResponseEvent + messageMetadataEvent + 基础 toolUseEvent；
// reasoningContentEvent / thinking 块 Phase 5+ 完善。
func (s *GatewayService) handleKiroStreamingResponse(
	ctx context.Context,
	resp *http.Response,
	c *gin.Context,
	account *Account,
	startTime time.Time,
	model string,
	registry *KiroToolNameRegistry,
) (*streamingResult, error) {
	w := c.Writer
	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil, errors.New("streaming not supported")
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	usage := &ClaudeUsage{}
	var firstTokenMs *int
	clientDisconnected := false

	messageID := "msg_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	state := newKiroSSEState(messageID, model, registry)

	decoder := NewAWSEventStreamDecoder(resp.Body)
	frameCh := make(chan *AWSEventStreamFrame, 16)
	errCh := make(chan error, 1)
	doneRead := make(chan struct{})

	go func() {
		defer close(doneRead)
		for {
			frame, err := decoder.NextFrame()
			if err != nil {
				if !errors.Is(err, io.EOF) {
					errCh <- err
				}
				close(frameCh)
				return
			}
			select {
			case frameCh <- frame:
			case <-ctx.Done():
				close(frameCh)
				return
			}
		}
	}()

	writeSSE := func(eventName string, dataObj any) bool {
		if clientDisconnected {
			return false
		}
		raw, err := json.Marshal(dataObj)
		if err != nil {
			return false
		}
		if _, err := fmt.Fprintf(w, "event: %s\ndata: %s\n\n", eventName, raw); err != nil {
			clientDisconnected = true
			return false
		}
		if firstTokenMs == nil {
			ms := int(time.Since(startTime).Milliseconds())
			firstTokenMs = &ms
		}
		flusher.Flush()
		return true
	}

	// 先发 message_start
	writeSSE("message_start", map[string]any{
		"type": "message_start",
		"message": map[string]any{
			"id":            messageID,
			"type":          "message",
			"role":          "assistant",
			"model":         model,
			"content":       []any{},
			"stop_reason":   nil,
			"stop_sequence": nil,
			"usage": map[string]any{
				"input_tokens":  0,
				"output_tokens": 0,
			},
		},
	})

	for frame := range frameCh {
		if frame == nil {
			continue
		}
		if frame.MessageType == "exception" || frame.MessageType == "error" || frame.ExceptionType != "" {
			logger.LegacyPrintf("service.gateway", "[Kiro] account=%d upstream stream exception: type=%s exception=%s payload=%s",
				account.ID, frame.MessageType, frame.ExceptionType, string(frame.Payload))
			break
		}
		switch frame.EventType {
		case "assistantResponseEvent":
			text := jsonStringField(frame.Payload, "content")
			if text == "" {
				continue
			}
			state.ensureTextBlockOpen(writeSSE)
			writeSSE("content_block_delta", map[string]any{
				"type":  "content_block_delta",
				"index": state.activeIndex,
				"delta": map[string]any{"type": "text_delta", "text": text},
			})
		case "toolUseEvent":
			state.handleToolUseEvent(frame.Payload, usage, writeSSE)
		case "messageMetadataEvent":
			applyKiroMetadataToUsage(frame.Payload, usage)
		case "meteringEvent":
			// 内部计费，不下发 SSE
		default:
			// 包括 reasoningContentEvent / supplementaryWebLinksEvent / contextUsageEvent / ...
			// Phase 5 实装；此处先忽略
		}
	}

	select {
	case err := <-errCh:
		if err != nil && !errors.Is(err, io.EOF) {
			logger.LegacyPrintf("service.gateway", "[Kiro] account=%d stream decode error: %v", account.ID, err)
		}
	default:
	}

	// 关闭打开的 content block
	state.flushPendingToolUse(writeSSE)
	state.closeActiveBlock(writeSSE)

	stopReason := "end_turn"
	if state.toolUseCount > 0 {
		stopReason = "tool_use"
	}
	writeSSE("message_delta", map[string]any{
		"type": "message_delta",
		"delta": map[string]any{
			"stop_reason":   stopReason,
			"stop_sequence": nil,
		},
		"usage": map[string]any{
			"input_tokens":              usage.InputTokens,
			"output_tokens":             usage.OutputTokens,
			"cache_creation_input_tokens": usage.CacheCreationInputTokens,
			"cache_read_input_tokens":     usage.CacheReadInputTokens,
		},
	})
	writeSSE("message_stop", map[string]any{"type": "message_stop"})

	// 等读取 goroutine 真正退出，避免 race
	<-doneRead

	return &streamingResult{usage: usage, firstTokenMs: firstTokenMs, clientDisconnect: clientDisconnected}, nil
}

// handleKiroNonStreamingResponse 聚合 stream 帧成一个 Anthropic Messages 响应 JSON
//
// Kiro 上游永远是 stream 协议（AWS Event Stream），即使客户端请求 stream=false。
// 我们仍然解码所有帧，但聚合成单个 JSON 响应。
func (s *GatewayService) handleKiroNonStreamingResponse(
	ctx context.Context,
	resp *http.Response,
	c *gin.Context,
	model string,
	registry *KiroToolNameRegistry,
) (*ClaudeUsage, error) {
	usage := &ClaudeUsage{}
	decoder := NewAWSEventStreamDecoder(resp.Body)

	type aggToolUse struct {
		ID       string
		Name     string
		InputBuf strings.Builder
		Done     bool
	}
	textBuf := strings.Builder{}
	toolUses := []*aggToolUse{}
	var current *aggToolUse

	for {
		frame, err := decoder.NextFrame()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, fmt.Errorf("kiro non-stream decode: %w", err)
		}
		if frame == nil {
			continue
		}
		if frame.MessageType == "exception" || frame.MessageType == "error" || frame.ExceptionType != "" {
			return nil, fmt.Errorf("kiro upstream stream exception: type=%s exception=%s payload=%s",
				frame.MessageType, frame.ExceptionType, string(frame.Payload))
		}
		switch frame.EventType {
		case "assistantResponseEvent":
			textBuf.WriteString(jsonStringField(frame.Payload, "content"))
		case "toolUseEvent":
			id := jsonStringField(frame.Payload, "toolUseId")
			name := jsonStringField(frame.Payload, "name")
			input := jsonStringField(frame.Payload, "input")
			stop := jsonBoolField(frame.Payload, "stop")

			if current == nil || current.ID != id {
				current = &aggToolUse{ID: id, Name: name}
				toolUses = append(toolUses, current)
			}
			if current.Name == "" && name != "" {
				current.Name = name
			}
			if input != "" {
				current.InputBuf.WriteString(input)
			}
			if stop {
				current.Done = true
				current = nil
			}
		case "messageMetadataEvent":
			applyKiroMetadataToUsage(frame.Payload, usage)
		}
	}

	// 构造 Anthropic Messages JSON 响应
	contentBlocks := []map[string]any{}
	if textBuf.Len() > 0 {
		contentBlocks = append(contentBlocks, map[string]any{
			"type": "text",
			"text": textBuf.String(),
		})
	}
	for _, tu := range toolUses {
		var inputObj any
		raw := tu.InputBuf.String()
		if raw == "" {
			inputObj = map[string]any{}
		} else if err := json.Unmarshal([]byte(raw), &inputObj); err != nil {
			inputObj = map[string]any{"_partial_input": raw, "_error": err.Error()}
		}
		nameForClient := tu.Name
		if registry != nil {
			if orig, ok := registry.Decode(tu.Name); ok {
				nameForClient = orig
			}
		}
		contentBlocks = append(contentBlocks, map[string]any{
			"type":  "tool_use",
			"id":    tu.ID,
			"name":  nameForClient,
			"input": inputObj,
		})
	}

	stopReason := "end_turn"
	if len(toolUses) > 0 {
		stopReason = "tool_use"
	}

	out := map[string]any{
		"id":            "msg_" + strings.ReplaceAll(uuid.NewString(), "-", ""),
		"type":          "message",
		"role":          "assistant",
		"model":         model,
		"content":       contentBlocks,
		"stop_reason":   stopReason,
		"stop_sequence": nil,
		"usage": map[string]any{
			"input_tokens":              usage.InputTokens,
			"output_tokens":             usage.OutputTokens,
			"cache_creation_input_tokens": usage.CacheCreationInputTokens,
			"cache_read_input_tokens":     usage.CacheReadInputTokens,
		},
	}
	c.JSON(http.StatusOK, out)
	_ = ctx
	return usage, nil
}

// applyKiroMetadataToUsage 把 messageMetadataEvent.tokenUsage 套进 ClaudeUsage
func applyKiroMetadataToUsage(payload []byte, usage *ClaudeUsage) {
	var data struct {
		TokenUsage struct {
			UncachedInputTokens     int `json:"uncachedInputTokens"`
			CacheReadInputTokens    int `json:"cacheReadInputTokens"`
			CacheWriteInputTokens   int `json:"cacheWriteInputTokens"`
			OutputTokens            int `json:"outputTokens"`
			TotalTokens             int `json:"totalTokens"`
		} `json:"tokenUsage"`
	}
	if err := json.Unmarshal(payload, &data); err != nil {
		return
	}
	usage.InputTokens = data.TokenUsage.UncachedInputTokens
	usage.OutputTokens = data.TokenUsage.OutputTokens
	usage.CacheReadInputTokens = data.TokenUsage.CacheReadInputTokens
	usage.CacheCreationInputTokens = data.TokenUsage.CacheWriteInputTokens
}

// kiroSSEState 流式输出的连续状态
//
// 同时只能开一个 content block：text 或 tool_use。新的 tool_use 出现时关掉 text 块。
// activeIndex 沿用 Anthropic SSE 的 content block 索引。
type kiroSSEState struct {
	messageID    string
	model        string
	registry     *KiroToolNameRegistry
	activeIndex  int
	textOpen     bool
	pendingTool  *kiroPendingToolUse
	toolUseCount int
}

type kiroPendingToolUse struct {
	ToolUseID string
	Name      string
	Index     int
	InputBuf  strings.Builder
	Started   bool
}

func newKiroSSEState(messageID, model string, registry *KiroToolNameRegistry) *kiroSSEState {
	return &kiroSSEState{messageID: messageID, model: model, registry: registry, activeIndex: -1}
}

// ensureTextBlockOpen 保证当前有一个 text 块在写
func (s *kiroSSEState) ensureTextBlockOpen(write func(string, any) bool) {
	if s.textOpen {
		return
	}
	if s.pendingTool != nil {
		s.flushPendingToolUse(write)
	}
	s.activeIndex++
	write("content_block_start", map[string]any{
		"type":          "content_block_start",
		"index":         s.activeIndex,
		"content_block": map[string]any{"type": "text", "text": ""},
	})
	s.textOpen = true
}

// closeActiveBlock 关掉当前块（text 或 tool_use）
func (s *kiroSSEState) closeActiveBlock(write func(string, any) bool) {
	if s.activeIndex < 0 {
		return
	}
	if s.textOpen || s.pendingTool != nil {
		write("content_block_stop", map[string]any{
			"type":  "content_block_stop",
			"index": s.activeIndex,
		})
	}
	s.textOpen = false
	s.pendingTool = nil
}

// handleToolUseEvent 处理 toolUseEvent，多帧累积 input
//
// 单帧字段：toolUseId / name / input(片段) / stop(bool)
func (s *kiroSSEState) handleToolUseEvent(payload []byte, _ *ClaudeUsage, write func(string, any) bool) {
	id := jsonStringField(payload, "toolUseId")
	name := jsonStringField(payload, "name")
	inputFrag := jsonStringField(payload, "input")
	stop := jsonBoolField(payload, "stop")

	// 新工具调用出现，关掉之前的 text/tool 块
	if s.pendingTool == nil || s.pendingTool.ToolUseID != id {
		if s.textOpen {
			s.closeActiveBlock(write)
		}
		if s.pendingTool != nil {
			s.flushPendingToolUse(write)
		}
		s.activeIndex++
		s.pendingTool = &kiroPendingToolUse{
			ToolUseID: id,
			Name:      name,
			Index:     s.activeIndex,
		}
	}
	if s.pendingTool.Name == "" && name != "" {
		s.pendingTool.Name = name
	}

	if !s.pendingTool.Started {
		nameForClient := s.pendingTool.Name
		if s.registry != nil {
			if orig, ok := s.registry.Decode(s.pendingTool.Name); ok {
				nameForClient = orig
			}
		}
		write("content_block_start", map[string]any{
			"type":  "content_block_start",
			"index": s.pendingTool.Index,
			"content_block": map[string]any{
				"type":  "tool_use",
				"id":    s.pendingTool.ToolUseID,
				"name":  nameForClient,
				"input": map[string]any{},
			},
		})
		s.pendingTool.Started = true
	}

	if inputFrag != "" {
		s.pendingTool.InputBuf.WriteString(inputFrag)
		write("content_block_delta", map[string]any{
			"type":  "content_block_delta",
			"index": s.pendingTool.Index,
			"delta": map[string]any{
				"type":         "input_json_delta",
				"partial_json": inputFrag,
			},
		})
	}

	if stop {
		write("content_block_stop", map[string]any{
			"type":  "content_block_stop",
			"index": s.pendingTool.Index,
		})
		s.toolUseCount++
		s.pendingTool = nil
	}
}

// flushPendingToolUse 把没收到 stop 的 tool_use 强制收尾
func (s *kiroSSEState) flushPendingToolUse(write func(string, any) bool) {
	if s.pendingTool == nil || !s.pendingTool.Started {
		s.pendingTool = nil
		return
	}
	write("content_block_stop", map[string]any{
		"type":  "content_block_stop",
		"index": s.pendingTool.Index,
	})
	s.toolUseCount++
	s.pendingTool = nil
}

// jsonStringField 安全提取 JSON 中的 string 字段
func jsonStringField(b []byte, key string) string {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(b, &m); err != nil {
		return ""
	}
	raw, ok := m[key]
	if !ok {
		return ""
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s
	}
	return string(raw)
}

// jsonBoolField 安全提取 JSON 中的 bool 字段
func jsonBoolField(b []byte, key string) bool {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(b, &m); err != nil {
		return false
	}
	raw, ok := m[key]
	if !ok {
		return false
	}
	var v bool
	if err := json.Unmarshal(raw, &v); err == nil {
		return v
	}
	return false
}
