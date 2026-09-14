package stream

import (
	"context"
	"errors"
	"io"
	"net"
	"strings"
	"sync"
	"time"
)

const (
	DefaultDialTimeout = 10 * time.Second
	// DefaultFlushIdle is how long the reader waits for more bytes before it
	// calls what it has one frame. TCP carries no message boundary, so one
	// logical reply can arrive as several segments; without this a reply would
	// be shown split at whatever offsets the network happened to choose.
	DefaultFlushIdle = 50 * time.Millisecond
	// DefaultMaxFrameBytes bounds a single frame, as --max-response-bytes
	// bounds a response body.
	DefaultMaxFrameBytes = 64 * 1024
)

type TCPConfig struct {
	DialTimeout   time.Duration
	FlushIdle     time.Duration
	MaxFrameBytes int
}

func (config TCPConfig) withDefaults() TCPConfig {
	if config.DialTimeout <= 0 {
		config.DialTimeout = DefaultDialTimeout
	}
	if config.FlushIdle <= 0 {
		config.FlushIdle = DefaultFlushIdle
	}
	if config.MaxFrameBytes <= 0 {
		config.MaxFrameBytes = DefaultMaxFrameBytes
	}
	return config
}

type TCPSession struct {
	conn   net.Conn
	config TCPConfig

	frames       chan Frame
	done         chan struct{}
	emitMutex    sync.Mutex
	framesClosed bool

	closeOnce  sync.Once
	closeError error

	errorMutex sync.Mutex
	err        error
}

// DialTCP opens a raw socket. The address may carry a tcp:// scheme, which is
// how a .socket file writes it, or be a bare host:port.
//
// Only the dial observes ctx and the configured timeout. The session outlives
// both, so a cancelled ctx after this returns does not close the connection;
// call Close for that.
func DialTCP(ctx context.Context, address string, config TCPConfig) (*TCPSession, error) {
	config = config.withDefaults()
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, config.DialTimeout)
	defer cancel()

	dialer := net.Dialer{}
	conn, err := dialer.DialContext(ctx, "tcp", strings.TrimPrefix(address, "tcp://"))
	if err != nil {
		return nil, err
	}

	session := &TCPSession{
		conn:   conn,
		config: config,
		frames: make(chan Frame, 64),
		done:   make(chan struct{}),
	}
	go session.readLoop()
	return session, nil
}

func (session *TCPSession) Frames() <-chan Frame { return session.frames }

func (session *TCPSession) Err() error {
	session.errorMutex.Lock()
	defer session.errorMutex.Unlock()
	return session.err
}

func (session *TCPSession) Send(ctx context.Context, frame Frame) error {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return err
		}
	}
	if _, err := session.conn.Write(frame.Payload); err != nil {
		session.setErr(err)
		return err
	}
	session.emit(Frame{
		At:        time.Now(),
		Direction: Sent,
		Opcode:    Bytes,
		Payload:   append([]byte(nil), frame.Payload...),
	})
	return nil
}

func (session *TCPSession) Close() error {
	session.closeOnce.Do(func() {
		close(session.done)
		session.closeError = session.conn.Close()
	})
	return session.closeError
}

func (session *TCPSession) setErr(err error) {
	session.errorMutex.Lock()
	defer session.errorMutex.Unlock()
	if session.err == nil {
		session.err = err
	}
}

func (session *TCPSession) emit(frame Frame) {
	session.emitMutex.Lock()
	defer session.emitMutex.Unlock()
	if session.framesClosed {
		return
	}
	select {
	case session.frames <- frame:
	case <-session.done:
	}
}

func (session *TCPSession) closeFrames() {
	session.emitMutex.Lock()
	defer session.emitMutex.Unlock()
	if !session.framesClosed {
		session.framesClosed = true
		close(session.frames)
	}
}

func (session *TCPSession) readLoop() {
	defer session.closeFrames()

	buffer := make([]byte, 32*1024)
	var pending []byte

	flush := func() {
		if len(pending) == 0 {
			return
		}
		payload := pending
		truncated := false
		if len(payload) > session.config.MaxFrameBytes {
			payload = payload[:session.config.MaxFrameBytes]
			truncated = true
		}
		session.emit(Frame{
			At:        time.Now(),
			Direction: Received,
			Opcode:    Bytes,
			Payload:   payload,
			Truncated: truncated,
		})
		pending = nil
	}

	for {
		// With nothing buffered there is nothing to flush, so the read blocks
		// until the peer speaks. With bytes in hand it waits only for the idle
		// window, and the timeout that follows is the message boundary.
		deadline := time.Time{}
		if len(pending) > 0 {
			deadline = time.Now().Add(session.config.FlushIdle)
		}
		if err := session.conn.SetReadDeadline(deadline); err != nil {
			flush()
			session.setErr(err)
			return
		}

		read, err := session.conn.Read(buffer)
		if read > 0 {
			pending = append(pending, buffer[:read]...)
			if len(pending) >= session.config.MaxFrameBytes {
				flush()
			}
			continue
		}
		if err == nil {
			continue
		}

		var netError net.Error
		if errors.As(err, &netError) && netError.Timeout() {
			flush()
			continue
		}
		flush()
		if !errors.Is(err, io.EOF) && !errors.Is(err, net.ErrClosed) {
			session.setErr(err)
		}
		return
	}
}
