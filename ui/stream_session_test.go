package ui

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Hecatoncheir/lazyrest/environment"
	"github.com/Hecatoncheir/lazyrest/finder"
	parserhttp "github.com/Hecatoncheir/lazyrest/parser/http"
	runnerstream "github.com/Hecatoncheir/lazyrest/runner/stream"
	"github.com/Hecatoncheir/lazyrest/ui/tree"
	"github.com/gdamore/tcell/v2"
)

// fakeStreamSession lets the TUI tests drive a connection without a server,
// the way scanFiles and loadEnvironment stand in for the filesystem.
type fakeStreamSession struct {
	frames    chan runnerstream.Frame
	sent      chan runnerstream.Frame
	closeOnce sync.Once
}

func newFakeStreamSession() *fakeStreamSession {
	return &fakeStreamSession{
		frames: make(chan runnerstream.Frame, 8),
		sent:   make(chan runnerstream.Frame, 8),
	}
}

func (session *fakeStreamSession) Frames() <-chan runnerstream.Frame { return session.frames }

func (session *fakeStreamSession) Send(_ context.Context, frame runnerstream.Frame) error {
	session.sent <- frame
	return nil
}

func (session *fakeStreamSession) Close() error {
	session.closeOnce.Do(func() { close(session.frames) })
	return nil
}

func (session *fakeStreamSession) Err() error { return nil }

func buildStreamApplication(t *testing.T) *Application {
	t.Helper()
	root := t.TempDir()
	application := BuildApplication(root, Config{Environment: environment.Config{Name: "test"}})
	application.scanFiles = func(context.Context) tree.ScanResult {
		return tree.ScanResult{Directory: finder.Directory{Name: filepath.Base(root), Path: root}}
	}
	application.loadEnvironment = func(string, environment.Config) (environment.Environment, error) {
		return environment.Environment{Name: "test", Values: map[string]string{}}, nil
	}
	return application
}

func streamSuite() parserhttp.HttpSuite {
	return parserhttp.HttpSuite{
		Name:      "live",
		Method:    "WEBSOCKET",
		Uri:       "wss://example.com/socket",
		Transport: parserhttp.TransportWebSocket,
	}
}

func TestTUIShowsFramesFromAStream(t *testing.T) {
	application := buildStreamApplication(t)
	session := newFakeStreamSession()
	application.dialStream = func(context.Context, parserhttp.HttpSuite) (runnerstream.Session, error) {
		return session, nil
	}

	screen, _ := runTestApplication(t, application)
	application.Element.QueueUpdateDraw(func() { onSuiteRun(application)(streamSuite()) })

	waitFor(t, "the stream pane taking the response slot", func() bool {
		return application.Workspace.ResponseElement() == application.Stream.Element
	})

	session.frames <- runnerstream.Frame{
		At:        time.Now(),
		Direction: runnerstream.Received,
		Opcode:    runnerstream.Text,
		Payload:   []byte("hello-from-the-stream"),
	}
	waitForScreenText(t, application, screen, "hello-from-the-stream")
}

func TestTUIReportsAStreamThatCannotConnect(t *testing.T) {
	application := buildStreamApplication(t)
	application.dialStream = func(context.Context, parserhttp.HttpSuite) (runnerstream.Session, error) {
		return nil, errors.New("connection refused by the peer")
	}

	screen, _ := runTestApplication(t, application)
	application.Element.QueueUpdateDraw(func() { onSuiteRun(application)(streamSuite()) })

	waitForScreenText(t, application, screen, "connection refused by the peer")
	waitFor(t, "the failed phase", func() bool {
		return application.Model.Snapshot().Request.Phase == PhaseFailed
	})
}

func TestTUIStreamEscapeRestoresTheResponsePane(t *testing.T) {
	application := buildStreamApplication(t)
	session := newFakeStreamSession()
	application.dialStream = func(context.Context, parserhttp.HttpSuite) (runnerstream.Session, error) {
		return session, nil
	}

	runTestApplication(t, application)
	application.Element.QueueUpdateDraw(func() { onSuiteRun(application)(streamSuite()) })
	waitFor(t, "the stream pane taking the response slot", func() bool {
		return application.Workspace.ResponseElement() == application.Stream.Element
	})

	application.Element.QueueUpdateDraw(func() { onStreamEscape(application)() })
	waitFor(t, "the response pane coming back", func() bool {
		return application.Workspace.ResponseElement() == application.Producer.Element
	})

	// Leaving the pane has to hang up, not merely hide the log.
	waitFor(t, "the session being closed", func() bool {
		select {
		case _, open := <-session.frames:
			return !open
		default:
			return false
		}
	})
}

// An ordinary request after a stream must get the response pane back.
func TestTUIOrdinaryRequestRestoresTheResponsePaneAfterAStream(t *testing.T) {
	application := buildStreamApplication(t)
	application.dialStream = func(context.Context, parserhttp.HttpSuite) (runnerstream.Session, error) {
		return newFakeStreamSession(), nil
	}

	runTestApplication(t, application)
	application.Element.QueueUpdateDraw(func() { onSuiteRun(application)(streamSuite()) })
	waitFor(t, "the stream pane taking the response slot", func() bool {
		return application.Workspace.ResponseElement() == application.Stream.Element
	})

	plain := parserhttp.HttpSuite{Name: "plain", Method: "GET", Uri: "http://127.0.0.1:1/"}
	application.Element.QueueUpdateDraw(func() { onSuiteRun(application)(plain) })
	waitFor(t, "the response pane coming back", func() bool {
		return application.Workspace.ResponseElement() == application.Producer.Element
	})
}

func TestTUIStreamPaneKeysFollowTheKeymap(t *testing.T) {
	application := buildStreamApplication(t)
	session := newFakeStreamSession()
	application.dialStream = func(context.Context, parserhttp.HttpSuite) (runnerstream.Session, error) {
		return session, nil
	}

	screen, _ := runTestApplication(t, application)
	application.Element.QueueUpdateDraw(func() { onSuiteRun(application)(streamSuite()) })
	waitFor(t, "the stream pane taking the response slot", func() bool {
		return application.Workspace.ResponseElement() == application.Stream.Element
	})

	session.frames <- runnerstream.Frame{
		At:        time.Now(),
		Direction: runnerstream.Received,
		Opcode:    runnerstream.Text,
		Payload:   []byte("before-the-pause"),
	}
	waitForScreenText(t, application, screen, "before-the-pause")

	screen.InjectKey(tcell.KeyRune, 'f', tcell.ModNone)
	waitFor(t, "the pane pausing", func() bool { return !application.Stream.Following() })

	screen.InjectKey(tcell.KeyRune, 'f', tcell.ModNone)
	waitFor(t, "the pane following again", func() bool { return application.Stream.Following() })

	screen.InjectKey(tcell.KeyRune, 'c', tcell.ModNone)
	waitFor(t, "the log being cleared", func() bool {
		return !strings.Contains(applicationText(application, screen), "before-the-pause")
	})
}
