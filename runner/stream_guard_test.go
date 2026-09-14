package runner

import (
	"context"
	"errors"
	"testing"

	parser "github.com/Hecatoncheir/lazyrest/parser/http"
)

// Routing a stream to the one shot runner would otherwise fail obscurely: the
// request would be built with a ws:// URL and bounded by the run timeout.
func TestExecuteRefusesAStreamTransport(t *testing.T) {
	for _, transport := range []parser.Transport{parser.TransportWebSocket, parser.TransportTCP} {
		suite := parser.HttpSuite{
			Name:      "stream",
			Method:    "WEBSOCKET",
			Uri:       "wss://example.com/socket",
			Transport: transport,
		}
		runner := NewFromSuite(suite)
		_, err := runner.Execute(context.Background(), nil)
		if !errors.Is(err, ErrStreamTransport) {
			t.Fatalf("%v: err = %v, want ErrStreamTransport", transport, err)
		}
	}
}

func TestExecuteStillRunsAnOrdinaryRequest(t *testing.T) {
	suite := parser.HttpSuite{Name: "plain", Method: "GET", Uri: "http://127.0.0.1:1/"}
	runner := NewFromSuite(suite)
	_, err := runner.Execute(context.Background(), nil)
	if errors.Is(err, ErrStreamTransport) {
		t.Fatal("an ordinary request was refused as a stream")
	}
}
