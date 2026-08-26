package hurl

import (
	"fmt"
	nethttp "net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Hecatoncheir/lazyrest/parser/http"
)

type Parser struct{}

func NewParser() (*Parser, error) {
	return &Parser{}, nil
}

func (p *Parser) GetSuitesFromFile(filePath string) ([]http.HttpSuite, error) {
	return p.GetSuitesFromFileWithOptions(filePath, http.ParseOptions{})
}

func (p *Parser) GetSuitesFromFileWithOptions(filePath string, options http.ParseOptions) ([]http.HttpSuite, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("hurl path is not a regular file: %s", filePath)
	}
	contents, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	variables := http.ResolveVariables(options.Variables)
	secrets := http.ResolveSecretValues(options)

	entries := splitEntries(string(contents))
	if len(entries) == 0 {
		// Nothing was recognized, so the file is offered as a whole.
		return []http.HttpSuite{{
			Name:         filepath.Base(filePath),
			Method:       "HURL",
			Uri:          filePath,
			Header:       nethttp.Header{},
			IsHurl:       true,
			HurlFilePath: filePath,
			Variables:    variables,
			SecretValues: secrets,
		}}, nil
	}

	suites := make([]http.HttpSuite, 0, len(entries))
	for _, current := range entries {
		name := current.Name
		if name == "" {
			name = strings.TrimSpace(current.Method + " " + current.Uri)
		}
		suites = append(suites, http.HttpSuite{
			Name:         name,
			Method:       current.Method,
			Uri:          current.Uri,
			Header:       nethttp.Header{},
			Body:         current.Text,
			IsHurl:       true,
			HurlFilePath: filePath,
			HurlEntry:    current.Number,
			Variables:    variables,
			SecretValues: secrets,
		})
	}
	return suites, nil
}
