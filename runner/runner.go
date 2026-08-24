package runner

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	parser "github.com/Hecatoncheir/lazyrest/parser/http"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

type ProgressCallback func(current, total int64)

const (
	DefaultTimeout          = 30 * time.Second
	DefaultMaxResponseBytes = int64(10 << 20)
	maxStderrBytes          = int64(1 << 20)
)

type Config struct {
	Client           *http.Client
	Timeout          time.Duration
	MaxResponseBytes int64
	HurlExecutable   string
}

type progressReader struct {
	r        io.Reader
	total    int64
	current  int64
	callback ProgressCallback
}

func (pr *progressReader) Read(p []byte) (n int, err error) {
	n, err = pr.r.Read(p)
	pr.current += int64(n)
	if pr.callback != nil {
		pr.callback(pr.current, pr.total)
	}
	return n, err
}

type limitedBuffer struct {
	buffer    bytes.Buffer
	limit     int64
	truncated bool
}

func (b *limitedBuffer) Write(p []byte) (int, error) {
	written := len(p)
	remaining := b.limit - int64(b.buffer.Len())
	if remaining <= 0 {
		b.truncated = b.truncated || len(p) > 0
		return written, nil
	}
	if int64(len(p)) > remaining {
		p = p[:remaining]
		b.truncated = true
	}
	_, _ = b.buffer.Write(p)
	return written, nil
}

func (b *limitedBuffer) String() string {
	return b.buffer.String()
}

func (b *limitedBuffer) Len() int {
	return b.buffer.Len()
}

type Runner struct {
	suite            parser.HttpSuite
	client           *http.Client
	timeout          time.Duration
	maxResponseBytes int64
	hurlExecutable   string
}

func NewFromSuite(suite parser.HttpSuite) Runner {
	return NewFromSuiteWithConfig(suite, Config{})
}

func NewFromSuiteWithConfig(suite parser.HttpSuite, config Config) Runner {
	client := config.Client
	if client == nil {
		client = &http.Client{}
	}
	timeout := config.Timeout
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	maxResponseBytes := config.MaxResponseBytes
	if maxResponseBytes <= 0 {
		maxResponseBytes = DefaultMaxResponseBytes
	}
	hurlExecutable := config.HurlExecutable
	if hurlExecutable == "" {
		hurlExecutable = "hurl"
	}

	return Runner{
		suite:            suite,
		client:           client,
		timeout:          timeout,
		maxResponseBytes: maxResponseBytes,
		hurlExecutable:   hurlExecutable,
	}
}

func (runner *Runner) Execute(ctx context.Context, onProgress ProgressCallback) (Response, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, runner.timeout)
	defer cancel()

	if runner.suite.IsHurl {
		return runner.executeHurl(ctx)
	}

	method := runner.suite.Method
	url := runner.suite.Uri
	requestBody := runner.suite.Body
	requestBodyReader := bytes.NewBuffer([]byte(requestBody))

	request, err := http.NewRequestWithContext(ctx, method, url, requestBodyReader)
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

	begin := time.Now()
	result, err := runner.client.Do(request)
	if err != nil {
		return Response{}, err
	}

	defer result.Body.Close()

	total := int64(result.ContentLength)
	var bodyReader io.Reader = result.Body
	if onProgress != nil {
		bodyReader = &progressReader{
			r:        result.Body,
			total:    total,
			callback: onProgress,
		}
	}

	responseBody, err := io.ReadAll(io.LimitReader(bodyReader, runner.maxResponseBytes+1))
	if err != nil {
		return Response{}, err
	}
	truncated := int64(len(responseBody)) > runner.maxResponseBytes
	if truncated {
		responseBody = responseBody[:runner.maxResponseBytes]
	}

	diff := time.Since(begin)

	response := Response{
		Body:          string(responseBody),
		ContentLength: len(responseBody),
		Code:          fmt.Sprintf("%d %s", result.StatusCode, http.StatusText(result.StatusCode)),
		Time:          diff,
		Truncated:     truncated,
	}
	return response, nil
}

func (runner *Runner) executeHurl(ctx context.Context) (Response, error) {
	if runner.suite.HurlFilePath == "" {
		return Response{}, errors.New("Hurl file path is empty")
	}
	executable, err := exec.LookPath(runner.hurlExecutable)
	if err != nil {
		return Response{}, fmt.Errorf("Hurl executable %q was not found: %w", runner.hurlExecutable, err)
	}

	cmd := exec.CommandContext(ctx, executable, "--json", runner.suite.HurlFilePath)
	stdout := &limitedBuffer{limit: runner.maxResponseBytes}
	stderr := &limitedBuffer{limit: maxStderrBytes}
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	begin := time.Now()
	err = cmd.Run()
	diff := time.Since(begin)
	if ctx.Err() != nil {
		return Response{}, ctx.Err()
	}
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
		Time:          diff,
		ContentLength: stdout.Len(),
		Truncated:     stdout.truncated,
	}, nil
}
