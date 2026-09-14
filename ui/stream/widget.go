package stream

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/Hecatoncheir/lazyrest/locale"
	parserhttp "github.com/Hecatoncheir/lazyrest/parser/http"
	runnerstream "github.com/Hecatoncheir/lazyrest/runner/stream"
	"github.com/Hecatoncheir/lazyrest/ui/syntax"
	"github.com/Hecatoncheir/lazyrest/ui/theme"

	"github.com/rivo/tview"
)

// DefaultMaxVisibleFrames bounds how much is rendered at once. The log may hold
// far more; drawing all of it on every repaint would undo the coalescing.
const DefaultMaxVisibleFrames = 300

// hexPreviewBytes is how much of an unprintable payload is shown as hex.
const hexPreviewBytes = 48

type Parameters struct {
	Theme            theme.Theme
	Locale           *locale.Translator
	MaxVisibleFrames int
}

type Widget struct {
	Element *tview.TextView

	theme      theme.ProducerTheme
	syntax     syntax.Palette
	locale     *locale.Translator
	log        *Log
	secrets    []string
	following  bool
	maxVisible int
}

func NewWidget() *Widget {
	return &Widget{following: true}
}

func (widget *Widget) Build(parameters Parameters) tview.Primitive {
	if parameters.Locale == nil {
		parameters.Locale = locale.English()
	}
	widget.locale = parameters.Locale
	widget.theme = parameters.Theme.Producer
	widget.syntax = parameters.Theme.Syntax
	widget.maxVisible = parameters.MaxVisibleFrames
	if widget.maxVisible <= 0 {
		widget.maxVisible = DefaultMaxVisibleFrames
	}

	element := tview.NewTextView()
	element.SetDynamicColors(true)
	element.SetWrap(true)
	element.SetBorder(true)
	element.SetBackgroundColor(widget.theme.Background)
	element.SetTextColor(widget.theme.Foreground)
	element.SetBorderColor(widget.theme.Border)
	element.SetTitleColor(widget.theme.Title)

	widget.Element = element
	widget.updateTitle()
	widget.Render()
	return element
}

func (widget *Widget) SetLog(log *Log) {
	widget.log = log
	widget.Render()
}

// SetSecretValues installs the values that must never be shown. A stream
// carries credentials as readily as a response body does, so redaction happens
// here as it does when Producer renders history.
func (widget *Widget) SetSecretValues(secrets []string) {
	widget.secrets = secrets
}

func (widget *Widget) Following() bool { return widget.following }

func (widget *Widget) SetFollowing(following bool) {
	widget.following = following
	widget.updateTitle()
	widget.Render()
}

func (widget *Widget) ToggleFollowing() { widget.SetFollowing(!widget.following) }

func (widget *Widget) updateTitle() {
	if widget.Element == nil || widget.locale == nil {
		return
	}
	state := widget.locale.Text("stream_paused")
	if widget.following {
		state = widget.locale.Text("stream_following")
	}
	widget.Element.SetTitle(fmt.Sprintf(" %s — %s ", widget.locale.Text("stream"), state))
}

// Render rebuilds the visible text from the log. Paused keeps the current text
// so a reader can study a frame while the connection keeps talking.
func (widget *Widget) Render() {
	if widget.Element == nil {
		return
	}
	if !widget.following {
		return
	}
	widget.Element.SetText(widget.text())
	widget.Element.ScrollToEnd()
}

func (widget *Widget) text() string {
	if widget.log == nil || widget.log.Len() == 0 {
		return tview.Escape(widget.locale.Text("stream_waiting"))
	}

	var builder strings.Builder
	if dropped := widget.log.Dropped(); dropped > 0 {
		builder.WriteString("[gray]")
		builder.WriteString(tview.Escape(widget.locale.Format("stream_frames_dropped", dropped)))
		builder.WriteString("[-]\n")
	}
	for _, frame := range widget.log.Tail(widget.maxVisible) {
		builder.WriteString(widget.line(frame))
		builder.WriteByte('\n')
	}
	return builder.String()
}

func (widget *Widget) line(frame runnerstream.Frame) string {
	var builder strings.Builder
	builder.WriteString("[gray]")
	builder.WriteString(frame.At.Format("15:04:05.000"))
	builder.WriteString("[-] ")
	builder.WriteString(directionMark(frame.Direction))
	builder.WriteString(" [gray]")
	builder.WriteString(frame.Opcode.String())
	builder.WriteByte(' ')
	builder.WriteString(humanBytes(len(frame.Payload)))
	if frame.Truncated {
		builder.WriteString("+")
	}
	builder.WriteString("[-] ")
	builder.WriteString(widget.payload(frame))
	return builder.String()
}

func directionMark(direction runnerstream.Direction) string {
	if direction == runnerstream.Sent {
		return "[green]→[-]"
	}
	return "[blue]←[-]"
}

func humanBytes(size int) string {
	if size < 1024 {
		return fmt.Sprintf("%dB", size)
	}
	return fmt.Sprintf("%.1fKB", float64(size)/1024)
}

// payload redacts first, then decides how to show what is left. Binary and
// anything that is not printable text becomes hex: a raw socket is under no
// obligation to carry text.
func (widget *Widget) payload(frame runnerstream.Frame) string {
	if len(frame.Payload) == 0 {
		return ""
	}
	if frame.Opcode == runnerstream.Binary || !printable(frame.Payload) {
		return "[gray]" + tview.Escape(hexPreview(frame.Payload)) + "[-]"
	}

	text := parserhttp.RedactSecrets(string(frame.Payload), widget.secrets)
	if looksLikeJSON(text) {
		// syntax.Highlight escapes its own output; escaping again would show
		// the tags rather than apply them.
		return syntax.Highlight(text, syntax.LanguageJSON, widget.syntax)
	}
	return tview.Escape(text)
}

func printable(payload []byte) bool {
	if !utf8.Valid(payload) {
		return false
	}
	for _, character := range string(payload) {
		if character == '\n' || character == '\r' || character == '\t' {
			continue
		}
		if character < 0x20 || character == 0x7f {
			return false
		}
	}
	return true
}

func hexPreview(payload []byte) string {
	shown := payload
	suffix := ""
	if len(shown) > hexPreviewBytes {
		shown = shown[:hexPreviewBytes]
		suffix = " …"
	}
	parts := make([]string, len(shown))
	for index, value := range shown {
		parts[index] = fmt.Sprintf("%02x", value)
	}
	return strings.Join(parts, " ") + suffix
}

func looksLikeJSON(text string) bool {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return false
	}
	return trimmed[0] == '{' || trimmed[0] == '['
}
