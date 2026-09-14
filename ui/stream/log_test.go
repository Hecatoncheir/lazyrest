package stream

import (
	"strconv"
	"sync"
	"testing"

	runnerstream "github.com/Hecatoncheir/lazyrest/runner/stream"
)

func frameOf(payload string) runnerstream.Frame {
	return runnerstream.Frame{Payload: []byte(payload)}
}

func payloads(frames []runnerstream.Frame) []string {
	out := make([]string, len(frames))
	for index, frame := range frames {
		out[index] = string(frame.Payload)
	}
	return out
}

func TestLogKeepsTheNewestFramesWithinTheCountBound(t *testing.T) {
	log := NewLog(3, 1<<20)
	for _, payload := range []string{"a", "b", "c", "d", "e"} {
		log.Append(frameOf(payload))
	}

	if log.Len() != 3 {
		t.Fatalf("length = %d, want 3", log.Len())
	}
	got := payloads(log.Tail(0))
	want := []string{"c", "d", "e"}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("tail = %v, want %v", got, want)
		}
	}
}

func TestLogEvictsOnTheByteBound(t *testing.T) {
	log := NewLog(100, 10)
	log.Append(frameOf("12345"))
	log.Append(frameOf("67890"))
	log.Append(frameOf("abcde"))

	if log.Len() != 2 {
		t.Fatalf("length = %d, want 2; the byte bound did not evict", log.Len())
	}
	if log.Bytes() > 10 {
		t.Fatalf("bytes = %d, want at most 10", log.Bytes())
	}
}

// A frame that alone exceeds the bound is still kept: emptying the log on
// arrival would be worse than overshooting by one frame.
func TestLogKeepsAFrameLargerThanTheByteBound(t *testing.T) {
	log := NewLog(100, 4)
	log.Append(frameOf("0123456789"))

	if log.Len() != 1 {
		t.Fatalf("length = %d, want 1", log.Len())
	}
	if got := payloads(log.Tail(0))[0]; got != "0123456789" {
		t.Fatalf("payload = %q", got)
	}
}

func TestLogCountsDroppedFrames(t *testing.T) {
	log := NewLog(2, 1<<20)
	for index := range 5 {
		log.Append(frameOf(strconv.Itoa(index)))
	}
	if log.Dropped() != 3 {
		t.Fatalf("dropped = %d, want 3", log.Dropped())
	}
}

func TestLogSequenceChangesOnlyOnAppend(t *testing.T) {
	log := NewLog(10, 1<<20)
	start := log.Sequence()

	if log.Sequence() != start || log.Len() != 0 {
		t.Fatal("reading the log moved the sequence")
	}
	log.Append(frameOf("a"))
	if log.Sequence() == start {
		t.Fatal("appending did not move the sequence")
	}

	afterAppend := log.Sequence()
	_ = log.Tail(0)
	_ = log.Dropped()
	if log.Sequence() != afterAppend {
		t.Fatal("reading the log moved the sequence")
	}
}

func TestLogTailReturnsTheNewestFramesOldestFirst(t *testing.T) {
	log := NewLog(10, 1<<20)
	for _, payload := range []string{"a", "b", "c", "d"} {
		log.Append(frameOf(payload))
	}
	got := payloads(log.Tail(2))
	if len(got) != 2 || got[0] != "c" || got[1] != "d" {
		t.Fatalf("tail(2) = %v, want [c d]", got)
	}
}

func TestLogClearResetsTheBounds(t *testing.T) {
	log := NewLog(4, 1<<20)
	for _, payload := range []string{"a", "b", "c", "d", "e"} {
		log.Append(frameOf(payload))
	}
	log.Clear()

	if log.Len() != 0 || log.Bytes() != 0 || log.Dropped() != 0 {
		t.Fatalf("after clear: length %d, bytes %d, dropped %d", log.Len(), log.Bytes(), log.Dropped())
	}
	if len(log.Tail(0)) != 0 {
		t.Fatal("tail is not empty after clear")
	}
}

func TestLogIsSafeForConcurrentUse(t *testing.T) {
	log := NewLog(64, 1<<20)
	var waiting sync.WaitGroup

	for writer := range 4 {
		waiting.Add(1)
		go func() {
			defer waiting.Done()
			for index := range 200 {
				log.Append(frameOf(strconv.Itoa(writer*1000 + index)))
			}
		}()
	}
	for range 4 {
		waiting.Add(1)
		go func() {
			defer waiting.Done()
			for range 200 {
				_ = log.Tail(16)
				_ = log.Sequence()
				_ = log.Len()
			}
		}()
	}
	waiting.Wait()

	if log.Len() != 64 {
		t.Fatalf("length = %d, want the bound of 64", log.Len())
	}
}
