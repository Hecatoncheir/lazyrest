package stream

import (
	"context"
	"net"
	"testing"
	"time"
)

// startServer accepts exactly one connection and hands it to handle.
func startServer(t *testing.T, handle func(net.Conn)) string {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	t.Cleanup(func() { _ = listener.Close() })

	accepted := make(chan struct{})
	go func() {
		defer close(accepted)
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer func() { _ = conn.Close() }()
		handle(conn)
	}()
	t.Cleanup(func() {
		select {
		case <-accepted:
		case <-time.After(2 * time.Second):
		}
	})
	return listener.Addr().String()
}

func nextFrame(t *testing.T, frames <-chan Frame) Frame {
	t.Helper()
	select {
	case frame, ok := <-frames:
		if !ok {
			t.Fatal("frames closed before a frame arrived")
		}
		return frame
	case <-time.After(5 * time.Second):
		t.Fatal("no frame arrived")
	}
	return Frame{}
}

func TestTCPSessionCoalescesABurstIntoOneFrame(t *testing.T) {
	address := startServer(t, func(conn net.Conn) {
		_, _ = conn.Write([]byte("he"))
		_, _ = conn.Write([]byte("llo"))
		time.Sleep(time.Second)
	})

	session, err := DialTCP(context.Background(), address, TCPConfig{FlushIdle: 300 * time.Millisecond})
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { _ = session.Close() })

	frame := nextFrame(t, session.Frames())
	if string(frame.Payload) != "hello" {
		t.Fatalf("payload = %q, want %q; the burst was not coalesced", frame.Payload, "hello")
	}
	if frame.Direction != Received {
		t.Fatalf("direction = %v, want Received", frame.Direction)
	}
}

func TestTCPSessionSplitsWhenThePeerPauses(t *testing.T) {
	address := startServer(t, func(conn net.Conn) {
		_, _ = conn.Write([]byte("first"))
		time.Sleep(400 * time.Millisecond)
		_, _ = conn.Write([]byte("second"))
		time.Sleep(time.Second)
	})

	session, err := DialTCP(context.Background(), address, TCPConfig{FlushIdle: 20 * time.Millisecond})
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { _ = session.Close() })

	if got := string(nextFrame(t, session.Frames()).Payload); got != "first" {
		t.Fatalf("first payload = %q, want %q", got, "first")
	}
	if got := string(nextFrame(t, session.Frames()).Payload); got != "second" {
		t.Fatalf("second payload = %q, want %q", got, "second")
	}
}

func TestTCPSessionTruncatesAtTheFrameLimit(t *testing.T) {
	address := startServer(t, func(conn net.Conn) {
		_, _ = conn.Write([]byte("0123456789"))
		time.Sleep(time.Second)
	})

	session, err := DialTCP(context.Background(), address, TCPConfig{MaxFrameBytes: 4})
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { _ = session.Close() })

	frame := nextFrame(t, session.Frames())
	if !frame.Truncated {
		t.Fatal("frame is not marked truncated")
	}
	if len(frame.Payload) != 4 {
		t.Fatalf("payload length = %d, want 4", len(frame.Payload))
	}
}

func TestTCPSessionRecordsWhatItSent(t *testing.T) {
	received := make(chan string, 1)
	address := startServer(t, func(conn net.Conn) {
		buffer := make([]byte, 16)
		read, err := conn.Read(buffer)
		if err != nil {
			return
		}
		received <- string(buffer[:read])
		time.Sleep(500 * time.Millisecond)
	})

	session, err := DialTCP(context.Background(), address, TCPConfig{})
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { _ = session.Close() })

	if err := session.Send(context.Background(), Frame{Payload: []byte("PING")}); err != nil {
		t.Fatalf("send: %v", err)
	}

	frame := nextFrame(t, session.Frames())
	if frame.Direction != Sent {
		t.Fatalf("direction = %v, want Sent", frame.Direction)
	}
	if string(frame.Payload) != "PING" {
		t.Fatalf("payload = %q, want %q", frame.Payload, "PING")
	}

	select {
	case got := <-received:
		if got != "PING" {
			t.Fatalf("server read %q, want %q", got, "PING")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("the server never received the frame")
	}
}

func TestTCPSessionClosesFramesWhenThePeerHangsUp(t *testing.T) {
	address := startServer(t, func(conn net.Conn) {
		_, _ = conn.Write([]byte("bye"))
	})

	session, err := DialTCP(context.Background(), address, TCPConfig{FlushIdle: 20 * time.Millisecond})
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { _ = session.Close() })

	if got := string(nextFrame(t, session.Frames()).Payload); got != "bye" {
		t.Fatalf("payload = %q, want %q", got, "bye")
	}
	select {
	case _, ok := <-session.Frames():
		if ok {
			t.Fatal("an unexpected frame arrived after the peer hung up")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("frames was not closed after the peer hung up")
	}
	if err := session.Err(); err != nil {
		t.Fatalf("a clean hang up reported an error: %v", err)
	}
}

// The dial timeout must bound only the dial. This is the whole reason a stream
// is not a runner.Runner, which times out its entire execution.
func TestTCPSessionOutlivesTheDialTimeout(t *testing.T) {
	address := startServer(t, func(conn net.Conn) {
		time.Sleep(300 * time.Millisecond)
		_, _ = conn.Write([]byte("late"))
		time.Sleep(time.Second)
	})

	session, err := DialTCP(context.Background(), address, TCPConfig{
		DialTimeout: 50 * time.Millisecond,
		FlushIdle:   20 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { _ = session.Close() })

	if got := string(nextFrame(t, session.Frames()).Payload); got != "late" {
		t.Fatalf("payload = %q, want %q", got, "late")
	}
}

func TestDialTCPAcceptsATcpScheme(t *testing.T) {
	address := startServer(t, func(conn net.Conn) {
		time.Sleep(200 * time.Millisecond)
	})

	session, err := DialTCP(context.Background(), "tcp://"+address, TCPConfig{})
	if err != nil {
		t.Fatalf("dial with a tcp:// scheme: %v", err)
	}
	_ = session.Close()
}

func TestDialTCPFailsOnAClosedPort(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	address := listener.Addr().String()
	_ = listener.Close()

	if _, err := DialTCP(context.Background(), address, TCPConfig{DialTimeout: time.Second}); err == nil {
		t.Fatal("dialling a closed port returned no error")
	}
}
