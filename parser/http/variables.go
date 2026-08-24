package http

import (
	"regexp"
	"slices"
	"strconv"
	"strings"
)

var variablePattern = regexp.MustCompile(`\{\{\s*([A-Za-z_][A-Za-z0-9_.-]*)\s*\}\}`)

func parseVariableDeclaration(content string) (string, string, bool) {
	content = strings.TrimSpace(content)
	content = strings.TrimPrefix(content, "@")
	name, value, found := strings.Cut(content, "=")
	if !found {
		return "", "", false
	}
	name = strings.TrimSpace(name)
	value = strings.TrimSpace(value)
	if name == "" {
		return "", "", false
	}
	if unquoted, err := strconv.Unquote(value); err == nil {
		value = unquoted
	}
	return name, value, true
}

func resolveSuiteVariables(suite *HttpSuite, variables map[string]string) []string {
	missingSet := make(map[string]struct{})
	resolve := func(value string) string {
		return variablePattern.ReplaceAllStringFunc(value, func(match string) string {
			parts := variablePattern.FindStringSubmatch(match)
			name := parts[1]
			if replacement, ok := variables[name]; ok {
				return replacement
			}
			missingSet[name] = struct{}{}
			return match
		})
	}

	suite.Uri = resolve(suite.Uri)
	suite.Body = resolve(suite.Body)
	resolvedHeaders := make(map[string]string, len(suite.Header))
	for key, value := range suite.Header {
		resolvedHeaders[resolve(key)] = resolve(value)
	}
	suite.Header = resolvedHeaders

	missing := make([]string, 0, len(missingSet))
	for name := range missingSet {
		missing = append(missing, name)
	}
	slices.Sort(missing)
	return missing
}
