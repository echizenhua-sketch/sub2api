package service

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestNeedsToolContinuationSignals(t *testing.T) {
	// 覆盖所有触发续链的信号来源，确保判定逻辑完整。
	cases := []struct {
		name string
		body map[string]any
		want bool
	}{
		{name: "nil", body: nil, want: false},
		{name: "previous_response_id", body: map[string]any{"previous_response_id": "resp_1"}, want: true},
		{name: "previous_response_id_blank", body: map[string]any{"previous_response_id": "  "}, want: false},
		{name: "function_call_output", body: map[string]any{"input": []any{map[string]any{"type": "function_call_output"}}}, want: true},
		{name: "tool_search_output", body: map[string]any{"input": []any{map[string]any{"type": "tool_search_output"}}}, want: true},
		{name: "custom_tool_call_output", body: map[string]any{"input": []any{map[string]any{"type": "custom_tool_call_output"}}}, want: true},
		{name: "mcp_tool_call_output", body: map[string]any{"input": []any{map[string]any{"type": "mcp_tool_call_output"}}}, want: true},
		{name: "item_reference", body: map[string]any{"input": []any{map[string]any{"type": "item_reference"}}}, want: true},
		{name: "tools", body: map[string]any{"tools": []any{map[string]any{"type": "function"}}}, want: true},
		{name: "tools_empty", body: map[string]any{"tools": []any{}}, want: false},
		{name: "tools_invalid", body: map[string]any{"tools": "bad"}, want: false},
		{name: "tool_choice", body: map[string]any{"tool_choice": "auto"}, want: true},
		{name: "tool_choice_object", body: map[string]any{"tool_choice": map[string]any{"type": "function"}}, want: true},
		{name: "tool_choice_empty_object", body: map[string]any{"tool_choice": map[string]any{}}, want: false},
		{name: "none", body: map[string]any{"input": []any{map[string]any{"type": "text", "text": "hi"}}}, want: false},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, NeedsToolContinuation(tt.body))
		})
	}
}

func TestHasFunctionCallOutput(t *testing.T) {
	// 所有 Codex 工具输出都应视为续链输出，避免 WS 续链时丢失 previous_response_id。
	require.False(t, HasFunctionCallOutput(nil))
	for _, typ := range []string{
		"function_call_output",
		"tool_search_output",
		"custom_tool_call_output",
		"mcp_tool_call_output",
	} {
		require.True(t, HasFunctionCallOutput(map[string]any{
			"input": []any{map[string]any{"type": typ}},
		}), typ)
	}
	require.False(t, HasFunctionCallOutput(map[string]any{
		"input": "text",
	}))
}

func TestHasToolCallContext(t *testing.T) {
	// 工具调用上下文必须包含 call_id，才能作为可关联上下文。
	require.False(t, HasToolCallContext(nil))
	for _, typ := range []string{
		"tool_call",
		"function_call",
		"local_shell_call",
		"tool_search_call",
		"custom_tool_call",
		"mcp_tool_call",
	} {
		require.True(t, HasToolCallContext(map[string]any{
			"input": []any{map[string]any{"type": typ, "call_id": "call_1"}},
		}), typ)
	}
	require.False(t, HasToolCallContext(map[string]any{
		"input": []any{map[string]any{"type": "tool_call"}},
	}))
}

func TestFunctionCallOutputCallIDs(t *testing.T) {
	// 仅提取工具输出的非空 call_id，去重后返回。
	require.Empty(t, FunctionCallOutputCallIDs(nil))
	callIDs := FunctionCallOutputCallIDs(map[string]any{
		"input": []any{
			map[string]any{"type": "function_call_output", "call_id": "call_1"},
			map[string]any{"type": "tool_search_output", "call_id": "call_search"},
			map[string]any{"type": "custom_tool_call_output", "call_id": "call_custom"},
			map[string]any{"type": "mcp_tool_call_output", "call_id": "call_mcp"},
			map[string]any{"type": "function_call_output", "call_id": ""},
			map[string]any{"type": "function_call_output", "call_id": "call_1"},
		},
	})
	require.ElementsMatch(t, []string{"call_1", "call_search", "call_custom", "call_mcp"}, callIDs)
}

func TestHasFunctionCallOutputMissingCallID(t *testing.T) {
	require.False(t, HasFunctionCallOutputMissingCallID(nil))
	require.True(t, HasFunctionCallOutputMissingCallID(map[string]any{
		"input": []any{map[string]any{"type": "function_call_output"}},
	}))
	require.True(t, HasFunctionCallOutputMissingCallID(map[string]any{
		"input": []any{map[string]any{"type": "tool_search_output"}},
	}))
	require.False(t, HasFunctionCallOutputMissingCallID(map[string]any{
		"input": []any{map[string]any{"type": "tool_search_output", "call_id": "call_1"}},
	}))
}

func TestHasItemReferenceForCallIDs(t *testing.T) {
	// item_reference 需要覆盖所有 call_id 才视为可关联上下文。
	require.False(t, HasItemReferenceForCallIDs(nil, []string{"call_1"}))
	require.False(t, HasItemReferenceForCallIDs(map[string]any{}, []string{"call_1"}))
	req := map[string]any{
		"input": []any{
			map[string]any{"type": "item_reference", "id": "call_1"},
			map[string]any{"type": "item_reference", "id": "call_2"},
		},
	}
	require.True(t, HasItemReferenceForCallIDs(req, []string{"call_1"}))
	require.True(t, HasItemReferenceForCallIDs(req, []string{"call_1", "call_2"}))
	require.False(t, HasItemReferenceForCallIDs(req, []string{"call_1", "call_3"}))
}

func TestValidateFunctionCallOutputContextBytesMatchesMapValidation(t *testing.T) {
	// handler 预校验走 raw JSON 扫描，语义必须与 service 内部 map 校验保持一致。
	cases := []struct {
		name string
		body map[string]any
	}{
		{
			name: "no_input",
			body: map[string]any{"model": "gpt-5.4"},
		},
		{
			name: "missing_call_id",
			body: map[string]any{"input": []any{map[string]any{"type": "function_call_output"}}},
		},
		{
			name: "call_id_without_reference",
			body: map[string]any{"input": []any{map[string]any{"type": "function_call_output", "call_id": "call_1"}}},
		},
		{
			name: "matching_reference",
			body: map[string]any{"input": []any{
				map[string]any{"type": "function_call_output", "call_id": "call_1"},
				map[string]any{"type": "item_reference", "id": "call_1"},
			}},
		},
		{
			name: "partial_reference",
			body: map[string]any{"input": []any{
				map[string]any{"type": "function_call_output", "call_id": "call_1"},
				map[string]any{"type": "tool_search_output", "call_id": "call_2"},
				map[string]any{"type": "item_reference", "id": "call_1"},
			}},
		},
		{
			name: "tool_context",
			body: map[string]any{"input": []any{
				map[string]any{"type": "function_call_output", "call_id": "call_1"},
				map[string]any{"type": "function_call", "call_id": "call_1"},
			}},
		},
		{
			name: "all_codex_tool_outputs",
			body: map[string]any{"input": []any{
				map[string]any{"type": "function_call_output", "call_id": "call_function"},
				map[string]any{"type": "tool_search_output", "call_id": "call_search"},
				map[string]any{"type": "custom_tool_call_output", "call_id": "call_custom"},
				map[string]any{"type": "mcp_tool_call_output", "call_id": "call_mcp"},
				map[string]any{"type": "item_reference", "id": "call_function"},
				map[string]any{"type": "item_reference", "id": "call_search"},
				map[string]any{"type": "item_reference", "id": "call_custom"},
				map[string]any{"type": "item_reference", "id": "call_mcp"},
			}},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			require.Equal(t, ValidateFunctionCallOutputContext(tt.body), ValidateFunctionCallOutputContextBytes(bodyBytes))
		})
	}
}

func TestAnalyzeToolCallOutputContextCoverageBytes(t *testing.T) {
	cases := []struct {
		name         string
		body         map[string]any
		hasOutput    bool
		coversAllIDs bool
	}{
		{
			name:         "no_input",
			body:         map[string]any{"model": "gpt-5.1"},
			hasOutput:    false,
			coversAllIDs: false,
		},
		{
			name: "no_tool_output",
			body: map[string]any{"input": []any{
				map[string]any{"type": "message", "content": "hi"},
			}},
			hasOutput:    false,
			coversAllIDs: false,
		},
		{
			name: "all_outputs_covered_by_context",
			body: map[string]any{"input": []any{
				map[string]any{"type": "function_call", "call_id": "call_a"},
				map[string]any{"type": "function_call_output", "call_id": "call_a"},
			}},
			hasOutput:    true,
			coversAllIDs: true,
		},
		{
			name: "all_outputs_covered_by_item_reference",
			body: map[string]any{"input": []any{
				map[string]any{"type": "function_call_output", "call_id": "call_a"},
				map[string]any{"type": "item_reference", "id": "call_a"},
			}},
			hasOutput:    true,
			coversAllIDs: true,
		},
		{
			// 关键回归用例：input 内存在某一个上下文项，但另一个输出的 call_id
			// 只能由上游会话链（previous_response_id）解析——不可剥离。
			name: "partial_coverage_not_movable",
			body: map[string]any{"input": []any{
				map[string]any{"type": "function_call", "call_id": "call_a"},
				map[string]any{"type": "function_call_output", "call_id": "call_a"},
				map[string]any{"type": "function_call_output", "call_id": "call_b"},
			}},
			hasOutput:    true,
			coversAllIDs: false,
		},
		{
			name: "unrelated_context_does_not_cover",
			body: map[string]any{"input": []any{
				map[string]any{"type": "function_call", "call_id": "call_x"},
				map[string]any{"type": "function_call_output", "call_id": "call_b"},
			}},
			hasOutput:    true,
			coversAllIDs: false,
		},
		{
			name: "output_missing_call_id_not_movable",
			body: map[string]any{"input": []any{
				map[string]any{"type": "function_call", "call_id": "call_a"},
				map[string]any{"type": "function_call_output"},
				map[string]any{"type": "function_call_output", "call_id": "call_a"},
			}},
			hasOutput:    true,
			coversAllIDs: false,
		},
		{
			name: "mixed_context_and_reference_cover_all",
			body: map[string]any{"input": []any{
				map[string]any{"type": "function_call", "call_id": "call_a"},
				map[string]any{"type": "function_call_output", "call_id": "call_a"},
				map[string]any{"type": "function_call_output", "call_id": "call_b"},
				map[string]any{"type": "item_reference", "id": "call_b"},
			}},
			hasOutput:    true,
			coversAllIDs: true,
		},
		{
			name: "all_codex_output_types_covered",
			body: map[string]any{"input": []any{
				map[string]any{"type": "tool_search_output", "call_id": "call_s"},
				map[string]any{"type": "tool_search_call", "call_id": "call_s"},
				map[string]any{"type": "mcp_tool_call_output", "call_id": "call_m"},
				map[string]any{"type": "mcp_tool_call", "call_id": "call_m"},
			}},
			hasOutput:    true,
			coversAllIDs: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)

			coverage := AnalyzeToolCallOutputContextCoverageBytes(bodyBytes)
			require.Equal(t, tt.hasOutput, coverage.HasFunctionCallOutput, "HasFunctionCallOutput")
			require.Equal(t, tt.coversAllIDs, coverage.ContextCoversAllCallIDs, "ContextCoversAllCallIDs")
		})
	}
}

func TestRewriteOpenAIHTTPContinuationRequest_ReplaysPreviousOutputWithoutForgingReferences(t *testing.T) {
	state := OpenAIResponseContinuation{
		ResponseID: "resp_previous",
		AccountID:  42,
		ReplayInput: []json.RawMessage{
			json.RawMessage(`{"type":"reasoning","id":"rs_previous","encrypted_content":"opaque"}`),
			json.RawMessage(`{"type":"function_call","id":"fc_item_a","call_id":"call_a","name":"first","arguments":"{}"}`),
			json.RawMessage(`{"type":"custom_tool_call","id":"ctc_item_b","call_id":"call_b","name":"exec","input":"pwd"}`),
			json.RawMessage(`{"type":"tool_search_call","id":"tsc_item_c","call_id":"call_c","arguments":{"query":"git"}}`),
			json.RawMessage(`{"type":"mcp_tool_call","id":"mcp_item_d","call_id":"call_d","name":"list"}`),
		},
	}
	body := []byte(`{
		"model":"gpt-5.6-sol",
		"previous_response_id":"resp_previous",
		"input":[
			{"type":"function_call_output","call_id":"call_a","output":"a"},
			{"type":"custom_tool_call_output","call_id":"call_b","output":"b"},
			{"type":"tool_search_output","call_id":"call_c","output":"c"},
			{"type":"mcp_tool_call_output","call_id":"call_d","output":"d"}
		]
	}`)

	rewritten, err := RewriteOpenAIHTTPContinuationRequest(body, state)
	require.NoError(t, err)
	require.False(t, gjson.GetBytes(rewritten, "previous_response_id").Exists())
	require.Len(t, gjson.GetBytes(rewritten, "input").Array(), 9)
	require.Equal(t, "fc_item_a", gjson.GetBytes(rewritten, "input.1.id").String())
	require.Equal(t, "call_a", gjson.GetBytes(rewritten, "input.1.call_id").String())
	require.NotEqual(t, gjson.GetBytes(rewritten, "input.1.id").String(), gjson.GetBytes(rewritten, "input.1.call_id").String(), "item id must not be synthesized from call_id")
	require.False(t, gjson.GetBytes(rewritten, `input.#(type=="item_reference")`).Exists(), "replay must preserve real prior items instead of forging item_reference ids")
}

func TestRewriteOpenAIHTTPContinuationRequest_RejectsMissingParallelToolContext(t *testing.T) {
	state := OpenAIResponseContinuation{
		ResponseID: "resp_previous",
		AccountID:  42,
		ReplayInput: []json.RawMessage{
			json.RawMessage(`{"type":"function_call","id":"fc_item_a","call_id":"call_a","name":"first","arguments":"{}"}`),
		},
	}
	body := []byte(`{
		"model":"gpt-5.6-sol",
		"previous_response_id":"resp_previous",
		"input":[
			{"type":"function_call_output","call_id":"call_a","output":"a"},
			{"type":"function_call_output","call_id":"call_missing","output":"missing"}
		]
	}`)

	_, err := RewriteOpenAIHTTPContinuationRequest(body, state)
	require.Error(t, err)
	require.Contains(t, err.Error(), "call_missing")
}

func TestExtractOpenAIResponseReplayInput_CoversTerminalAndOutputItemEvents(t *testing.T) {
	terminal := []byte(`{"type":"response.completed","response":{"id":"resp_1","output":[{"type":"function_call","id":"fc_1","call_id":"call_1","name":"one","arguments":"{}"},{"type":"message","id":"msg_1","role":"assistant","content":[]}]}}`)
	items := ExtractOpenAIResponseReplayInput(terminal)
	require.Len(t, items, 2)
	require.Equal(t, "fc_1", gjson.GetBytes(items[0], "id").String())
	require.Equal(t, "msg_1", gjson.GetBytes(items[1], "id").String())

	itemDone := []byte(`{"type":"response.output_item.done","item":{"type":"custom_tool_call","id":"ctc_1","call_id":"call_custom","name":"exec","input":"pwd"}}`)
	items = ExtractOpenAIResponseReplayInput(itemDone)
	require.Len(t, items, 1)
	require.Equal(t, "ctc_1", gjson.GetBytes(items[0], "id").String())
}

func TestPrepareOpenAIHTTPContinuationRequest_ExplicitAndImplicitRecovery(t *testing.T) {
	svc := &OpenAIGatewayService{}
	store := svc.getOpenAIWSStateStore()
	scope := OpenAIResponseContinuationScope{GroupID: 9, APIKeyID: 90, UserID: 900}
	state := OpenAIResponseContinuation{
		ResponseID: "resp_prepare",
		AccountID:  77,
		ReplayInput: []json.RawMessage{
			json.RawMessage(`{"type":"function_call","id":"fc_prepare","call_id":"call_prepare","name":"lookup","arguments":"{}"}`),
		},
	}
	store.BindResponseContinuation(scope, "session_prepare", state, time.Minute)

	explicitBody := []byte(`{"model":"gpt-5.6-sol","previous_response_id":"resp_prepare","input":[{"type":"function_call_output","call_id":"call_prepare","output":"ok"}]}`)
	explicit, err := svc.PrepareOpenAIHTTPContinuationRequest(context.Background(), scope, "session_prepare", explicitBody)
	require.NoError(t, err)
	require.True(t, explicit.Replayed)
	require.Equal(t, "resp_prepare", explicit.RoutingResponseID)
	require.Equal(t, int64(77), explicit.AccountID)
	require.False(t, gjson.GetBytes(explicit.Body, "previous_response_id").Exists())

	implicitBody := []byte(`{"model":"gpt-5.6-sol","input":[{"type":"function_call_output","call_id":"call_prepare","output":"ok"}]}`)
	implicit, err := svc.PrepareOpenAIHTTPContinuationRequest(context.Background(), scope, "session_prepare", implicitBody)
	require.NoError(t, err)
	require.True(t, implicit.Replayed)
	require.Equal(t, "resp_prepare", implicit.RoutingResponseID)
	require.Equal(t, int64(77), implicit.AccountID)
}

func TestPrepareOpenAIHTTPContinuationRequest_RecoversByCallIDWhenSessionChanges(t *testing.T) {
	svc := &OpenAIGatewayService{}
	store := svc.getOpenAIWSStateStore()
	scope := OpenAIResponseContinuationScope{GroupID: 10, APIKeyID: 8, UserID: 1}
	state := OpenAIResponseContinuation{
		ResponseID: "resp_vscode",
		AccountID:  15932,
		ReplayInput: []json.RawMessage{
			json.RawMessage(`{"type":"function_call","id":"fc_vscode","call_id":"call_vscode","name":"lookup","arguments":"{}"}`),
		},
	}
	store.BindResponseContinuation(scope, "first_turn_content_hash", state, time.Minute)

	body := []byte(`{"model":"grok-4.5-con","input":[{"type":"function_call_output","call_id":"call_vscode","output":"ok"}]}`)
	prepared, err := svc.PrepareOpenAIHTTPContinuationRequest(context.Background(), scope, "second_turn_content_hash", body)
	require.NoError(t, err)
	require.True(t, prepared.Replayed)
	require.Equal(t, "resp_vscode", prepared.RoutingResponseID)
	require.Equal(t, int64(15932), prepared.AccountID)
	require.Equal(t, "fc_vscode", gjson.GetBytes(prepared.Body, "input.0.id").String())
	require.Equal(t, "call_vscode", gjson.GetBytes(prepared.Body, "input.1.call_id").String())
}

func TestPrepareOpenAIHTTPContinuationRequest_ParallelCallIDsMustResolveSameResponse(t *testing.T) {
	svc := &OpenAIGatewayService{}
	store := svc.getOpenAIWSStateStore()
	scope := OpenAIResponseContinuationScope{GroupID: 10, APIKeyID: 8, UserID: 1}
	store.BindResponseContinuation(scope, "session_a", OpenAIResponseContinuation{
		ResponseID: "resp_a",
		AccountID:  100,
		ReplayInput: []json.RawMessage{
			json.RawMessage(`{"type":"function_call","id":"fc_a","call_id":"call_a","name":"first","arguments":"{}"}`),
		},
	}, time.Minute)
	store.BindResponseContinuation(scope, "session_b", OpenAIResponseContinuation{
		ResponseID: "resp_b",
		AccountID:  200,
		ReplayInput: []json.RawMessage{
			json.RawMessage(`{"type":"function_call","id":"fc_b","call_id":"call_b","name":"second","arguments":"{}"}`),
		},
	}, time.Minute)

	body := []byte(`{"model":"grok-4.5-con","input":[{"type":"function_call_output","call_id":"call_a","output":"a"},{"type":"function_call_output","call_id":"call_b","output":"b"}]}`)
	_, err := svc.PrepareOpenAIHTTPContinuationRequest(context.Background(), scope, "unrelated_session", body)
	require.Error(t, err)
	require.Contains(t, err.Error(), "different prior responses")
}

func TestPrepareOpenAIHTTPContinuationRequest_ParallelCallIDsRecoverSameResponse(t *testing.T) {
	svc := &OpenAIGatewayService{}
	store := svc.getOpenAIWSStateStore()
	scope := OpenAIResponseContinuationScope{GroupID: 10, APIKeyID: 8, UserID: 1}
	store.BindResponseContinuation(scope, "first_turn_hash", OpenAIResponseContinuation{
		ResponseID: "resp_parallel",
		AccountID:  15932,
		ReplayInput: []json.RawMessage{
			json.RawMessage(`{"type":"function_call","id":"fc_a","call_id":"call_a","name":"first","arguments":"{}"}`),
			json.RawMessage(`{"type":"custom_tool_call","id":"ctc_b","call_id":"call_b","name":"second","input":"pwd"}`),
		},
	}, time.Minute)

	body := []byte(`{"model":"grok-4.5-con","input":[{"type":"function_call_output","call_id":"call_a","output":"a"},{"type":"custom_tool_call_output","call_id":"call_b","output":"b"}]}`)
	prepared, err := svc.PrepareOpenAIHTTPContinuationRequest(context.Background(), scope, "second_turn_hash", body)
	require.NoError(t, err)
	require.True(t, prepared.Replayed)
	require.Equal(t, "resp_parallel", prepared.RoutingResponseID)
	require.Equal(t, int64(15932), prepared.AccountID)
	require.Len(t, gjson.GetBytes(prepared.Body, "input").Array(), 4)
}

func TestPrepareOpenAIHTTPContinuationRequest_DuplicateCallIDFailsClosed(t *testing.T) {
	svc := &OpenAIGatewayService{}
	store := svc.getOpenAIWSStateStore()
	scope := OpenAIResponseContinuationScope{GroupID: 10, APIKeyID: 8, UserID: 1}
	for _, state := range []OpenAIResponseContinuation{
		{
			ResponseID: "resp_first",
			AccountID:  100,
			ReplayInput: []json.RawMessage{
				json.RawMessage(`{"type":"function_call","id":"fc_first","call_id":"call_reused","name":"first","arguments":"{}"}`),
			},
		},
		{
			ResponseID: "resp_second",
			AccountID:  200,
			ReplayInput: []json.RawMessage{
				json.RawMessage(`{"type":"function_call","id":"fc_second","call_id":"call_reused","name":"second","arguments":"{}"}`),
			},
		},
	} {
		store.BindResponseContinuation(scope, state.ResponseID, state, time.Minute)
	}

	body := []byte(`{"model":"grok-4.5-con","input":[{"type":"function_call_output","call_id":"call_reused","output":"ok"}]}`)
	_, err := svc.PrepareOpenAIHTTPContinuationRequest(context.Background(), scope, "unrelated_session", body)
	require.Error(t, err)
	require.Contains(t, err.Error(), "ambiguous")
}

func TestPrepareOpenAIHTTPContinuationRequest_CallIDDoesNotCrossScope(t *testing.T) {
	svc := &OpenAIGatewayService{}
	store := svc.getOpenAIWSStateStore()
	scope := OpenAIResponseContinuationScope{GroupID: 10, APIKeyID: 8, UserID: 1}
	store.BindResponseContinuation(scope, "session", OpenAIResponseContinuation{
		ResponseID: "resp_scoped_call",
		AccountID:  15932,
		ReplayInput: []json.RawMessage{
			json.RawMessage(`{"type":"function_call","id":"fc_scoped","call_id":"call_scoped","name":"lookup","arguments":"{}"}`),
		},
	}, time.Minute)

	body := []byte(`{"model":"grok-4.5-con","input":[{"type":"function_call_output","call_id":"call_scoped","output":"ok"}]}`)
	wrongScope := OpenAIResponseContinuationScope{GroupID: 10, APIKeyID: 9, UserID: 1}
	_, err := svc.PrepareOpenAIHTTPContinuationRequest(context.Background(), wrongScope, "unrelated_session", body)
	require.Error(t, err)
	require.Contains(t, err.Error(), "prior response context was not found")
}

func TestPrepareOpenAIHTTPContinuationRequest_MissingOrWrongScopeFailsClosed(t *testing.T) {
	svc := &OpenAIGatewayService{}
	scope := OpenAIResponseContinuationScope{GroupID: 9, APIKeyID: 90, UserID: 900}
	body := []byte(`{"model":"gpt-5.6-sol","previous_response_id":"resp_missing","input":[{"type":"message","role":"user","content":[{"type":"input_text","text":"continue"}]}]}`)

	_, err := svc.PrepareOpenAIHTTPContinuationRequest(context.Background(), scope, "session_missing", body)
	require.Error(t, err)
	require.Contains(t, err.Error(), "previous_response_id context was not found")
}
