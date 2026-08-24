package http

import (
	"context"
	"os"

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
	source, err := os.ReadFile(filePath)
	if err != nil {
		return ParseResult{}, err
	}
	tree, err := getTree(ctx, source, parser.treesitter)
	if err != nil {
		return ParseResult{}, err
	}
	defer tree.Close()

	suites, diagnostics := getSuites(source, tree)
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
