// Package stream keeps what a live connection has said and decides how often
// that is drawn.
//
// Frames deliberately never reach ui.Model. A Model snapshot deep copies every
// suite in the file, and refreshStatus takes one on each call; routing a busy
// socket through it would clone the whole request list per frame. The log below
// carries its own lock instead, and the Model holds only the connection
// summary, which changes rarely.
package stream

import (
	"sync"

	runnerstream "github.com/Hecatoncheir/lazyrest/runner/stream"
)

const (
	DefaultMaxFrames = 2000
	DefaultMaxBytes  = 4 << 20
)

// Log is a bounded ring of frames. It is bounded twice: by count, so a chatty
// connection cannot grow it without limit, and by total bytes, so a few large
// frames cannot either.
type Log struct {
	mutex    sync.RWMutex
	buffer   []runnerstream.Frame
	start    int
	count    int
	bytes    int
	maxBytes int
	dropped  uint64
	sequence uint64
}

func NewLog(maxFrames, maxBytes int) *Log {
	if maxFrames <= 0 {
		maxFrames = DefaultMaxFrames
	}
	if maxBytes <= 0 {
		maxBytes = DefaultMaxBytes
	}
	return &Log{buffer: make([]runnerstream.Frame, maxFrames), maxBytes: maxBytes}
}

func (log *Log) Append(frame runnerstream.Frame) {
	log.mutex.Lock()
	defer log.mutex.Unlock()

	if log.count == len(log.buffer) {
		log.evictOldest()
	}
	log.buffer[(log.start+log.count)%len(log.buffer)] = frame
	log.count++
	log.bytes += len(frame.Payload)

	// The newest frame is always kept, even when it alone exceeds the byte
	// bound: a log that answered an arrival by emptying itself would be worse
	// than one that overshoots by a single frame.
	for log.bytes > log.maxBytes && log.count > 1 {
		log.evictOldest()
	}
	log.sequence++
}

func (log *Log) evictOldest() {
	oldest := log.buffer[log.start]
	log.bytes -= len(oldest.Payload)
	log.buffer[log.start] = runnerstream.Frame{}
	log.start = (log.start + 1) % len(log.buffer)
	log.count--
	log.dropped++
}

// Sequence changes only when a frame is appended, so a renderer can tell that
// nothing has happened without copying anything.
func (log *Log) Sequence() uint64 {
	log.mutex.RLock()
	defer log.mutex.RUnlock()
	return log.sequence
}

// Dropped is how many frames the bounds discarded. The UI reports it rather
// than pretending the log is complete.
func (log *Log) Dropped() uint64 {
	log.mutex.RLock()
	defer log.mutex.RUnlock()
	return log.dropped
}

func (log *Log) Len() int {
	log.mutex.RLock()
	defer log.mutex.RUnlock()
	return log.count
}

func (log *Log) Bytes() int {
	log.mutex.RLock()
	defer log.mutex.RUnlock()
	return log.bytes
}

// Tail returns the newest count frames, oldest first. A count of zero or less
// returns everything held.
func (log *Log) Tail(count int) []runnerstream.Frame {
	log.mutex.RLock()
	defer log.mutex.RUnlock()

	if count <= 0 || count > log.count {
		count = log.count
	}
	frames := make([]runnerstream.Frame, count)
	first := log.count - count
	for index := range frames {
		frames[index] = log.buffer[(log.start+first+index)%len(log.buffer)]
	}
	return frames
}

func (log *Log) Clear() {
	log.mutex.Lock()
	defer log.mutex.Unlock()
	for index := range log.buffer {
		log.buffer[index] = runnerstream.Frame{}
	}
	log.start = 0
	log.count = 0
	log.bytes = 0
	log.dropped = 0
	log.sequence++
}
