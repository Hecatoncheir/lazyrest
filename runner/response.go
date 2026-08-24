package runner

import (
	"fmt"
	"net/http"
	"time"

	"github.com/hako/durafmt"
)

type Response struct {
	Code          string
	Time          time.Duration
	ContentLength int
	Body          string
	Truncated     bool
	Header        http.Header
	Protocol      string
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
