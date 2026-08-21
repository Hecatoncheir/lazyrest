package hurl

import (
	"lazyrest/parser/http"
)

type Parser struct{}

func NewParser() (Parser, error) {
	return Parser{}, nil
}

func (p Parser) GetSuitesFromFile(filePath string) ([]http.HttpSuite, error) {
	hp, err := http.NewParser()
	if err != nil {
		return nil, err
	}
	suites, err := hp.GetSuitesFromFile(filePath)
	if err != nil {
		return nil, err
	}

	for i := range suites {
		suites[i].IsHurl = true
	}

	return suites, nil
}
