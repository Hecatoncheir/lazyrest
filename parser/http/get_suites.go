package http

import (
	"fmt"
	"strings"
	sitter "github.com/smacker/go-tree-sitter"
)

func getSuites(source []byte, tree sitter.Tree) ([]HttpSuite, error) {
	suites := []HttpSuite{}

	rootNode := tree.RootNode()
	fmt.Printf("DEBUG: rootNode child count: %d\n", rootNode.ChildCount())

	for i := 0; i < int(rootNode.ChildCount()); i++ {
		node := rootNode.Child(i)
		nodeType := node.Type()
		fmt.Printf("DEBUG: rootNode child %d type: %s\n", i, nodeType)

		if nodeType == "request" {
			suite, err := getSuite(source, node)
			if err != nil {
				continue
			}
			suites = append(suites, suite)
		} else if nodeType == "ERROR" && len(suites) > 0 {
			lastSuite := &suites[len(suites)-1]
			if lastSuite.Body == "" {
				lastSuite.Body = strings.TrimSpace(node.Content(source))
			}
		} else if nodeType == "ERROR" {
			fmt.Printf("DEBUG: ERROR node found at index %d. Content: %q\n", i, node.Content(source))
		}
	}

	return suites, nil
}
