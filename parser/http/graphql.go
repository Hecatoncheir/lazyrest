package http

import (
	"encoding/json"
	"regexp"
	"strings"
)

// BodyTypeGraphQL marks a request whose body is a GraphQL document.
const BodyTypeGraphQL = "graphql"

// graphQLRequestHeader marks a request as GraphQL in the editors that share the
// .http format, which is the only portable way to say so.
const graphQLRequestHeader = "X-REQUEST-TYPE"

// operationPattern finds the named operations of a GraphQL document.
var operationPattern = regexp.MustCompile(`(?m)^[ \t]*(?:query|mutation|subscription)[ \t\n]+([A-Za-z_][A-Za-z0-9_]*)`)

// applyGraphQL splits a GraphQL request into the parts it is sent as. The query
// stays in the body so that substitution, redaction, and search keep working on
// it unchanged.
func applyGraphQL(suite *HttpSuite) {
	if !declaresGraphQL(*suite) {
		return
	}
	suite.BodyType = BodyTypeGraphQL
	suite.Body, suite.GraphQLVariables = splitGraphQLBody(suite.Body)
	suite.GraphQLOperation = graphQLOperationName(suite.Body)
}

func declaresGraphQL(suite HttpSuite) bool {
	if suite.BodyType == BodyTypeGraphQL {
		return true
	}
	for key, values := range suite.Header {
		if !strings.EqualFold(key, graphQLRequestHeader) {
			continue
		}
		for _, value := range values {
			if strings.EqualFold(strings.TrimSpace(value), BodyTypeGraphQL) {
				return true
			}
		}
	}
	return false
}

// splitGraphQLBody separates the query from the JSON object of variables that
// may follow it, the layout every editor of the .http format uses.
func splitGraphQLBody(body string) (string, string) {
	lines := strings.SplitAfter(body, "\n")
	for index := len(lines) - 1; index > 0; index-- {
		if !strings.HasPrefix(lines[index], "{") {
			continue
		}
		variables := strings.TrimSpace(strings.Join(lines[index:], ""))
		if !isJSONObject(variables) {
			continue
		}
		query := strings.TrimSpace(strings.Join(lines[:index], ""))
		if query == "" {
			continue
		}
		return query, variables
	}
	return strings.TrimSpace(body), ""
}

// graphQLOperationName returns the operation to run, but only when the document
// names exactly one. Picking one of several would be a guess, and a server
// rejects a multi-operation document without a name anyway.
func graphQLOperationName(query string) string {
	matches := operationPattern.FindAllStringSubmatch(query, -1)
	if len(matches) != 1 {
		return ""
	}
	return matches[0][1]
}

func isJSONObject(text string) bool {
	if !strings.HasPrefix(text, "{") {
		return false
	}
	var object map[string]any
	return json.Unmarshal([]byte(text), &object) == nil
}
