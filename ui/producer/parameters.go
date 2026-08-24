package producer

import (
	"github.com/Hecatoncheir/lazyrest/runner"
	"github.com/Hecatoncheir/lazyrest/ui/theme"
	"github.com/rivo/tview"
)

type Parameters struct {
	Theme                 theme.Theme
	OnEscapeCallback      OnEscapeCallbackType
	OnRunFinishedCallback func(error)
	App                   *tview.Application
	RunnerConfig          runner.Config
}
