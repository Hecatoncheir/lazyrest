package ui

import (
	"github.com/Hecatoncheir/lazyrest/environment"
	"github.com/Hecatoncheir/lazyrest/keymap"
	parserhttp "github.com/Hecatoncheir/lazyrest/parser/http"
	"github.com/Hecatoncheir/lazyrest/runner"
)

type Config struct {
	Runner          runner.Config
	ParseOptions    parserhttp.ParseOptions
	EnvironmentName string
	Environment     environment.Config
	Keybindings     *keymap.Bindings
}
