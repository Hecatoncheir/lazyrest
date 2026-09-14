package http

import (
	"path/filepath"
	"strings"
)

// SocketFileExtension marks a file of raw socket requests.
//
// What decides the file is not whether the protocol is HTTP. It is whether the
// request wants the surrounding HTTP session: its variables, its cookies, and
// the responses earlier requests captured. A WebSocket upgrades from that
// session, and an MQTT broker commonly takes as its password a token an HTTP
// login returned, so both live in .http and can write
// {{login.response.body.$.token}}. A raw socket authenticates inside its own
// protocol and wants none of it, so it gets a file of its own — and gives up
// the chaining, which a reference keyed by source file cannot cross.
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
