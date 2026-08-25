package tree

import (
	"github.com/Hecatoncheir/lazyrest/keymap"
	"github.com/Hecatoncheir/lazyrest/ui/theme"
)

type Parameters struct {
	RootDirectoryPath    string
	FilesExtension       []string
	Theme                theme.Theme
	OnSelectFileCallback OnSelectFileCallbackType
	OnReloadCallback     func()
	Keybindings          *keymap.Bindings
}
