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
	Suite          storedSuite    `json:"suite"`
	Response       storedResponse `json:"response"`
	Error          string         `json:"error,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	DetailsOmitted bool           `json:"details_omitted,omitempty"`
}

// storedSuite is an explicit allowlist of request fields that are safe and
// useful to restore. Keeping it separate from HttpSuite prevents a newly added
// runtime field from silently becoming part of the on-disk history format.
type storedSuite struct {
	Name             string       `json:"Name"`
	Method           string       `json:"Method"`
	URI              string       `json:"Uri"`
	Header           storedHeader `json:"Header"`
	Body             string       `json:"Body"`
	BodyType         string       `json:"BodyType"`
	IsHurl           bool         `json:"IsHurl"`
	HurlEntry        int          `json:"HurlEntry,omitempty"`
	GraphQLVariables string       `json:"GraphQLVariables,omitempty"`
	GraphQLOperation string       `json:"GraphQLOperation,omitempty"`
}

// storedResponse serves the same purpose for responses. Its JSON field names
// match the history written before the allowlist was introduced.
type storedResponse struct {
	Code          string        `json:"Code"`
	StatusCode    int           `json:"StatusCode"`
	Time          time.Duration `json:"Time"`
	ContentLength int           `json:"ContentLength"`
	Body          string        `json:"Body"`
	Truncated     bool          `json:"Truncated"`
	Header        storedHeader  `json:"Header"`
	Protocol      string        `json:"Protocol"`
	GraphQLErrors []string      `json:"GraphQLErrors,omitempty"`
}

type storedHeader http.Header

func newStoredSuite(suite parserhttp.HttpSuite) storedSuite {
	return storedSuite{
		Name:             suite.Name,
		Method:           suite.Method,
		URI:              suite.Uri,
		Header:           cloneStoredHeader(suite.Header),
		Body:             suite.Body,
		BodyType:         suite.BodyType,
		IsHurl:           suite.IsHurl,
		HurlEntry:        suite.HurlEntry,
		GraphQLVariables: suite.GraphQLVariables,
		GraphQLOperation: suite.GraphQLOperation,
	}
}

func (stored storedSuite) suite() parserhttp.HttpSuite {
	return parserhttp.HttpSuite{
		Name:             stored.Name,
		Method:           stored.Method,
		Uri:              stored.URI,
		Header:           http.Header(stored.Header).Clone(),
		Body:             stored.Body,
		BodyType:         stored.BodyType,
		IsHurl:           stored.IsHurl,
		HurlEntry:        stored.HurlEntry,
		GraphQLVariables: stored.GraphQLVariables,
		GraphQLOperation: stored.GraphQLOperation,
	}
}

func newStoredResponse(response runner.Response) storedResponse {
	return storedResponse{
		Code:          response.Code,
		StatusCode:    response.StatusCode,
		Time:          response.Time,
		ContentLength: response.ContentLength,
		Body:          response.Body,
		Truncated:     response.Truncated,
		Header:        cloneStoredHeader(response.Header),
		Protocol:      response.Protocol,
		GraphQLErrors: append([]string(nil), response.GraphQLErrors...),
	}
}

func (stored storedResponse) response() runner.Response {
	return runner.Response{
		Code:          stored.Code,
		StatusCode:    stored.StatusCode,
		Time:          stored.Time,
		ContentLength: stored.ContentLength,
		Body:          stored.Body,
		Truncated:     stored.Truncated,
		Header:        http.Header(stored.Header).Clone(),
		Protocol:      stored.Protocol,
		GraphQLErrors: append([]string(nil), stored.GraphQLErrors...),
	}
}

func cloneStoredHeader(header http.Header) storedHeader {
	return storedHeader(header.Clone())
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
	restored := make([]HistoryEntry, 0, len(stored.Entries))
	rewriteMetadata := false
	for _, entry := range stored.Entries {
		var entryErr error
		if entry.Error != "" {
			entryErr = errors.New(entry.Error)
		}
		historyEntry := HistoryEntry{
			Suite:          entry.Suite.suite(),
			Response:       entry.Response.response(),
			Err:            entryErr,
			CreatedAt:      entry.CreatedAt,
			DetailsOmitted: entry.DetailsOmitted,
		}
		if widget.historyModeValue() == HistoryMetadata && !historyEntry.DetailsOmitted {
			historyEntry = metadataOnlyHistoryEntry(historyEntry)
			rewriteMetadata = true
		}
		restored = append(restored, historyEntry)
	}
	widget.historyDataMutex.Lock()
	widget.history = restored
	widget.historyIndex = len(widget.history) - 1
	widget.historyDataMutex.Unlock()
	if rewriteMetadata {
		widget.persistHistory()
	}
	return nil
}

// persistHistory writes the history without blocking the UI: encoding and
// writing hundreds of kilobytes on the draw goroutine froze the interface after
// every request. Only the newest snapshot reaches the file.
func (widget *Producer) persistHistory() {
	if widget.historyPath == "" {
		return
	}

	widget.historyMutex.Lock()
	widget.historyRequested++
	generation := widget.historyRequested
	stored := widget.buildStoredHistory()
	widget.historyMutex.Unlock()

	widget.historyWrites.Add(1)
	go func() {
		defer widget.historyWrites.Done()
		widget.historyMutex.Lock()
		if generation <= widget.historyWritten {
			widget.historyMutex.Unlock()
			return
		}
		widget.historyWritten = generation
		err := writeHistory(widget.historyPath, stored)
		widget.historyMutex.Unlock()
		if err != nil {
			widget.reportHistoryError("persist", err)
		}
	}()
}

func (widget *Producer) saveHistory() error {
	if widget.historyPath == "" {
		return nil
	}
	return writeHistory(widget.historyPath, widget.buildStoredHistory())
}

func (widget *Producer) buildStoredHistory() storedHistory {
	widget.historyDataMutex.RLock()
	defer widget.historyDataMutex.RUnlock()
	stored := storedHistory{Version: historyVersion, Entries: make([]storedHistoryEntry, 0, len(widget.history))}
	for _, entry := range widget.history {
		if widget.historyModeValue() == HistoryMetadata {
			entry = metadataOnlyHistoryEntry(entry)
		}
		errorText := ""
		if entry.Err != nil {
			errorText = entry.Err.Error()
		}
		suite, response := boundedEntryBodies(entry.Suite, entry.Response)
		stored.Entries = append(stored.Entries, storedHistoryEntry{
			Suite:          newStoredSuite(suite),
			Response:       newStoredResponse(response),
			Error:          errorText,
			CreatedAt:      entry.CreatedAt,
			DetailsOmitted: entry.DetailsOmitted,
		})
	}
	return stored
}

func metadataOnlyHistoryEntry(entry HistoryEntry) HistoryEntry {
	entry.Suite = parserhttp.HttpSuite{
		Name:     entry.Suite.Name,
		Method:   entry.Suite.Method,
		BodyType: entry.Suite.BodyType,
		IsHurl:   entry.Suite.IsHurl,
	}
	entry.Response = runner.Response{
		Code:          entry.Response.Code,
		StatusCode:    entry.Response.StatusCode,
		Time:          entry.Response.Time,
		ContentLength: entry.Response.ContentLength,
		Truncated:     entry.Response.Truncated,
		Protocol:      entry.Response.Protocol,
	}
	if entry.Err != nil {
		entry.Err = errors.New("request failed; details were not persisted")
	}
	entry.DetailsOmitted = true
	entry.request = nil
	return entry
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
	secrets = append(secrets, parserhttp.SensitiveResponseValues(response.Body, response.Header)...)
	suite.Name = redactSecrets(suite.Name, secrets)
	suite.Method = redactSecrets(suite.Method, secrets)
	suite.Uri = redactSecrets(suite.Uri, secrets)
	suite.Body = redactSecrets(suite.Body, secrets)
	suite.BodyType = redactSecrets(suite.BodyType, secrets)
	suite.GraphQLVariables = redactSecrets(suite.GraphQLVariables, secrets)
	suite.GraphQLOperation = redactSecrets(suite.GraphQLOperation, secrets)
	suite.HurlFilePath = ""
	suite.Variables = nil
	suite.Header = sanitizeHeaders(suite.Header, secrets)
	suite.SecretValues = nil
	response.Code = redactSecrets(response.Code, secrets)
	response.Body = redactSecrets(response.Body, secrets)
	response.Header = sanitizeHeaders(response.Header, secrets)
	response.Protocol = redactSecrets(response.Protocol, secrets)
	response.GraphQLErrors = append([]string(nil), response.GraphQLErrors...)
	for index, message := range response.GraphQLErrors {
		response.GraphQLErrors[index] = redactSecrets(message, secrets)
	}
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
		if parserhttp.IsSensitiveHeader(key) {
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
