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
	return Runner{
		suite: suite,
	}
}

func (runner *Runner) Execute() (Response, error) {
	method := runner.suite.Method
	url := runner.suite.Uri
	requestBody := runner.suite.Body
	requestBodyReader := bytes.NewBuffer([]byte(requestBody))

	request, err := http.NewRequest(method, url, requestBodyReader)
	if err != nil {
		return Response{}, err
	}

	bodyType := runner.suite.BodyType
	contentType := ""
	switch bodyType {
	case "json":
		contentType = "application/json"
	case "xml":
		contentType = "application/xml"
	case "graphql":
		contentType = "application/graphql"
	default:
		contentType = bodyType // Fallback to what's provided
	}

	if contentType != "" && contentType != "raw" {
		value := fmt.Sprintf("%v; charset=utf-8", contentType)
		request.Header.Add("Content-Type", value)
	}

	requestHeader := runner.suite.Header
	for key, value := range requestHeader {
		// If Content-Type is already set by bodyType, we should be careful not to duplicate or conflict.
		// However, the user might want to override it via headers.
		if strings.EqualFold(key, "Content-Type") {
			// If already set from bodyType, and user provided one, we use the user's one.
			request.Header.Set("Content-Type", value)
		} else {
			request.Header.Add(key, value)
		}
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
		return Response{}, err
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
