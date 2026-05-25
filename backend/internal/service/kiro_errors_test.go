package service

import "testing"

func TestClassifyKiroError(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		status int
		body   string
		want   KiroErrorClass
	}{
		{"401 → auth_needs_refresh", 401, "", KiroErrAuthNeedsRefresh},
		{"402 → recoverable", 402, "", KiroErrRecoverable},
		{"403 → recoverable", 403, "", KiroErrRecoverable},
		{"429 → recoverable", 429, "", KiroErrRecoverable},
		{"422 → fatal", 422, "", KiroErrFatal},
		{"400 generic → fatal", 400, `{"message":"bad request"}`, KiroErrFatal},
		{"400 CONTENT_LENGTH_EXCEEDS_THRESHOLD → fatal", 400, `{"__type":"ValidationException","message":"CONTENT_LENGTH_EXCEEDS_THRESHOLD"}`, KiroErrFatal},
		{"500 → fatal", 500, "", KiroErrFatal},
		{"503 → fatal", 503, "", KiroErrFatal},
		{"200 → unknown", 200, "", KiroErrUnknown},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := classifyKiroError(tt.status, []byte(tt.body))
			if got != tt.want {
				t.Errorf("status=%d body=%q: want %v got %v", tt.status, tt.body, tt.want, got)
			}
		})
	}
}

func TestIsKiroQuotaExhausted(t *testing.T) {
	t.Parallel()
	cases := map[int]bool{
		402: true,
		429: true,
		401: false,
		403: false,
		500: false,
		200: false,
	}
	for status, want := range cases {
		if got := IsKiroQuotaExhausted(status); got != want {
			t.Errorf("status=%d: want %v got %v", status, want, got)
		}
	}
}

func TestKiroErrorClassString(t *testing.T) {
	t.Parallel()
	pairs := map[KiroErrorClass]string{
		KiroErrUnknown:          "unknown",
		KiroErrRecoverable:      "recoverable",
		KiroErrAuthNeedsRefresh: "auth_needs_refresh",
		KiroErrFatal:            "fatal",
	}
	for c, want := range pairs {
		if got := c.String(); got != want {
			t.Errorf("%d: want %q got %q", c, want, got)
		}
	}
}
