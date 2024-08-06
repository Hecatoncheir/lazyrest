package http

import (
	"lazyrest/parser/http/treesitter"

	sitter "github.com/smacker/go-tree-sitter"
)

func getParser() sitter.Parser {
	language := treesitter.Language()
	lang := sitter.NewLanguage(language)
	parser := sitter.NewParser()
	parser.SetLanguage(lang)
	return *parser
}
