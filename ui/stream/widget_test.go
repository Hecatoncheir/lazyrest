package stream

import (
	"strings"
	"testing"
	"time"

	runnerstream "github.com/Hecatoncheir/lazyrest/runner/stream"
	"github.com/Hecatoncheir/lazyrest/ui/theme"
)

func buildWidget(t *testing.T, log *Log) *Widget {
	t.Helper()
	widget := NewWidget()
	widget.Build(Parameters{Theme: theme.NewDefault()})
	widget.SetLog(log)
	return widget
}

func at(second int) time.Time {
	return time.Date(2026, 9, 14, 10, 0, second, 0, time.UTC)
}

func TestWidgetShowsAWaitingStateWhenEmpty(t *testing.T) {
	widget := buildWidget(t, NewLog(10, 1<<20))
	if got := widget.Element.GetText(true); !strings.Contains(got, "Waiting for frames") {
		t.Fatalf("text = %q, want the waiting state", got)
	}
}

func TestWidgetRendersFramesOldestFirst(t *testing.T) {
	log := NewLog(10, 1<<20)
	log.Append(runnerstream.Frame{At: at(1), Direction: runnerstream.Sent, Opcode: runnerstream.Text, Payload: []byte("subscribe")})
	log.Append(runnerstream.Frame{At: at(2), Direction: runnerstream.Received, Opcode: runnerstream.Text, Payload: []byte("welcome")})

	widget := buildWidget(t, log)
	text := widget.Element.GetText(true)

	sent := strings.Index(text, "subscribe")
	received := strings.Index(text, "welcome")
	if sent < 0 || received < 0 {
		t.Fatalf("text = %q, want both payloads", text)
	}
	if sent > received {
		t.Fatal("frames are rendered newest first, want oldest first")
	}
	if !strings.Contains(text, "10:00:01.000") {
		t.Fatalf("text = %q, want a timestamp", text)
	}
}

// A stream carries credentials as readily as a response body does.
func TestWidgetRedactsSecretsInPayloads(t *testing.T) {
	log := NewLog(10, 1<<20)
	log.Append(runnerstream.Frame{
		At:      at(1),
		Opcode:  runnerstream.Text,
		Payload: []byte(`{"token":"super-secret-value"}`),
	})

	widget := NewWidget()
	widget.Build(Parameters{Theme: theme.NewDefault()})
	widget.SetSecretValues([]string{"super-secret-value"})
	widget.SetLog(log)

	if text := widget.Element.GetText(true); strings.Contains(text, "super-secret-value") {
		t.Fatalf("the secret reached the screen: %q", text)
	}
}

func TestWidgetShowsBinaryPayloadAsHex(t *testing.T) {
	log := NewLog(10, 1<<20)
	log.Append(runnerstream.Frame{At: at(1), Opcode: runnerstream.Binary, Payload: []byte{0x01, 0xff, 0x10}})

	widget := buildWidget(t, log)
	if text := widget.Element.GetText(true); !strings.Contains(text, "01 ff 10") {
		t.Fatalf("text = %q, want hex", text)
	}
}

// A raw socket is under no obligation to carry text.
func TestWidgetShowsUnprintableBytesAsHex(t *testing.T) {
	log := NewLog(10, 1<<20)
	log.Append(runnerstream.Frame{At: at(1), Opcode: runnerstream.Bytes, Payload: []byte{0x00, 0x01, 0x02}})

	widget := buildWidget(t, log)
	if text := widget.Element.GetText(true); !strings.Contains(text, "00 01 02") {
		t.Fatalf("text = %q, want hex", text)
	}
}

func TestWidgetKeepsPrintableRawBytesAsText(t *testing.T) {
	log := NewLog(10, 1<<20)
	log.Append(runnerstream.Frame{At: at(1), Opcode: runnerstream.Bytes, Payload: []byte("+PONG")})

	widget := buildWidget(t, log)
	if text := widget.Element.GetText(true); !strings.Contains(text, "+PONG") {
		t.Fatalf("text = %q, want the text kept as text", text)
	}
}

func TestWidgetReportsDroppedFrames(t *testing.T) {
	log := NewLog(2, 1<<20)
	for _, payload := range []string{"a", "b", "c", "d"} {
		log.Append(runnerstream.Frame{At: at(1), Opcode: runnerstream.Text, Payload: []byte(payload)})
	}

	widget := buildWidget(t, log)
	if text := widget.Element.GetText(true); !strings.Contains(text, "2 earlier frames dropped") {
		t.Fatalf("text = %q, want the dropped notice", text)
	}
}

// Pausing is what makes the pane readable while the connection keeps talking.
func TestWidgetPausedKeepsTheCurrentText(t *testing.T) {
	log := NewLog(10, 1<<20)
	log.Append(runnerstream.Frame{At: at(1), Opcode: runnerstream.Text, Payload: []byte("first")})

	widget := buildWidget(t, log)
	widget.SetFollowing(false)

	log.Append(runnerstream.Frame{At: at(2), Opcode: runnerstream.Text, Payload: []byte("second")})
	widget.Render()

	text := widget.Element.GetText(true)
	if !strings.Contains(text, "first") {
		t.Fatalf("text = %q, want the frame from before the pause", text)
	}
	if strings.Contains(text, "second") {
		t.Fatal("a frame that arrived while paused was drawn")
	}

	widget.SetFollowing(true)
	if text := widget.Element.GetText(true); !strings.Contains(text, "second") {
		t.Fatalf("text = %q, want the missed frame after resuming", text)
	}
}

func TestWidgetTitleShowsTheFollowingState(t *testing.T) {
	widget := buildWidget(t, NewLog(10, 1<<20))

	if title := widget.Element.GetTitle(); !strings.Contains(title, "following") {
		t.Fatalf("title = %q, want the following state", title)
	}
	widget.ToggleFollowing()
	if title := widget.Element.GetTitle(); !strings.Contains(title, "paused") {
		t.Fatalf("title = %q, want the paused state", title)
	}
}

func TestWidgetMarksATruncatedFrame(t *testing.T) {
	log := NewLog(10, 1<<20)
	log.Append(runnerstream.Frame{At: at(1), Opcode: runnerstream.Text, Payload: []byte("abcd"), Truncated: true})

	widget := buildWidget(t, log)
	if text := widget.Element.GetText(true); !strings.Contains(text, "4B+") {
		t.Fatalf("text = %q, want a truncated size marker", text)
	}
}
