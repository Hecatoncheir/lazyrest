package http

import (
	"log"
	"strings"
	sitter "github.com/smacker/go-tree-sitter"
)

func getSuites(source []byte, tree sitter.Tree) ([]HttpSuite, error) {
	suites := []HttpSuite{}

	rootNode := tree.RootNode()

	for i := 0; i < int(rootNode.ChildCount()); i++ {
		node := rootNode.Child(i)
		nodeType := node.Type()

		if nodeType == "request" {
			suite, err := getSuite(source, node)
			if err != nil {
				log.Printf("Error parsing suite at index %d: %v", i, err)
				continue
			}
			suites = append(suites, suite)
		} else if nodeType == "ERROR" && len(suites) > 0 {
			lastSuite := &suites[len(suites)-1]
			if lastSuite.Body == "" {
				lastSuite.Body = strings.TrimSpace(node.Content(source))
			}
		} else if nodeType == "ERROR" {
			log.Printf("Error node found at index %d: %q", i, node.Content(source))
		}
	}

	return suites, nil
}
