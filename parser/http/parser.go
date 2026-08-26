package http

import (
	"context"
	"os"
	"path/filepath"
)

// Parser reads .http files. It holds no state, so one parser can be reused for
// any number of files.
type Parser struct{}

func NewParser() (*Parser, error) {
	return &Parser{}, nil
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
	if err := ctx.Err(); err != nil {
		return ParseResult{}, err
	}
	options.baseDirectory = filepath.Dir(filePath)

	suites, diagnostics := parseDocument(string(source), options)
	for index := range suites {
		suites[index].SourceFilePath = filepath.Clean(filePath)
	}
	return ParseResult{Suites: suites, Diagnostics: diagnostics}, nil
}

// Reset and Close keep the parser interchangeable with one that holds
// resources.
func (parser *Parser) Reset() {}

func (parser *Parser) Close() {}
