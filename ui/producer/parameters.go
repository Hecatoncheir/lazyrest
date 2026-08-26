package producer

import (
	"github.com/Hecatoncheir/lazyrest/keymap"
	"github.com/Hecatoncheir/lazyrest/locale"
	"github.com/Hecatoncheir/lazyrest/runner"
	"github.com/Hecatoncheir/lazyrest/ui/theme"
	"github.com/rivo/tview"
)

type Parameters struct {
	Theme                  theme.Theme
	OnEscapeCallback       OnEscapeCallbackType
	OnProgressCallback     func(current, total int64)
	OnRunFinishedCallback  func(runner.Response, error)
	OnCopyBodyCallback     func()
	OnCopyResponseCallback func()
	OnSaveResponseCallback func()
	App                    *tview.Application
	RunnerConfig           runner.Config
	Keybindings            *keymap.Bindings
	Locale                 *locale.Translator
	HistoryPath            string
}
