// Command echo-server answers a raw TCP client so the stream pane can be
// driven without depending on an external service. It replies +PONG to a PING
// and echoes anything else, and sends a tick while the client is idle so the
// pane has live traffic to show.
package main

import (
	"fmt"
	"net"
	"strings"
	"time"
)

const address = "127.0.0.1:7788"

func main() {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		panic(err)
	}
	fmt.Println("listening on " + address)
	for {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		go serve(conn)
	}
}

func serve(conn net.Conn) {
	defer func() { _ = conn.Close() }()

	buffer := make([]byte, 4096)
	ticks := 0
	for {
		_ = conn.SetReadDeadline(time.Now().Add(700 * time.Millisecond))
		read, err := conn.Read(buffer)
		if read > 0 {
			command := strings.TrimSpace(string(buffer[:read]))
			if strings.HasPrefix(command, "PING") {
				_, _ = conn.Write([]byte("+PONG\r\n"))
			} else {
				_, _ = conn.Write([]byte(`{"echo":"` + command + `"}`))
			}
			continue
		}
		if err != nil {
			if netError, ok := err.(net.Error); !ok || !netError.Timeout() {
				return
			}
		}
		ticks++
		_, _ = conn.Write([]byte(fmt.Sprintf(`{"tick":%d,"at":"%s"}`, ticks, time.Now().Format("15:04:05"))))
	}
}
