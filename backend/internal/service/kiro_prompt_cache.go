package service

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"sync"
	"time"
	"unicode"
)

// Kiro Prompt Cache 模拟
//
// Kiro/CodeWhisperer 上游不支持 Anthropic 的 cache_control 语义。为了让客户端
// （Claude Code 等）的费用估算合理，反代侧在内存里模拟 prompt cache：
//
//   - fingerprint：history 内容的滚动 SHA256，配合 cache_control 断点累加
//   - TTL：默认 5min；cache_control.ttl="1h" 时改为 1h
//   - 阈值：Opus 系列 4096 tokens / 其他 1024 tokens（不达阈值不命中）
//   - 命中上限：min(matchedTokens, totalInputTokens * 0.85)
//   - token 估算：CJK *0.6 + 其他 *0.3（无真实 token 时的兜底）
//
// 参考 chaogei/Kiro-account-manager src/main/proxy/promptCacheTracker.ts。

const (
	kiroCacheTTLDefault = 5 * time.Minute
	kiroCacheTTL1h      = 1 * time.Hour

	kiroCacheMinTokensOpus  = 4096
	kiroCacheMinTokensOther = 1024

	kiroCacheHitRatioCap = 0.85
)

// KiroPromptCacheEntry 一个 cache 断点的记录
type kiroPromptCacheEntry struct {
	cumulativeTokens int
	expiresAt        time.Time
}

// KiroPromptCache 进程内单例 cache。线程安全。
type KiroPromptCache struct {
	mu      sync.Mutex
	entries map[string]*kiroPromptCacheEntry
}

// NewKiroPromptCache 创建一个空 cache
func NewKiroPromptCache() *KiroPromptCache {
	return &KiroPromptCache{entries: map[string]*kiroPromptCacheEntry{}}
}

// Probe 探测 cache 命中
//
// 当前实现：用 history 全量内容算单个 fingerprint，命中即返回 cumulativeTokens；
// 未命中则写入新条目（创建 cache）。返回值是要回填给客户端的命中 token 数。
//
// modelID 用于阈值判断：包含 "opus" 走 4096 阈值，其他走 1024。
// totalInputTokens 用于命中上限封顶（0.85）。
func (c *KiroPromptCache) Probe(modelID string, history []KiroHistoryItem, totalInputTokens int, ttl time.Duration) int {
	if len(history) == 0 || totalInputTokens <= 0 {
		return 0
	}
	fp := fingerprintKiroHistory(history)
	cumulative := EstimateKiroHistoryTokens(history)
	threshold := kiroCacheMinTokensOther
	if strings.Contains(strings.ToLower(modelID), "opus") {
		threshold = kiroCacheMinTokensOpus
	}
	if cumulative < threshold {
		return 0
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	if entry, ok := c.entries[fp]; ok && entry.expiresAt.After(now) {
		// 命中；刷新 TTL
		entry.expiresAt = now.Add(ttl)
		hit := entry.cumulativeTokens
		cap := int(float64(totalInputTokens) * kiroCacheHitRatioCap)
		if hit > cap {
			hit = cap
		}
		return hit
	}

	// 未命中；新建 entry
	c.entries[fp] = &kiroPromptCacheEntry{
		cumulativeTokens: cumulative,
		expiresAt:        now.Add(ttl),
	}
	return 0
}

// Purge 移除已过期 entry。可由后台 goroutine 定期调用；当前实现不主动启动。
func (c *KiroPromptCache) Purge() {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := time.Now()
	for k, e := range c.entries {
		if !e.expiresAt.After(now) {
			delete(c.entries, k)
		}
	}
}

// fingerprintKiroHistory 对 history 做 sha256 取前 32 字符 hex
//
// 与 conversationId 派生算法保持一致风格，但这里取全量 history 而不是前 2 条。
func fingerprintKiroHistory(history []KiroHistoryItem) string {
	var sb strings.Builder
	for _, h := range history {
		if h.UserInputMessage != nil {
			sb.WriteString("U|")
			sb.WriteString(h.UserInputMessage.Content)
		} else if h.AssistantResponseMessage != nil {
			sb.WriteString("A|")
			sb.WriteString(h.AssistantResponseMessage.Content)
		}
		sb.WriteByte('\n')
	}
	sum := sha256.Sum256([]byte(sb.String()))
	return hex.EncodeToString(sum[:])[:32]
}

// EstimateKiroHistoryTokens 估算 history 总 token 数
//
// CJK 字符 *0.6，其他字符 *0.3，向上取整。
func EstimateKiroHistoryTokens(history []KiroHistoryItem) int {
	total := 0
	for _, h := range history {
		if h.UserInputMessage != nil {
			total += estimateTokens(h.UserInputMessage.Content)
		} else if h.AssistantResponseMessage != nil {
			total += estimateTokens(h.AssistantResponseMessage.Content)
		}
	}
	return total
}

// estimateTokens 字符级 token 估算
func estimateTokens(text string) int {
	if text == "" {
		return 0
	}
	cjk := 0
	other := 0
	for _, r := range text {
		if isCJK(r) {
			cjk++
		} else if !unicode.IsSpace(r) {
			other++
		}
	}
	tokens := float64(cjk)*0.6 + float64(other)*0.3
	if tokens < 1 {
		return 1
	}
	return int(tokens + 0.5)
}

// isCJK 是否为 CJK 范围字符
func isCJK(r rune) bool {
	switch {
	case r >= 0x4E00 && r <= 0x9FFF: // CJK Unified
		return true
	case r >= 0x3400 && r <= 0x4DBF: // CJK Ext A
		return true
	case r >= 0x3040 && r <= 0x30FF: // Hiragana / Katakana
		return true
	case r >= 0xAC00 && r <= 0xD7AF: // Hangul Syllables
		return true
	}
	return false
}
