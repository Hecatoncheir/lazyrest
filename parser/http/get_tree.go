package http

import (
	"context"

	sitter "github.com/smacker/go-tree-sitter"
)

func getTree(ctx context.Context, source []byte, parser *sitter.Parser) (*sitter.Tree, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	tree, err := parser.ParseCtx(ctx, nil, source)
	if err != nil {
		return nil, err
	}
	return tree, nil
}
