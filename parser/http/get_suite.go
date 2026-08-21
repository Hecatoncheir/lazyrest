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
			suite.Header[strings.TrimSpace(key)] = strings.TrimSpace(val)
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
