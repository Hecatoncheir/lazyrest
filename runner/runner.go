package runner

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	parser "github.com/Hecatoncheir/lazyrest/parser/http"
	"io"
	"net/http"
	"os"
	"os/exec"
	"slices"
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

	requestHeader := runner.suite.Header
	// A Content-Type declared by the request itself wins over the one implied
	// by the body type.
	if _, declared := headerValue(requestHeader, "Content-Type"); contentType != "" && contentType != "raw" && !declared {
		value := fmt.Sprintf("%v; charset=utf-8", contentType)
		request.Header.Set("Content-Type", value)
	}
	for key, values := range requestHeader {
		for _, value := range values {
			request.Header.Add(key, value)
		}
	}
	// net/http takes the Host header from the request field, not from the
	// header map.
	if host, declared := headerValue(requestHeader, "Host"); declared && host != "" {
		request.Host = host
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
		StatusCode:    result.StatusCode,
		Time:          diff,
		Truncated:     truncated,
		Header:        result.Header.Clone(),
		Protocol:      result.Proto,
	}
	return response, nil
}

// headerValue returns the first value declared for name, matching the name
// case-insensitively, and whether it was declared at all.
func headerValue(header http.Header, name string) (string, bool) {
	for key, values := range header {
		if !strings.EqualFold(key, name) {
			continue
		}
		if len(values) == 0 {
			return "", true
		}
		return values[0], true
	}
	return "", false
}

func (runner *Runner) executeHurl(ctx context.Context) (Response, error) {
	if runner.suite.HurlFilePath == "" {
		return Response{}, errors.New("Hurl file path is empty")
	}
	executable, err := exec.LookPath(runner.hurlExecutable)
	if err != nil {
		return Response{}, fmt.Errorf("Hurl executable %q was not found: %w", runner.hurlExecutable, err)
	}

	variablesPath, removeVariables, err := writeHurlVariables(runner.suite.Variables)
	if err != nil {
		return Response{}, err
	}
	defer removeVariables()

	arguments := []string{"--json"}
	if variablesPath != "" {
		arguments = append(arguments, "--variables-file", variablesPath)
	}
	arguments = append(arguments, runner.suite.HurlFilePath)

	cmd := exec.CommandContext(ctx, executable, arguments...)
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

// writeHurlVariables hands the environment to Hurl through a file instead of
// the command line, which would expose secret values to every process able to
// list the process table. The returned function removes the file.
func writeHurlVariables(variables map[string]string) (string, func(), error) {
	nothingToRemove := func() {}
	lines := make([]string, 0, len(variables))
	for name, value := range variables {
		if name == "" || strings.ContainsAny(name, "=\r\n") || strings.ContainsAny(value, "\r\n") {
			continue
		}
		lines = append(lines, name+"="+value)
	}
	if len(lines) == 0 {
		return "", nothingToRemove, nil
	}
	slices.Sort(lines)

	file, err := os.CreateTemp("", "lazyrest-variables-*.properties")
	if err != nil {
		return "", nothingToRemove, fmt.Errorf("create Hurl variables file: %w", err)
	}
	remove := func() { _ = os.Remove(file.Name()) }
	if err := file.Chmod(0o600); err != nil {
		file.Close()
		remove()
		return "", nothingToRemove, fmt.Errorf("secure Hurl variables file: %w", err)
	}
	if _, err := file.WriteString(strings.Join(lines, "\n") + "\n"); err != nil {
		file.Close()
		remove()
		return "", nothingToRemove, fmt.Errorf("write Hurl variables file: %w", err)
	}
	if err := file.Close(); err != nil {
		remove()
		return "", nothingToRemove, fmt.Errorf("close Hurl variables file: %w", err)
	}
	return file.Name(), remove, nil
}
