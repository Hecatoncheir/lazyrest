package runner

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Response struct {
	Code          string
	StatusCode    int
	Time          time.Duration
	ContentLength int
	Body          string
	Truncated     bool
	Header        http.Header
	Protocol      string
	// GraphQLErrors holds the errors a GraphQL response reported alongside its
	// 200 status.
	GraphQLErrors []string
}

func (response Response) IsSuccessful() bool {
	if len(response.GraphQLErrors) > 0 {
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
