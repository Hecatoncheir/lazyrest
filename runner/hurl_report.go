package runner

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const maxHurlReportBytes = int64(16 << 20)

type hurlReportFile struct {
	Success bool              `json:"success"`
	Entries []hurlReportEntry `json:"entries"`
}

type hurlReportEntry struct {
	Asserts []hurlReportAssert `json:"asserts"`
	Calls   []hurlReportCall   `json:"calls"`
}

type hurlReportAssert struct {
	Line    int    `json:"line"`
	Message string `json:"message"`
	Success bool   `json:"success"`
}

type hurlReportCall struct {
	Response hurlReportResponse `json:"response"`
}

type hurlReportResponse struct {
	Body        string             `json:"body"`
	Headers     []hurlReportHeader `json:"headers"`
	HTTPVersion string             `json:"http_version"`
	Status      int                `json:"status"`
}

type hurlReportHeader struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func responseFromHurlReport(directory string, elapsed time.Duration, maxBodyBytes int64) (Response, bool, error) {
	contents, err := readLimitedFile(filepath.Join(directory, "report.json"), maxHurlReportBytes)
	if errors.Is(err, os.ErrNotExist) {
		return Response{}, false, nil
	}
	if err != nil {
		return Response{}, true, fmt.Errorf("read Hurl JSON report: %w", err)
	}
	var files []hurlReportFile
	if err := json.Unmarshal(contents, &files); err != nil {
		return Response{}, true, fmt.Errorf("parse Hurl JSON report: %w", err)
	}
	if len(files) == 0 || len(files[len(files)-1].Entries) == 0 {
		return Response{}, false, nil
	}
	file := files[len(files)-1]
	entry := file.Entries[len(file.Entries)-1]
	if len(entry.Calls) == 0 {
		return Response{}, false, nil
	}
	call := entry.Calls[len(entry.Calls)-1]
	body, fullLength, truncated, err := readHurlBody(directory, call.Response.Body, maxBodyBytes)
	if err != nil {
		return Response{}, true, err
	}
	headers := make(http.Header)
	for _, header := range call.Response.Headers {
		headers.Add(header.Name, header.Value)
	}
	failures := make([]string, 0)
	for _, currentEntry := range file.Entries {
		for _, assertion := range currentEntry.Asserts {
			if assertion.Success {
				continue
			}
			message := strings.TrimSpace(assertion.Message)
			if message == "" {
				message = fmt.Sprintf("assertion at line %d failed", assertion.Line)
			}
			failures = append(failures, message)
		}
	}
	status := call.Response.Status
	return Response{
		Body:            string(body),
		Code:            fmt.Sprintf("%d %s", status, http.StatusText(status)),
		StatusCode:      status,
		Time:            elapsed,
		ContentLength:   fullLength,
		StoredLength:    len(body),
		Truncated:       truncated,
		Header:          headers,
		Protocol:        call.Response.HTTPVersion,
		AssertionErrors: failures,
	}, true, nil
}

func readHurlBody(reportDirectory, relativePath string, limit int64) ([]byte, int, bool, error) {
	if relativePath == "" {
		return nil, 0, false, nil
	}
	cleaned := filepath.Clean(relativePath)
	if filepath.IsAbs(cleaned) || cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) {
		return nil, 0, false, fmt.Errorf("hurl report body path escapes the report directory: %s", relativePath)
	}
	path := filepath.Join(reportDirectory, cleaned)
	info, err := os.Stat(path)
	if err != nil {
		return nil, 0, false, fmt.Errorf("stat Hurl response body: %w", err)
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, 0, false, fmt.Errorf("read Hurl response body: %w", err)
	}
	contents, err := io.ReadAll(io.LimitReader(file, limit+1))
	closeErr := file.Close()
	if err != nil {
		return nil, 0, false, fmt.Errorf("read Hurl response body: %w", err)
	}
	if closeErr != nil {
		return nil, 0, false, fmt.Errorf("close Hurl response body: %w", closeErr)
	}
	truncated := int64(len(contents)) > limit
	if truncated {
		contents = contents[:limit]
	}
	return contents, int(info.Size()), truncated, nil
}

func readLimitedFile(path string, limit int64) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()
	contents, err := io.ReadAll(io.LimitReader(file, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(contents)) > limit {
		return nil, fmt.Errorf("file is larger than %d bytes", limit)
	}
	return contents, nil
}
