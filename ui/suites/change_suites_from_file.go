package suites

import (
	"lazyrest/finder"
	"lazyrest/parser/hurl"
	"lazyrest/parser/http"
	"strings"

	"github.com/rivo/tview"
)

type OnSuiteSelectCallbackType func(suite http.HttpSuite)

func (widget *Suites) ChangeSuitesFromFile(file finder.File) {
	element := widget.Element.(*tview.List)
	element.Clear()

	var suites []http.HttpSuite
	var err error

	if strings.HasSuffix(file.Path, ".hurl") {
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
			suites, err = parser.GetSuitesFromFile(file.Path)
		}
	}

	if err != nil {
		element.AddItem(err.Error(), "", '0', nil)
	}

	widget.suites = suites

	for _, suite := range suites {
		theme := widget.theme
		element.AddItem(suite.Method+" "+suite.Uri, suite.Body, 0, func() {
			widget.onSuiteSelectCallback(suite)
		}).
			SetWrapAround(true).
			SetHighlightFullLine(true).
			SetMainTextColor(theme.SuiteForeground).
			SetSecondaryTextColor(theme.SuiteForeground).
			SetSelectedTextColor(theme.SuiteFocusForeground).
			SetSelectedBackgroundColor(theme.SuiteFocusBackground).
			SetBackgroundColor(theme.SuiteBackground)
	}
}
