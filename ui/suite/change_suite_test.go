package suite

import (
	"strings"
	"testing"

	"github.com/Hecatoncheir/lazyrest/parser/http"
	"github.com/Hecatoncheir/lazyrest/ui/theme"
	"github.com/rivo/tview"
)

func TestChangeSuiteRedactsEnvironmentSecrets(t *testing.T) {
	widget := New()
	widget.Build(Parameters{
		Theme:            theme.NewDefault(),
		OnEscapeCallback: func() {},
		OnRunCallback:    func(http.HttpSuite) {},
	})
	widget.ChangeSuite(http.HttpSuite{
		Method:       "POST",
		Uri:          "https://example.com/private-token",
		Body:         `{"token":"private-token"}`,
		SecretValues: []string{"private-token"},
	})

	text := widget.Element.(*tview.TextView).GetText(false)
	if strings.Contains(text, "private-token") || !strings.Contains(text, "<redacted>") {
		t.Fatalf("suite secret was not redacted: %q", text)
	}
}
