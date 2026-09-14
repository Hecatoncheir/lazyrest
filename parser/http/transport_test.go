package http

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTransportForURI(t *testing.T) {
	cases := []struct {
		uri  string
		want Transport
	}{
		{"ws://example.com/socket", TransportWebSocket},
		{"wss://example.com/socket", TransportWebSocket},
		{"WSS://EXAMPLE.COM/socket", TransportWebSocket},
		{"  wss://example.com/socket  ", TransportWebSocket},
		{"tcp://127.0.0.1:6379", TransportTCP},
		{"mqtt://127.0.0.1:1883", TransportMQTT},
		{"MQTT://broker", TransportMQTT},
		{"mqtts://broker:8883", TransportMQTT},
		{"http://example.com", TransportHTTP},
		{"https://example.com", TransportHTTP},
		{"example.com/path", TransportHTTP},
		{"", TransportHTTP},
		// A host that merely starts with the scheme letters is not a scheme.
		{"wsserver.example.com", TransportHTTP},
	}
	for _, testCase := range cases {
		if got := TransportForURI(testCase.uri); got != testCase.want {
			t.Errorf("TransportForURI(%q) = %v, want %v", testCase.uri, got, testCase.want)
		}
	}
}

func TestTransportForRequestUsesTheFileExtension(t *testing.T) {
	cases := []struct {
		name     string
		uri      string
		filePath string
		want     Transport
	}{
		{"websocket in a request file", "wss://example.com", "/x/api.http", TransportWebSocket},
		{"plain request in a request file", "https://example.com", "/x/api.http", TransportHTTP},
		{"bare address in a socket file", "127.0.0.1:6379", "/x/cache.socket", TransportTCP},
		{"scheme wins in a socket file", "tcp://127.0.0.1:6379", "/x/cache.socket", TransportTCP},
		{"extension is matched case insensitively", "127.0.0.1:6379", "/x/cache.SOCKET", TransportTCP},
	}
	for _, testCase := range cases {
		if got := TransportForRequest(testCase.uri, testCase.filePath); got != testCase.want {
			t.Errorf("%s: got %v, want %v", testCase.name, got, testCase.want)
		}
	}
}

func TestTransportIsStream(t *testing.T) {
	if TransportHTTP.IsStream() {
		t.Error("an ordinary request reports itself as a stream")
	}
	if !TransportWebSocket.IsStream() || !TransportTCP.IsStream() || !TransportMQTT.IsStream() {
		t.Error("a stream transport does not report itself as one")
	}
}

func parseSource(t *testing.T, fileName, source string) []HttpSuite {
	t.Helper()
	path := filepath.Join(t.TempDir(), fileName)
	if err := os.WriteFile(path, []byte(source), 0o600); err != nil {
		t.Fatalf("write %s: %v", fileName, err)
	}
	parser, err := NewParser()
	if err != nil {
		t.Fatalf("new parser: %v", err)
	}
	suites, err := parser.GetSuitesFromFile(path)
	if err != nil {
		t.Fatalf("parse %s: %v", fileName, err)
	}
	return suites
}

func TestParseMarksAWebSocketRequestInARequestFile(t *testing.T) {
	suites := parseSource(t, "api.http", `### login
POST https://example.com/login

### stream
WEBSOCKET wss://example.com/socket
`)
	if len(suites) != 2 {
		t.Fatalf("parsed %d suites, want 2", len(suites))
	}
	if suites[0].Transport != TransportHTTP {
		t.Errorf("login transport = %v, want http", suites[0].Transport)
	}
	if suites[1].Transport != TransportWebSocket {
		t.Errorf("stream transport = %v, want websocket", suites[1].Transport)
	}
}

func TestParseMarksEveryRequestInASocketFile(t *testing.T) {
	suites := parseSource(t, "cache.socket", `### ping
SOCKET tcp://127.0.0.1:6379

PING
`)
	if len(suites) != 1 {
		t.Fatalf("parsed %d suites, want 1", len(suites))
	}
	if suites[0].Transport != TransportTCP {
		t.Errorf("transport = %v, want tcp", suites[0].Transport)
	}
}
