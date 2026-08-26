package progress

import (
	"strings"
	"testing"

	"github.com/Hecatoncheir/lazyrest/locale"
)

func TestBodyShowsAPercentageWhenTheTotalIsKnown(t *testing.T) {
	cases := []struct {
		name           string
		current, total int64
		want           string
	}{
		{name: "nothing read yet", current: 0, total: 100, want: "[----------] 0%"},
		{name: "half way", current: 50, total: 100, want: "[=====-----] 50%"},
		{name: "complete", current: 100, total: 100, want: "[==========] 100%"},
		// A server that reports less than it sends must not overflow the bar.
		{name: "more than announced", current: 150, total: 100, want: "[==========] 100%"},
		{name: "a negative count", current: -5, total: 100, want: "[----------] 0%"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if got := Body(testCase.current, testCase.total, 10, 3); got != testCase.want {
				t.Fatalf("got %q, want %q", got, testCase.want)
			}
		})
	}
}

func TestBodyCountsBytesWhenTheTotalIsUnknown(t *testing.T) {
	// A chunked response reports no length, so there is nothing to divide by.
	got := Body(2048, -1, 10, 3)
	if !strings.Contains(got, "2048") {
		t.Fatalf("the byte count is missing: %q", got)
	}
	if strings.Contains(got, "%") {
		t.Fatalf("a percentage was shown without a total: %q", got)
	}
}

func TestBodyLocalizedTranslatesTheByteCount(t *testing.T) {
	russian, err := locale.New("ru", nil)
	if err != nil {
		t.Fatal(err)
	}
	english := BodyLocalized(2048, 0, 10, 3, locale.English())
	translated := BodyLocalized(2048, 0, 10, 3, russian)
	if english == translated {
		t.Fatalf("the byte count was not translated: %q", translated)
	}
	if !strings.Contains(translated, "2048") {
		t.Fatalf("the byte count is missing: %q", translated)
	}
}

func TestIndeterminateMovesAndStaysInsideTheBar(t *testing.T) {
	const width, pulse = 12, 4
	seen := map[string]struct{}{}
	for frame := range 64 {
		bar := Indeterminate(frame, width, pulse)
		if len(bar) != width+2 {
			t.Fatalf("frame %d has width %d, want %d: %q", frame, len(bar)-2, width, bar)
		}
		if !strings.HasPrefix(bar, "[") || !strings.HasSuffix(bar, "]") {
			t.Fatalf("frame %d is not bracketed: %q", frame, bar)
		}
		if filled := strings.Count(bar, "=") + strings.Count(bar, ">") + strings.Count(bar, "<"); filled != pulse {
			t.Fatalf("frame %d shows %d filled cells, want %d: %q", frame, filled, pulse, bar)
		}
		seen[bar] = struct{}{}
	}
	if len(seen) < 2 {
		t.Fatal("the bar never moved")
	}
}

func TestIndeterminateHandlesDegenerateSizes(t *testing.T) {
	if got := Indeterminate(3, 0, 3); got != "[]" {
		t.Errorf("a bar with no width: %q", got)
	}
	// A pulse at least as wide as the bar fills it, leaving nothing to move.
	if got := Indeterminate(3, 4, 9); got != "[====]" {
		t.Errorf("a pulse wider than the bar: %q", got)
	}
	if got := Indeterminate(-7, 12, 4); len(got) != 14 {
		t.Errorf("a negative frame: %q", got)
	}
}
