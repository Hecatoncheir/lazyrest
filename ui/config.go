package ui

import (
	"github.com/Hecatoncheir/lazyrest/parser/http"
	"github.com/Hecatoncheir/lazyrest/runner"
)

type Config struct {
	Runner          runner.Config
	ParseOptions    http.ParseOptions
	EnvironmentName string
}
