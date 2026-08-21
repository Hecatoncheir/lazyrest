package producer

import "lazyrest/ui/theme"
import "github.com/rivo/tview"

type Parameters struct {
	Theme            theme.Theme
	OnEscapeCallback OnEscapeCallbackType
	App              *tview.Application
}
