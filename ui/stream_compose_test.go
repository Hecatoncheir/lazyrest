package ui

import (
	"context"
	"strings"
	"testing"
	"time"

	parserhttp "github.com/Hecatoncheir/lazyrest/parser/http"
	runnerstream "github.com/Hecatoncheir/lazyrest/runner/stream"

	"github.com/gdamore/tcell/v2"
)

func TestExpandEscapesForARawSocket(t *testing.T) {
	cases := []struct{ in, want string }{
		{"PING", "PING"},
		{`PING\r\n`, "PING\r\n"},
		{`a\tb`, "a\tb"},
		{`back\\slash`, `back\slash`},
		// A STOMP frame ends on a null byte.
		{`CONNECT\n\n\0`, "CONNECT\n\n\x00"},
		// An unknown escape is left exactly as typed rather than guessed at.
		{`\q`, `\q`},
		// A trailing backslash has nothing to escape.
		{`trailing\`, `trailing\`},
		{"", ""},
	}
	for _, testCase := range cases {
		if got := expandEscapes(testCase.in, rawSocketEscapes); got != testCase.want {
			t.Errorf("expandEscapes(%q) = %q, want %q", testCase.in, got, testCase.want)
		}
	}
}

// A WebSocket message is sent as written so JSON keeps its own escapes. The
// null byte is the exception: JSON has no \0 escape, so it can only be the
// terminator a STOMP frame ends on.
func TestExpandEscapesForAWebSocket(t *testing.T) {
	cases := []struct{ in, want string }{
		{`{"text":"line\nbreak"}`, `{"text":"line\nbreak"}`},
		{`a\tb`, `a\tb`},
		{`back\\slash`, `back\\slash`},
		{`CONNECT\n\n\0`, "CONNECT\\n\\n\x00"},
		{"", ""},
	}
	for _, testCase := range cases {
		if got := expandEscapes(testCase.in, webSocketEscapes); got != testCase.want {
			t.Errorf("expandEscapes(%q) = %q, want %q", testCase.in, got, testCase.want)
		}
	}
}

func startStreamFor(t *testing.T, suite parserhttp.HttpSuite) (*Application, *fakeStreamSession, tcell.SimulationScreen) {
	t.Helper()
	application := buildStreamApplication(t)
	session := newFakeStreamSession()
	application.dialStream = func(context.Context, parserhttp.HttpSuite) (runnerstream.Session, error) {
		return session, nil
	}
	screen, _ := runTestApplication(t, application)
	application.Element.QueueUpdateDraw(func() { onSuiteRun(application)(suite) })
	waitFor(t, "the stream pane taking the response slot", func() bool {
		return application.Workspace.ResponseElement() == application.Stream.Element
	})
	return application, session, screen
}

func nextSent(t *testing.T, session *fakeStreamSession) runnerstream.Frame {
	t.Helper()
	select {
	case frame := <-session.sent:
		return frame
	case <-time.After(5 * time.Second):
		t.Fatal("nothing was sent")
	}
	return runnerstream.Frame{}
}

func TestTUIStreamComposerSendsAFrame(t *testing.T) {
	application, session, screen := startStreamFor(t, streamSuite())

	screen.InjectKey(tcell.KeyRune, 's', tcell.ModNone)
	waitFor(t, "the composer opening", func() bool {
		return application.Model.CurrentOverlay() == OverlaySendFrame
	})

	application.Element.QueueUpdateDraw(func() {
		application.SendFrame.SetText(`{"action":"subscribe"}`)
		application.sendFrame()
	})

	frame := nextSent(t, session)
	if string(frame.Payload) != `{"action":"subscribe"}` {
		t.Fatalf("payload = %q", frame.Payload)
	}
	if frame.Opcode != runnerstream.Text {
		t.Fatalf("opcode = %v, want text for a websocket", frame.Opcode)
	}
	waitFor(t, "the composer closing", func() bool {
		return application.Model.CurrentOverlay() == OverlayNone
	})
}

// A raw socket is line oriented, and an input field cannot hold a newline.
func TestTUIStreamComposerExpandsEscapesForARawSocket(t *testing.T) {
	suite := parserhttp.HttpSuite{
		Name:      "cache",
		Method:    "SOCKET",
		Uri:       "tcp://127.0.0.1:6379",
		Transport: parserhttp.TransportTCP,
	}
	application, session, _ := startStreamFor(t, suite)

	application.Element.QueueUpdateDraw(func() {
		application.SendFrame.SetText(`PING\r\n`)
		application.sendFrame()
	})

	frame := nextSent(t, session)
	if string(frame.Payload) != "PING\r\n" {
		t.Fatalf("payload = %q, want a real CRLF", frame.Payload)
	}
	if frame.Opcode != runnerstream.Bytes {
		t.Fatalf("opcode = %v, want bytes for a raw socket", frame.Opcode)
	}
}

// Expanding escapes in a WebSocket message would corrupt JSON that contains a
// legitimate \n inside a string.
func TestTUIStreamComposerLeavesAWebSocketPayloadAlone(t *testing.T) {
	application, session, _ := startStreamFor(t, streamSuite())

	application.Element.QueueUpdateDraw(func() {
		application.SendFrame.SetText(`{"text":"line\nbreak"}`)
		application.sendFrame()
	})

	frame := nextSent(t, session)
	if string(frame.Payload) != `{"text":"line\nbreak"}` {
		t.Fatalf("payload = %q, want the escape left as typed", frame.Payload)
	}
}

func TestTUIStreamComposerRefusesAnEmptyFrame(t *testing.T) {
	application, session, _ := startStreamFor(t, streamSuite())

	application.Element.QueueUpdateDraw(func() {
		application.SendFrame.SetText("   ")
		application.sendFrame()
	})

	select {
	case frame := <-session.sent:
		t.Fatalf("an empty frame was sent: %q", frame.Payload)
	case <-time.After(300 * time.Millisecond):
	}
}

func TestTUIStreamComposerWillNotOpenWithoutAConnection(t *testing.T) {
	application := buildStreamApplication(t)
	runTestApplication(t, application)

	application.Element.QueueUpdateDraw(func() { application.openSendFrame() })
	waitFor(t, "the composer staying shut", func() bool {
		return application.Model.CurrentOverlay() == OverlayNone
	})
}

// The body of a stream request is its opening message. Both examples in
// example/ rely on this, and without it they promise what does not happen.
func TestTUIStreamSendsTheRequestBodyOnConnect(t *testing.T) {
	suite := streamSuite()
	suite.Body = `{"action":"subscribe","symbol":"BTC"}`
	_, session, _ := startStreamFor(t, suite)

	frame := nextSent(t, session)
	if string(frame.Payload) != suite.Body {
		t.Fatalf("payload = %q, want the request body", frame.Payload)
	}
	if frame.Opcode != runnerstream.Text {
		t.Fatalf("opcode = %v, want text for a websocket", frame.Opcode)
	}
}

func TestTUIStreamSendsARawSocketBodyAsBytes(t *testing.T) {
	suite := parserhttp.HttpSuite{
		Name:      "cache",
		Method:    "SOCKET",
		Uri:       "tcp://127.0.0.1:6379",
		Transport: parserhttp.TransportTCP,
		Body:      "PING\r\n",
	}
	_, session, _ := startStreamFor(t, suite)

	frame := nextSent(t, session)
	if string(frame.Payload) != "PING\r\n" {
		t.Fatalf("payload = %q", frame.Payload)
	}
	if frame.Opcode != runnerstream.Bytes {
		t.Fatalf("opcode = %v, want bytes for a raw socket", frame.Opcode)
	}
}

func TestTUIStreamSendsNothingWhenTheBodyIsEmpty(t *testing.T) {
	suite := streamSuite()
	suite.Body = "   \n"
	_, session, _ := startStreamFor(t, suite)

	select {
	case frame := <-session.sent:
		t.Fatalf("an empty body was sent as a frame: %q", frame.Payload)
	case <-time.After(300 * time.Millisecond):
	}
}

// The earlier test asserted on the model, which reported the overlay closed
// while the page was still drawn. Assert on what is actually on screen.
func TestTUIStreamComposerLeavesTheScreenAfterSending(t *testing.T) {
	application, session, screen := startStreamFor(t, streamSuite())

	screen.InjectKey(tcell.KeyRune, 's', tcell.ModNone)
	waitForScreenText(t, application, screen, "Send frame")

	application.Element.QueueUpdateDraw(func() {
		application.SendFrame.SetText("ping")
		application.sendFrame()
	})
	nextSent(t, session)

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if !strings.Contains(applicationText(application, screen), "Send frame") {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("the composer is still drawn after the frame was sent")
}

// The parser trims the trailing newline off a body, so a .socket file cannot
// express the CRLF a line oriented protocol ends on. Escapes cover for that.
func TestTUIStreamExpandsEscapesInARawSocketBody(t *testing.T) {
	suite := parserhttp.HttpSuite{
		Name:      "cache",
		Method:    "SOCKET",
		Uri:       "tcp://127.0.0.1:6379",
		Transport: parserhttp.TransportTCP,
		Body:      `PING\r\n`,
	}
	_, session, _ := startStreamFor(t, suite)

	frame := nextSent(t, session)
	if string(frame.Payload) != "PING\r\n" {
		t.Fatalf("payload = %q, want a real CRLF", frame.Payload)
	}
}

// A WebSocket message needs no terminator, and expanding escapes would corrupt
// JSON that legitimately contains one.
func TestTUIStreamLeavesAWebSocketBodyAsWritten(t *testing.T) {
	suite := streamSuite()
	suite.Body = `{"text":"line\nbreak"}`
	_, session, _ := startStreamFor(t, suite)

	frame := nextSent(t, session)
	if string(frame.Payload) != `{"text":"line\nbreak"}` {
		t.Fatalf("payload = %q, want the escape left as written", frame.Payload)
	}
}

// The whole point of the null byte: a STOMP frame cannot be sent without one.
func TestTUIStreamSendsAStompFrameOverAWebSocket(t *testing.T) {
	suite := streamSuite()
	suite.Body = `CONNECT\naccept-version:1.2\nhost:/\n\n\0`
	_, session, _ := startStreamFor(t, suite)

	frame := nextSent(t, session)
	want := "CONNECT\\naccept-version:1.2\\nhost:/\\n\\n\x00"
	if string(frame.Payload) != want {
		t.Fatalf("payload = %q, want %q", frame.Payload, want)
	}
	if frame.Opcode != runnerstream.Text {
		t.Fatalf("opcode = %v, want text", frame.Opcode)
	}
}

func TestTUIStreamSendsAStompFrameOverARawSocket(t *testing.T) {
	suite := parserhttp.HttpSuite{
		Name:      "stomp",
		Method:    "SOCKET",
		Uri:       "tcp://127.0.0.1:61613",
		Transport: parserhttp.TransportTCP,
		Body:      `CONNECT\naccept-version:1.2\nhost:/\n\n\0`,
	}
	_, session, _ := startStreamFor(t, suite)

	frame := nextSent(t, session)
	want := "CONNECT\naccept-version:1.2\nhost:/\n\n\x00"
	if string(frame.Payload) != want {
		t.Fatalf("payload = %q, want %q", frame.Payload, want)
	}
}
