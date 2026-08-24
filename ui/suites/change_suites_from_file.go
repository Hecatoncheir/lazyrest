package suites

import (
	"context"
	"github.com/Hecatoncheir/lazyrest/finder"
	"github.com/Hecatoncheir/lazyrest/parser/http"
	"github.com/Hecatoncheir/lazyrest/parser/hurl"
	"path/filepath"
	"strings"

	"github.com/rivo/tview"
)

type OnSuiteSelectCallbackType func(suite http.HttpSuite)

func (widget *Suites) ChangeSuitesFromFile(file finder.File) {
	element := widget.Element.(*tview.List)
	element.Clear()
	widget.searchQuery = ""
	widget.searchMode = false

	var suites []http.HttpSuite
	var diagnostics []http.Diagnostic
	var err error

	if strings.EqualFold(filepath.Ext(file.Path), ".hurl") {
		parser, pErr := hurl.NewParser()
		if pErr != nil {
			err = pErr
		} else {
			suites, err = parser.GetSuitesFromFile(file.Path)
		}
	} else {
		parser, pErr := http.NewParser()
		if pErr != nil {
			err = pErr
		} else {
			defer parser.Close()
			result, parseErr := parser.ParseFile(context.Background(), file.Path)
			suites = result.Suites
			diagnostics = result.Diagnostics
			err = parseErr
		}
	}

	if err != nil {
		element.AddItem("Error: "+err.Error(), "", 0, nil)
		widget.suites = nil
		widget.diagnostics = nil
		return
	}

	widget.suites = suites
	widget.diagnostics = diagnostics
	widget.render()
}
