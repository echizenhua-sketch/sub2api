package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	openAIWSResponseAccountCachePrefix = "openai:response:"
	openAIWSStateStoreCleanupInterval  = time.Minute
	openAIWSStateStoreCleanupMaxPerMap = 512
	openAIWSStateStoreMaxEntriesPerMap = 65536
	openAIWSStateStoreRedisTimeout     = 3 * time.Second
)

type openAIWSAccountBinding struct {
	accountID int64
	expiresAt time.Time
}

type openAIWSConnBinding struct {
	connID    string
	expiresAt time.Time
}

type openAIWSTurnStateBinding struct {
	turnState string
	expiresAt time.Time
}

type openAIWSSessionConnBinding struct {
	connID    string
	expiresAt time.Time
}

type openAIResponseContinuationBinding struct {
	state     OpenAIResponseContinuation
	expiresAt time.Time
	ambiguous bool
}

// OpenAIResponseStateStore 管理跨 HTTP/WS Responses 传输共用的续链状态。
// - response_id -> account_id 用于续链路由
// - response_id -> conn_id 用于连接内上下文复用
//
// response_id -> account_id 优先走 GatewayCache（Redis），同时维护本地热缓存。
// response_id -> conn_id 仅在本进程内有效。
type OpenAIResponseStateStore interface {
	BindResponseAccount(ctx context.Context, groupID int64, responseID string, accountID int64, ttl time.Duration) error
	GetResponseAccount(ctx context.Context, groupID int64, responseID string) (int64, error)
	DeleteResponseAccount(ctx context.Context, groupID int64, responseID string) error

	BindResponseConn(responseID, connID string, ttl time.Duration)
	GetResponseConn(responseID string) (string, bool)
	DeleteResponseConn(responseID string)

	BindSessionTurnState(groupID int64, sessionHash, turnState string, ttl time.Duration)
	GetSessionTurnState(groupID int64, sessionHash string) (string, bool)
	DeleteSessionTurnState(groupID int64, sessionHash string)

	BindSessionConn(groupID int64, sessionHash, connID string, ttl time.Duration)
	GetSessionConn(groupID int64, sessionHash string) (string, bool)
	DeleteSessionConn(groupID int64, sessionHash string)

	BindResponseContinuation(scope OpenAIResponseContinuationScope, sessionHash string, state OpenAIResponseContinuation, ttl time.Duration)
	GetResponseContinuation(scope OpenAIResponseContinuationScope, responseID string) (OpenAIResponseContinuation, bool)
	GetSessionContinuation(scope OpenAIResponseContinuationScope, sessionHash string) (OpenAIResponseContinuation, bool)
	GetCallIDContinuation(scope OpenAIResponseContinuationScope, callIDs []string) (OpenAIResponseContinuation, bool, error)
}

// OpenAIWSStateStore remains as a compatibility alias for existing WS callers.
type OpenAIWSStateStore = OpenAIResponseStateStore

type defaultOpenAIWSStateStore struct {
	cache GatewayCache

	responseToAccountMu   sync.RWMutex
	responseToAccount     map[string]openAIWSAccountBinding
	responseToConnMu      sync.RWMutex
	responseToConn        map[string]openAIWSConnBinding
	sessionToTurnStateMu  sync.RWMutex
	sessionToTurnState    map[string]openAIWSTurnStateBinding
	sessionToConnMu       sync.RWMutex
	sessionToConn         map[string]openAIWSSessionConnBinding
	continuationMu        sync.RWMutex
	responseContinuations map[string]openAIResponseContinuationBinding
	sessionContinuations  map[string]openAIResponseContinuationBinding
	callIDContinuations   map[string]openAIResponseContinuationBinding

	lastCleanupUnixNano atomic.Int64
}

// NewOpenAIWSStateStore 创建默认 WS 状态存储。
func NewOpenAIResponseStateStore(cache GatewayCache) OpenAIResponseStateStore {
	store := &defaultOpenAIWSStateStore{
		cache:                 cache,
		responseToAccount:     make(map[string]openAIWSAccountBinding, 256),
		responseToConn:        make(map[string]openAIWSConnBinding, 256),
		sessionToTurnState:    make(map[string]openAIWSTurnStateBinding, 256),
		sessionToConn:         make(map[string]openAIWSSessionConnBinding, 256),
		responseContinuations: make(map[string]openAIResponseContinuationBinding, 256),
		sessionContinuations:  make(map[string]openAIResponseContinuationBinding, 256),
		callIDContinuations:   make(map[string]openAIResponseContinuationBinding, 256),
	}
	store.lastCleanupUnixNano.Store(time.Now().UnixNano())
	return store
}

func NewOpenAIWSStateStore(cache GatewayCache) OpenAIWSStateStore {
	return NewOpenAIResponseStateStore(cache)
}

func (s *defaultOpenAIWSStateStore) BindResponseAccount(ctx context.Context, groupID int64, responseID string, accountID int64, ttl time.Duration) error {
	id := normalizeOpenAIWSResponseID(responseID)
	if id == "" || accountID <= 0 {
		return nil
	}
	ttl = normalizeOpenAIWSTTL(ttl)
	s.maybeCleanup()

	expiresAt := time.Now().Add(ttl)
	mapKey := openAIWSResponseAccountMapKey(groupID, id)
	s.responseToAccountMu.Lock()
	ensureBindingCapacity(s.responseToAccount, mapKey, openAIWSStateStoreMaxEntriesPerMap)
	s.responseToAccount[mapKey] = openAIWSAccountBinding{accountID: accountID, expiresAt: expiresAt}
	s.responseToAccountMu.Unlock()

	if s.cache == nil {
		return nil
	}
	cacheKey := openAIWSResponseAccountCacheKey(id)
	cacheCtx, cancel := withOpenAIWSStateStoreRedisTimeout(ctx)
	defer cancel()
	return s.cache.SetSessionAccountID(cacheCtx, groupID, cacheKey, accountID, ttl)
}

func (s *defaultOpenAIWSStateStore) GetResponseAccount(ctx context.Context, groupID int64, responseID string) (int64, error) {
	id := normalizeOpenAIWSResponseID(responseID)
	if id == "" {
		return 0, nil
	}
	s.maybeCleanup()

	now := time.Now()
	mapKey := openAIWSResponseAccountMapKey(groupID, id)
	s.responseToAccountMu.RLock()
	if binding, ok := s.responseToAccount[mapKey]; ok {
		if now.Before(binding.expiresAt) {
			accountID := binding.accountID
			s.responseToAccountMu.RUnlock()
			return accountID, nil
		}
	}
	s.responseToAccountMu.RUnlock()

	if s.cache == nil {
		return 0, nil
	}

	cacheKey := openAIWSResponseAccountCacheKey(id)
	cacheCtx, cancel := withOpenAIWSStateStoreRedisTimeout(ctx)
	defer cancel()
	accountID, err := s.cache.GetSessionAccountID(cacheCtx, groupID, cacheKey)
	if err != nil || accountID <= 0 {
		// 缓存读取失败不阻断主流程，按未命中降级。
		return 0, nil
	}
	return accountID, nil
}

func (s *defaultOpenAIWSStateStore) DeleteResponseAccount(ctx context.Context, groupID int64, responseID string) error {
	id := normalizeOpenAIWSResponseID(responseID)
	if id == "" {
		return nil
	}
	s.responseToAccountMu.Lock()
	delete(s.responseToAccount, openAIWSResponseAccountMapKey(groupID, id))
	s.responseToAccountMu.Unlock()

	if s.cache == nil {
		return nil
	}
	cacheCtx, cancel := withOpenAIWSStateStoreRedisTimeout(ctx)
	defer cancel()
	return s.cache.DeleteSessionAccountID(cacheCtx, groupID, openAIWSResponseAccountCacheKey(id))
}

func (s *defaultOpenAIWSStateStore) BindResponseConn(responseID, connID string, ttl time.Duration) {
	id := normalizeOpenAIWSResponseID(responseID)
	conn := strings.TrimSpace(connID)
	if id == "" || conn == "" {
		return
	}
	ttl = normalizeOpenAIWSTTL(ttl)
	s.maybeCleanup()

	s.responseToConnMu.Lock()
	ensureBindingCapacity(s.responseToConn, id, openAIWSStateStoreMaxEntriesPerMap)
	s.responseToConn[id] = openAIWSConnBinding{
		connID:    conn,
		expiresAt: time.Now().Add(ttl),
	}
	s.responseToConnMu.Unlock()
}

func (s *defaultOpenAIWSStateStore) GetResponseConn(responseID string) (string, bool) {
	id := normalizeOpenAIWSResponseID(responseID)
	if id == "" {
		return "", false
	}
	s.maybeCleanup()

	now := time.Now()
	s.responseToConnMu.RLock()
	binding, ok := s.responseToConn[id]
	s.responseToConnMu.RUnlock()
	if !ok || now.After(binding.expiresAt) || strings.TrimSpace(binding.connID) == "" {
		return "", false
	}
	return binding.connID, true
}

func (s *defaultOpenAIWSStateStore) DeleteResponseConn(responseID string) {
	id := normalizeOpenAIWSResponseID(responseID)
	if id == "" {
		return
	}
	s.responseToConnMu.Lock()
	delete(s.responseToConn, id)
	s.responseToConnMu.Unlock()
}

func (s *defaultOpenAIWSStateStore) BindSessionTurnState(groupID int64, sessionHash, turnState string, ttl time.Duration) {
	key := openAIWSSessionTurnStateKey(groupID, sessionHash)
	state := strings.TrimSpace(turnState)
	if key == "" || state == "" {
		return
	}
	ttl = normalizeOpenAIWSTTL(ttl)
	s.maybeCleanup()

	s.sessionToTurnStateMu.Lock()
	ensureBindingCapacity(s.sessionToTurnState, key, openAIWSStateStoreMaxEntriesPerMap)
	s.sessionToTurnState[key] = openAIWSTurnStateBinding{
		turnState: state,
		expiresAt: time.Now().Add(ttl),
	}
	s.sessionToTurnStateMu.Unlock()
}

func (s *defaultOpenAIWSStateStore) GetSessionTurnState(groupID int64, sessionHash string) (string, bool) {
	key := openAIWSSessionTurnStateKey(groupID, sessionHash)
	if key == "" {
		return "", false
	}
	s.maybeCleanup()

	now := time.Now()
	s.sessionToTurnStateMu.RLock()
	binding, ok := s.sessionToTurnState[key]
	s.sessionToTurnStateMu.RUnlock()
	if !ok || now.After(binding.expiresAt) || strings.TrimSpace(binding.turnState) == "" {
		return "", false
	}
	return binding.turnState, true
}

func (s *defaultOpenAIWSStateStore) DeleteSessionTurnState(groupID int64, sessionHash string) {
	key := openAIWSSessionTurnStateKey(groupID, sessionHash)
	if key == "" {
		return
	}
	s.sessionToTurnStateMu.Lock()
	delete(s.sessionToTurnState, key)
	s.sessionToTurnStateMu.Unlock()
}

func (s *defaultOpenAIWSStateStore) BindSessionConn(groupID int64, sessionHash, connID string, ttl time.Duration) {
	key := openAIWSSessionTurnStateKey(groupID, sessionHash)
	conn := strings.TrimSpace(connID)
	if key == "" || conn == "" {
		return
	}
	ttl = normalizeOpenAIWSTTL(ttl)
	s.maybeCleanup()

	s.sessionToConnMu.Lock()
	ensureBindingCapacity(s.sessionToConn, key, openAIWSStateStoreMaxEntriesPerMap)
	s.sessionToConn[key] = openAIWSSessionConnBinding{
		connID:    conn,
		expiresAt: time.Now().Add(ttl),
	}
	s.sessionToConnMu.Unlock()
}

func (s *defaultOpenAIWSStateStore) GetSessionConn(groupID int64, sessionHash string) (string, bool) {
	key := openAIWSSessionTurnStateKey(groupID, sessionHash)
	if key == "" {
		return "", false
	}
	s.maybeCleanup()

	now := time.Now()
	s.sessionToConnMu.RLock()
	binding, ok := s.sessionToConn[key]
	s.sessionToConnMu.RUnlock()
	if !ok || now.After(binding.expiresAt) || strings.TrimSpace(binding.connID) == "" {
		return "", false
	}
	return binding.connID, true
}

func (s *defaultOpenAIWSStateStore) DeleteSessionConn(groupID int64, sessionHash string) {
	key := openAIWSSessionTurnStateKey(groupID, sessionHash)
	if key == "" {
		return
	}
	s.sessionToConnMu.Lock()
	delete(s.sessionToConn, key)
	s.sessionToConnMu.Unlock()
}

func (s *defaultOpenAIWSStateStore) BindResponseContinuation(scope OpenAIResponseContinuationScope, sessionHash string, state OpenAIResponseContinuation, ttl time.Duration) {
	if !scope.valid() || state.AccountID <= 0 || len(state.ReplayInput) == 0 {
		return
	}
	state.ResponseID = strings.TrimSpace(state.ResponseID)
	state.ReplayInput = cloneOpenAIReplayInput(state.ReplayInput)
	if len(state.ReplayInput) == 0 {
		return
	}
	ttl = normalizeOpenAIWSTTL(ttl)
	s.maybeCleanup()
	binding := openAIResponseContinuationBinding{state: state, expiresAt: time.Now().Add(ttl)}

	s.continuationMu.Lock()
	if state.ResponseID != "" {
		key := openAIResponseContinuationKey(scope, state.ResponseID)
		ensureBindingCapacity(s.responseContinuations, key, openAIWSStateStoreMaxEntriesPerMap)
		s.responseContinuations[key] = binding
	}
	if key := openAISessionContinuationKey(scope, sessionHash); key != "" {
		ensureBindingCapacity(s.sessionContinuations, key, openAIWSStateStoreMaxEntriesPerMap)
		s.sessionContinuations[key] = binding
	}
	if state.ResponseID != "" {
		now := time.Now()
		for _, item := range state.ReplayInput {
			callID := openAIContinuationContextCallID(item)
			if callID == "" {
				continue
			}
			key := openAICallIDContinuationKey(scope, callID)
			ensureBindingCapacity(s.callIDContinuations, key, openAIWSStateStoreMaxEntriesPerMap)
			next := binding
			if existing, exists := s.callIDContinuations[key]; exists && now.Before(existing.expiresAt) {
				if existing.ambiguous || !sameOpenAIResponseContinuation(existing.state, state) {
					next.ambiguous = true
					if existing.expiresAt.After(next.expiresAt) {
						next.expiresAt = existing.expiresAt
					}
				}
			}
			s.callIDContinuations[key] = next
		}
	}
	s.continuationMu.Unlock()
}

func (s *defaultOpenAIWSStateStore) GetResponseContinuation(scope OpenAIResponseContinuationScope, responseID string) (OpenAIResponseContinuation, bool) {
	if !scope.valid() || strings.TrimSpace(responseID) == "" {
		return OpenAIResponseContinuation{}, false
	}
	s.maybeCleanup()
	return s.getResponseContinuation(openAIResponseContinuationKey(scope, responseID))
}

func (s *defaultOpenAIWSStateStore) GetSessionContinuation(scope OpenAIResponseContinuationScope, sessionHash string) (OpenAIResponseContinuation, bool) {
	key := openAISessionContinuationKey(scope, sessionHash)
	if key == "" {
		return OpenAIResponseContinuation{}, false
	}
	s.maybeCleanup()
	return s.getResponseContinuation(key)
}

func (s *defaultOpenAIWSStateStore) GetCallIDContinuation(scope OpenAIResponseContinuationScope, callIDs []string) (OpenAIResponseContinuation, bool, error) {
	if !scope.valid() || len(callIDs) == 0 {
		return OpenAIResponseContinuation{}, false, nil
	}
	s.maybeCleanup()

	now := time.Now()
	seen := make(map[string]struct{}, len(callIDs))
	found := 0
	missing := 0
	var selected OpenAIResponseContinuation

	s.continuationMu.RLock()
	for _, rawCallID := range callIDs {
		callID := strings.TrimSpace(rawCallID)
		if callID == "" {
			continue
		}
		if _, exists := seen[callID]; exists {
			continue
		}
		seen[callID] = struct{}{}

		binding, ok := s.callIDContinuations[openAICallIDContinuationKey(scope, callID)]
		if !ok || !now.Before(binding.expiresAt) {
			missing++
			continue
		}
		if binding.ambiguous {
			s.continuationMu.RUnlock()
			return OpenAIResponseContinuation{}, false, fmt.Errorf("call_id continuation context is ambiguous")
		}
		if found == 0 {
			selected = binding.state
		} else if !sameOpenAIResponseContinuation(selected, binding.state) {
			s.continuationMu.RUnlock()
			return OpenAIResponseContinuation{}, false, fmt.Errorf("call_ids resolve to different prior responses")
		}
		found++
	}
	s.continuationMu.RUnlock()

	if found == 0 {
		return OpenAIResponseContinuation{}, false, nil
	}
	if missing > 0 {
		return OpenAIResponseContinuation{}, false, fmt.Errorf("call_ids did not all resolve to prior response context")
	}
	selected.ReplayInput = cloneOpenAIReplayInput(selected.ReplayInput)
	return selected, true, nil
}

func sameOpenAIResponseContinuation(a, b OpenAIResponseContinuation) bool {
	return a.AccountID > 0 && a.AccountID == b.AccountID &&
		strings.TrimSpace(a.ResponseID) != "" && strings.TrimSpace(a.ResponseID) == strings.TrimSpace(b.ResponseID)
}

func (s *defaultOpenAIWSStateStore) getResponseContinuation(key string) (OpenAIResponseContinuation, bool) {
	now := time.Now()
	s.continuationMu.RLock()
	binding, ok := s.responseContinuations[key]
	if !ok {
		binding, ok = s.sessionContinuations[key]
	}
	s.continuationMu.RUnlock()
	if !ok || !now.Before(binding.expiresAt) {
		return OpenAIResponseContinuation{}, false
	}
	state := binding.state
	state.ReplayInput = cloneOpenAIReplayInput(state.ReplayInput)
	return state, true
}

func (s *defaultOpenAIWSStateStore) maybeCleanup() {
	if s == nil {
		return
	}
	now := time.Now()
	last := time.Unix(0, s.lastCleanupUnixNano.Load())
	if now.Sub(last) < openAIWSStateStoreCleanupInterval {
		return
	}
	if !s.lastCleanupUnixNano.CompareAndSwap(last.UnixNano(), now.UnixNano()) {
		return
	}

	// 增量限额清理，避免高规模下一次性全量扫描导致长时间阻塞。
	s.responseToAccountMu.Lock()
	cleanupExpiredAccountBindings(s.responseToAccount, now, openAIWSStateStoreCleanupMaxPerMap)
	s.responseToAccountMu.Unlock()

	s.responseToConnMu.Lock()
	cleanupExpiredConnBindings(s.responseToConn, now, openAIWSStateStoreCleanupMaxPerMap)
	s.responseToConnMu.Unlock()

	s.sessionToTurnStateMu.Lock()
	cleanupExpiredTurnStateBindings(s.sessionToTurnState, now, openAIWSStateStoreCleanupMaxPerMap)
	s.sessionToTurnStateMu.Unlock()

	s.sessionToConnMu.Lock()
	cleanupExpiredSessionConnBindings(s.sessionToConn, now, openAIWSStateStoreCleanupMaxPerMap)
	s.sessionToConnMu.Unlock()

	s.continuationMu.Lock()
	cleanupExpiredResponseContinuationBindings(s.responseContinuations, now, openAIWSStateStoreCleanupMaxPerMap)
	cleanupExpiredResponseContinuationBindings(s.sessionContinuations, now, openAIWSStateStoreCleanupMaxPerMap)
	cleanupExpiredResponseContinuationBindings(s.callIDContinuations, now, openAIWSStateStoreCleanupMaxPerMap)
	s.continuationMu.Unlock()
}

func cleanupExpiredAccountBindings(bindings map[string]openAIWSAccountBinding, now time.Time, maxScan int) {
	if len(bindings) == 0 || maxScan <= 0 {
		return
	}
	scanned := 0
	for key, binding := range bindings {
		if now.After(binding.expiresAt) {
			delete(bindings, key)
		}
		scanned++
		if scanned >= maxScan {
			break
		}
	}
}

func cleanupExpiredConnBindings(bindings map[string]openAIWSConnBinding, now time.Time, maxScan int) {
	if len(bindings) == 0 || maxScan <= 0 {
		return
	}
	scanned := 0
	for key, binding := range bindings {
		if now.After(binding.expiresAt) {
			delete(bindings, key)
		}
		scanned++
		if scanned >= maxScan {
			break
		}
	}
}

func cleanupExpiredTurnStateBindings(bindings map[string]openAIWSTurnStateBinding, now time.Time, maxScan int) {
	if len(bindings) == 0 || maxScan <= 0 {
		return
	}
	scanned := 0
	for key, binding := range bindings {
		if now.After(binding.expiresAt) {
			delete(bindings, key)
		}
		scanned++
		if scanned >= maxScan {
			break
		}
	}
}

func cleanupExpiredSessionConnBindings(bindings map[string]openAIWSSessionConnBinding, now time.Time, maxScan int) {
	if len(bindings) == 0 || maxScan <= 0 {
		return
	}
	scanned := 0
	for key, binding := range bindings {
		if now.After(binding.expiresAt) {
			delete(bindings, key)
		}
		scanned++
		if scanned >= maxScan {
			break
		}
	}
}

func cleanupExpiredResponseContinuationBindings(bindings map[string]openAIResponseContinuationBinding, now time.Time, maxScan int) {
	if len(bindings) == 0 || maxScan <= 0 {
		return
	}
	scanned := 0
	for key, binding := range bindings {
		if now.After(binding.expiresAt) {
			delete(bindings, key)
		}
		scanned++
		if scanned >= maxScan {
			break
		}
	}
}

func ensureBindingCapacity[T any](bindings map[string]T, incomingKey string, maxEntries int) {
	if len(bindings) < maxEntries || maxEntries <= 0 {
		return
	}
	if _, exists := bindings[incomingKey]; exists {
		return
	}
	// 固定上限保护：淘汰任意一项，优先保证内存有界。
	for key := range bindings {
		delete(bindings, key)
		return
	}
}

func normalizeOpenAIWSResponseID(responseID string) string {
	return strings.TrimSpace(responseID)
}

func openAIWSResponseAccountCacheKey(responseID string) string {
	sum := sha256.Sum256([]byte(responseID))
	return openAIWSResponseAccountCachePrefix + hex.EncodeToString(sum[:])
}

// openAIWSResponseAccountMapKey 本地热缓存按分组隔离的 key，与 Redis 层保持一致，避免跨组命中。
func openAIWSResponseAccountMapKey(groupID int64, responseID string) string {
	return fmt.Sprintf("%d:%s", groupID, responseID)
}

func normalizeOpenAIWSTTL(ttl time.Duration) time.Duration {
	if ttl <= 0 {
		return time.Hour
	}
	return ttl
}

func openAIWSSessionTurnStateKey(groupID int64, sessionHash string) string {
	hash := strings.TrimSpace(sessionHash)
	if hash == "" {
		return ""
	}
	return fmt.Sprintf("%d:%s", groupID, hash)
}

func openAIResponseContinuationScopeKey(scope OpenAIResponseContinuationScope) string {
	return fmt.Sprintf("%d:%d:%d", scope.GroupID, scope.APIKeyID, scope.UserID)
}

func openAIResponseContinuationKey(scope OpenAIResponseContinuationScope, responseID string) string {
	return "response:" + openAIResponseContinuationScopeKey(scope) + ":" + strings.TrimSpace(responseID)
}

func openAISessionContinuationKey(scope OpenAIResponseContinuationScope, sessionHash string) string {
	hash := strings.TrimSpace(sessionHash)
	if !scope.valid() || hash == "" {
		return ""
	}
	return "session:" + openAIResponseContinuationScopeKey(scope) + ":" + hash
}

func openAICallIDContinuationKey(scope OpenAIResponseContinuationScope, callID string) string {
	id := strings.TrimSpace(callID)
	if !scope.valid() || id == "" {
		return ""
	}
	return "call:" + openAIResponseContinuationScopeKey(scope) + ":" + id
}

func withOpenAIWSStateStoreRedisTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithTimeout(ctx, openAIWSStateStoreRedisTimeout)
}
