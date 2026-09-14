package stream

import (
	"context"
	"time"

	runnerstream "github.com/Hecatoncheir/lazyrest/runner/stream"
)

// DefaultRedrawInterval is how often a live connection may repaint. A socket
// can deliver thousands of frames a second; a terminal cannot, and queueing a
// draw per frame would back up the tview event loop. Producer already takes
// this approach for its progress animation.
const DefaultRedrawInterval = 60 * time.Millisecond

// Follow drains session into log until the connection ends or ctx is done,
// calling onChange at most once per interval and only when frames actually
// arrived. It blocks; run it on a goroutine.
//
// onChange runs on that goroutine, so a caller driving tview must do its work
// inside QueueUpdateDraw.
func Follow(
	ctx context.Context,
	session runnerstream.Session,
	log *Log,
	interval time.Duration,
	onChange func(),
) {
	if interval <= 0 {
		interval = DefaultRedrawInterval
	}
	if ctx == nil {
		ctx = context.Background()
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	frames := session.Frames()
	var drawn uint64

	redraw := func() {
		if sequence := log.Sequence(); sequence != drawn {
			drawn = sequence
			if onChange != nil {
				onChange()
			}
		}
	}

	for {
		select {
		case <-ctx.Done():
			redraw()
			return
		case frame, open := <-frames:
			if !open {
				// The last frames still deserve one paint.
				redraw()
				return
			}
			log.Append(frame)
		case <-ticker.C:
			redraw()
		}
	}
}
