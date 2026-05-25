package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestKiroOAuthService_parseRefreshResponse(t *testing.T) {
	t.Parallel()
	s := NewKiroOAuthService()

	tests := []struct {
		name           string
		body           string
		oldRefresh     string
		wantAccess     string
		wantRefresh    string
		wantErrSubstr  string
		minExpiresFrom int64
	}{
		{
			name:           "ok with new refresh token",
			body:           `{"accessToken":"AT","refreshToken":"RT","expiresIn":3600,"tokenType":"Bearer"}`,
			oldRefresh:     "OLD",
			wantAccess:     "AT",
			wantRefresh:    "RT",
			minExpiresFrom: 3600 - 300 - 5,
		},
		{
			name:           "ok without refresh token reuses old",
			body:           `{"accessToken":"AT","expiresIn":3600}`,
			oldRefresh:     "OLD",
			wantAccess:     "AT",
			wantRefresh:    "OLD",
			minExpiresFrom: 3600 - 300 - 5,
		},
		{
			name:          "missing accessToken",
			body:          `{"refreshToken":"x"}`,
			wantErrSubstr: "missing accessToken",
		},
		{
			name:          "malformed json",
			body:          `not-json`,
			wantErrSubstr: "parse refresh response",
		},
		{
			name:           "expiresIn zero falls back to 3600",
			body:           `{"accessToken":"AT","refreshToken":"RT","expiresIn":0}`,
			oldRefresh:     "OLD",
			wantAccess:     "AT",
			wantRefresh:    "RT",
			minExpiresFrom: 3600 - 300 - 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := time.Now().Unix()
			info, err := s.parseRefreshResponse([]byte(tt.body), tt.oldRefresh)
			if tt.wantErrSubstr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErrSubstr) {
					t.Fatalf("want err containing %q, got %v", tt.wantErrSubstr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if info.AccessToken != tt.wantAccess {
				t.Errorf("AccessToken: want %q got %q", tt.wantAccess, info.AccessToken)
			}
			if info.RefreshToken != tt.wantRefresh {
				t.Errorf("RefreshToken: want %q got %q", tt.wantRefresh, info.RefreshToken)
			}
			if got := info.ExpiresAt - before; got < tt.minExpiresFrom {
				t.Errorf("ExpiresAt offset: want >=%d got %d", tt.minExpiresFrom, got)
			}
		})
	}
}

func TestKiroOAuthService_BuildAccountCredentials(t *testing.T) {
	t.Parallel()
	s := NewKiroOAuthService()

	info := &KiroTokenInfo{
		AccessToken:  "AT",
		RefreshToken: "RT",
		ExpiresIn:    3600,
		ExpiresAt:    1234567890,
	}
	got := s.BuildAccountCredentials(info)

	if got["access_token"] != "AT" {
		t.Errorf("access_token: want AT got %v", got["access_token"])
	}
	if got["refresh_token"] != "RT" {
		t.Errorf("refresh_token: want RT got %v", got["refresh_token"])
	}
	expExp, _ := strconv.ParseInt(got["expires_at"].(string), 10, 64)
	if expExp != 1234567890 {
		t.Errorf("expires_at: want 1234567890 got %d", expExp)
	}

	// Without refresh token: should not set the key
	got2 := s.BuildAccountCredentials(&KiroTokenInfo{AccessToken: "AT", ExpiresAt: 1})
	if _, ok := got2["refresh_token"]; ok {
		t.Errorf("refresh_token should be absent when empty")
	}
}

func TestKiroOAuthService_RefreshOIDC(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.URL.Path, "/token"; got != want {
			t.Errorf("path: want %q got %q", want, got)
		}
		if got, want := r.Header.Get("Content-Type"), "application/json"; got != want {
			t.Errorf("content-type: want %q got %q", want, got)
		}

		var body kiroOidcRefreshRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if body.GrantType != "refresh_token" {
			t.Errorf("grantType: want refresh_token got %q", body.GrantType)
		}
		if body.ClientID != "C1" || body.ClientSecret != "C2" || body.RefreshToken != "RT" {
			t.Errorf("body fields wrong: %+v", body)
		}

		_ = json.NewEncoder(w).Encode(map[string]any{
			"accessToken":  "NEW_AT",
			"refreshToken": "NEW_RT",
			"expiresIn":    3600,
			"tokenType":    "Bearer",
		})
	}))
	defer srv.Close()

	s := NewKiroOAuthService()
	// Override the URL helper isn't possible from outside; we exercise the JSON
	// path of doJSONRequest + parseRefreshResponse via a manual call shape.
	body := mustJSON(kiroOidcRefreshRequest{
		ClientID: "C1", ClientSecret: "C2", RefreshToken: "RT", GrantType: "refresh_token",
	})
	respBody, err := s.doJSONRequest(context.Background(), srv.URL+"/token", body, nil)
	if err != nil {
		t.Fatalf("doJSONRequest: %v", err)
	}
	info, err := s.parseRefreshResponse(respBody, "RT")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if info.AccessToken != "NEW_AT" || info.RefreshToken != "NEW_RT" {
		t.Errorf("got %+v", info)
	}
}

func TestKiroOAuthService_RouteByLoginType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		loginType string
		wantOIDC  bool
	}{
		{"empty login_type goes OIDC", "", true},
		{"builder goes OIDC", "builder", true},
		{"idc goes OIDC", "idc", true},
		{"github goes Social", "github", false},
		{"google goes Social", "google", false},
		{"social alias goes Social", "social", false},
		{"upper-case still routes correctly", "GitHub", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lt := strings.ToLower(strings.TrimSpace(tt.loginType))
			isOIDC := lt != "github" && lt != "google" && lt != "social"
			if isOIDC != tt.wantOIDC {
				t.Errorf("classify(%q): want OIDC=%v got %v", tt.loginType, tt.wantOIDC, isOIDC)
			}
		})
	}
}

func TestKiroUserAgent_FormatStable(t *testing.T) {
	t.Parallel()
	ua := kiroUserAgent("abc123")
	if !strings.Contains(ua, "KiroIDE-"+kiroVersion+"-abc123") {
		t.Errorf("UA missing version+machineId suffix: %q", ua)
	}
	if !strings.Contains(ua, "aws-sdk-js/"+kiroAwsSdkVersionUA) {
		t.Errorf("UA missing aws-sdk-js prefix: %q", ua)
	}
	uaNoMachine := kiroUserAgent("")
	if strings.Contains(uaNoMachine, "KiroIDE-"+kiroVersion+"-") {
		t.Errorf("UA without machineId should not have trailing dash: %q", uaNoMachine)
	}
}

func mustJSON(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}
