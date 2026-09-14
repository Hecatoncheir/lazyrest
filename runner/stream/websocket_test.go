package stream

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
)

func startWebSocketServer(t *testing.T, handle func(conn *websocket.Conn, request *http.Request)) string {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		conn, err := websocket.Accept(writer, request, &websocket.AcceptOptions{InsecureSkipVerify: true})
		if err != nil {
			return
		}
		defer func() { _ = conn.CloseNow() }()
		handle(conn, request)
	}))
	t.Cleanup(server.Close)
	return "ws" + strings.TrimPrefix(server.URL, "http")
}

func TestWebSocketSessionReadsATextMessage(t *testing.T) {
	url := startWebSocketServer(t, func(conn *websocket.Conn, _ *http.Request) {
		_ = conn.Write(context.Background(), websocket.MessageText, []byte(`{"ok":true}`))
		time.Sleep(500 * time.Millisecond)
	})

	session, err := DialWebSocket(context.Background(), url, WebSocketConfig{})
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { _ = session.Close() })

	frame := nextFrame(t, session.Frames())
	if frame.Opcode != Text {
		t.Fatalf("opcode = %v, want text", frame.Opcode)
	}
	if string(frame.Payload) != `{"ok":true}` {
		t.Fatalf("payload = %q", frame.Payload)
	}
}

func TestWebSocketSessionRecordsWhatItSent(t *testing.T) {
	received := make(chan string, 1)
	url := startWebSocketServer(t, func(conn *websocket.Conn, _ *http.Request) {
		_, payload, err := conn.Read(context.Background())
		if err != nil {
			return
		}
		received <- string(payload)
		time.Sleep(500 * time.Millisecond)
	})

	session, err := DialWebSocket(context.Background(), url, WebSocketConfig{})
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { _ = session.Close() })

	if err := session.Send(context.Background(), Frame{Opcode: Text, Payload: []byte("subscribe")}); err != nil {
		t.Fatalf("send: %v", err)
	}
	frame := nextFrame(t, session.Frames())
	if frame.Direction != Sent || string(frame.Payload) != "subscribe" {
		t.Fatalf("logged frame = %v %q", frame.Direction, frame.Payload)
	}
	select {
	case got := <-received:
		if got != "subscribe" {
			t.Fatalf("server read %q", got)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("the server never received the message")
	}
}

func TestWebSocketSessionTruncatesAtTheFrameLimit(t *testing.T) {
	url := startWebSocketServer(t, func(conn *websocket.Conn, _ *http.Request) {
		_ = conn.Write(context.Background(), websocket.MessageText, []byte("0123456789"))
		time.Sleep(500 * time.Millisecond)
	})

	session, err := DialWebSocket(context.Background(), url, WebSocketConfig{MaxFrameBytes: 4})
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { _ = session.Close() })

	frame := nextFrame(t, session.Frames())
	if !frame.Truncated || len(frame.Payload) != 4 {
		t.Fatalf("truncated = %v, length = %d, want true and 4", frame.Truncated, len(frame.Payload))
	}
}

// The handshake has to carry what the request file declared, otherwise keeping
// WebSocket in .http buys nothing.
func TestWebSocketSessionCarriesHandshakeHeaders(t *testing.T) {
	authorization := make(chan string, 1)
	url := startWebSocketServer(t, func(conn *websocket.Conn, request *http.Request) {
		authorization <- request.Header.Get("Authorization")
		time.Sleep(300 * time.Millisecond)
	})

	header := http.Header{}
	header.Set("Authorization", "Bearer token-from-an-earlier-request")

	session, err := DialWebSocket(context.Background(), url, WebSocketConfig{Header: header})
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { _ = session.Close() })

	select {
	case got := <-authorization:
		if got != "Bearer token-from-an-earlier-request" {
			t.Fatalf("Authorization = %q", got)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("the handshake never reached the server")
	}
}

func TestWebSocketSessionLogsAPingFromTheServer(t *testing.T) {
	url := startWebSocketServer(t, func(conn *websocket.Conn, _ *http.Request) {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = conn.Ping(ctx)
		time.Sleep(300 * time.Millisecond)
	})

	session, err := DialWebSocket(context.Background(), url, WebSocketConfig{})
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { _ = session.Close() })

	frame := nextFrame(t, session.Frames())
	if frame.Opcode != Ping {
		t.Fatalf("opcode = %v, want ping; control frames are not reaching the log", frame.Opcode)
	}
}

func TestWebSocketSessionReportsACleanClose(t *testing.T) {
	url := startWebSocketServer(t, func(conn *websocket.Conn, _ *http.Request) {
		_ = conn.Close(websocket.StatusNormalClosure, "done")
	})

	session, err := DialWebSocket(context.Background(), url, WebSocketConfig{})
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { _ = session.Close() })

	frame := nextFrame(t, session.Frames())
	if frame.Opcode != Close {
		t.Fatalf("opcode = %v, want close", frame.Opcode)
	}
	select {
	case _, ok := <-session.Frames():
		if ok {
			t.Fatal("an unexpected frame followed the close")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("frames was not closed")
	}
	if err := session.Err(); err != nil {
		t.Fatalf("a normal closure reported an error: %v", err)
	}
}

// The same invariant the TCP session has: the dial bound must not reach the
// session. DialWebSocket cancels its dial context on return, so a connection
// tied to it would die immediately.
func TestWebSocketSessionOutlivesTheDialTimeout(t *testing.T) {
	url := startWebSocketServer(t, func(conn *websocket.Conn, _ *http.Request) {
		time.Sleep(300 * time.Millisecond)
		_ = conn.Write(context.Background(), websocket.MessageText, []byte("late"))
		time.Sleep(500 * time.Millisecond)
	})

	session, err := DialWebSocket(context.Background(), url, WebSocketConfig{DialTimeout: 100 * time.Millisecond})
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { _ = session.Close() })

	if got := string(nextFrame(t, session.Frames()).Payload); got != "late" {
		t.Fatalf("payload = %q, want %q", got, "late")
	}
}

func TestDialWebSocketFailsOnAPlainHTTPEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)

	url := "ws" + strings.TrimPrefix(server.URL, "http")
	if _, err := DialWebSocket(context.Background(), url, WebSocketConfig{DialTimeout: 2 * time.Second}); err == nil {
		t.Fatal("dialling an endpoint that does not upgrade returned no error")
	}
}
