package ui

import (
	"context"
	"fmt"
	"strings"

	parserhttp "github.com/Hecatoncheir/lazyrest/parser/http"
	runnerstream "github.com/Hecatoncheir/lazyrest/runner/stream"
	uistream "github.com/Hecatoncheir/lazyrest/ui/stream"

	"github.com/rivo/tview"
)

// dialStreamSuite opens what a stream suite describes. It is installed as
// Application.dialStream, which tests replace the way they replace scanFiles
// and loadEnvironment.
func dialStreamSuite(config Config) func(context.Context, parserhttp.HttpSuite) (runnerstream.Session, error) {
	return func(ctx context.Context, suite parserhttp.HttpSuite) (runnerstream.Session, error) {
		switch suite.Transport {
		case parserhttp.TransportWebSocket:
			return runnerstream.DialWebSocket(ctx, suite.Uri, runnerstream.WebSocketConfig{
				DialTimeout: config.Runner.Timeout,
				Header:      suite.Header,
				// The same client, and so the same cookie jar, as ordinary
				// requests. A socket opened after a login belongs to that
				// session.
				HTTPClient: config.Runner.Client,
			})
		case parserhttp.TransportTCP:
			return runnerstream.DialTCP(ctx, suite.Uri, runnerstream.TCPConfig{
				DialTimeout: config.Runner.Timeout,
			})
		case parserhttp.TransportMQTT:
			// The request headers describe the session: which topics to watch,
			// where a composed frame is published, who to connect as.
			mqtt := runnerstream.MQTTConfigFromHeader(suite.Header)
			mqtt.DialTimeout = config.Runner.Timeout
			mqtt.InsecureSkipVerify = config.Runner.InsecureSkipVerify
			return runnerstream.DialMQTT(ctx, suite.Uri, mqtt)
		default:
			return nil, fmt.Errorf("%q is not a stream", suite.Uri)
		}
	}
}

// startStream replaces whatever the response slot held with the live frame log
// and opens the connection. Any stream already running is closed first.
func (application *Application) startStream(suite parserhttp.HttpSuite) {
	application.stopStream()

	log := uistream.NewLog(0, 0)
	application.Stream.SetSecretValues(suite.SecretValues)
	application.Stream.SetError(nil)
	application.Stream.SetFollowing(true)
	application.Stream.SetLog(log)

	application.Workspace.SetResponseElement(application.Stream.Element)
	application.Element.SetFocus(application.Stream.Element)

	application.Model.update(func(state *State) {
		state.Request = TaskState{Phase: PhaseLoading}
	})
	application.refreshStatus()

	ctx, cancel := context.WithCancel(context.Background())
	application.streamMutex.Lock()
	application.streamCancel = cancel
	application.streamTransport = suite.Transport
	application.streamMutex.Unlock()

	go application.runStream(ctx, cancel, suite, log)
}

func (application *Application) runStream(
	ctx context.Context,
	cancel context.CancelFunc,
	suite parserhttp.HttpSuite,
	log *uistream.Log,
) {
	session, err := application.dialStream(ctx, suite)
	if err != nil {
		cancel()
		application.Element.QueueUpdateDraw(func() {
			application.Model.update(func(state *State) {
				state.Request = TaskState{Phase: PhaseFailed, Error: err.Error(), Outcome: OutcomeFailure}
			})
			application.refreshStatus()
			application.Stream.SetError(err)
		})
		return
	}

	application.streamMutex.Lock()
	application.streamSession = session
	application.streamMutex.Unlock()

	application.Element.QueueUpdateDraw(func() {
		application.Model.update(func(state *State) {
			state.Request = TaskState{Phase: PhaseReady, Outcome: OutcomeSuccess}
		})
		application.refreshStatus()
	})

	// The body of a stream request is what it says on connecting: a
	// subscription is normally the first thing sent.
	if strings.TrimSpace(suite.Body) != "" {
		opening := runnerstream.Frame{
			Opcode:  runnerstream.Text,
			Payload: []byte(expandEscapes(suite.Body, escapesFor(suite.Transport))),
		}
		if suite.Transport == parserhttp.TransportTCP {
			opening.Opcode = runnerstream.Bytes
		}
		if err := session.Send(ctx, opening); err != nil {
			application.Element.QueueUpdateDraw(func() {
				application.showStreamError(err.Error())
			})
		}
	}

	uistream.Follow(ctx, session, log, 0, func() {
		application.Element.QueueUpdateDraw(application.Stream.Render)
	})

	// Follow returns when the connection ends. Whatever the peer said last
	// still deserves to be on screen, and a failure has to be reported.
	failure := session.Err()
	application.Element.QueueUpdateDraw(func() {
		if failure != nil {
			application.Model.update(func(state *State) {
				state.Request = TaskState{Phase: PhaseFailed, Error: failure.Error(), Outcome: OutcomeFailure}
			})
			application.refreshStatus()
		}
		application.Stream.Render()
	})
}

// stopStream closes the running connection, if any. It is safe to call when
// none is running, and is called before starting another and when the
// application shuts down.
func (application *Application) stopStream() {
	application.streamMutex.Lock()
	cancel := application.streamCancel
	session := application.streamSession
	application.streamCancel = nil
	application.streamSession = nil
	application.streamTransport = parserhttp.TransportHTTP
	application.streamMutex.Unlock()

	if cancel != nil {
		cancel()
	}
	if session != nil {
		_ = session.Close()
	}
}

// responseElement is whatever fills the response slot right now: Producer's
// view, or the stream pane while a connection is open.
func (application *Application) responseElement() tview.Primitive {
	if application.Workspace == nil {
		return application.Producer.Element
	}
	return application.Workspace.ResponseElement()
}

// showResponsePane puts the ordinary response back in the slot.
func (application *Application) showResponsePane() {
	if application.Workspace == nil || application.Producer == nil {
		return
	}
	application.Workspace.SetResponseElement(application.Producer.Element)
}

func onStreamEscape(application *Application) func() {
	return func() {
		application.stopStream()
		application.showResponsePane()
		if application.Suite != nil {
			application.Element.SetFocus(application.Suite.Element)
		}
		application.Model.update(func(state *State) {
			state.Request = TaskState{}
		})
		application.refreshStatus()
	}
}
