package http

import (
	"strings"
	sitter "github.com/smacker/go-tree-sitter"
)

func getSuite(source []byte, node *sitter.Node) (HttpSuite, error) {
	suite := NewHttpSuite()

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		nodeType := child.Type()
		value := child.Content(source)

		switch nodeType {
		case "method":
			suite.Method = value
		case "target_url":
			suite.Uri = value
		case "header":
			key, val, isFound := strings.Cut(value, ":")
			if !isFound {
				continue
			}
			trimmedVal := strings.TrimSpace(val)
			if strings.Contains(trimmedVal, "\n") {
				parts := strings.SplitN(trimmedVal, "\n", 2)
				suite.Header[strings.TrimSpace(key)] = strings.TrimSpace(parts[0])
				if len(parts) > 1 {
					suite.Body = strings.TrimSpace(parts[1])
				}
			} else {
				suite.Header[strings.TrimSpace(key)] = trimmedVal
			}
		case "body":
			suite.Body = value
		case "text_body":
			suite.Body = value
		case "raw_body":
			suite.Body = value
		case "json":
			suite.BodyType = "json"
			suite.Body = value
		case "xml":
			suite.BodyType = "xml"
			suite.Body = value
		case "graphql":
			suite.BodyType = "graphql"
			suite.Body = value
		case "json_body":
			suite.BodyType = "json"
			suite.Body = value
		case "xml_body":
			suite.BodyType = "xml"
			suite.Body = value
		case "graphql_body":
			suite.BodyType = "graphql"
			suite.Body = value
		default:
			if suite.Body == "" && nodeType != "method" && nodeType != "target_url" && nodeType != "header" {
				suite.Body = value
			}
		}
	}

	return suite, nil
}
