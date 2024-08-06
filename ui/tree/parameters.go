package tree

import "lazyrest/ui/theme"

type Parameters struct {
	RootDirectoryPath    string
	FilesExtension       string
	Theme                theme.Theme
	OnSelectFileCallback OnSelectFileCallbackType
}
