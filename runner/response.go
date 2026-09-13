package runner

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Response struct {
	Code       string
	StatusCode int
	Time       time.Duration
	// ContentLength is the complete response size when it is known. When
	// ContentLengthLowerBound is true, the server did not declare a size and
	// reading stopped at the configured response limit.
	ContentLength int
	// StoredLength is the number of body bytes retained in Body. It can be
	// smaller than ContentLength when the response was truncated.
	StoredLength            int
	ContentLengthLowerBound bool
	Body                    string
	Truncated               bool
	Header                  http.Header
	Protocol                string
	// Failed preserves a failed outcome when detailed error data is omitted,
	// for example by metadata-only persistent history.
	Failed bool
	// GraphQLErrors holds the errors a GraphQL response reported alongside its
	// 200 status.
	GraphQLErrors []string
	// AssertionErrors holds failed Hurl assertions for the selected exchange.
	AssertionErrors []string
}

func (response Response) IsSuccessful() bool {
	if response.Failed {
		return false
	}
	if len(response.GraphQLErrors) > 0 {
		return false
	}
	if len(response.AssertionErrors) > 0 {
		return false
	}
	statusCode := response.StatusCode
	if statusCode == 0 {
		code, _, _ := strings.Cut(response.Code, " ")
		statusCode, _ = strconv.Atoi(code)
	}
	if statusCode != 0 {
		return statusCode >= http.StatusOK && statusCode < http.StatusMultipleChoices
	}
	return response.Code == "OK"
}
