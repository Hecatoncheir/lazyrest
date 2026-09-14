// Package stream runs the long lived connections a request file can open: a
// WebSocket upgraded from HTTP, or a raw TCP socket.
//
// A stream differs from runner.Runner in the one way that matters. A Runner
// bounds its whole execution with the configured timeout and returns a single
// terminal Response. A stream has no terminal response, so only the dial is
// bounded; the session that follows lasts until one side closes it.
package stream

import (
	"context"
	"time"
)

// Direction records which side of the connection a frame came from.
type Direction uint8

const (
	Sent Direction = iota
	Received
)

func (direction Direction) String() string {
	if direction == Sent {
		return "sent"
	}
	return "received"
}

// Opcode names what a frame carries. A WebSocket declares this; a raw socket
// does not, so every frame it produces is Bytes.
type Opcode uint8

const (
	Bytes Opcode = iota
	Text
	Binary
	Ping
	Pong
	Close
)

func (opcode Opcode) String() string {
	switch opcode {
	case Text:
		return "text"
	case Binary:
		return "binary"
	case Ping:
		return "ping"
	case Pong:
		return "pong"
	case Close:
		return "close"
	default:
		return "bytes"
	}
}

// Frame is one message in either direction. Payload is raw: redaction and
// escaping belong to whoever renders or exports it, never to the transport.
type Frame struct {
	At        time.Time
	Direction Direction
	Opcode    Opcode
	Payload   []byte
	// Truncated marks a frame cut short by the configured byte limit, the way
	// runner.Response.Truncated marks a bounded body.
	Truncated bool
}

// Session is what the UI drives. Frames is closed when the connection ends,
// after which Err reports why if it was not a clean close.
type Session interface {
	Frames() <-chan Frame
	Send(ctx context.Context, frame Frame) error
	Close() error
	Err() error
}
