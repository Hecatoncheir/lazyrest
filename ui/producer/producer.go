package producer

import (
	"context"
	"github.com/Hecatoncheir/lazyrest/parser/http"
	"github.com/Hecatoncheir/lazyrest/runner"
	"github.com/Hecatoncheir/lazyrest/ui/theme"
	"sync"

	"github.com/rivo/tview"
)

func New() *Producer {
	return &Producer{}
}

func (widget *Producer) IsSearching() bool {
	return widget.searchMode
}

type Producer struct {
	Element          tview.Primitive
	theme            theme.ProducerTheme
	suite            http.HttpSuite
	onEscapeCallback OnEscapeCallbackType
	app              *tview.Application
	runMutex         sync.Mutex
	runID            uint64
	cancelRun        context.CancelFunc
	history          []HistoryEntry
	historyIndex     int
	historyVisible   bool
	currentText      string
	searchMode       bool
	searchQuery      string
	runnerConfig     runner.Config
	bodyViewMode     BodyViewMode
}

func (widget *Producer) StartRun() (context.Context, uint64) {
	widget.runMutex.Lock()
	defer widget.runMutex.Unlock()
	if widget.cancelRun != nil {
		widget.cancelRun()
	}
	widget.runID++
	ctx, cancel := context.WithCancel(context.Background())
	widget.cancelRun = cancel
	return ctx, widget.runID
}

func (widget *Producer) IsCurrentRun(runID uint64) bool {
	widget.runMutex.Lock()
	defer widget.runMutex.Unlock()
	return widget.runID == runID
}

func (widget *Producer) FinishRun(runID uint64) bool {
	widget.runMutex.Lock()
	defer widget.runMutex.Unlock()
	if widget.runID != runID {
		return false
	}
	if widget.cancelRun != nil {
		widget.cancelRun()
		widget.cancelRun = nil
	}
	return true
}

func (widget *Producer) IsRunning() bool {
	widget.runMutex.Lock()
	defer widget.runMutex.Unlock()
	return widget.cancelRun != nil
}

func (widget *Producer) CancelActive() {
	widget.runMutex.Lock()
	defer widget.runMutex.Unlock()
	widget.runID++
	if widget.cancelRun != nil {
		widget.cancelRun()
		widget.cancelRun = nil
	}
}

func (widget *Producer) Build(parameters Parameters) tview.Primitive {
	widget.onEscapeCallback = parameters.OnEscapeCallback
	widget.app = parameters.App
	widget.runnerConfig = parameters.RunnerConfig
	widget.bodyViewMode = BodyViewPretty
	theme := parameters.Theme.Producer
	widget.theme = theme

	element := tview.NewTextView()
	element.
		SetDynamicColors(true).
		SetTextColor(theme.Foreground).
		SetBackgroundColor(theme.Background)

	element.SetFocusFunc(func() {
		element.
			SetTitleColor(theme.TitleFocus).
			SetBackgroundColor(theme.BackgroundFocus).
			SetBorderColor(theme.BorderFocus)
	})

	element.SetBlurFunc(func() {
		element.
			SetTitleColor(theme.Title).
			SetBackgroundColor(theme.Background).
			SetBorderColor(theme.Border)
	})

	box := tview.NewBox().
		SetBorder(true).
		SetBorderColor(theme.Border).
		SetBackgroundColor(theme.Background).
		SetTitle("Producer").
		SetInputCapture(onInputCallback(widget))

	box.SetFocusFunc(func() {
		box.
			SetTitleColor(theme.TitleFocus).
			SetBackgroundColor(theme.BackgroundFocus).
			SetBorderColor(theme.BorderFocus)
	})

	box.SetBlurFunc(func() {
		box.
			SetTitleColor(theme.Title).
			SetBackgroundColor(theme.Background).
			SetBorderColor(theme.Border)
	})

	element.Box = box

	widget.Element = element
	widget.updateTitle()
	return element
}
