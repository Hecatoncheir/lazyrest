package producer

import parserhttp "github.com/Hecatoncheir/lazyrest/parser/http"

// CapturedResponses returns safe metadata for named responses in this session.
func (widget *Producer) CapturedResponses() []parserhttp.CapturedResponse {
	if widget == nil {
		return nil
	}
	return widget.responses.Captures()
}

// ClearCapturedResponses discards response-reference state for this session.
func (widget *Producer) ClearCapturedResponses() int {
	if widget == nil {
		return 0
	}
	return widget.responses.Clear()
}
