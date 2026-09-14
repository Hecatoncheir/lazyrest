package suites

import "testing"

func TestFramePreview(t *testing.T) {
	cases := []struct {
		name string
		body string
		want string
	}{
		{
			name: "a STOMP frame is reduced to its command",
			body: "CONNECT\naccept-version:1.2\nhost:/\n\n\\0",
			want: "CONNECT …",
		},
		{
			name: "a one line frame is left whole",
			body: `PING\r\n`,
			want: `PING\r\n`,
		},
		{
			name: "a CRLF frame keeps no stray carriage return",
			body: "SUBSCRIBE\r\nid:0\r\n",
			want: "SUBSCRIBE …",
		},
		// A lone brace would say less than the squashed text it replaced.
		{
			name: "a JSON payload is left for the caller to squash",
			body: "{\n  \"hello\": \"world\"\n}",
			want: "{\n  \"hello\": \"world\"\n}",
		},
		{
			name: "an array payload is left alone too",
			body: "[\n  1\n]",
			want: "[\n  1\n]",
		},
		{
			name: "trailing blank lines do not count as a second line",
			body: "DISCONNECT\n\n",
			want: "DISCONNECT",
		},
		{name: "an empty body stays empty", body: "", want: ""},
	}
	for _, testCase := range cases {
		if got := framePreview(testCase.body); got != testCase.want {
			t.Errorf("%s: framePreview(%q) = %q, want %q", testCase.name, testCase.body, got, testCase.want)
		}
	}
}
