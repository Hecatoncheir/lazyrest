package suites

import "lazyrest/ui/theme"

type Parameters struct {
	Theme                     theme.Theme
	OnEscapeCallback          OnEscapeCallbackType
	OnSuiteSelectCallbackType OnSuiteSelectCallbackType
}
