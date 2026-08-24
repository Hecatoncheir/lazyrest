package producer

import (
	"context"
	"errors"
	"github.com/Hecatoncheir/lazyrest/parser/http"
	"github.com/Hecatoncheir/lazyrest/runner"
	"strings"
	"testing"
)

func TestStartRunCancelsPreviousRun(t *testing.T) {
	widget := New()
	firstContext, firstID := widget.StartRun()
	secondContext, secondID := widget.StartRun()

	select {
	case <-firstContext.Done():
	default:
		t.Fatal("previous run was not cancelled")
	}
	if widget.IsCurrentRun(firstID) || !widget.IsCurrentRun(secondID) {
		t.Fatal("current run identifier was not updated")
	}

	widget.CancelActive()
	if !errors.Is(secondContext.Err(), context.Canceled) {
		t.Fatalf("active run was not cancelled: %v", secondContext.Err())
	}
}

func TestRenderExecutionResult_RedactsSensitiveHeaders(t *testing.T) {
	suite := http.HttpSuite{
		Method: "GET",
		Uri:    "https://example.com",
		Header: map[string]string{
			"Authorization": "Bearer secret",
			"Accept":        "application/json",
		},
	}
	text := renderExecutionResult(suite, runner.Response{Code: "200 OK"}, nil)

	if strings.Contains(text, "Bearer secret") {
		t.Fatal("sensitive header was rendered")
	}
	if !strings.Contains(text, "Authorization: <redacted>") {
		t.Fatalf("redaction marker is missing: %q", text)
	}
	if !strings.Contains(text, "Accept: application/json") {
		t.Fatalf("regular header is missing: %q", text)
	}
}
