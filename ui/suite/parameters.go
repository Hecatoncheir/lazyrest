package suite

import "lazyrest/ui/theme"

type Parameters struct {
	Theme            theme.Theme
	OnEscapeCallback OnEscapeCallbackType
	OnRunCallback    OnRunCallbackType
}
