package http

import (
	"context"
	"os"
	"path/filepath"

	sitter "github.com/smacker/go-tree-sitter"
)

func NewParser() (*Parser, error) {
	return &Parser{treesitter: getParser()}, nil
}

type Parser struct {
	treesitter *sitter.Parser
}

func (parser *Parser) GetSuitesFromFile(filePath string) ([]HttpSuite, error) {
	result, err := parser.ParseFile(context.Background(), filePath)
	return result.Suites, err
}

func (parser *Parser) ParseFile(ctx context.Context, filePath string) (ParseResult, error) {
	return parser.ParseFileWithOptions(ctx, filePath, ParseOptions{})
}

func (parser *Parser) ParseFileWithOptions(ctx context.Context, filePath string, options ParseOptions) (ParseResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return ParseResult{}, err
	}
	source, err := os.ReadFile(filePath)
	if err != nil {
		return ParseResult{}, err
	}
	options.baseDirectory = filepath.Dir(filePath)
	tree, err := getTree(ctx, source, parser.treesitter)
	if err != nil {
		return ParseResult{}, err
	}
	defer tree.Close()

	suites, diagnostics := getSuites(ctx, source, tree, options)
	return ParseResult{Suites: suites, Diagnostics: diagnostics}, nil
}

func (parser *Parser) Reset() {
	parser.treesitter.Reset()
}

func (parser *Parser) Close() {
	if parser == nil || parser.treesitter == nil {
		return
	}
	parser.treesitter.Close()
	parser.treesitter = nil
}
