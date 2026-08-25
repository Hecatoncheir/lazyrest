package runner

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/hako/durafmt"
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

// Response code: 200 (OK); Time: 7035ms (7 s 35 ms); Content length: 389 bytes (389 B)
func (response *Response) ToMiniString() string {
	timeFormat := fmt.Sprintf(
		"Time: %vms (%v); ",
		response.Time.Milliseconds(),
		durafmt.Parse(response.Time).String(),
	)

	text := fmt.Sprintf(
		"Response code: %v; \n"+
			timeFormat+" \n"+
			"Content length: %v \n",
		response.Code,
		response.ContentLength,
	)
	if response.Truncated {
		text += "Response body was truncated.\n"
	}
	return text
}
