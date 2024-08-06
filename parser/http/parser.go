package http

import (
	"os"

	sitter "github.com/smacker/go-tree-sitter"
)

func NewParser() (Parser, error) {
	httpParser := Parser{}
	parser := getParser()
	httpParser.treesitter = parser
	return httpParser, nil
}

type Parser struct {
	treesitter sitter.Parser
}

func (parser Parser) GetSuitesFromFile(filePath string) ([]HttpSuite, error) {
	source, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	tree, err := getTree(source, parser.treesitter)
	if err != nil {
		return nil, err
	}

	suites, err := getSuites(source, tree)
	if err != nil {
		return nil, err
	}

	return suites, nil
}

func (parser Parser) Reset() {
	parser.treesitter.Reset()
}
