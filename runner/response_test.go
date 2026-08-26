package runner

import (
	"net/http"
	"testing"
)

func TestResponseIsSuccessful(t *testing.T) {
	tests := []struct {
		name     string
		response Response
		want     bool
	}{
		{name: "HTTP success", response: Response{StatusCode: http.StatusNoContent}, want: true},
		{name: "HTTP redirect", response: Response{StatusCode: http.StatusFound}},
		{name: "HTTP client error", response: Response{StatusCode: http.StatusNotFound}},
		{name: "code fallback", response: Response{Code: "201 Created"}, want: true},
		{name: "Hurl success", response: Response{Code: "OK"}, want: true},
		{name: "Hurl failure", response: Response{Code: "FAILED"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := test.response.IsSuccessful(); got != test.want {
				t.Fatalf("IsSuccessful() = %v, want %v", got, test.want)
			}
		})
	}
}
