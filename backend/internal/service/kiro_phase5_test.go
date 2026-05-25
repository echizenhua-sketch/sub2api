package service

import (
	"strings"
	"testing"
	"time"
)

func TestKiroToolNameRegistry_ShortNamePassthrough(t *testing.T) {
	t.Parallel()
	r := NewKiroToolNameRegistry()
	encoded := r.Encode("get_time")
	if encoded != "get_time" {
		t.Errorf("short name should be passthrough: got %q", encoded)
	}
	orig, ok := r.Decode("get_time")
	if !ok || orig != "get_time" {
		t.Errorf("decode short name: %q ok=%v", orig, ok)
	}
}

func TestKiroToolNameRegistry_LongNameTruncation(t *testing.T) {
	t.Parallel()
	r := NewKiroToolNameRegistry()
	long := strings.Repeat("abcdef_", 20) // 140 chars
	encoded := r.Encode(long)
	if len(encoded) > 64 {
		t.Errorf("encoded length must be <=64: got %d (%q)", len(encoded), encoded)
	}
	if !strings.Contains(encoded, "_") {
		t.Errorf("encoded must contain _hash suffix")
	}
	orig, ok := r.Decode(encoded)
	if !ok || orig != long {
		t.Errorf("decode round-trip failed: %q→%q (ok=%v)", encoded, orig, ok)
	}
}

func TestKiroToolNameRegistry_InvalidCharsReplaced(t *testing.T) {
	t.Parallel()
	r := NewKiroToolNameRegistry()
	in := "weather.api/get current"
	encoded := r.Encode(in)
	for _, ch := range encoded {
		ok := (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_' || ch == '-'
		if !ok {
			t.Errorf("encoded contains invalid char %q in %q", ch, encoded)
			return
		}
	}
}

func TestKiroToolNameRegistry_StableForSameInput(t *testing.T) {
	t.Parallel()
	r := NewKiroToolNameRegistry()
	a := r.Encode("super.long.dotted.tool.name.over.sixty.four.chars.demo.demo.demo.demo")
	b := r.Encode("super.long.dotted.tool.name.over.sixty.four.chars.demo.demo.demo.demo")
	if a != b {
		t.Errorf("same input should yield same encoded: %q vs %q", a, b)
	}
}

func TestTruncateKiroToolDescription(t *testing.T) {
	t.Parallel()
	short := "abc"
	if got := TruncateKiroToolDescription(short); got != short {
		t.Errorf("short desc shouldn't change: %q", got)
	}
	long := strings.Repeat("x", kiroToolDescriptionMaxLen+50)
	got := TruncateKiroToolDescription(long)
	if len(got) != kiroToolDescriptionMaxLen+3 {
		t.Errorf("truncate length: want %d got %d", kiroToolDescriptionMaxLen+3, len(got))
	}
	if !strings.HasSuffix(got, "...") {
		t.Errorf("must end with ...")
	}
}

func TestKiroPromptCache_BelowThresholdMisses(t *testing.T) {
	t.Parallel()
	c := NewKiroPromptCache()
	history := []KiroHistoryItem{
		{UserInputMessage: &KiroUserInputMessage{Content: "hi"}},
	}
	hit := c.Probe("claude-sonnet-4-5", history, 100, kiroCacheTTLDefault)
	if hit != 0 {
		t.Errorf("below threshold should miss: got %d", hit)
	}
}

func TestKiroPromptCache_HitsOnSecondProbe(t *testing.T) {
	t.Parallel()
	c := NewKiroPromptCache()
	bigText := strings.Repeat("hello world ", 500) // ~6000 chars * 0.3 ≈ 1800 tokens
	history := []KiroHistoryItem{
		{UserInputMessage: &KiroUserInputMessage{Content: bigText}},
	}
	if hit := c.Probe("claude-sonnet-4-5", history, 2000, kiroCacheTTLDefault); hit != 0 {
		t.Errorf("first probe should miss: got %d", hit)
	}
	hit := c.Probe("claude-sonnet-4-5", history, 2000, kiroCacheTTLDefault)
	if hit == 0 {
		t.Errorf("second probe should hit (got 0)")
	}
	cap := int(2000 * 0.85)
	if hit > cap {
		t.Errorf("hit must be capped at 0.85 of input: got %d cap %d", hit, cap)
	}
}

func TestKiroPromptCache_OpusHigherThreshold(t *testing.T) {
	t.Parallel()
	c := NewKiroPromptCache()
	mid := strings.Repeat("x", 5000) // ~5000*0.3 = 1500 tokens; >1024 but <4096
	h := []KiroHistoryItem{{UserInputMessage: &KiroUserInputMessage{Content: mid}}}

	// Sonnet: should pass threshold (1024) → second probe hits
	c.Probe("claude-sonnet-4-5", h, 5000, kiroCacheTTLDefault)
	if c.Probe("claude-sonnet-4-5", h, 5000, kiroCacheTTLDefault) == 0 {
		t.Errorf("sonnet 1500 tokens should hit on 2nd probe")
	}

	// Opus 4.7: 1500 < 4096 → never hits
	c2 := NewKiroPromptCache()
	c2.Probe("claude-opus-4-7", h, 5000, kiroCacheTTLDefault)
	if hit := c2.Probe("claude-opus-4-7", h, 5000, kiroCacheTTLDefault); hit != 0 {
		t.Errorf("opus 1500 tokens should NOT hit (threshold 4096): got %d", hit)
	}
}

func TestKiroPromptCache_TTLExpires(t *testing.T) {
	t.Parallel()
	c := NewKiroPromptCache()
	h := []KiroHistoryItem{{UserInputMessage: &KiroUserInputMessage{Content: strings.Repeat("y", 5000)}}}
	c.Probe("claude-sonnet-4-5", h, 5000, 1*time.Millisecond)
	time.Sleep(5 * time.Millisecond)
	if hit := c.Probe("claude-sonnet-4-5", h, 5000, 1*time.Millisecond); hit != 0 {
		t.Errorf("after TTL, should miss: got %d", hit)
	}
}

func TestEstimateTokens_CJKAndASCII(t *testing.T) {
	t.Parallel()
	cjk := estimateTokens("你好世界")    // 4 CJK * 0.6 = 2.4 → 2
	ascii := estimateTokens("hello") // 5 *0.3 = 1.5 → 2
	if cjk < 2 || cjk > 3 {
		t.Errorf("cjk: %d", cjk)
	}
	if ascii < 1 || ascii > 2 {
		t.Errorf("ascii: %d", ascii)
	}
	if estimateTokens("") != 0 {
		t.Errorf("empty: nonzero")
	}
}

func TestSanitize_RemovesOrphanToolResult(t *testing.T) {
	t.Parallel()
	// 上一条不是 assistant 的 toolUse → tool_result 是孤儿，应被丢弃
	msgs := []KiroHistoryItem{
		{UserInputMessage: &KiroUserInputMessage{Content: "hi"}},
		{UserInputMessage: &KiroUserInputMessage{
			UserInputMessageContext: &KiroUserInputMessageContext{
				ToolResults: []KiroToolResult{{ToolUseID: "tu_orphan", Content: []KiroToolResultBlock{{Text: "x"}}}},
			},
		}},
		{AssistantResponseMessage: &KiroAssistantResponseMessage{Content: "ok"}},
		{UserInputMessage: &KiroUserInputMessage{Content: "follow-up"}},
	}
	out := sanitizeKiroConversation(msgs, "m", kiroPrimaryEndpoint())

	// 检查输出里没有保留 tu_orphan 的 toolResult
	for _, m := range out {
		if m.UserInputMessage != nil && m.UserInputMessage.UserInputMessageContext != nil {
			for _, tr := range m.UserInputMessage.UserInputMessageContext.ToolResults {
				if tr.ToolUseID == "tu_orphan" {
					t.Errorf("orphan toolResult should have been removed")
				}
			}
		}
	}
}

func TestSanitize_FillsMissingToolResultPlaceholder(t *testing.T) {
	t.Parallel()
	// assistant 有 tool_use 但下条 user 没给 tool_result → 应插占位
	msgs := []KiroHistoryItem{
		{UserInputMessage: &KiroUserInputMessage{Content: "hi"}},
		{AssistantResponseMessage: &KiroAssistantResponseMessage{
			Content:  "calling tool",
			ToolUses: []KiroToolUse{{ToolUseID: "tu_a", Name: "get_x"}},
		}},
		{UserInputMessage: &KiroUserInputMessage{Content: "next"}}, // 没带 tu_a 的 result
	}
	out := sanitizeKiroConversation(msgs, "m", kiroPrimaryEndpoint())

	var found bool
	for _, m := range out {
		if m.UserInputMessage != nil && m.UserInputMessage.UserInputMessageContext != nil {
			for _, tr := range m.UserInputMessage.UserInputMessageContext.ToolResults {
				if tr.ToolUseID == "tu_a" && tr.Status == "error" {
					found = true
				}
			}
		}
	}
	if !found {
		t.Errorf("missing tool_result placeholder not inserted")
	}
}

func TestSanitize_RemovesEmptyUserExceptFirst(t *testing.T) {
	t.Parallel()
	msgs := []KiroHistoryItem{
		{UserInputMessage: &KiroUserInputMessage{Content: ""}}, // 首条空，保留
		{AssistantResponseMessage: &KiroAssistantResponseMessage{Content: "ok"}},
		{UserInputMessage: &KiroUserInputMessage{Content: ""}}, // 中间空，应删
		{AssistantResponseMessage: &KiroAssistantResponseMessage{Content: "ok2"}},
		{UserInputMessage: &KiroUserInputMessage{Content: "real"}},
	}
	out := sanitizeKiroConversation(msgs, "m", kiroPrimaryEndpoint())

	emptyUserCount := 0
	for _, m := range out {
		if m.UserInputMessage != nil && strings.TrimSpace(m.UserInputMessage.Content) == "" &&
			(m.UserInputMessage.UserInputMessageContext == nil ||
				len(m.UserInputMessage.UserInputMessageContext.ToolResults) == 0) {
			emptyUserCount++
		}
	}
	if emptyUserCount != 1 {
		t.Errorf("expected exactly 1 empty user (the first one); got %d", emptyUserCount)
	}
}
