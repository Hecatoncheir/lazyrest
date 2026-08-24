package producer

import (
	"strings"
	"testing"
)

func TestFormatIndeterminateProgressBarMoves(t *testing.T) {
	first := formatIndeterminateProgressBar(0)
	second := formatIndeterminateProgressBar(4)
	returning := formatIndeterminateProgressBar(progressBarWidth)

	if first == second || second == returning {
		t.Fatalf("progress pulse did not move: %q, %q, %q", first, second, returning)
	}
	for _, bar := range []string{first, second, returning} {
		if len(bar) != progressBarWidth+2 {
			t.Fatalf("unexpected progress bar width: %q", bar)
		}
		if !strings.ContainsAny(bar, "<>") {
			t.Fatalf("progress bar has no direction marker: %q", bar)
		}
	}
}

func TestFormatProgressBar(t *testing.T) {
	if got := formatProgressBar(5, 10); got != "[==========----------] 50%" {
		t.Fatalf("unexpected determinate progress: %q", got)
	}
	if got := formatProgressBar(15, 10); got != "[====================] 100%" {
		t.Fatalf("progress was not clamped: %q", got)
	}
	if got := formatProgressBar(2048, -1); !strings.Contains(got, "2048 bytes") || !strings.HasPrefix(got, "[") {
		t.Fatalf("unexpected unknown-length progress: %q", got)
	}
}
