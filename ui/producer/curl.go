package producer

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	parserhttp "github.com/Hecatoncheir/lazyrest/parser/http"
	"github.com/Hecatoncheir/lazyrest/runner"
)

var ErrHurlCurlUnsupported = errors.New("a Hurl workflow cannot be represented as one cURL command")

// CurlCommand renders the request currently shown in Producer as a POSIX-shell
// command. The command intentionally contains the actual request values because
// copying it is an explicit export action.
func (widget *Producer) CurlCommand() (string, error) {
	suite, ok := widget.CurrentRequest()
	if !ok {
		return "", errors.New("no runnable request is available")
	}
	return BuildCurlCommand(suite)
}

func BuildCurlCommand(suite parserhttp.HttpSuite) (string, error) {
	if suite.IsHurl {
		return "", ErrHurlCurlUnsupported
	}
	if strings.TrimSpace(suite.Method) == "" || strings.TrimSpace(suite.Uri) == "" {
		return "", errors.New("request method and URL are required")
	}
	body, impliedContentType, err := runner.RequestPayload(suite)
	if err != nil {
		return "", fmt.Errorf("build request payload: %w", err)
	}

	headers := suite.Header.Clone()
	if headers == nil {
		headers = make(map[string][]string)
	}
	if impliedContentType != "" && impliedContentType != "raw" && !hasHeader(headers, "Content-Type") {
		headers.Set("Content-Type", impliedContentType+"; charset=utf-8")
	}
	names := make([]string, 0, len(headers))
	for name := range headers {
		names = append(names, name)
	}
	slices.SortFunc(names, func(left, right string) int {
		return strings.Compare(strings.ToLower(left), strings.ToLower(right))
	})

	arguments := []string{
		"curl",
		"--globoff",
		"--request " + shellQuote(suite.Method),
		"--url " + shellQuote(suite.Uri),
	}
	for _, name := range names {
		for _, value := range headers[name] {
			arguments = append(arguments, "--header "+shellQuote(name+": "+value))
		}
	}
	if body != "" {
		arguments = append(arguments, "--data-raw "+shellQuote(body))
	}
	return strings.Join(arguments, " \\\n  "), nil
}

func hasHeader(headers map[string][]string, name string) bool {
	for candidate := range headers {
		if strings.EqualFold(candidate, name) {
			return true
		}
	}
	return false
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}
