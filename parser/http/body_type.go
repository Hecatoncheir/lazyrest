package http

import "strings"

const (
	// BodyTypeJSON and BodyTypeXML name the formats a body is recognized as,
	// next to BodyTypeGraphQL.
	BodyTypeJSON = "json"
	BodyTypeXML  = "xml"
)

// graphQLKeywords open a GraphQL document that does not start with a selection
// set.
var graphQLKeywords = []string{"query", "mutation", "subscription", "fragment"}

// detectBodyType names the format of a body. What the request declares wins,
// because it is a statement of intent; otherwise the body speaks for itself.
func detectBodyType(suite HttpSuite) string {
	if declared := bodyTypeFromContentType(suite.Header.Get("Content-Type")); declared != "" {
		return declared
	}
	return sniffBodyType(suite.Body)
}

func bodyTypeFromContentType(contentType string) string {
	contentType = strings.ToLower(contentType)
	switch {
	case contentType == "":
		return ""
	// A GraphQL response type carries "json" too, so it is matched first.
	case strings.Contains(contentType, BodyTypeGraphQL):
		return BodyTypeGraphQL
	case strings.Contains(contentType, BodyTypeJSON):
		return BodyTypeJSON
	case strings.Contains(contentType, BodyTypeXML):
		return BodyTypeXML
	}
	return ""
}

// sniffBodyType reads the format from the shape of the body, which is what an
// editor of this format falls back to when no header says otherwise.
func sniffBodyType(body string) string {
	body = strings.TrimSpace(body)
	if body == "" {
		return ""
	}

	switch body[0] {
	case '<':
		return BodyTypeXML
	case '[':
		return BodyTypeJSON
	case '{':
		// A JSON object names its members with strings, while a GraphQL
		// selection set opens with a field name.
		rest := strings.TrimLeft(body[1:], " \t\r\n")
		if rest == "" || strings.HasPrefix(rest, "\"") || strings.HasPrefix(rest, "}") {
			return BodyTypeJSON
		}
		return BodyTypeGraphQL
	}

	for _, keyword := range graphQLKeywords {
		if strings.HasPrefix(body, keyword) {
			return BodyTypeGraphQL
		}
	}
	return ""
}
