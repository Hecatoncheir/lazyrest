package ui

import (
	"github.com/Hecatoncheir/lazyrest/environment"
	"github.com/Hecatoncheir/lazyrest/keymap"
	"github.com/Hecatoncheir/lazyrest/locale"
	parserhttp "github.com/Hecatoncheir/lazyrest/parser/http"
	"github.com/Hecatoncheir/lazyrest/runner"
	"github.com/Hecatoncheir/lazyrest/ui/theme"
)

type Config struct {
	Runner          runner.Config
	ParseOptions    parserhttp.ParseOptions
	EnvironmentName string
	Environment     environment.Config
	Keybindings     *keymap.Bindings
	Locale          *locale.Translator
	Theme           theme.Theme
	ConfigPath      string
	ConfigPaths     []string
	Ignore          []string
	HistoryPath     string
	HistoryBodies   bool
}
