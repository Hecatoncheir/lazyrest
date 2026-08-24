package suites

import "github.com/Hecatoncheir/lazyrest/ui/theme"

type Parameters struct {
	Theme                     theme.Theme
	OnEscapeCallback          OnEscapeCallbackType
	OnSuiteSelectCallbackType OnSuiteSelectCallbackType
}
