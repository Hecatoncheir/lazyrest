package http

import (
	sitter "github.com/smacker/go-tree-sitter"
)

// bodyTypeByNode maps the grammar's body nodes to the content type they imply.
var bodyTypeByNode = map[string]string{
	"json":         "json",
	"json_body":    "json",
	"xml":          "xml",
	"xml_body":     "xml",
	"graphql":      "graphql",
	"graphql_body": "graphql",
}

// getSuite builds a suite from the raw text of a request region. The grammar is
// consulted only for the body type and for a request line the text scan could
// not read.
func getSuite(source []byte, node *sitter.Node, regionText string) HttpSuite {
	parsed := parseRequestText(regionText)

	suite := NewHttpSuite()
	suite.Method = parsed.Method
	suite.Uri = parsed.Uri
	suite.Header = parsed.Header
	suite.Body = parsed.Body

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch nodeType := child.Type(); nodeType {
		case "method":
			if suite.Method == "" {
				suite.Method = child.Content(source)
			}
		case "target_url":
			if suite.Uri == "" {
				suite.Uri = child.Content(source)
			}
		default:
			if bodyType, found := bodyTypeByNode[nodeType]; found && suite.BodyType == "" {
				suite.BodyType = bodyType
			}
		}
	}

	return suite
}
