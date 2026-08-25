package suites

import (
	"github.com/Hecatoncheir/lazyrest/keymap"
	"github.com/Hecatoncheir/lazyrest/locale"
	"github.com/Hecatoncheir/lazyrest/parser/http"
	"github.com/Hecatoncheir/lazyrest/ui/theme"
)

type Parameters struct {
	Theme                     theme.Theme
	OnEscapeCallback          OnEscapeCallbackType
	OnSuiteSelectCallbackType OnSuiteSelectCallbackType
	ParseOptions              http.ParseOptions
	Keybindings               *keymap.Bindings
	Locale                    *locale.Translator
}
