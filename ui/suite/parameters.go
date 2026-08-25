package suite

import (
	"github.com/Hecatoncheir/lazyrest/keymap"
	"github.com/Hecatoncheir/lazyrest/locale"
	"github.com/Hecatoncheir/lazyrest/ui/theme"
)

type Parameters struct {
	Theme            theme.Theme
	OnEscapeCallback OnEscapeCallbackType
	OnRunCallback    OnRunCallbackType
	Keybindings      *keymap.Bindings
	Locale           *locale.Translator
}
