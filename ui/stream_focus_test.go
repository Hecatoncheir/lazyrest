package ui

import (
	"context"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Hecatoncheir/lazyrest/locale"
	parserhttp "github.com/Hecatoncheir/lazyrest/parser/http"
	"github.com/Hecatoncheir/lazyrest/runner"
	runnerstream "github.com/Hecatoncheir/lazyrest/runner/stream"
	"github.com/Hecatoncheir/lazyrest/ui/theme"

	"github.com/gdamore/tcell/v2"
)

// Navigating by name would send focus to Producer's view even while the stream
// pane occupies the slot, so focus would land on something not on screen.
func TestTUIFocusRightReachesTheStreamPane(t *testing.T) {
	application, _, screen := startStreamFor(t, streamSuite())

	application.Element.QueueUpdateDraw(func() {
		application.Element.SetFocus(application.Suite.Element)
	})
	waitFor(t, "focus on the request pane", func() bool {
		return application.Suite.Element.HasFocus()
	})

	screen.InjectKey(tcell.KeyCtrlL, 0, tcell.ModCtrl)
	waitFor(t, "focus reaching the stream pane", func() bool {
		return application.Stream.Element.HasFocus()
	})
}

func TestTUIFocusLeftLeavesTheStreamPane(t *testing.T) {
	application, _, screen := startStreamFor(t, streamSuite())

	waitFor(t, "focus on the stream pane", func() bool {
		return application.Stream.Element.HasFocus()
	})

	screen.InjectKey(tcell.KeyCtrlH, 0, tcell.ModCtrl)
	waitFor(t, "focus leaving the stream pane", func() bool {
		return application.Suite.Element.HasFocus()
	})
}

// Without this the pane keeps the colours and keys it was built with while
// every other pane follows the new theme.
func TestStreamPaneFollowsATheme(t *testing.T) {
	application := hintsApplication(t)
	before := application.Stream.Element.GetBackgroundColor()

	other := theme.NewDefault()
	other.Producer.Background = tcell.NewRGBColor(1, 2, 3)
	application.Stream.ApplySettings(other, application.config.Locale, application.config.Keybindings)

	after := application.Stream.Element.GetBackgroundColor()
	if after == before {
		t.Fatal("the stream pane ignored the new theme")
	}
	if after != tcell.NewRGBColor(1, 2, 3) {
		t.Fatalf("background = %v, want the theme colour", after)
	}
}

func TestStreamPaneFollowsALanguage(t *testing.T) {
	application := hintsApplication(t)

	russian, err := locale.New("ru", nil)
	if err != nil {
		t.Fatalf("load russian: %v", err)
	}
	application.Stream.ApplySettings(theme.NewDefault(), russian, application.config.Keybindings)

	if title := application.Stream.Element.GetTitle(); title == "" || !containsRussian(title) {
		t.Fatalf("title = %q, want the translated pane name", title)
	}
}

func containsRussian(text string) bool {
	for _, character := range text {
		if character >= 'А' && character <= 'я' {
			return true
		}
	}
	return false
}

// Picking a history entry shows a past response, so the response pane has to
// return to the slot; otherwise the entry is rendered off screen and the live
// connection keeps running invisibly.
func TestTUIHistoryRestoresTheResponsePaneOverAStream(t *testing.T) {
	root := t.TempDir()
	client := &http.Client{Transport: uiRoundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode:    http.StatusOK,
			Status:        "200 OK",
			Header:        http.Header{"Content-Type": []string{"application/json"}},
			Body:          io.NopCloser(strings.NewReader(`{"ok":true}`)),
			ContentLength: -1,
			Proto:         "HTTP/1.1",
		}, nil
	})}
	application := BuildApplication(root, Config{
		HistoryPath: filepath.Join(root, "state", "history.json"),
		Runner:      runner.Config{Client: client, Timeout: 10 * time.Second},
	})
	session := newFakeStreamSession()
	application.dialStream = func(context.Context, parserhttp.HttpSuite) (runnerstream.Session, error) {
		return session, nil
	}
	runTestApplication(t, application)

	application.Element.QueueUpdateDraw(func() {
		onSuiteRun(application)(parserhttp.HttpSuite{
			Name: "plain", Method: http.MethodGet, Uri: "https://example.test/x",
			Header: http.Header{}, SourceFilePath: filepath.Join(root, "requests.http"),
		})
	})
	waitFor(t, "a history entry", func() bool {
		return len(application.Producer.HistorySummaries()) == 1
	})

	application.Element.QueueUpdateDraw(func() { onSuiteRun(application)(streamSuite()) })
	waitFor(t, "the stream pane taking the response slot", func() bool {
		return application.Workspace.ResponseElement() == application.Stream.Element
	})

	application.Element.QueueUpdateDraw(func() { application.selectHistory(0) })
	waitFor(t, "the response pane coming back", func() bool {
		return application.Workspace.ResponseElement() == application.Producer.Element
	})
	waitFor(t, "the connection being closed", func() bool {
		select {
		case _, open := <-session.frames:
			return !open
		default:
			return false
		}
	})
}
