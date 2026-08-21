package runner

import (
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
