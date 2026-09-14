package stream

import (
	"context"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	runnerstream "github.com/Hecatoncheir/lazyrest/runner/stream"
)

// fakeSession stands in for a real connection. That the interface substitutes
// this cleanly is what will let the UI tests drive a stream without a server.
type fakeSession struct {
	frames chan runnerstream.Frame
}

func newFakeSession(buffer int) *fakeSession {
	return &fakeSession{frames: make(chan runnerstream.Frame, buffer)}
}

func (session *fakeSession) Frames() <-chan runnerstream.Frame { return session.frames }

func (session *fakeSession) Send(context.Context, runnerstream.Frame) error { return nil }

func (session *fakeSession) Close() error { return nil }

func (session *fakeSession) Err() error { return nil }

// The point of the whole design: frames arrive far faster than a terminal can
// paint, so redraws must be coalesced rather than queued one per frame.
func TestFollowCoalescesManyFramesIntoFewRedraws(t *testing.T) {
	const frameCount = 500

	session := newFakeSession(frameCount)
	for index := range frameCount {
		session.frames <- frameOf(strconv.Itoa(index))
	}
	close(session.frames)

	log := NewLog(frameCount, 1<<20)
	var redraws atomic.Int64

	Follow(context.Background(), session, log, 50*time.Millisecond, func() {
		redraws.Add(1)
	})

	if log.Len() != frameCount {
		t.Fatalf("log holds %d frames, want %d", log.Len(), frameCount)
	}
	drawn := redraws.Load()
	t.Logf("%d redraws for %d frames", drawn, frameCount)
	if drawn < 1 {
		t.Fatal("the arrived frames were never painted")
	}
	if drawn > 10 {
		t.Fatalf("%d redraws for %d frames; they are not being coalesced", drawn, frameCount)
	}
}

// Nothing arriving must cost nothing. This is what Sequence buys.
func TestFollowDoesNotRedrawWhileIdle(t *testing.T) {
	session := newFakeSession(1)
	log := NewLog(16, 1<<20)
	var redraws atomic.Int64

	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()

	Follow(ctx, session, log, 10*time.Millisecond, func() { redraws.Add(1) })

	if drawn := redraws.Load(); drawn != 0 {
		t.Fatalf("%d redraws while idle, want 0", drawn)
	}
}

func TestFollowPaintsTheFinalFramesWhenTheConnectionEnds(t *testing.T) {
	session := newFakeSession(2)
	session.frames <- frameOf("last")
	close(session.frames)

	log := NewLog(16, 1<<20)
	var redraws atomic.Int64

	// An interval far longer than the test would never tick, so a paint here
	// can only come from the close path.
	Follow(context.Background(), session, log, time.Hour, func() { redraws.Add(1) })

	if redraws.Load() != 1 {
		t.Fatalf("redraws = %d, want exactly 1 on close", redraws.Load())
	}
	if log.Len() != 1 {
		t.Fatalf("log holds %d frames, want 1", log.Len())
	}
}

func TestFollowStopsWhenTheContextIsDone(t *testing.T) {
	session := newFakeSession(1)
	log := NewLog(16, 1<<20)

	ctx, cancel := context.WithCancel(context.Background())
	finished := make(chan struct{})
	go func() {
		defer close(finished)
		Follow(ctx, session, log, 10*time.Millisecond, nil)
	}()

	cancel()
	select {
	case <-finished:
	case <-time.After(5 * time.Second):
		t.Fatal("Follow did not return after its context was cancelled")
	}
}
