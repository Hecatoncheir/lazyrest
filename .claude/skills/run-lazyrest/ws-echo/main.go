// Command ws-echo answers a WebSocket client so the stream pane can be driven
// without depending on an external service. It echoes what it is sent and
// pushes a tick while the client is idle, so the pane has live traffic to show.
//
// It uses the same WebSocket library lazyrest does, so it needs no module of
// its own.
package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/coder/websocket"
)

const address = "127.0.0.1:7799"

func main() {
	handler := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		conn, err := websocket.Accept(writer, request, &websocket.AcceptOptions{InsecureSkipVerify: true})
		if err != nil {
			return
		}
		defer func() { _ = conn.CloseNow() }()
		serve(conn)
	})
	fmt.Println("ws-echo listening on ws://" + address)
	_ = http.ListenAndServe(address, handler) //nolint:gosec // a local test server
}

func serve(conn *websocket.Conn) {
	ctx := context.Background()
	go tick(ctx, conn)

	for {
		_, payload, err := conn.Read(ctx)
		if err != nil {
			return
		}
		if err := conn.Write(ctx, websocket.MessageText, []byte(`{"echo":`+string(payload)+`}`)); err != nil {
			return
		}
	}
}

func tick(ctx context.Context, conn *websocket.Conn) {
	for count := 1; ; count++ {
		time.Sleep(900 * time.Millisecond)
		message := fmt.Sprintf(`{"tick":%d}`, count)
		if err := conn.Write(ctx, websocket.MessageText, []byte(message)); err != nil {
			return
		}
	}
}
