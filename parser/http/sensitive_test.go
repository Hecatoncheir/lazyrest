package http

import (
	"net/http"
	"slices"
	"testing"
)

func TestSensitiveResponseValuesFindsRuntimeCredentials(t *testing.T) {
	body := `{
  "accessToken": "access-123",
  "nested": {"client_secret": "client-456"},
  "sessions": ["session-a", "session-b"],
  "user": {"id": 42}
}`
	values := SensitiveResponseValues(body, http.Header{
		"Set-Cookie": {"sid=cookie-789"},
		"X-Trace":    {"trace-ignored"},
	})
	for _, expected := range []string{"access-123", "client-456", "session-a", "session-b", "sid=cookie-789"} {
		if !slices.Contains(values, expected) {
			t.Errorf("sensitive response value %q was not found: %#v", expected, values)
		}
	}
	for _, ordinary := range []string{"42", "trace-ignored"} {
		if slices.Contains(values, ordinary) {
			t.Errorf("ordinary response value %q was marked sensitive: %#v", ordinary, values)
		}
	}
}

func TestIsSensitiveHeader(t *testing.T) {
	for _, name := range []string{"Authorization", "Set-Cookie", "X-Api-Key", "X-Session-Token", "X_Client_Secret"} {
		if !IsSensitiveHeader(name) {
			t.Errorf("sensitive header %q was not recognized", name)
		}
	}
	for _, name := range []string{"Accept", "Content-Type", "X-Trace-ID"} {
		if IsSensitiveHeader(name) {
			t.Errorf("ordinary header %q was marked sensitive", name)
		}
	}
}
