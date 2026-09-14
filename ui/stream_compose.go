package ui

import (
	"context"
	"strings"

	"github.com/Hecatoncheir/lazyrest/keymap"
	parserhttp "github.com/Hecatoncheir/lazyrest/parser/http"
	runnerstream "github.com/Hecatoncheir/lazyrest/runner/stream"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func (application *Application) buildSendFrameInput() {
	translator := application.config.Locale
	input := tview.NewInputField().
		SetLabel(translator.Text("frame") + ": ")
	input.SetBorder(true).
		SetTitle(translator.Text("send_frame"))
	input.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			application.sendFrame()
		}
	})
	// The recall keys are arrows rather than letters: anything printable would
	// have to be typed into this very field.
	input.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		bindings := application.config.Keybindings
		if bindings == nil {
			return event
		}
		switch {
		case bindings.Matches(keymap.StreamRecallPrevious, event):
			application.recallSentFrame(-1)
			return nil
		case bindings.Matches(keymap.StreamRecallNext, event):
			application.recallSentFrame(1)
			return nil
		}
		return event
	})
	application.SendFrame = input
	application.applySendFrameTheme()
}

func (application *Application) applySendFrameTheme() {
	if application.SendFrame == nil {
		return
	}
	uiTheme := application.theme.Suites
	application.SendFrame.
		SetLabelColor(uiTheme.SuiteForeground).
		SetFieldTextColor(uiTheme.SuiteFocusForeground).
		SetFieldBackgroundColor(uiTheme.SuiteFocusBackground)
	application.SendFrame.SetBackgroundColor(uiTheme.BackgroundFocus)
	application.SendFrame.SetBorderColor(uiTheme.BorderFocus)
	application.SendFrame.SetTitleColor(uiTheme.TitleFocus)
}

// showStreamError puts a transient message in the footer. It leaves the frame
// log alone on purpose: losing what the connection has said, in order to report
// that nothing is connected, would be a poor trade.
func (application *Application) showStreamError(message string) {
	application.showExportError(message)
}

func (application *Application) currentStream() (runnerstream.Session, parserhttp.Transport) {
	application.streamMutex.Lock()
	defer application.streamMutex.Unlock()
	return application.streamSession, application.streamTransport
}

func (application *Application) openSendFrame() {
	if session, _ := application.currentStream(); session == nil {
		application.showStreamError(application.config.Locale.Text("no_stream_to_send"))
		return
	}
	application.resetSentFrameCursor()
	application.SendFrame.SetText("")
	application.openOverlay(OverlaySendFrame)
}

// SentFrameHistoryLimit bounds what the composer remembers.
const SentFrameHistoryLimit = 50

// rememberSentFrame keeps what was sent so the composer can offer it again.
// The history lives in memory only, and is never written to disk: a frame
// carries credentials as readily as a request body does.
func (application *Application) rememberSentFrame(text string) {
	application.sentFramesMutex.Lock()
	defer application.sentFramesMutex.Unlock()

	// Resending the same frame is ordinary — a keepalive, a repeated poll — and
	// recording each one would push everything else out of reach.
	if count := len(application.sentFrames); count > 0 && application.sentFrames[count-1] == text {
		return
	}
	application.sentFrames = append(application.sentFrames, text)
	if excess := len(application.sentFrames) - SentFrameHistoryLimit; excess > 0 {
		application.sentFrames = append([]string(nil), application.sentFrames[excess:]...)
	}
}

func (application *Application) resetSentFrameCursor() {
	application.sentFramesMutex.Lock()
	defer application.sentFramesMutex.Unlock()
	application.sentFrameCursor = len(application.sentFrames)
}

// SentFrames is what the composer will offer, oldest first.
func (application *Application) SentFrames() []string {
	application.sentFramesMutex.Lock()
	defer application.sentFramesMutex.Unlock()
	return append([]string(nil), application.sentFrames...)
}

// recallSentFrame walks the history. The position one past the newest entry is
// the empty draft, so walking forward off the end clears the field rather than
// sticking on the last frame sent.
func (application *Application) recallSentFrame(delta int) {
	application.sentFramesMutex.Lock()
	cursor := application.sentFrameCursor + delta
	if cursor < 0 {
		cursor = 0
	}
	if cursor > len(application.sentFrames) {
		cursor = len(application.sentFrames)
	}
	application.sentFrameCursor = cursor
	text := ""
	if cursor < len(application.sentFrames) {
		text = application.sentFrames[cursor]
	}
	application.sentFramesMutex.Unlock()

	if application.SendFrame != nil {
		application.SendFrame.SetText(text)
	}
}

func (application *Application) sendFrame() {
	translator := application.config.Locale
	text := application.SendFrame.GetText()
	if strings.TrimSpace(text) == "" {
		application.showStreamError(translator.Text("frame_required"))
		return
	}

	session, transport := application.currentStream()
	application.closeOverlay()
	if session == nil {
		application.showStreamError(translator.Text("no_stream_to_send"))
		return
	}

	// Remembered as typed, before any escape is expanded: what is offered back
	// should be what was written.
	application.rememberSentFrame(text)

	frame := runnerstream.Frame{Opcode: runnerstream.Text, Payload: []byte(text)}
	if transport == parserhttp.TransportTCP {
		// A raw socket is line oriented more often than not, and an input field
		// cannot hold a real newline, so escapes are expanded here. A WebSocket
		// message is a whole message on its own, and expanding escapes there
		// would corrupt JSON that legitimately contains a \n.
		frame = runnerstream.Frame{Opcode: runnerstream.Bytes, Payload: []byte(expandEscapes(text))}
	}

	// Sending writes to the network, which must not happen on the draw
	// goroutine.
	go func() {
		if err := session.Send(context.Background(), frame); err != nil {
			application.Element.QueueUpdateDraw(func() {
				application.showStreamError(err.Error())
			})
		}
	}()
}

// expandEscapes turns the escapes a line oriented protocol needs into the bytes
// they stand for. An unknown escape is left exactly as it was typed.
func expandEscapes(text string) string {
	var builder strings.Builder
	for index := 0; index < len(text); index++ {
		if text[index] != '\\' || index+1 >= len(text) {
			builder.WriteByte(text[index])
			continue
		}
		index++
		switch text[index] {
		case 'n':
			builder.WriteByte('\n')
		case 'r':
			builder.WriteByte('\r')
		case 't':
			builder.WriteByte('\t')
		case '\\':
			builder.WriteByte('\\')
		default:
			builder.WriteByte('\\')
			builder.WriteByte(text[index])
		}
	}
	return builder.String()
}
