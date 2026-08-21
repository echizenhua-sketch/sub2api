package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tidwall/gjson"
)

type openAIHTTPContinuationContextKey struct{}

type openAIHTTPContinuationContext struct {
	Scope       OpenAIResponseContinuationScope
	SessionHash string
}

// OpenAIHTTPContinuationPreparation is the result of resolving and replaying
// an HTTP Responses continuation before account selection.
type OpenAIHTTPContinuationPreparation struct {
	Body              []byte
	RoutingResponseID string
	AccountID         int64
	Replayed          bool
}

func WithOpenAIHTTPContinuationContext(ctx context.Context, scope OpenAIResponseContinuationScope, sessionHash string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, openAIHTTPContinuationContextKey{}, openAIHTTPContinuationContext{
		Scope:       scope,
		SessionHash: strings.TrimSpace(sessionHash),
	})
}

func openAIHTTPContinuationContextFromContext(ctx context.Context) (openAIHTTPContinuationContext, bool) {
	if ctx == nil {
		return openAIHTTPContinuationContext{}, false
	}
	value, ok := ctx.Value(openAIHTTPContinuationContextKey{}).(openAIHTTPContinuationContext)
	return value, ok && value.Scope.valid()
}

// PrepareOpenAIHTTPContinuationRequest restores the prior response output for
// HTTP clients. Explicit previous_response_id lookups and implicit tool-output
// recovery are both scoped to group, API key and user.
func (s *OpenAIGatewayService) PrepareOpenAIHTTPContinuationRequest(
	ctx context.Context,
	scope OpenAIResponseContinuationScope,
	sessionHash string,
	body []byte,
) (OpenAIHTTPContinuationPreparation, error) {
	result := OpenAIHTTPContinuationPreparation{Body: body}
	validation := ValidateFunctionCallOutputContextBytes(body)
	if validation.HasFunctionCallOutputMissingCallID {
		return result, fmt.Errorf("function_call_output requires a non-empty call_id")
	}

	previousResponseID := strings.TrimSpace(gjson.GetBytes(body, "previous_response_id").String())
	needsImplicitRecovery := validation.HasFunctionCallOutput && !validation.HasToolCallContext && !validation.HasItemReference
	if previousResponseID == "" && !needsImplicitRecovery {
		return result, nil
	}
	if !scope.valid() {
		return result, fmt.Errorf("Responses continuation scope is incomplete")
	}
	store := s.getOpenAIWSStateStore()
	if store == nil {
		return result, fmt.Errorf("Responses continuation state is unavailable")
	}

	var state OpenAIResponseContinuation
	var ok bool
	if previousResponseID != "" {
		state, ok = store.GetResponseContinuation(scope, previousResponseID)
		if !ok {
			return result, fmt.Errorf("previous_response_id context was not found for this API key, user, and group")
		}
	} else {
		var lookupErr error
		state, ok, lookupErr = store.GetCallIDContinuation(scope, functionCallOutputCallIDsBytes(body))
		if lookupErr != nil {
			return result, fmt.Errorf("function_call_output %w", lookupErr)
		}
		if !ok {
			state, ok = store.GetSessionContinuation(scope, sessionHash)
		}
		if !ok {
			return result, fmt.Errorf("function_call_output prior response context was not found for this API key, user, group, and session")
		}
	}

	rewritten, err := RewriteOpenAIHTTPContinuationRequest(body, state)
	if err != nil {
		return result, err
	}
	result.Body = rewritten
	result.RoutingResponseID = state.ResponseID
	result.AccountID = state.AccountID
	result.Replayed = true
	return result, nil
}

func (s *OpenAIGatewayService) bindHTTPResponseContinuation(
	ctx context.Context,
	account *Account,
	responseID string,
	replayInput []json.RawMessage,
) {
	continuationContext, ok := openAIHTTPContinuationContextFromContext(ctx)
	if !ok || account == nil || account.ID <= 0 || len(replayInput) == 0 {
		return
	}
	store := s.getOpenAIWSStateStore()
	if store == nil {
		return
	}
	store.BindResponseContinuation(
		continuationContext.Scope,
		continuationContext.SessionHash,
		OpenAIResponseContinuation{
			ResponseID:  strings.TrimSpace(responseID),
			AccountID:   account.ID,
			ReplayInput: replayInput,
		},
		s.openAIWSResponseStickyTTL(),
	)
}
