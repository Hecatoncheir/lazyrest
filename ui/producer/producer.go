package producer

import (
	"context"
	"github.com/Hecatoncheir/lazyrest/keymap"
	"github.com/Hecatoncheir/lazyrest/locale"
	"github.com/Hecatoncheir/lazyrest/parser/http"
	"github.com/Hecatoncheir/lazyrest/runner"
	"github.com/Hecatoncheir/lazyrest/ui/syntax"
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
	onProgress       func(current, total int64)
	onRunFinished    func(runner.Response, error)
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
	keybindings      *keymap.Bindings
	locale           *locale.Translator
	historyPath      string
	syntax           syntax.Palette
	historyMutex     sync.Mutex
	historyRequested uint64
	historyWritten   uint64
	historyWrites    sync.WaitGroup
}

// WaitForHistory blocks until every pending history write has finished. It is
// called before the application exits so that the last run is not lost.
func (widget *Producer) WaitForHistory() {
	if widget == nil {
		return
	}
	widget.historyWrites.Wait()
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
	if parameters.Keybindings == nil {
		parameters.Keybindings = keymap.Default()
	}
	if parameters.Locale == nil {
		parameters.Locale = locale.English()
	}
	widget.onEscapeCallback = parameters.OnEscapeCallback
	widget.onProgress = parameters.OnProgressCallback
	widget.onRunFinished = parameters.OnRunFinishedCallback
	widget.app = parameters.App
	widget.runnerConfig = parameters.RunnerConfig
	widget.keybindings = parameters.Keybindings
	widget.locale = parameters.Locale
	widget.historyPath = parameters.HistoryPath
	widget.syntax = parameters.Theme.Syntax
	_ = widget.loadHistory()
	widget.bodyViewMode = BodyViewPretty
	theme := parameters.Theme.Producer
	widget.theme = theme

	element := tview.NewTextView()
	element.
		SetDynamicColors(true).
		SetTextColor(theme.Foreground).
		SetBackgroundColor(theme.Background)

	box := tview.NewBox().
		SetBorder(true).
		SetBorderColor(theme.Border).
		SetBackgroundColor(theme.Background).
		SetTitle(widget.locale.Text("producer")).
		SetInputCapture(onInputCallback(widget))

	element.Box = box
	element.SetFocusFunc(func() { widget.applyTheme(element, true) })
	element.SetBlurFunc(func() { widget.applyTheme(element, false) })

	widget.Element = element
	widget.updateTitle()
	return element
}
