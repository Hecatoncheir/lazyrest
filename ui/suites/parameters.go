package suites

import (
	"github.com/Hecatoncheir/lazyrest/parser/http"
	"github.com/Hecatoncheir/lazyrest/ui/theme"
)

type Parameters struct {
	Theme                     theme.Theme
	OnEscapeCallback          OnEscapeCallbackType
	OnSuiteSelectCallbackType OnSuiteSelectCallbackType
	ParseOptions              http.ParseOptions
}
