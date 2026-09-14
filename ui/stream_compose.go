package ui

import (
	"context"
	"strings"

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
	application.SendFrame.SetText("")
	application.openOverlay(OverlaySendFrame)
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
