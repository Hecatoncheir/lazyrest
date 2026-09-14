package http

import (
	"path/filepath"
	"strings"
)

// SocketFileExtension marks a file of raw socket requests. WebSocket stays in
// .http because it upgrades from an HTTP session and wants the same headers,
// cookies and captured responses. A raw socket shares none of that, so it gets
// a file of its own.
const SocketFileExtension = ".socket"

// Transport names how a request reaches its peer. Everything but TransportHTTP
// is a long lived connection, which runner.Runner cannot execute: it bounds a
// whole run with one timeout and returns a single terminal response.
type Transport uint8

const (
	TransportHTTP Transport = iota
	TransportWebSocket
	TransportTCP
	TransportMQTT
)

func (transport Transport) String() string {
	switch transport {
	case TransportWebSocket:
		return "websocket"
	case TransportTCP:
		return "tcp"
	case TransportMQTT:
		return "mqtt"
	default:
		return "http"
	}
}

// IsStream reports whether the request opens a connection that outlives a
// single request and response.
func (transport Transport) IsStream() bool {
	return transport != TransportHTTP
}

// TransportForURI derives the transport from the scheme alone, the way the
// JetBrains HTTP client does: ws:// and wss:// are WebSocket, tcp:// is a raw
// socket, and everything else is an ordinary request.
func TransportForURI(uri string) Transport {
	trimmed := strings.TrimSpace(uri)
	switch {
	case hasScheme(trimmed, "ws"), hasScheme(trimmed, "wss"):
		return TransportWebSocket
	case hasScheme(trimmed, "tcp"):
		return TransportTCP
	case hasScheme(trimmed, "mqtt"), hasScheme(trimmed, "mqtts"):
		return TransportMQTT
	default:
		return TransportHTTP
	}
}

func hasScheme(uri, scheme string) bool {
	prefix := scheme + "://"
	return len(uri) >= len(prefix) && strings.EqualFold(uri[:len(prefix)], prefix)
}

// TransportForRequest adds what the file extension implies. Inside a .socket
// file a bare host:port is a raw socket, so the scheme may be left out.
func TransportForRequest(uri, filePath string) Transport {
	if transport := TransportForURI(uri); transport.IsStream() {
		return transport
	}
	if strings.EqualFold(filepath.Ext(filePath), SocketFileExtension) {
		return TransportTCP
	}
	return TransportHTTP
}
