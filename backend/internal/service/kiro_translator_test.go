package service

import (
	"encoding/json"
	"strings"
	"testing"
)

// 简短工具：构造一个 Kiro account
func newTestKiroAccount(loginType string) *Account {
	a := &Account{
		ID:       42,
		Platform: PlatformKiro,
		Type:     AccountTypeOAuth,
	}
	a.Credentials = map[string]any{
		"access_token":  "AT",
		"refresh_token": "RT",
		"client_id":     "CID",
		"client_secret": "CSEC",
		"login_type":    loginType,
	}
	return a
}

func TestAnthropicToKiro_BasicSystemAndUser(t *testing.T) {
	t.Parallel()

	body := mustJSON(map[string]any{
		"model":  "claude-sonnet-4-5",
		"system": "you are helpful",
		"messages": []any{
			map[string]any{"role": "user", "content": "hi"},
		},
		"max_tokens":  256,
		"temperature": 0.7,
	})

	acc := newTestKiroAccount("builder")
	payload, err := AnthropicToKiro(body, acc, kiroPrimaryEndpoint())
	if err != nil {
		t.Fatalf("translate: %v", err)
	}

	// system 应该作为 history 头部 user 消息 + 固定 assistant 回复
	if len(payload.ConversationState.History) < 2 {
		t.Fatalf("history too short: %d", len(payload.ConversationState.History))
	}
	if got := payload.ConversationState.History[0].UserInputMessage.Content; got != "you are helpful" {
		t.Errorf("system→user[0]: %q", got)
	}
	if got := payload.ConversationState.History[1].AssistantResponseMessage.Content; got != "I will follow these instructions." {
		t.Errorf("system→assistant[1]: %q", got)
	}

	// 最后一条用户消息成为 currentMessage
	if got := payload.ConversationState.CurrentMessage.UserInputMessage.Content; got != "hi" {
		t.Errorf("currentMessage.content: %q", got)
	}
	if payload.ConversationState.CurrentMessage.UserInputMessage.ModelID != "claude-sonnet-4-5" {
		t.Errorf("modelId not propagated")
	}

	// inferenceConfig
	if payload.InferenceConfig == nil || payload.InferenceConfig.MaxTokens != 256 || payload.InferenceConfig.Temperature != 0.7 {
		t.Errorf("inferenceConfig wrong: %+v", payload.InferenceConfig)
	}

	// profileArn 默认 builder
	if !strings.Contains(payload.ProfileArn, "AAAACCCCXXXX") {
		t.Errorf("profileArn for builder should be AAAA: %q", payload.ProfileArn)
	}

	// agentTaskType=vibe（非 CLI 端点）
	if payload.ConversationState.AgentTaskType != "vibe" {
		t.Errorf("agentTaskType: %q", payload.ConversationState.AgentTaskType)
	}
}

func TestAnthropicToKiro_ToolUseAndResult(t *testing.T) {
	t.Parallel()
	body := mustJSON(map[string]any{
		"model": "claude-sonnet-4-5",
		"messages": []any{
			map[string]any{"role": "user", "content": "what's the weather"},
			map[string]any{
				"role": "assistant",
				"content": []any{
					map[string]any{"type": "text", "text": "let me check"},
					map[string]any{
						"type":  "tool_use",
						"id":    "tu_1",
						"name":  "get_weather",
						"input": map[string]any{"city": "Beijing"},
					},
				},
			},
			map[string]any{
				"role": "user",
				"content": []any{
					map[string]any{
						"type":         "tool_result",
						"tool_use_id":  "tu_1",
						"content":      "sunny",
					},
				},
			},
		},
	})

	payload, err := AnthropicToKiro(body, newTestKiroAccount("builder"), kiroPrimaryEndpoint())
	if err != nil {
		t.Fatalf("translate: %v", err)
	}

	// history 中应该有 assistant + toolUses
	hist := payload.ConversationState.History
	var foundToolUse bool
	for _, h := range hist {
		if h.AssistantResponseMessage != nil && len(h.AssistantResponseMessage.ToolUses) == 1 {
			tu := h.AssistantResponseMessage.ToolUses[0]
			if tu.ToolUseID == "tu_1" && tu.Name == "get_weather" && tu.Input["city"] == "Beijing" {
				foundToolUse = true
			}
		}
	}
	if !foundToolUse {
		t.Errorf("tool_use not found in history")
	}

	// currentMessage 应该是带 toolResults 的 user
	curCtx := payload.ConversationState.CurrentMessage.UserInputMessage.UserInputMessageContext
	if curCtx == nil || len(curCtx.ToolResults) != 1 {
		t.Fatalf("currentMessage missing toolResults: %+v", curCtx)
	}
	tr := curCtx.ToolResults[0]
	if tr.ToolUseID != "tu_1" || tr.Content[0].Text != "sunny" {
		t.Errorf("toolResult: %+v", tr)
	}
}

func TestAnthropicToKiro_ToolsOnlyOnCurrentMessage(t *testing.T) {
	t.Parallel()
	body := mustJSON(map[string]any{
		"model": "claude-sonnet-4-5",
		"messages": []any{
			map[string]any{"role": "user", "content": "hi"},
		},
		"tools": []any{
			map[string]any{
				"name":        "get_time",
				"description": "Get current time",
				"input_schema": map[string]any{
					"type":       "object",
					"properties": map[string]any{},
				},
			},
		},
	})
	p, err := AnthropicToKiro(body, newTestKiroAccount("builder"), kiroPrimaryEndpoint())
	if err != nil {
		t.Fatalf("translate: %v", err)
	}
	curCtx := p.ConversationState.CurrentMessage.UserInputMessage.UserInputMessageContext
	if curCtx == nil || len(curCtx.Tools) != 1 {
		t.Fatalf("tools not on currentMessage: %+v", curCtx)
	}
	if curCtx.Tools[0].ToolSpecification.Name != "get_time" {
		t.Errorf("tool name: %+v", curCtx.Tools[0])
	}
}

func TestSanitizeKiroConversation_StartAndEndWithUser(t *testing.T) {
	t.Parallel()

	// 以 assistant 起 → 前面插一个 user "Hello" 占位
	msgs := []KiroHistoryItem{
		{AssistantResponseMessage: &KiroAssistantResponseMessage{Content: "asst"}},
		{UserInputMessage: &KiroUserInputMessage{Content: "u"}},
	}
	out := sanitizeKiroConversation(msgs, "m", kiroPrimaryEndpoint())
	if out[0].UserInputMessage == nil {
		t.Errorf("must start with user, got %+v", out[0])
	}
	if out[len(out)-1].UserInputMessage == nil {
		t.Errorf("must end with user, got %+v", out[len(out)-1])
	}
}

func TestSanitizeKiroConversation_StrictAlternation(t *testing.T) {
	t.Parallel()

	// 连续两个 user，应该插一个 assistant 占位
	msgs := []KiroHistoryItem{
		{UserInputMessage: &KiroUserInputMessage{Content: "u1"}},
		{UserInputMessage: &KiroUserInputMessage{Content: "u2"}},
	}
	out := sanitizeKiroConversation(msgs, "m", kiroPrimaryEndpoint())

	roles := []string{}
	for _, h := range out {
		if h.UserInputMessage != nil {
			roles = append(roles, "U")
		} else {
			roles = append(roles, "A")
		}
	}
	got := strings.Join(roles, "")
	if !strings.HasPrefix(got, "UAU") {
		t.Errorf("alternation failed: got %s", got)
	}
}

func TestKiroConversationID_StableFromHistory(t *testing.T) {
	t.Parallel()
	h := []KiroHistoryItem{
		{UserInputMessage: &KiroUserInputMessage{Content: "hello"}},
		{AssistantResponseMessage: &KiroAssistantResponseMessage{Content: "world"}},
	}
	id1 := kiroConversationID(h)
	id2 := kiroConversationID(h)
	if id1 != id2 {
		t.Errorf("not stable: %q vs %q", id1, id2)
	}
	if len(id1) != 32 {
		t.Errorf("expected 32 chars: got %d", len(id1))
	}
}

func TestResolveKiroProfileArn_LoginTypeRouting(t *testing.T) {
	t.Parallel()
	tests := []struct {
		loginType string
		want      string
	}{
		{"builder", "AAAACCCCXXXX"},
		{"idc", "AAAACCCCXXXX"},
		{"github", "EHGA3GRVQMUK"},
		{"google", "EHGA3GRVQMUK"},
		{"", "AAAACCCCXXXX"},
	}
	for _, tt := range tests {
		t.Run(tt.loginType, func(t *testing.T) {
			arn := resolveKiroProfileArn(newTestKiroAccount(tt.loginType))
			if !strings.Contains(arn, tt.want) {
				t.Errorf("login_type=%q: want %q in %q", tt.loginType, tt.want, arn)
			}
		})
	}

	// 显式 profile_arn 覆盖
	a := newTestKiroAccount("builder")
	a.Credentials["profile_arn"] = "arn:aws:iam::custom"
	if got := resolveKiroProfileArn(a); got != "arn:aws:iam::custom" {
		t.Errorf("explicit profile_arn not honored: %q", got)
	}
}

func TestResolveKiroMachineID_FallbackSha256(t *testing.T) {
	t.Parallel()
	a := newTestKiroAccount("builder")
	id1 := resolveKiroMachineID(a)
	id2 := resolveKiroMachineID(a)
	if id1 != id2 || len(id1) != 64 {
		t.Errorf("machineId fallback unstable: %q (%d) vs %q", id1, len(id1), id2)
	}

	// 显式覆盖
	a.Credentials["machine_id"] = "deadbeef"
	if got := resolveKiroMachineID(a); got != "deadbeef" {
		t.Errorf("explicit machine_id not honored: %q", got)
	}
}

func TestApplyKiroMetadataToUsage(t *testing.T) {
	t.Parallel()
	payload, _ := json.Marshal(map[string]any{
		"tokenUsage": map[string]any{
			"uncachedInputTokens":   100,
			"cacheReadInputTokens":  50,
			"cacheWriteInputTokens": 30,
			"outputTokens":          200,
			"totalTokens":           380,
		},
	})
	u := &ClaudeUsage{}
	applyKiroMetadataToUsage(payload, u)
	if u.InputTokens != 100 || u.OutputTokens != 200 || u.CacheReadInputTokens != 50 || u.CacheCreationInputTokens != 30 {
		t.Errorf("usage mapping wrong: %+v", u)
	}
}

func TestJSONFieldHelpers(t *testing.T) {
	t.Parallel()
	body := []byte(`{"x":"yes","n":true,"obj":{"k":"v"}}`)
	if got := jsonStringField(body, "x"); got != "yes" {
		t.Errorf("string: %q", got)
	}
	if got := jsonStringField(body, "missing"); got != "" {
		t.Errorf("missing string: %q", got)
	}
	if !jsonBoolField(body, "n") {
		t.Errorf("bool: false")
	}
	if jsonBoolField(body, "missing") {
		t.Errorf("missing bool: true")
	}
}
