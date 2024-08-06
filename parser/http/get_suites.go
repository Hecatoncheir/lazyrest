package http

import sitter "github.com/smacker/go-tree-sitter"

func getSuites(source []byte, tree sitter.Tree) ([]HttpSuite, error) {
	suites := []HttpSuite{}

	rootNode := tree.RootNode()
	for i := 0; i < int(rootNode.ChildCount()); i++ {
		node := rootNode.Child(i)
		nodeType := node.Type()
		if nodeType == "request" {
			suite, err := getSuite(source, node)
			if err != nil {
				continue
			}
			suites = append(suites, suite)
		}
	}

	return suites, nil
}
