package http

import (
	"encoding/json"
	"fmt"
	nethttp "net/http"
	"regexp"
	"strconv"
	"strings"
)

// responseReferencePattern matches a reference to what an earlier request
// answered, written the way the editors that share this format write it:
// {{login.response.body.$.token}} or {{login.response.headers.X-Token}}.
var responseReferencePattern = regexp.MustCompile(`\{\{\s*([^{}]+?)\.response\.(body|headers)(?:\.([^{}]*?))?\s*\}\}`)

// pathSegmentPattern splits a JSONPath step into its name and its indices, so
// that `items[0]` reads as the member `items` followed by index 0.
var pathSegmentPattern = regexp.MustCompile(`\[(\d+)\]`)

// ResponseValue is what a request answered, kept so that the requests after it
// can refer to it.
type ResponseValue struct {
	Body   string
	Header nethttp.Header
}

type responseKey struct {
	sourceFilePath string
	name           string
}

// ResponseStore holds the last answer of every named request, scoped to the
// file that declared it.
type ResponseStore map[responseKey]ResponseValue

// Record keeps a response under the source file and name of its request.
func (store ResponseStore) Record(suite HttpSuite, response ResponseValue) {
	store[responseKey{
		sourceFilePath: suite.SourceFilePath,
		name:           strings.TrimSpace(suite.Name),
	}] = response
}

// HasResponseReference reports whether text refers to an earlier response.
func HasResponseReference(text string) bool {
	return responseReferencePattern.MatchString(text)
}

// ResolveResponseReferences replaces every reference to an earlier response in
// a suite. It returns the references it could not resolve, which stay in the
// text so that the request shows what it was waiting for.
func ResolveResponseReferences(suite *HttpSuite, store ResponseStore) []string {
	unresolved := map[string]struct{}{}
	resolve := func(text string) string {
		return responseReferencePattern.ReplaceAllStringFunc(text, func(match string) string {
			parts := responseReferencePattern.FindStringSubmatch(match)
			value, err := lookupResponseValue(store, suite.SourceFilePath, parts[1], parts[2], parts[3])
			if err != nil {
				unresolved[strings.TrimSpace(match)+": "+err.Error()] = struct{}{}
				return match
			}
			return value
		})
	}

	suite.Uri = resolve(suite.Uri)
	suite.Body = resolve(suite.Body)
	suite.GraphQLVariables = resolve(suite.GraphQLVariables)
	resolved := make(nethttp.Header, len(suite.Header))
	for name, values := range suite.Header {
		for _, value := range values {
			resolved.Add(resolve(name), resolve(value))
		}
	}
	suite.Header = resolved

	result := make([]string, 0, len(unresolved))
	for reference := range unresolved {
		result = append(result, reference)
	}
	return result
}

func lookupResponseValue(store ResponseStore, sourceFilePath, name, kind, path string) (string, error) {
	trimmedName := strings.TrimSpace(name)
	response, recorded := store[responseKey{sourceFilePath: sourceFilePath, name: trimmedName}]
	if !recorded {
		return "", fmt.Errorf("%q has not been run yet", trimmedName)
	}

	if kind == "headers" {
		value := response.Header.Get(path)
		if value == "" {
			return "", fmt.Errorf("the response carries no %q header", path)
		}
		return value, nil
	}

	if path == "" || path == "$" || path == "*" {
		return response.Body, nil
	}
	return valueAtJSONPath(response.Body, path)
}

// valueAtJSONPath reads a value out of a JSON body with the subset of JSONPath
// these files use: member names and array indices under a `$` root.
func valueAtJSONPath(body, path string) (string, error) {
	var document any
	if err := json.Unmarshal([]byte(body), &document); err != nil {
		return "", fmt.Errorf("the response body is not JSON")
	}

	current := document
	for _, segment := range splitJSONPath(path) {
		next, err := stepInto(current, segment)
		if err != nil {
			return "", err
		}
		current = next
	}
	return renderJSONValue(current), nil
}

// splitJSONPath turns `$.data.items[0].id` into the steps to walk.
func splitJSONPath(path string) []string {
	path = strings.TrimPrefix(strings.TrimSpace(path), "$")
	path = strings.TrimPrefix(path, ".")
	segments := []string{}
	for _, part := range strings.Split(path, ".") {
		if part == "" {
			continue
		}
		indices := pathSegmentPattern.FindAllStringSubmatch(part, -1)
		name := pathSegmentPattern.ReplaceAllString(part, "")
		if name != "" {
			segments = append(segments, name)
		}
		for _, index := range indices {
			segments = append(segments, "["+index[1]+"]")
		}
	}
	return segments
}

func stepInto(value any, segment string) (any, error) {
	if strings.HasPrefix(segment, "[") {
		index, err := strconv.Atoi(strings.Trim(segment, "[]"))
		if err != nil {
			return nil, fmt.Errorf("invalid index %s", segment)
		}
		list, ok := value.([]any)
		if !ok {
			return nil, fmt.Errorf("%s is not an array", segment)
		}
		if index < 0 || index >= len(list) {
			return nil, fmt.Errorf("index %d is out of range", index)
		}
		return list[index], nil
	}

	object, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%q is not an object member", segment)
	}
	member, present := object[segment]
	if !present {
		return nil, fmt.Errorf("the response body has no %q", segment)
	}
	return member, nil
}

// renderJSONValue writes a value the way it would be used in a request: a
// string as itself, anything else as its JSON form.
func renderJSONValue(value any) string {
	if text, ok := value.(string); ok {
		return text
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return fmt.Sprintf("%v", value)
	}
	return string(encoded)
}
