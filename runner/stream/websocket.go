package stream

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/coder/websocket"
)

const (
	// DefaultReadLimitBytes is the protocol bound: a message larger than this
	// fails the connection, because the library cannot hand back half a
	// message. It is deliberately far above DefaultMaxFrameBytes, which only
	// bounds what is kept for display. A debugging tool should show that an
	// oversized message arrived, not drop the connection at the display bound.
	DefaultReadLimitBytes int64 = 1 << 20
)

type WebSocketConfig struct {
	DialTimeout   time.Duration
	MaxFrameBytes int
	// ReadLimitBytes bounds what the protocol will accept at all.
	ReadLimitBytes int64
	// Header carries the handshake headers of the request that opened this
	// stream, so an Authorization built from {{variables}} reaches the server.
	Header       http.Header
	Subprotocols []string
	// HTTPClient shares the session cookie jar with ordinary requests. This is
	// the reason a WebSocket belongs in a .http file: it upgrades from the same
	// session as the login that preceded it.
	HTTPClient *http.Client
}

func (config WebSocketConfig) withDefaults() WebSocketConfig {
	if config.DialTimeout <= 0 {
		config.DialTimeout = DefaultDialTimeout
	}
	if config.MaxFrameBytes <= 0 {
		config.MaxFrameBytes = DefaultMaxFrameBytes
	}
	if config.ReadLimitBytes <= 0 {
		config.ReadLimitBytes = DefaultReadLimitBytes
	}
	return config
}

type WebSocketSession struct {
	conn   *websocket.Conn
	config WebSocketConfig

	readCtx    context.Context
	cancelRead context.CancelFunc

	frames       chan Frame
	done         chan struct{}
	emitMutex    sync.Mutex
	framesClosed bool

	closeOnce  sync.Once
	closeError error

	errorMutex sync.Mutex
	err        error
}

// DialWebSocket upgrades a ws:// or wss:// URL.
//
// Only the handshake observes ctx and the configured timeout. Reads that follow
// run on a context of the session's own, so a dial bound of a few seconds does
// not end a stream that stays open for hours.
func DialWebSocket(ctx context.Context, url string, config WebSocketConfig) (*WebSocketSession, error) {
	config = config.withDefaults()
	if ctx == nil {
		ctx = context.Background()
	}
	dialCtx, cancelDial := context.WithTimeout(ctx, config.DialTimeout)
	defer cancelDial()

	readCtx, cancelRead := context.WithCancel(context.Background())
	session := &WebSocketSession{
		config:     config,
		readCtx:    readCtx,
		cancelRead: cancelRead,
		frames:     make(chan Frame, 64),
		done:       make(chan struct{}),
	}

	options := &websocket.DialOptions{
		HTTPClient:   config.HTTPClient,
		HTTPHeader:   config.Header,
		Subprotocols: config.Subprotocols,
		// Control frames are what you look at when a connection dies quietly,
		// so they belong in the log rather than being handled invisibly. The
		// callbacks run synchronously on the read path, so neither may block.
		OnPingReceived: func(_ context.Context, payload []byte) bool {
			session.tryEmit(Frame{At: time.Now(), Direction: Received, Opcode: Ping, Payload: clone(payload)})
			return true
		},
		OnPongReceived: func(_ context.Context, payload []byte) {
			session.tryEmit(Frame{At: time.Now(), Direction: Received, Opcode: Pong, Payload: clone(payload)})
		},
	}

	conn, _, err := websocket.Dial(dialCtx, url, options)
	if err != nil {
		cancelRead()
		return nil, err
	}
	conn.SetReadLimit(config.ReadLimitBytes)
	session.conn = conn

	go session.readLoop()
	return session, nil
}

func clone(payload []byte) []byte { return append([]byte(nil), payload...) }

func (session *WebSocketSession) Frames() <-chan Frame { return session.frames }

func (session *WebSocketSession) Err() error {
	session.errorMutex.Lock()
	defer session.errorMutex.Unlock()
	return session.err
}

func (session *WebSocketSession) Send(ctx context.Context, frame Frame) error {
	if ctx == nil {
		ctx = context.Background()
	}
	messageType := websocket.MessageText
	if frame.Opcode == Binary {
		messageType = websocket.MessageBinary
	}
	if frame.Opcode == Ping {
		if err := session.conn.Ping(ctx); err != nil {
			session.setErr(err)
			return err
		}
		session.emit(Frame{At: time.Now(), Direction: Sent, Opcode: Ping})
		return nil
	}
	if err := session.conn.Write(ctx, messageType, frame.Payload); err != nil {
		session.setErr(err)
		return err
	}
	session.emit(Frame{
		At:        time.Now(),
		Direction: Sent,
		Opcode:    opcodeFor(messageType),
		Payload:   clone(frame.Payload),
	})
	return nil
}

func (session *WebSocketSession) Close() error {
	session.closeOnce.Do(func() {
		close(session.done)
		session.closeError = session.conn.Close(websocket.StatusNormalClosure, "")
		session.cancelRead()
	})
	return session.closeError
}

func opcodeFor(messageType websocket.MessageType) Opcode {
	if messageType == websocket.MessageBinary {
		return Binary
	}
	return Text
}

func (session *WebSocketSession) setErr(err error) {
	session.errorMutex.Lock()
	defer session.errorMutex.Unlock()
	if session.err == nil {
		session.err = err
	}
}

func (session *WebSocketSession) emit(frame Frame) {
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

// tryEmit never blocks. It backs the control frame callbacks, which the library
// runs synchronously on the read path: waiting there would stall the protocol,
// so a control frame is dropped rather than allowed to wedge the connection.
func (session *WebSocketSession) tryEmit(frame Frame) {
	session.emitMutex.Lock()
	defer session.emitMutex.Unlock()
	if session.framesClosed {
		return
	}
	select {
	case session.frames <- frame:
	default:
	}
}

func (session *WebSocketSession) closeFrames() {
	session.emitMutex.Lock()
	defer session.emitMutex.Unlock()
	if !session.framesClosed {
		session.framesClosed = true
		close(session.frames)
	}
}

func (session *WebSocketSession) readLoop() {
	defer session.closeFrames()

	for {
		messageType, payload, err := session.conn.Read(session.readCtx)
		if err != nil {
			session.recordTermination(err)
			return
		}
		truncated := false
		if len(payload) > session.config.MaxFrameBytes {
			payload = payload[:session.config.MaxFrameBytes]
			truncated = true
		}
		session.emit(Frame{
			At:        time.Now(),
			Direction: Received,
			Opcode:    opcodeFor(messageType),
			Payload:   payload,
			Truncated: truncated,
		})
	}
}

// recordTermination turns the read error into the last entry of the log. A
// close handshake is the normal end of a stream, not a failure, so it is
// reported as a Close frame and leaves Err nil.
func (session *WebSocketSession) recordTermination(err error) {
	status := websocket.CloseStatus(err)
	if status != -1 {
		session.emit(Frame{
			At:        time.Now(),
			Direction: Received,
			Opcode:    Close,
			Payload:   []byte(status.String()),
		})
		if status != websocket.StatusNormalClosure && status != websocket.StatusGoingAway {
			session.setErr(err)
		}
		return
	}
	select {
	case <-session.done:
		return
	default:
	}
	if !errors.Is(err, context.Canceled) {
		session.setErr(err)
	}
}
