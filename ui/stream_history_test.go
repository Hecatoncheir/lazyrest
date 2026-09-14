package ui

import (
	"strconv"
	"testing"
	"time"

	parserhttp "github.com/Hecatoncheir/lazyrest/parser/http"

	"github.com/gdamore/tcell/v2"
)

// composerText reads the field on the draw goroutine and waits for the read to
// happen. tview is not safe to touch from the test goroutine.
func composerText(t *testing.T, application *Application) string {
	t.Helper()
	var text string
	done := make(chan struct{})
	application.Element.QueueUpdateDraw(func() {
		text = application.SendFrame.GetText()
		close(done)
	})
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("the draw goroutine never ran the read")
	}
	return text
}

// sendThrough drives the real path: open the composer, type, send. Calling
// sendFrame directly leaves focus wherever it was, and the pane never gets the
// key that opens the composer next time.
func sendThrough(
	t *testing.T,
	application *Application,
	session *fakeStreamSession,
	screen tcell.SimulationScreen,
	text string,
) {
	t.Helper()
	screen.InjectKey(tcell.KeyRune, 's', tcell.ModNone)
	waitFor(t, "the composer opening", func() bool {
		return application.Model.CurrentOverlay() == OverlaySendFrame
	})
	application.Element.QueueUpdateDraw(func() {
		application.SendFrame.SetText(text)
		application.sendFrame()
	})
	nextSent(t, session)
	waitFor(t, "the composer closing", func() bool {
		return application.Model.CurrentOverlay() == OverlayNone
	})
}

func TestTUIComposerRecallsFramesSentEarlier(t *testing.T) {
	application, session, screen := startStreamFor(t, streamSuite())
	sendThrough(t, application, session, screen, `{"action":"subscribe"}`)
	sendThrough(t, application, session, screen, `{"action":"ping"}`)

	screen.InjectKey(tcell.KeyRune, 's', tcell.ModNone)
	waitFor(t, "the composer opening", func() bool {
		return application.Model.CurrentOverlay() == OverlaySendFrame
	})
	if got := composerText(t, application); got != "" {
		t.Fatalf("the composer opened holding %q, want an empty draft", got)
	}

	screen.InjectKey(tcell.KeyUp, 0, tcell.ModNone)
	waitFor(t, "the newest frame", func() bool { return composerText(t, application) == `{"action":"ping"}` })

	screen.InjectKey(tcell.KeyUp, 0, tcell.ModNone)
	waitFor(t, "the older frame", func() bool { return composerText(t, application) == `{"action":"subscribe"}` })

	screen.InjectKey(tcell.KeyDown, 0, tcell.ModNone)
	waitFor(t, "the newer frame again", func() bool { return composerText(t, application) == `{"action":"ping"}` })

	// Walking forward off the end returns to the draft rather than sticking.
	screen.InjectKey(tcell.KeyDown, 0, tcell.ModNone)
	waitFor(t, "the empty draft", func() bool { return composerText(t, application) == "" })
}

func TestTUIComposerRecallStopsAtTheOldestFrame(t *testing.T) {
	application, session, screen := startStreamFor(t, streamSuite())
	sendThrough(t, application, session, screen, "only")

	screen.InjectKey(tcell.KeyRune, 's', tcell.ModNone)
	waitFor(t, "the composer opening", func() bool {
		return application.Model.CurrentOverlay() == OverlaySendFrame
	})
	for range 5 {
		screen.InjectKey(tcell.KeyUp, 0, tcell.ModNone)
	}
	waitFor(t, "the oldest frame", func() bool { return composerText(t, application) == "only" })
}

// Resending the same frame is ordinary; recording each would push everything
// else out of reach.
func TestComposerHistoryIgnoresConsecutiveRepeats(t *testing.T) {
	application := hintsApplication(t)
	application.rememberSentFrame("PING")
	application.rememberSentFrame("PING")
	application.rememberSentFrame("PONG")
	application.rememberSentFrame("PING")

	history := application.SentFrames()
	want := []string{"PING", "PONG", "PING"}
	if len(history) != len(want) {
		t.Fatalf("history = %v, want %v", history, want)
	}
	for index := range want {
		if history[index] != want[index] {
			t.Fatalf("history = %v, want %v", history, want)
		}
	}
}

func TestComposerHistoryIsBounded(t *testing.T) {
	application := hintsApplication(t)
	for index := range SentFrameHistoryLimit + 10 {
		application.rememberSentFrame(strconv.Itoa(index))
	}

	history := application.SentFrames()
	if len(history) != SentFrameHistoryLimit {
		t.Fatalf("history holds %d frames, want the bound of %d", len(history), SentFrameHistoryLimit)
	}
	if history[0] != strconv.Itoa(10) {
		t.Fatalf("oldest kept frame = %q, want the eleventh", history[0])
	}
	if newest := history[len(history)-1]; newest != strconv.Itoa(SentFrameHistoryLimit+9) {
		t.Fatalf("newest frame = %q", newest)
	}
}

// What is offered back should be what was written, not what went on the wire.
func TestComposerHistoryKeepsTheTextAsTyped(t *testing.T) {
	suite := streamSuite()
	suite.Transport = parserhttp.TransportTCP
	application, session, screen := startStreamFor(t, suite)

	sendThrough(t, application, session, screen, `PING\r\n`)

	history := application.SentFrames()
	if len(history) != 1 || history[0] != `PING\r\n` {
		t.Fatalf("history = %#v, want the escape as typed", history)
	}
}
