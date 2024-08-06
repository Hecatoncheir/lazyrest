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
			suite.method = value
		case "target_url":
			suite.uri = value
		case "header":
			key, value, isFound := strings.Cut(value, ":")
			if !isFound {
				continue
			}
			suite.header[key] = value
		case "xml_body":
			suite.bodyType = "xml"
			suite.body = value
		case "json_body":
			suite.bodyType = "json"
			suite.body = value
		case "graphql_body":
			suite.bodyType = "graphql"
			suite.body = value
		}
	}

	return suite, nil
}
