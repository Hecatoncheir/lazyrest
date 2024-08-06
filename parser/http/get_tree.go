package http

import (
	"context"

	sitter "github.com/smacker/go-tree-sitter"
)

func getTree(source []byte, parser sitter.Parser) (sitter.Tree, error) {
	context := context.Background()
	tree, err := parser.ParseCtx(context, nil, source)
	if err != nil {
		return sitter.Tree{}, err
	}
	return *tree, nil
}
