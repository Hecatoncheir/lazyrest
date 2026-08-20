package http

import (
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

func getSuite(source []byte, node *sitter.Node) (HttpSuite, error) {
	suite := NewHttpSuite()

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		value := child.Content(source)

		switch child.Type() {
		case "method":
			suite.Method = value
		case "target_url":
			suite.Uri = value
		case "header":
			key, value, isFound := strings.Cut(value, ":")
			if !isFound {
				continue
			}
			suite.Header[key] = value
		case "xml_body":
			suite.BodyType = "xml"
			suite.Body = value
		case "json_body":
			suite.BodyType = "json"
			suite.Body = value
		case "graphql_body":
			suite.BodyType = "graphql"
			suite.Body = value
		}
	}

	return suite, nil
}
