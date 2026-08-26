package producer

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"
	"unicode/utf8"

	parserhttp "github.com/Hecatoncheir/lazyrest/parser/http"
	"github.com/Hecatoncheir/lazyrest/runner"
)

// historyVersion is the format written by this version of lazyrest. Version 1
// stored a single value per request header; it is still read.
const historyVersion = 2

// maxHistoryBodyBytes caps the body kept per entry. Without it a run of large
// responses would grow the history file to hundreds of megabytes and make every
// following run re-encode all of it.
const maxHistoryBodyBytes = 64 << 10

type storedHistory struct {
	Version int                  `json:"version"`
	Entries []storedHistoryEntry `json:"entries"`
}

type storedHistoryEntry struct {
	Suite     storedSuite     `json:"suite"`
	Response  runner.Response `json:"response"`
	Error     string          `json:"error,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
}

// storedSuite shadows the header of the embedded suite so that both the current
// and the version 1 header format can be read.
type storedSuite struct {
	parserhttp.HttpSuite
	Header storedHeader
}

type storedHeader http.Header

func newStoredSuite(suite parserhttp.HttpSuite) storedSuite {
	return storedSuite{HttpSuite: suite, Header: storedHeader(suite.Header)}
}

func (stored storedSuite) suite() parserhttp.HttpSuite {
	restored := stored.HttpSuite
	restored.Header = http.Header(stored.Header)
	return restored
}

func (header *storedHeader) UnmarshalJSON(contents []byte) error {
	var current map[string][]string
	if err := json.Unmarshal(contents, &current); err == nil {
		*header = storedHeader(current)
		return nil
	}
	var legacy map[string]string
	if err := json.Unmarshal(contents, &legacy); err != nil {
		return err
	}
	converted := make(storedHeader, len(legacy))
	for name, value := range legacy {
		converted[name] = []string{value}
	}
	*header = converted
	return nil
}

func (widget *Producer) loadHistory() error {
	if widget.historyPath == "" {
		return nil
	}
	contents, err := os.ReadFile(widget.historyPath)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read history: %w", err)
	}
	var stored storedHistory
	if err := json.Unmarshal(contents, &stored); err != nil {
		return fmt.Errorf("parse history: %w", err)
	}
	if stored.Version < 1 || stored.Version > historyVersion {
		return fmt.Errorf("unsupported history version %d", stored.Version)
	}
	if len(stored.Entries) > maxHistoryEntries {
		stored.Entries = stored.Entries[len(stored.Entries)-maxHistoryEntries:]
	}
	widget.history = make([]HistoryEntry, 0, len(stored.Entries))
	for _, entry := range stored.Entries {
		var entryErr error
		if entry.Error != "" {
			entryErr = errors.New(entry.Error)
		}
		widget.history = append(widget.history, HistoryEntry{Suite: entry.Suite.suite(), Response: entry.Response, Err: entryErr, CreatedAt: entry.CreatedAt})
	}
	widget.historyIndex = len(widget.history) - 1
	return nil
}

// persistHistory writes the history without blocking the UI: encoding and
// writing hundreds of kilobytes on the draw goroutine froze the interface after
// every request. Only the newest snapshot reaches the file.
func (widget *Producer) persistHistory() {
	if widget.historyPath == "" {
		return
	}
	stored := widget.buildStoredHistory()

	widget.historyMutex.Lock()
	widget.historyRequested++
	generation := widget.historyRequested
	widget.historyMutex.Unlock()

	widget.historyWrites.Add(1)
	go func() {
		defer widget.historyWrites.Done()
		widget.historyMutex.Lock()
		defer widget.historyMutex.Unlock()
		if generation <= widget.historyWritten {
			return
		}
		widget.historyWritten = generation
		_ = writeHistory(widget.historyPath, stored)
	}()
}

func (widget *Producer) saveHistory() error {
	if widget.historyPath == "" {
		return nil
	}
	return writeHistory(widget.historyPath, widget.buildStoredHistory())
}

func (widget *Producer) buildStoredHistory() storedHistory {
	stored := storedHistory{Version: historyVersion, Entries: make([]storedHistoryEntry, 0, len(widget.history))}
	for _, entry := range widget.history {
		errorText := ""
		if entry.Err != nil {
			errorText = entry.Err.Error()
		}
		suite, response := boundedEntryBodies(entry.Suite, entry.Response)
		stored.Entries = append(stored.Entries, storedHistoryEntry{Suite: newStoredSuite(suite), Response: response, Error: errorText, CreatedAt: entry.CreatedAt})
	}
	return stored
}

// boundedEntryBodies limits what a single entry contributes to the file. The
// in-memory entry keeps its full bodies so that the pane still shows them.
func boundedEntryBodies(suite parserhttp.HttpSuite, response runner.Response) (parserhttp.HttpSuite, runner.Response) {
	if body, cut := truncateBody(suite.Body); cut {
		suite.Body = body + "\n... truncated"
	}
	if body, cut := truncateBody(response.Body); cut {
		response.Body = body
		response.Truncated = true
	}
	return suite, response
}

func truncateBody(body string) (string, bool) {
	if len(body) <= maxHistoryBodyBytes {
		return body, false
	}
	cut := body[:maxHistoryBodyBytes]
	for len(cut) > 0 && !utf8.ValidString(cut) {
		cut = cut[:len(cut)-1]
	}
	return cut, true
}

func writeHistory(path string, stored storedHistory) error {
	contents, err := json.MarshalIndent(stored, "", "  ")
	if err != nil {
		return fmt.Errorf("encode history: %w", err)
	}
	directory := filepath.Dir(path)
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return fmt.Errorf("create history directory: %w", err)
	}
	temporary, err := os.CreateTemp(directory, ".history-*.tmp")
	if err != nil {
		return fmt.Errorf("create history file: %w", err)
	}
	temporaryPath := temporary.Name()
	defer func() { _ = os.Remove(temporaryPath) }()
	if err := temporary.Chmod(0o600); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("secure history file: %w", err)
	}
	if _, err := temporary.Write(contents); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("write history: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close history: %w", err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("replace history: %w", err)
	}
	return nil
}

func sanitizedHistoryEntry(suite parserhttp.HttpSuite, response runner.Response, err error, createdAt time.Time) HistoryEntry {
	secrets := append([]string(nil), suite.SecretValues...)
	suite.Name = redactSecrets(suite.Name, secrets)
	suite.Uri = redactSecrets(suite.Uri, secrets)
	suite.Body = redactSecrets(suite.Body, secrets)
	suite.HurlFilePath = ""
	suite.Variables = nil
	suite.Header = sanitizeHeaders(suite.Header, secrets)
	suite.SecretValues = nil
	response.Body = redactSecrets(response.Body, secrets)
	response.Header = sanitizeHeaders(response.Header, secrets)
	var sanitizedError error
	if err != nil {
		sanitizedError = errors.New(redactSecrets(err.Error(), secrets))
	}
	return HistoryEntry{Suite: suite, Response: response, Err: sanitizedError, CreatedAt: createdAt}
}

func sanitizeHeaders(headers http.Header, secrets []string) http.Header {
	result := make(http.Header, len(headers))
	for key, values := range headers {
		cleanKey := redactSecrets(key, secrets)
		if isSensitiveHeader(key) {
			result[cleanKey] = []string{"<redacted>"}
			continue
		}
		cleanValues := make([]string, len(values))
		for index, value := range values {
			cleanValues[index] = redactSecrets(value, secrets)
		}
		result[cleanKey] = cleanValues
	}
	return result
}
