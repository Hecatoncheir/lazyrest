package runner

import (
	"bytes"
	"fmt"
	"io"
	parser "lazyrest/parser/http"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

type ProgressCallback func(current, total int64)

type progressReader struct {
	r        io.Reader
	total    int64
	current  int64
	callback ProgressCallback
}

func (pr *progressReader) Read(p []byte) (n int, err error) {
	n, err = pr.r.Read(p)
	pr.current += int64(n)
	if pr.callback != nil && pr.total > 0 {
		pr.callback(pr.current, pr.total)
	}
	return n, err
}

type Runner struct {
	suite parser.HttpSuite
}

func NewFromSuite(suite parser.HttpSuite) Runner {
	return Runner{
		suite: suite,
	}
}

func (runner *Runner) Execute(onProgress ProgressCallback) (Response, error) {
	if runner.suite.IsHurl {
		return runner.executeHurl(onProgress)
	}

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
		// If Content-Type is already set from bodyType, and user provided one, we use the user's one.
		if strings.EqualFold(key, "Content-Type") {
			request.Header.Set("Content-Type", value)
		} else {
			request.Header.Add(key, value)
		}
	}

	client := http.Client{}
	begin := time.Now()
	result, err := client.Do(request)
	if err != nil {
		return Response{}, err
	}
	end := time.Now()

	defer result.Body.Close()

	total := int64(result.ContentLength)
	var bodyReader io.Reader = result.Body
	if onProgress != nil && total > 0 {
		bodyReader = &progressReader{
			r:        result.Body,
			total:    total,
			callback: onProgress,
		}
	}

	responseBody, err := io.ReadAll(bodyReader)
	if err != nil {
		return Response{}, err
	}

	diff := end.Sub(begin)

	response := Response{
		Body:          string(responseBody),
		ContentLength: len(responseBody),
		Code:          fmt.Sprintf("%d %s", result.StatusCode, http.StatusText(result.StatusCode)),
		Time:          diff,
	}
	return response, nil
}

func (runner *Runner) executeHurl(onProgress ProgressCallback) (Response, error) {
	cmd := exec.Command("hurl", "--json", runner.suite.HurlFilePath)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	// Hurl returns non-zero exit code if assertions fail. 
	// Even if it's a "failure" in terms of assertions, we might still want to see the JSON output.
	if err != nil {
		if stdout.Len() == 0 {
			return Response{}, fmt.Errorf("hurl error: %v, stderr: %s", err, stderr.String())
		}
	}

	code := "OK"
	if err != nil {
		code = "FAILED"
	}

	return Response{
		Body:          stdout.String(),
		Code:          code,
		Time:          0,
		ContentLength: stdout.Len(),
	}, nil
}
