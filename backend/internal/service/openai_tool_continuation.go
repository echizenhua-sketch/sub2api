package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tidwall/gjson"
)

// ToolContinuationSignals 聚合工具续链相关信号，避免重复遍历 input。
type ToolContinuationSignals struct {
	HasFunctionCallOutput              bool
	HasFunctionCallOutputMissingCallID bool
	HasToolCallContext                 bool
	HasItemReference                   bool
	HasItemReferenceForAllCallIDs      bool
	FunctionCallOutputCallIDs          []string
}

// FunctionCallOutputValidation 汇总 function_call_output 关联性校验结果。
type FunctionCallOutputValidation struct {
	HasFunctionCallOutput              bool
	HasToolCallContext                 bool
	HasFunctionCallOutputMissingCallID bool
	HasItemReference                   bool
	HasItemReferenceForAllCallIDs      bool
}

// OpenAIResponseContinuationScope isolates HTTP continuation state between tenants.
type OpenAIResponseContinuationScope struct {
	GroupID  int64
	APIKeyID int64
	UserID   int64
}

func (s OpenAIResponseContinuationScope) valid() bool {
	return s.APIKeyID > 0 && s.UserID > 0
}

// OpenAIResponseContinuation contains the real output items needed to replay a
// previous Responses turn without treating call_id as an item id.
type OpenAIResponseContinuation struct {
	ResponseID  string
	AccountID   int64
	ReplayInput []json.RawMessage
}

func isCodexToolCallContextItemType(typ string) bool {
	switch strings.TrimSpace(typ) {
	case "tool_call",
		"function_call",
		"local_shell_call",
		"tool_search_call",
		"custom_tool_call",
		"mcp_tool_call":
		return true
	default:
		return false
	}
}

func isCodexToolCallOutputItemType(typ string) bool {
	switch strings.TrimSpace(typ) {
	case "function_call_output",
		"tool_search_output",
		"custom_tool_call_output",
		"mcp_tool_call_output":
		return true
	default:
		return false
	}
}

// NeedsToolContinuation 判定请求是否需要工具调用续链处理。
// 满足以下任一信号即视为续链：previous_response_id、input 内包含工具输出/item_reference、
// 或显式声明 tools/tool_choice。
func NeedsToolContinuation(reqBody map[string]any) bool {
	if reqBody == nil {
		return false
	}
	if hasNonEmptyString(reqBody["previous_response_id"]) {
		return true
	}
	if hasToolsSignal(reqBody) {
		return true
	}
	if hasToolChoiceSignal(reqBody) {
		return true
	}
	input, ok := reqBody["input"].([]any)
	if !ok {
		return false
	}
	for _, item := range input {
		itemMap, ok := item.(map[string]any)
		if !ok {
			continue
		}
		itemType, _ := itemMap["type"].(string)
		if isCodexToolCallItemType(itemType) || itemType == "item_reference" {
			return true
		}
	}
	return false
}

// AnalyzeToolContinuationSignals 单次遍历 input，提取工具输出/工具调用上下文/item_reference 相关信号。
// 字段名保留 FunctionCallOutput 是为了兼容既有调用点；语义覆盖 Codex 的所有工具输出
// （function_call_output/tool_search_output/custom_tool_call_output/mcp_tool_call_output）。
func AnalyzeToolContinuationSignals(reqBody map[string]any) ToolContinuationSignals {
	signals := ToolContinuationSignals{}
	if reqBody == nil {
		return signals
	}
	input, ok := reqBody["input"].([]any)
	if !ok {
		return signals
	}

	var callIDs map[string]struct{}
	var referenceIDs map[string]struct{}

	for _, item := range input {
		itemMap, ok := item.(map[string]any)
		if !ok {
			continue
		}
		itemType, _ := itemMap["type"].(string)
		switch {
		case isCodexToolCallContextItemType(itemType):
			callID, _ := itemMap["call_id"].(string)
			if strings.TrimSpace(callID) != "" {
				signals.HasToolCallContext = true
			}
		case isCodexToolCallOutputItemType(itemType):
			signals.HasFunctionCallOutput = true
			callID, _ := itemMap["call_id"].(string)
			callID = strings.TrimSpace(callID)
			if callID == "" {
				signals.HasFunctionCallOutputMissingCallID = true
				continue
			}
			if callIDs == nil {
				callIDs = make(map[string]struct{})
			}
			callIDs[callID] = struct{}{}
		case itemType == "item_reference":
			signals.HasItemReference = true
			idValue, _ := itemMap["id"].(string)
			idValue = strings.TrimSpace(idValue)
			if idValue == "" {
				continue
			}
			if referenceIDs == nil {
				referenceIDs = make(map[string]struct{})
			}
			referenceIDs[idValue] = struct{}{}
		}
	}

	if len(callIDs) == 0 {
		return signals
	}
	signals.FunctionCallOutputCallIDs = make([]string, 0, len(callIDs))
	allReferenced := len(referenceIDs) > 0
	for callID := range callIDs {
		signals.FunctionCallOutputCallIDs = append(signals.FunctionCallOutputCallIDs, callID)
		if allReferenced {
			if _, ok := referenceIDs[callID]; !ok {
				allReferenced = false
			}
		}
	}
	signals.HasItemReferenceForAllCallIDs = allReferenced
	return signals
}

// ValidateFunctionCallOutputContextBytes 基于 raw JSON 校验工具输出续链，避免 handler 预校验阶段全量解码大 input。
func ValidateFunctionCallOutputContextBytes(body []byte) FunctionCallOutputValidation {
	result := FunctionCallOutputValidation{}
	if len(body) == 0 {
		return result
	}
	// handler 热路径只读扫描 input，避免 GetBytes 为大 Responses body 复制整段 JSON。
	input := parseRawJSONView(body).Get("input")
	if !input.IsArray() {
		return result
	}

	var callIDs map[string]struct{}
	var referenceIDs map[string]struct{}
	input.ForEach(func(_, item gjson.Result) bool {
		if !item.IsObject() {
			return true
		}
		itemType := item.Get("type").String()
		switch {
		case isCodexToolCallOutputItemType(itemType):
			result.HasFunctionCallOutput = true
			callID := strings.TrimSpace(item.Get("call_id").String())
			if callID == "" {
				result.HasFunctionCallOutputMissingCallID = true
				return true
			}
			if callIDs == nil {
				callIDs = make(map[string]struct{})
			}
			callIDs[callID] = struct{}{}
		case isCodexToolCallContextItemType(itemType):
			if strings.TrimSpace(item.Get("call_id").String()) != "" {
				result.HasToolCallContext = true
			}
		case itemType == "item_reference":
			result.HasItemReference = true
			idValue := strings.TrimSpace(item.Get("id").String())
			if idValue == "" {
				return true
			}
			if referenceIDs == nil {
				referenceIDs = make(map[string]struct{})
			}
			referenceIDs[idValue] = struct{}{}
		}
		return !result.HasFunctionCallOutput || !result.HasToolCallContext
	})
	if !result.HasFunctionCallOutput || result.HasToolCallContext || len(callIDs) == 0 || len(referenceIDs) == 0 {
		return result
	}
	allReferenced := true
	for callID := range callIDs {
		if _, ok := referenceIDs[callID]; !ok {
			allReferenced = false
			break
		}
	}
	result.HasItemReferenceForAllCallIDs = allReferenced
	return result
}

func functionCallOutputCallIDsBytes(body []byte) []string {
	if len(body) == 0 {
		return nil
	}
	input := parseRawJSONView(body).Get("input")
	if !input.IsArray() {
		return nil
	}

	var callIDs []string
	seen := make(map[string]struct{})
	input.ForEach(func(_, item gjson.Result) bool {
		if !item.IsObject() || !isCodexToolCallOutputItemType(item.Get("type").String()) {
			return true
		}
		callID := strings.TrimSpace(item.Get("call_id").String())
		if callID == "" {
			return true
		}
		if _, exists := seen[callID]; exists {
			return true
		}
		seen[callID] = struct{}{}
		callIDs = append(callIDs, callID)
		return true
	})
	return callIDs
}

// RewriteOpenAIHTTPContinuationRequest replaces previous_response_id with the
// actual prior output items. This keeps HTTP forwarding stateless from the
// upstream's perspective and preserves the distinct item id and call_id fields.
func RewriteOpenAIHTTPContinuationRequest(body []byte, state OpenAIResponseContinuation) ([]byte, error) {
	if len(body) == 0 || !json.Valid(body) {
		return nil, fmt.Errorf("invalid Responses request body")
	}
	if len(state.ReplayInput) == 0 {
		return nil, fmt.Errorf("previous_response_id context has no replayable output")
	}

	var request map[string]json.RawMessage
	if err := json.Unmarshal(body, &request); err != nil {
		return nil, fmt.Errorf("decode Responses request: %w", err)
	}
	current, err := decodeOpenAIContinuationInput(request["input"])
	if err != nil {
		return nil, err
	}

	replay := cloneOpenAIReplayInput(state.ReplayInput)
	contextCallIDs := make(map[string]struct{})
	for _, item := range replay {
		if callID := openAIContinuationContextCallID(item); callID != "" {
			contextCallIDs[callID] = struct{}{}
		}
	}
	for _, item := range current {
		if callID := openAIContinuationContextCallID(item); callID != "" {
			contextCallIDs[callID] = struct{}{}
		}
	}
	for _, item := range current {
		itemType := strings.TrimSpace(gjson.GetBytes(item, "type").String())
		if !isCodexToolCallOutputItemType(itemType) {
			continue
		}
		callID := strings.TrimSpace(gjson.GetBytes(item, "call_id").String())
		if callID == "" {
			return nil, fmt.Errorf("%s requires a non-empty call_id", itemType)
		}
		if _, ok := contextCallIDs[callID]; !ok {
			return nil, fmt.Errorf("previous response context does not contain tool call %s", callID)
		}
	}

	seenIDs := make(map[string]struct{}, len(current))
	for _, item := range current {
		if id := strings.TrimSpace(gjson.GetBytes(item, "id").String()); id != "" {
			seenIDs[id] = struct{}{}
		}
	}
	combined := make([]json.RawMessage, 0, len(replay)+len(current))
	for _, item := range replay {
		id := strings.TrimSpace(gjson.GetBytes(item, "id").String())
		if id != "" {
			if _, exists := seenIDs[id]; exists {
				continue
			}
			seenIDs[id] = struct{}{}
		}
		combined = append(combined, append(json.RawMessage(nil), item...))
	}
	combined = append(combined, current...)
	encodedInput, err := json.Marshal(combined)
	if err != nil {
		return nil, fmt.Errorf("encode continuation input: %w", err)
	}
	request["input"] = encodedInput
	delete(request, "previous_response_id")
	rewritten, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("encode Responses continuation request: %w", err)
	}
	return rewritten, nil
}

func decodeOpenAIContinuationInput(raw json.RawMessage) ([]json.RawMessage, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}
	var items []json.RawMessage
	if err := json.Unmarshal(raw, &items); err == nil {
		return items, nil
	}
	var text string
	if err := json.Unmarshal(raw, &text); err != nil {
		return nil, fmt.Errorf("Responses continuation input must be a string or array")
	}
	message, err := json.Marshal(map[string]any{
		"type":    "message",
		"role":    "user",
		"content": []any{map[string]any{"type": "input_text", "text": text}},
	})
	if err != nil {
		return nil, err
	}
	return []json.RawMessage{message}, nil
}

func openAIContinuationContextCallID(item json.RawMessage) string {
	itemType := strings.TrimSpace(gjson.GetBytes(item, "type").String())
	if !isCodexToolCallContextItemType(itemType) {
		return ""
	}
	return strings.TrimSpace(gjson.GetBytes(item, "call_id").String())
}

// ExtractOpenAIResponseReplayInput reads complete output arrays and individual
// output_item events. The caller may feed every SSE payload to a collector.
func ExtractOpenAIResponseReplayInput(payload []byte) []json.RawMessage {
	if len(payload) == 0 || !gjson.ValidBytes(payload) {
		return nil
	}
	for _, path := range []string{"output", "response.output"} {
		output := gjson.GetBytes(payload, path)
		if !output.IsArray() {
			continue
		}
		items := make([]json.RawMessage, 0, len(output.Array()))
		for _, item := range output.Array() {
			if item.IsObject() {
				items = append(items, json.RawMessage(append([]byte(nil), item.Raw...)))
			}
		}
		if len(items) > 0 {
			return items
		}
	}
	item := gjson.GetBytes(payload, "item")
	if item.IsObject() {
		return []json.RawMessage{json.RawMessage(append([]byte(nil), item.Raw...))}
	}
	return nil
}

type openAIResponseReplayCollector struct {
	items []json.RawMessage
	index map[string]int
}

func (c *openAIResponseReplayCollector) AddPayload(payload []byte) {
	for _, item := range ExtractOpenAIResponseReplayInput(payload) {
		key := openAIReplayItemKey(item)
		if key != "" {
			if c.index == nil {
				c.index = make(map[string]int)
			}
			if index, ok := c.index[key]; ok {
				c.items[index] = append(json.RawMessage(nil), item...)
				continue
			}
			c.index[key] = len(c.items)
		}
		c.items = append(c.items, append(json.RawMessage(nil), item...))
	}
}

func (c *openAIResponseReplayCollector) Items() []json.RawMessage {
	return cloneOpenAIReplayInput(c.items)
}

func openAIReplayItemKey(item json.RawMessage) string {
	if id := strings.TrimSpace(gjson.GetBytes(item, "id").String()); id != "" {
		return "id:" + id
	}
	typ := strings.TrimSpace(gjson.GetBytes(item, "type").String())
	callID := strings.TrimSpace(gjson.GetBytes(item, "call_id").String())
	if typ != "" && callID != "" {
		return "call:" + typ + ":" + callID
	}
	return ""
}

func cloneOpenAIReplayInput(items []json.RawMessage) []json.RawMessage {
	if len(items) == 0 {
		return nil
	}
	cloned := make([]json.RawMessage, 0, len(items))
	for _, item := range items {
		if len(item) == 0 || !json.Valid(item) {
			continue
		}
		cloned = append(cloned, append(json.RawMessage(nil), item...))
	}
	return cloned
}

// ToolCallOutputContextCoverage 描述 input 中工具输出与可重建上下文的覆盖关系，
// 用于判断剥离 previous_response_id 后上游能否仅凭 input 重建工具续链。
type ToolCallOutputContextCoverage struct {
	HasFunctionCallOutput bool
	// ContextCoversAllCallIDs 表示每个工具输出的 call_id 都能在 input 内找到
	// 同 call_id 的工具调用上下文项或同 id 的 item_reference，且不存在缺失 call_id 的输出。
	// 任一输出无法由 input 自身重建时为 false，此时剥离 previous_response_id 会导致
	// 上游以 "No tool call found for function call output" 拒绝请求。
	ContextCoversAllCallIDs bool
}

// AnalyzeToolCallOutputContextCoverageBytes 全量扫描 input，按 call_id 精确匹配工具输出
// 与可重建上下文。不能复用 ValidateFunctionCallOutputContextBytes 的 HasToolCallContext：
// 该标志只代表"存在某一个上下文项"，部分覆盖的续链仍会被上游拒绝。
func AnalyzeToolCallOutputContextCoverageBytes(body []byte) ToolCallOutputContextCoverage {
	coverage := ToolCallOutputContextCoverage{}
	if len(body) == 0 {
		return coverage
	}
	input := parseRawJSONView(body).Get("input")
	if !input.IsArray() {
		return coverage
	}

	missingCallID := false
	var outputCallIDs map[string]struct{}
	var contextIDs map[string]struct{}
	input.ForEach(func(_, item gjson.Result) bool {
		if !item.IsObject() {
			return true
		}
		itemType := item.Get("type").String()
		switch {
		case isCodexToolCallOutputItemType(itemType):
			coverage.HasFunctionCallOutput = true
			callID := strings.TrimSpace(item.Get("call_id").String())
			if callID == "" {
				missingCallID = true
				return true
			}
			if outputCallIDs == nil {
				outputCallIDs = make(map[string]struct{})
			}
			outputCallIDs[callID] = struct{}{}
		case isCodexToolCallContextItemType(itemType):
			callID := strings.TrimSpace(item.Get("call_id").String())
			if callID == "" {
				return true
			}
			if contextIDs == nil {
				contextIDs = make(map[string]struct{})
			}
			contextIDs[callID] = struct{}{}
		case itemType == "item_reference":
			idValue := strings.TrimSpace(item.Get("id").String())
			if idValue == "" {
				return true
			}
			if contextIDs == nil {
				contextIDs = make(map[string]struct{})
			}
			contextIDs[idValue] = struct{}{}
		}
		return true
	})

	if !coverage.HasFunctionCallOutput || missingCallID {
		return coverage
	}
	for callID := range outputCallIDs {
		if _, ok := contextIDs[callID]; !ok {
			return coverage
		}
	}
	coverage.ContextCoversAllCallIDs = true
	return coverage
}

// ValidateFunctionCallOutputContext 为 handler 提供低开销校验结果：
// 1) 无工具输出直接返回
// 2) 若已存在工具调用上下文则提前返回
// 3) 仅在无工具上下文时才构建 call_id / item_reference 集合
// 字段名保留 FunctionCallOutput 是为了兼容既有调用点；语义覆盖所有 Codex 工具输出。
func ValidateFunctionCallOutputContext(reqBody map[string]any) FunctionCallOutputValidation {
	result := FunctionCallOutputValidation{}
	if reqBody == nil {
		return result
	}
	input, ok := reqBody["input"].([]any)
	if !ok {
		return result
	}

	for _, item := range input {
		itemMap, ok := item.(map[string]any)
		if !ok {
			continue
		}
		itemType, _ := itemMap["type"].(string)
		switch {
		case isCodexToolCallOutputItemType(itemType):
			result.HasFunctionCallOutput = true
		case isCodexToolCallContextItemType(itemType):
			callID, _ := itemMap["call_id"].(string)
			if strings.TrimSpace(callID) != "" {
				result.HasToolCallContext = true
			}
		}
		if result.HasFunctionCallOutput && result.HasToolCallContext {
			return result
		}
	}

	if !result.HasFunctionCallOutput || result.HasToolCallContext {
		return result
	}

	callIDs := make(map[string]struct{})
	referenceIDs := make(map[string]struct{})
	for _, item := range input {
		itemMap, ok := item.(map[string]any)
		if !ok {
			continue
		}
		itemType, _ := itemMap["type"].(string)
		switch {
		case isCodexToolCallOutputItemType(itemType):
			callID, _ := itemMap["call_id"].(string)
			callID = strings.TrimSpace(callID)
			if callID == "" {
				result.HasFunctionCallOutputMissingCallID = true
				continue
			}
			callIDs[callID] = struct{}{}
		case itemType == "item_reference":
			result.HasItemReference = true
			idValue, _ := itemMap["id"].(string)
			idValue = strings.TrimSpace(idValue)
			if idValue == "" {
				continue
			}
			referenceIDs[idValue] = struct{}{}
		}
	}

	if len(callIDs) == 0 || len(referenceIDs) == 0 {
		return result
	}
	allReferenced := true
	for callID := range callIDs {
		if _, ok := referenceIDs[callID]; !ok {
			allReferenced = false
			break
		}
	}
	result.HasItemReferenceForAllCallIDs = allReferenced
	return result
}

// HasFunctionCallOutput 判断 input 是否包含任意 Codex 工具输出，用于触发续链校验。
// 名称保留 function_call_output 是为了兼容既有调用点。
func HasFunctionCallOutput(reqBody map[string]any) bool {
	return AnalyzeToolContinuationSignals(reqBody).HasFunctionCallOutput
}

// HasToolCallContext 判断 input 是否包含带 call_id 的工具调用上下文，
// 用于判断工具输出是否具备可关联的上下文。
func HasToolCallContext(reqBody map[string]any) bool {
	return AnalyzeToolContinuationSignals(reqBody).HasToolCallContext
}

// FunctionCallOutputCallIDs 提取 input 中工具输出的 call_id 集合。
// 仅返回非空 call_id，用于与 item_reference.id 做匹配校验。
func FunctionCallOutputCallIDs(reqBody map[string]any) []string {
	return AnalyzeToolContinuationSignals(reqBody).FunctionCallOutputCallIDs
}

// HasFunctionCallOutputMissingCallID 判断是否存在缺少 call_id 的工具输出。
func HasFunctionCallOutputMissingCallID(reqBody map[string]any) bool {
	return AnalyzeToolContinuationSignals(reqBody).HasFunctionCallOutputMissingCallID
}

// HasItemReferenceForCallIDs 判断 item_reference.id 是否覆盖所有 call_id。
// 用于仅依赖引用项完成续链场景的校验。
func HasItemReferenceForCallIDs(reqBody map[string]any, callIDs []string) bool {
	if reqBody == nil || len(callIDs) == 0 {
		return false
	}
	input, ok := reqBody["input"].([]any)
	if !ok {
		return false
	}
	referenceIDs := make(map[string]struct{})
	for _, item := range input {
		itemMap, ok := item.(map[string]any)
		if !ok {
			continue
		}
		itemType, _ := itemMap["type"].(string)
		if itemType != "item_reference" {
			continue
		}
		idValue, _ := itemMap["id"].(string)
		idValue = strings.TrimSpace(idValue)
		if idValue == "" {
			continue
		}
		referenceIDs[idValue] = struct{}{}
	}
	if len(referenceIDs) == 0 {
		return false
	}
	for _, callID := range callIDs {
		if _, ok := referenceIDs[strings.TrimSpace(callID)]; !ok {
			return false
		}
	}
	return true
}

// hasNonEmptyString 判断字段是否为非空字符串。
func hasNonEmptyString(value any) bool {
	stringValue, ok := value.(string)
	return ok && strings.TrimSpace(stringValue) != ""
}

// hasToolsSignal 判断 tools 字段是否显式声明（存在且不为空）。
func hasToolsSignal(reqBody map[string]any) bool {
	raw, exists := reqBody["tools"]
	if !exists || raw == nil {
		return false
	}
	if tools, ok := raw.([]any); ok {
		return len(tools) > 0
	}
	return false
}

// hasToolChoiceSignal 判断 tool_choice 是否显式声明（非空或非 nil）。
func hasToolChoiceSignal(reqBody map[string]any) bool {
	raw, exists := reqBody["tool_choice"]
	if !exists || raw == nil {
		return false
	}
	switch value := raw.(type) {
	case string:
		return strings.TrimSpace(value) != ""
	case map[string]any:
		return len(value) > 0
	default:
		return false
	}
}
