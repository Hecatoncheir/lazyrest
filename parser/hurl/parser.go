package hurl

import (
	"fmt"
	"github.com/Hecatoncheir/lazyrest/parser/http"
	nethttp "net/http"
	"os"
	"path/filepath"
)

type Parser struct{}

func NewParser() (*Parser, error) {
	return &Parser{}, nil
}

func (p *Parser) GetSuitesFromFile(filePath string) ([]http.HttpSuite, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("hurl path is not a regular file: %s", filePath)
	}

	return []http.HttpSuite{{
		Name:         filepath.Base(filePath),
		Method:       "HURL",
		Uri:          filePath,
		Header:       nethttp.Header{},
		IsHurl:       true,
		HurlFilePath: filePath,
	}}, nil
}
