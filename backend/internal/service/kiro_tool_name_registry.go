package service

import (
	"fmt"
	"hash/fnv"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

// Kiro 工具名长度与字符集约束
//
// Kiro 上游对 toolSpecification.name 有限制：长度 ≤64，字符集 [a-zA-Z0-9_-]。
// 客户端（特别是 OpenAI/Anthropic 协议下）经常发更长或带其他字符的工具名，需要先做
// 可逆的截断映射，响应里再还原回去。
//
// 同时 toolSpecification.description >10237 字符要截断。
//
// 参考 chaogei/Kiro-account-manager src/main/proxy/toolNameRegistry.ts:5-50。

const (
	kiroToolNameMaxLen        = 64
	kiroToolDescriptionMaxLen = 10237
)

var kiroToolNameInvalidChars = regexp.MustCompile(`[^a-zA-Z0-9_\-]`)

// KiroToolNameRegistry 双向工具名映射，线程安全。
//
// 单个 Registry 实例服务于一次请求 / 一个会话即可，不需要全局共享。
type KiroToolNameRegistry struct {
	mu  sync.RWMutex
	fwd map[string]string // 原名 → 编码名
	rev map[string]string // 编码名 → 原名
}

// NewKiroToolNameRegistry 创建一个空的注册表
func NewKiroToolNameRegistry() *KiroToolNameRegistry {
	return &KiroToolNameRegistry{
		fwd: map[string]string{},
		rev: map[string]string{},
	}
}

// Encode 把一个原始工具名编码为 Kiro 上游可接受的名字
//
// 规则：
//  1. 已经在 fwd 缓存里直接返回
//  2. 全部字符在 [a-zA-Z0-9_-] 且长度 ≤64：原样返回
//  3. 否则：把非法字符替换为 _，再保留前 N 字符 + "_" + FNV-1a 32位 base36 哈希；
//     最终长度严格 ≤64
//  4. 同名碰撞：再加哈希后缀直到唯一
func (r *KiroToolNameRegistry) Encode(original string) string {
	r.mu.Lock()
	defer r.mu.Unlock()

	if v, ok := r.fwd[original]; ok {
		return v
	}

	encoded := encodeKiroToolName(original)

	// 罕见但要处理：编码后碰撞
	if existing, ok := r.rev[encoded]; ok && existing != original {
		// 给 encoded 再加一层哈希后缀去重
		encoded = appendHashSuffix(encoded, original+"|salt")
	}

	r.fwd[original] = encoded
	r.rev[encoded] = original
	return encoded
}

// Decode 把 Kiro 返回的工具名还原成原始名；缺失时返回 (encoded, false)
func (r *KiroToolNameRegistry) Decode(encoded string) (string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	v, ok := r.rev[encoded]
	return v, ok
}

// encodeKiroToolName 是 Encode 的纯函数实现，便于测试
func encodeKiroToolName(name string) string {
	if name == "" {
		return ""
	}
	if isValidKiroToolName(name) {
		return name
	}

	// 替换非法字符
	cleaned := kiroToolNameInvalidChars.ReplaceAllString(name, "_")
	suffix := "_" + fnvBase36(name)
	prefix := cleaned
	if l := kiroToolNameMaxLen - len(suffix); l > 0 && len(prefix) > l {
		prefix = prefix[:l]
	} else if l <= 0 {
		// suffix 自己就 ≥64，几乎不可能；兜底直接返回前 64
		if len(suffix) > kiroToolNameMaxLen {
			return suffix[:kiroToolNameMaxLen]
		}
		return suffix
	}
	return prefix + suffix
}

// isValidKiroToolName 长度 ≤64 且字符集 [a-zA-Z0-9_-]
func isValidKiroToolName(s string) bool {
	if len(s) == 0 || len(s) > kiroToolNameMaxLen {
		return false
	}
	for _, r := range s {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-') {
			return false
		}
	}
	return true
}

// fnvBase36 FNV-1a 32 位哈希再转 base36 字符串（小写字母数字）
//
// 与 TypeScript 实现保持一致：(name) => fnv1a32(name).toString(36)
func fnvBase36(s string) string {
	h := fnv.New32a()
	_, _ = h.Write([]byte(s))
	return strconv.FormatUint(uint64(h.Sum32()), 36)
}

// appendHashSuffix 用于碰撞场景：在已有 encoded 上再叠一层哈希
func appendHashSuffix(encoded, salt string) string {
	suffix := "_" + fnvBase36(salt)
	if len(encoded)+len(suffix) <= kiroToolNameMaxLen {
		return encoded + suffix
	}
	keep := kiroToolNameMaxLen - len(suffix)
	if keep <= 0 {
		return suffix[:kiroToolNameMaxLen]
	}
	return encoded[:keep] + suffix
}

// TruncateKiroToolDescription Kiro 对 description 长度的限制：>10237 截断 + "..."
func TruncateKiroToolDescription(desc string) string {
	if len(desc) <= kiroToolDescriptionMaxLen {
		return desc
	}
	return desc[:kiroToolDescriptionMaxLen] + "..."
}

// 用于把 fmt.Sprintf 风格的调试字符串挂在 panic 上，便于排错
func kiroToolDebugLabel(name, encoded string) string {
	if name == encoded {
		return name
	}
	return fmt.Sprintf("%s→%s", name, encoded)
}

// 由编辑期防御性使用：保证 fmt 引用，避免 lint 抱怨
var _ = strings.Builder{}
var _ = kiroToolDebugLabel
