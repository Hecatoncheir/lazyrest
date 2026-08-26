package producer

import (
	nethttp "net/http"
	"testing"

	parserhttp "github.com/Hecatoncheir/lazyrest/parser/http"
	"github.com/Hecatoncheir/lazyrest/runner"
)

func TestProducerExposesAndClearsCapturedResponseSummaries(t *testing.T) {
	widget := New()
	widget.recordResponse(parserhttp.HttpSuite{
		SourceFilePath: "requests.http",
		Name:           "login",
	}, runner.Response{
		Code:   "204 No Content",
		Header: nethttp.Header{"X-Session": {"secret"}},
		Body:   "token",
	}, nil)

	captures := widget.CapturedResponses()
	if len(captures) != 1 || captures[0].Name != "login" || captures[0].Status != "204 No Content" {
		t.Fatalf("unexpected captures: %+v", captures)
	}
	if removed := widget.ClearCapturedResponses(); removed != 1 || len(widget.CapturedResponses()) != 0 {
		t.Fatalf("clear removed %d captures, remaining=%v", removed, widget.CapturedResponses())
	}
}
