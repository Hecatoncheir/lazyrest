package tree

import (
	"github.com/Hecatoncheir/lazyrest/keymap"
	"github.com/Hecatoncheir/lazyrest/locale"
	"github.com/Hecatoncheir/lazyrest/ui/theme"
)

type Parameters struct {
	RootDirectoryPath    string
	FilesExtension       []string
	Ignore               []string
	Theme                theme.Theme
	OnSelectFileCallback OnSelectFileCallbackType
	OnReloadCallback     func()
	Keybindings          *keymap.Bindings
	Locale               *locale.Translator
}
