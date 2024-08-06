package runner

import (
	"bytes"
	"fmt"
	"io"
	parser "lazyrest/parser/http"
	"net/http"
	"strings"
	"time"
)

type Runner struct {
	suite parser.HttpSuite
}

func NewFromSuite(suite parser.HttpSuite) Runner {
	runner := Runner{
		suite: suite,
	}
	return runner
}

func (runner *Runner) Execute() (Response, error) {
	method := runner.suite.Method()
	url := runner.suite.Uri()
	requestBody := runner.suite.Body()
	requestBodyReader := bytes.NewBuffer([]byte(requestBody))

	request, err := http.NewRequest(method, url, requestBodyReader)
	if err != nil {
		return Response{}, err
	}

	bodyType := runner.suite.BodyType()
	value := fmt.Sprintf("%v; charset=utf-8", bodyType)
	request.Header.Add("Content-Type", value)

	requestHeader := runner.suite.Header()
	for key, value := range requestHeader {
		if key == "Content-Type" && !strings.Contains(value, "charset") {
			updatedValue := ""
			if strings.Contains(value, ";") {
				updatedValue = fmt.Sprintf("%v charset=utf-8", value)
			} else {
				updatedValue = fmt.Sprintf("%v; charset=utf-8", value)
			}
			request.Header.Add("Content-Type", updatedValue)
		}
		request.Header.Add(key, value)
	}

	client := http.Client{
		// Timeout: time.Duration(3 * time.Minute),
	}
	begin := time.Now()
	result, err := client.Do(request)
	end := time.Now()
	if err != nil {
		return Response{}, err
	}

	defer result.Body.Close()

	responseBody, err := io.ReadAll(result.Body)
	if err != nil {
		return Response{}, nil
	}

	diff := end.Sub(begin)

	response := Response{
		Body:          string(responseBody),
		ContentLength: len(responseBody),
		Code:          result.Status,
		Time:          diff,
	}
	return response, nil
}
