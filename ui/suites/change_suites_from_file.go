package suites

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/Hecatoncheir/lazyrest/finder"
	"github.com/Hecatoncheir/lazyrest/parser/http"
	"github.com/Hecatoncheir/lazyrest/parser/hurl"
	"github.com/rivo/tview"
)

type OnSuiteSelectCallbackType func(suite http.HttpSuite)

type LoadResult struct {
	Suites      []http.HttpSuite
	Diagnostics []http.Diagnostic
	Err         error
}

func (widget *Suites) StartLoad() (context.Context, uint64) {
	widget.loadMutex.Lock()
	defer widget.loadMutex.Unlock()
	if widget.cancelLoad != nil {
		widget.cancelLoad()
	}
	widget.loadID++
	ctx, cancel := context.WithCancel(context.Background())
	widget.cancelLoad = cancel
	return ctx, widget.loadID
}

func (widget *Suites) IsCurrentLoad(loadID uint64) bool {
	widget.loadMutex.Lock()
	defer widget.loadMutex.Unlock()
	return widget.loadID == loadID
}

func (widget *Suites) FinishLoad(loadID uint64) bool {
	widget.loadMutex.Lock()
	defer widget.loadMutex.Unlock()
	if widget.loadID != loadID {
		return false
	}
	if widget.cancelLoad != nil {
		widget.cancelLoad()
		widget.cancelLoad = nil
	}
	return true
}

func (widget *Suites) CancelLoad() {
	widget.loadMutex.Lock()
	defer widget.loadMutex.Unlock()
	widget.loadID++
	if widget.cancelLoad != nil {
		widget.cancelLoad()
		widget.cancelLoad = nil
	}
}

func (widget *Suites) ShowLoading(file finder.File) {
	element := widget.Element.(*tview.List)
	element.Clear()
	element.SetTitle("Suites — loading " + file.Name)
	element.AddItem("Loading...", file.Path, 0, nil)
	widget.searchQuery = ""
	widget.searchMode = false
}

func (widget *Suites) LoadSuitesFromFile(ctx context.Context, file finder.File) LoadResult {
	if err := ctx.Err(); err != nil {
		return LoadResult{Err: err}
	}

	if strings.EqualFold(filepath.Ext(file.Path), ".hurl") {
		parser, pErr := hurl.NewParser()
		if pErr != nil {
			return LoadResult{Err: pErr}
		}
		suites, err := parser.GetSuitesFromFile(file.Path)
		if err == nil {
			err = ctx.Err()
		}
		return LoadResult{Suites: suites, Err: err}
	}

	parser, err := http.NewParser()
	if err != nil {
		return LoadResult{Err: err}
	}
	defer parser.Close()
	result, err := parser.ParseFileWithOptions(ctx, file.Path, widget.parseOptions)
	return LoadResult{Suites: result.Suites, Diagnostics: result.Diagnostics, Err: err}
}

func (widget *Suites) ApplyLoadResult(result LoadResult) {
	element := widget.Element.(*tview.List)
	element.Clear()
	widget.searchQuery = ""
	widget.searchMode = false
	if result.Err != nil {
		element.SetTitle("Suites")
		element.AddItem("Error: "+result.Err.Error(), "", 0, nil)
		widget.suites = nil
		widget.diagnosticCount = 0
		return
	}

	widget.suites = result.Suites
	widget.diagnosticCount = len(result.Diagnostics)
	widget.render()
}

func (widget *Suites) Clear() {
	widget.CancelLoad()
	widget.suites = nil
	widget.diagnosticCount = 0
	widget.searchQuery = ""
	widget.searchMode = false
	if widget.Element != nil {
		widget.render()
	}
}
