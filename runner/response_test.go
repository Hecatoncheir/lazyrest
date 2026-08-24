package runner

import (
	"net/http"
	"testing"
	"time"
)

func TestResponse_ToMiniString(t *testing.T) {
	resp := &Response{
		Code:          "200",
		Time:          7035 * time.Millisecond,
		ContentLength: 389,
		Body:          "some body",
	}

	got := resp.ToMiniString()
	// The expected format is based on the implementation:
	// Response code: 200; \nTime: 7035ms (7 s 35 ms); \nContent length: 389 \n
	// Note the spaces and newlines in the implementation
	want := "Response code: 200; \nTime: 7035ms (7 seconds 35 milliseconds);  \nContent length: 389 \n"

	if got != want {
		t.Errorf("ToMiniString() = %q, want %q", got, want)
	}
}

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
