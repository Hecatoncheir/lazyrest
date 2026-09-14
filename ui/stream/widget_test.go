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

// A frame must occupy exactly one row. A payload ending in CRLF used to add a
// blank row, so the log no longer lined up with the frames in it.
func TestWidgetKeepsOneFramePerLine(t *testing.T) {
	log := NewLog(10, 1<<20)
	log.Append(runnerstream.Frame{At: at(1), Opcode: runnerstream.Bytes, Payload: []byte("+PONG\r\n")})
	log.Append(runnerstream.Frame{At: at(2), Opcode: runnerstream.Bytes, Payload: []byte("+OK\r\n")})

	widget := buildWidget(t, log)
	text := strings.TrimRight(widget.Element.GetText(true), "\n")

	if lines := strings.Split(text, "\n"); len(lines) != 2 {
		t.Fatalf("rendered %d lines for 2 frames:\n%s", len(lines), text)
	}
	if !strings.Contains(text, `+PONG\r\n`) {
		t.Fatalf("text = %q, want the terminator shown rather than applied", text)
	}
}

// A STOMP frame ends on a null byte. Dropping the whole frame into hex because
// of that one byte would make a text protocol unreadable.
func TestWidgetShowsANullTerminatedFrameAsText(t *testing.T) {
	log := NewLog(10, 1<<20)
	log.Append(runnerstream.Frame{
		At:      at(1),
		Opcode:  runnerstream.Bytes,
		Payload: []byte("CONNECTED\nversion:1.2\n\n\x00"),
	})

	widget := buildWidget(t, log)
	text := widget.Element.GetText(true)

	if !strings.Contains(text, "CONNECTED") {
		t.Fatalf("text = %q, want the frame shown as text", text)
	}
	if !strings.Contains(text, `\0`) {
		t.Fatalf("text = %q, want the terminator shown as an escape", text)
	}
	if strings.Contains(text, "43 4f 4e") {
		t.Fatalf("text = %q, want text rather than hex", text)
	}
	if lines := strings.Split(strings.TrimRight(text, "\n"), "\n"); len(lines) != 1 {
		t.Fatalf("rendered %d lines for one frame:\n%s", len(lines), text)
	}
}

// Allowing the null byte must not turn genuinely binary payloads into text.
func TestWidgetStillShowsBinaryAsHexAlongsideANull(t *testing.T) {
	log := NewLog(10, 1<<20)
	log.Append(runnerstream.Frame{At: at(1), Opcode: runnerstream.Bytes, Payload: []byte{0x00, 0x01, 0x02}})

	widget := buildWidget(t, log)
	if text := widget.Element.GetText(true); !strings.Contains(text, "00 01 02") {
		t.Fatalf("text = %q, want hex", text)
	}
}

// A payload without the topic it arrived on hides the part that matters most.
func TestWidgetShowsFrameAttributes(t *testing.T) {
	log := NewLog(10, 1<<20)
	log.Append(runnerstream.Frame{
		At:      at(1),
		Opcode:  runnerstream.Text,
		Payload: []byte(`{"celsius":21}`),
		Attributes: []runnerstream.Attribute{
			{Name: "topic", Value: "sensors/1/temp"},
			{Name: "qos", Value: "1"},
		},
	})

	widget := buildWidget(t, log)
	text := widget.Element.GetText(true)

	for _, expected := range []string{"topic=sensors/1/temp", "qos=1"} {
		if !strings.Contains(text, expected) {
			t.Errorf("text = %q, want %q", text, expected)
		}
	}
	// Order is the protocol's, not a map's.
	if strings.Index(text, "topic=") > strings.Index(text, "qos=") {
		t.Errorf("attributes are reordered: %q", text)
	}
	// They belong before the payload, which is read less often.
	if strings.Index(text, "qos=") > strings.Index(text, "celsius") {
		t.Errorf("attributes follow the payload: %q", text)
	}
}

// A topic is built from the same variables a body is, so it can carry a secret
// just as easily.
func TestWidgetRedactsSecretsInAttributes(t *testing.T) {
	log := NewLog(10, 1<<20)
	log.Append(runnerstream.Frame{
		At:         at(1),
		Opcode:     runnerstream.Text,
		Payload:    []byte("{}"),
		Attributes: []runnerstream.Attribute{{Name: "topic", Value: "devices/super-secret-id/state"}},
	})

	widget := NewWidget()
	widget.Build(Parameters{Theme: theme.NewDefault()})
	widget.SetSecretValues([]string{"super-secret-id"})
	widget.SetLog(log)

	if text := widget.Element.GetText(true); strings.Contains(text, "super-secret-id") {
		t.Fatalf("the secret reached the screen through an attribute: %q", text)
	}
}

func TestFrameAttributeLookup(t *testing.T) {
	frame := runnerstream.Frame{Attributes: []runnerstream.Attribute{{Name: "topic", Value: "a/b"}}}
	if value, found := frame.Attribute("topic"); !found || value != "a/b" {
		t.Fatalf("Attribute(topic) = %q, %v", value, found)
	}
	if _, found := frame.Attribute("qos"); found {
		t.Fatal("Attribute reported a value that was never recorded")
	}
}
