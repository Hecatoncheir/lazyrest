package stream

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/eclipse/paho.golang/packets"
	"github.com/eclipse/paho.golang/paho"
)

const DefaultMQTTKeepAlive uint16 = 30

// MQTTSubscription is one topic filter the session subscribes to on connecting.
type MQTTSubscription struct {
	Topic string
	QoS   byte
}

type MQTTConfig struct {
	DialTimeout   time.Duration
	MaxFrameBytes int

	ClientID  string
	Username  string
	Password  string
	KeepAlive uint16

	// InsecureSkipVerify accepts any certificate an mqtts:// broker presents,
	// matching what the rest of the session does with --insecure.
	InsecureSkipVerify bool

	Subscriptions []MQTTSubscription
	// PublishTopic is where a frame composed by hand is published, since the
	// composer sends a payload and MQTT needs somewhere to put it.
	PublishTopic string
	PublishQoS   byte
}

func (config MQTTConfig) withDefaults() MQTTConfig {
	if config.DialTimeout <= 0 {
		config.DialTimeout = DefaultDialTimeout
	}
	if config.MaxFrameBytes <= 0 {
		config.MaxFrameBytes = DefaultMaxFrameBytes
	}
	if config.KeepAlive == 0 {
		config.KeepAlive = DefaultMQTTKeepAlive
	}
	if config.ClientID == "" {
		config.ClientID = fmt.Sprintf("lazyrest-%d", time.Now().UnixNano()%1e9)
	}
	return config
}

type MQTTSession struct {
	client *paho.Client
	config MQTTConfig

	frames       chan Frame
	done         chan struct{}
	emitMutex    sync.Mutex
	framesClosed bool

	closeOnce  sync.Once
	closeError error

	errorMutex sync.Mutex
	err        error
}

// DialMQTT opens a connection to a broker, subscribes to what the request
// declared, and reports every message on the topics it was given.
//
// Only the dial and the CONNECT exchange observe ctx and the configured
// timeout. The session that follows lasts until one side closes it, so the
// connection is not bound by a timeout meant for a handshake.
func DialMQTT(ctx context.Context, address string, config MQTTConfig) (*MQTTSession, error) {
	config = config.withDefaults()
	if ctx == nil {
		ctx = context.Background()
	}
	dialCtx, cancelDial := context.WithTimeout(ctx, config.DialTimeout)
	defer cancelDial()

	endpoint, secure := mqttEndpoint(address)
	conn, err := dialMQTTConn(dialCtx, endpoint, secure, config)
	if err != nil {
		return nil, err
	}

	session := &MQTTSession{
		config: config,
		frames: make(chan Frame, 64),
		done:   make(chan struct{}),
	}
	session.client = paho.NewClient(paho.ClientConfig{
		Conn:     conn,
		ClientID: config.ClientID,
		OnPublishReceived: []func(paho.PublishReceived) (bool, error){
			func(received paho.PublishReceived) (bool, error) {
				session.emit(session.publishFrame(Received, received.Packet))
				return true, nil
			},
		},
		OnServerDisconnect: func(disconnect *paho.Disconnect) {
			session.emit(Frame{
				At:        time.Now(),
				Direction: Received,
				Opcode:    Close,
				Attributes: []Attribute{
					{Name: "packet", Value: "disconnect"},
					{Name: "reason", Value: strconv.Itoa(int(disconnect.ReasonCode))},
				},
			})
		},
	})

	connect := &paho.Connect{
		ClientID:   config.ClientID,
		KeepAlive:  config.KeepAlive,
		CleanStart: true,
	}
	if config.Username != "" {
		connect.Username = config.Username
		connect.UsernameFlag = true
	}
	if config.Password != "" {
		connect.Password = []byte(config.Password)
		connect.PasswordFlag = true
	}

	connack, err := session.client.Connect(dialCtx, connect)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	if connack != nil && connack.ReasonCode != 0 {
		_ = conn.Close()
		return nil, fmt.Errorf("the broker refused the connection: reason %d", connack.ReasonCode)
	}
	session.emit(Frame{
		At:         time.Now(),
		Direction:  Received,
		Opcode:     Event,
		Attributes: []Attribute{{Name: "packet", Value: "connack"}, {Name: "client-id", Value: config.ClientID}},
	})

	if err := session.subscribe(dialCtx); err != nil {
		_ = session.Close()
		return nil, err
	}

	go session.watch()
	return session, nil
}

// mqttEndpoint reads the scheme a request file writes and returns what to dial
// along with whether the connection is wrapped in TLS. A bare host:port is
// accepted too, and the default port follows the scheme: 1883 in the clear and
// 8883 secured, as the standard assigns them.
func mqttEndpoint(address string) (string, bool) {
	trimmed := strings.TrimSpace(address)
	secure := false

	switch {
	case hasScheme(trimmed, "mqtts"), hasScheme(trimmed, "ssl"), hasScheme(trimmed, "tls"):
		secure = true
		trimmed = trimmed[strings.Index(trimmed, "://")+len("://"):]
	case hasScheme(trimmed, "mqtt"), hasScheme(trimmed, "tcp"):
		trimmed = trimmed[strings.Index(trimmed, "://")+len("://"):]
	}

	trimmed = strings.TrimSuffix(trimmed, "/")
	if !strings.Contains(trimmed, ":") {
		if secure {
			trimmed += ":8883"
		} else {
			trimmed += ":1883"
		}
	}
	return trimmed, secure
}

func hasScheme(address, scheme string) bool {
	prefix := scheme + "://"
	return len(address) >= len(prefix) && strings.EqualFold(address[:len(prefix)], prefix)
}

// dialMQTTConn opens the connection the client will speak over. TLS is the
// caller's concern rather than the client's, which is why paho takes a net.Conn
// and not an address.
func dialMQTTConn(ctx context.Context, endpoint string, secure bool, config MQTTConfig) (net.Conn, error) {
	if !secure {
		dialer := net.Dialer{}
		return dialer.DialContext(ctx, "tcp", endpoint)
	}
	// InsecureSkipVerify follows the same --insecure the rest of the session
	// obeys, so a broker with a self signed certificate is reachable on the
	// same terms as a server with one.
	dialer := tls.Dialer{Config: &tls.Config{
		InsecureSkipVerify: config.InsecureSkipVerify, //nolint:gosec // the user asked for it
		MinVersion:         tls.VersionTLS12,
	}}
	conn, err := dialer.DialContext(ctx, "tcp", endpoint)
	if err != nil {
		return nil, err
	}
	// The client writes from more than one goroutine — a publish, a keepalive —
	// and a tls.Conn is not safe for concurrent writes the way a net.TCPConn
	// is. Without this the stream is corrupted at random and the session hangs
	// until its timeout.
	return packets.NewThreadSafeConn(conn), nil
}

func (session *MQTTSession) subscribe(ctx context.Context) error {
	if len(session.config.Subscriptions) == 0 {
		return nil
	}
	options := make([]paho.SubscribeOptions, 0, len(session.config.Subscriptions))
	for _, subscription := range session.config.Subscriptions {
		options = append(options, paho.SubscribeOptions{Topic: subscription.Topic, QoS: subscription.QoS})
	}
	if _, err := session.client.Subscribe(ctx, &paho.Subscribe{Subscriptions: options}); err != nil {
		return err
	}
	for _, subscription := range session.config.Subscriptions {
		attributes := []Attribute{
			{Name: "packet", Value: "suback"},
			{Name: "topic", Value: subscription.Topic},
		}
		session.emit(Frame{
			At:         time.Now(),
			Direction:  Received,
			Opcode:     Event,
			Attributes: appendQoS(attributes, subscription.QoS),
		})
	}
	return nil
}

// appendQoS records the quality of service only when it is not the default.
// Zero is what a broker assumes, and a topic leaves little room on a row: six
// characters saying nothing push the payload onto a line of its own.
func appendQoS(attributes []Attribute, qos byte) []Attribute {
	if qos == 0 {
		return attributes
	}
	return append(attributes, Attribute{Name: "qos", Value: strconv.Itoa(int(qos))})
}

func (session *MQTTSession) publishFrame(direction Direction, publish *paho.Publish) Frame {
	payload := publish.Payload
	truncated := false
	if len(payload) > session.config.MaxFrameBytes {
		payload = payload[:session.config.MaxFrameBytes]
		truncated = true
	}
	attributes := []Attribute{{Name: "topic", Value: publish.Topic}}
	attributes = appendQoS(attributes, publish.QoS)
	if publish.Retain {
		attributes = append(attributes, Attribute{Name: "retain", Value: "true"})
	}
	return Frame{
		At:         time.Now(),
		Direction:  direction,
		Opcode:     Text,
		Payload:    append([]byte(nil), payload...),
		Truncated:  truncated,
		Attributes: attributes,
	}
}

func (session *MQTTSession) Frames() <-chan Frame { return session.frames }

func (session *MQTTSession) Err() error {
	session.errorMutex.Lock()
	defer session.errorMutex.Unlock()
	return session.err
}

// Send publishes the frame. The topic comes from the frame when it names one
// and from the request otherwise, because a composer sends a payload and MQTT
// has nowhere to put one without a topic.
func (session *MQTTSession) Send(ctx context.Context, frame Frame) error {
	if ctx == nil {
		ctx = context.Background()
	}
	topic, named := frame.Attribute("topic")
	if !named {
		topic = session.config.PublishTopic
	}
	if topic == "" {
		return fmt.Errorf("no topic to publish to: name one in the request")
	}
	publish := &paho.Publish{Topic: topic, QoS: session.config.PublishQoS, Payload: frame.Payload}
	if _, err := session.client.Publish(ctx, publish); err != nil {
		session.setErr(err)
		return err
	}
	session.emit(session.publishFrame(Sent, publish))
	return nil
}

func (session *MQTTSession) Close() error {
	session.closeOnce.Do(func() {
		close(session.done)
		session.closeError = session.client.Disconnect(&paho.Disconnect{ReasonCode: 0})
		session.closeFrames()
	})
	return session.closeError
}

// watch closes the log when the client stops for any reason, so a broker that
// hangs up ends the pane's stream rather than leaving it waiting.
func (session *MQTTSession) watch() {
	select {
	case <-session.client.Done():
	case <-session.done:
	}
	session.closeFrames()
}

func (session *MQTTSession) setErr(err error) {
	session.errorMutex.Lock()
	defer session.errorMutex.Unlock()
	if session.err == nil {
		session.err = err
	}
}

func (session *MQTTSession) emit(frame Frame) {
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

func (session *MQTTSession) closeFrames() {
	session.emitMutex.Lock()
	defer session.emitMutex.Unlock()
	if !session.framesClosed {
		session.framesClosed = true
		close(session.frames)
	}
}
