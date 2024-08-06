package runner

import (
	"fmt"
	"time"

	"github.com/hako/durafmt"
)

type Response struct {
	Code          string
	Time          time.Duration
	ContentLength int
	Body          string
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
			// "Time: %v (); "+
			"Content length: %v \n",
		response.Code,
		response.ContentLength,
	)
	return text
}
