package stream

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"math/big"
	"net"
	"testing"
	"time"

	"github.com/eclipse/paho.golang/packets"
)

// fakeBroker answers just enough MQTT 5 to drive a session, using the packet
// layer that ships with the client so the encoding is never hand rolled.
type fakeBroker struct {
	address   string
	published chan *packets.Publish
	conn      chan net.Conn
}

func startBroker(t *testing.T) *fakeBroker {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	return serveBroker(t, listener)
}

// startTLSBroker serves the same broker behind a certificate it signs itself,
// which is what an mqtts:// session has to get through.
func startTLSBroker(t *testing.T) *fakeBroker {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	return serveBroker(t, tls.NewListener(listener, &tls.Config{
		Certificates: []tls.Certificate{selfSignedCertificate(t)},
		MinVersion:   tls.VersionTLS12,
	}))
}

func selfSignedCertificate(t *testing.T) tls.Certificate {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	certificate, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("create certificate: %v", err)
	}
	return tls.Certificate{Certificate: [][]byte{certificate}, PrivateKey: key}
}

func serveBroker(t *testing.T, listener net.Listener) *fakeBroker {
	t.Helper()
	t.Cleanup(func() { _ = listener.Close() })

	broker := &fakeBroker{
		address:   listener.Addr().String(),
		published: make(chan *packets.Publish, 8),
		conn:      make(chan net.Conn, 1),
	}
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		broker.conn <- conn
		defer func() { _ = conn.Close() }()
		for {
			packet, err := packets.ReadPacket(conn)
			if err != nil {
				return
			}
			switch packet.Type {
			case packets.CONNECT:
				reply := packets.NewControlPacket(packets.CONNACK)
				_, _ = reply.WriteTo(conn)
			case packets.SUBSCRIBE:
				subscribe := packet.Content.(*packets.Subscribe)
				reply := packets.NewControlPacket(packets.SUBACK)
				suback := reply.Content.(*packets.Suback)
				suback.PacketID = subscribe.PacketID
				suback.Reasons = []byte{0}
				_, _ = reply.WriteTo(conn)
			case packets.PUBLISH:
				broker.published <- packet.Content.(*packets.Publish)
			case packets.PINGREQ:
				reply := packets.NewControlPacket(packets.PINGRESP)
				_, _ = reply.WriteTo(conn)
			case packets.DISCONNECT:
				return
			}
		}
	}()
	return broker
}

func (broker *fakeBroker) deliver(t *testing.T, topic, payload string) {
	t.Helper()
	var conn net.Conn
	select {
	case conn = <-broker.conn:
		broker.conn <- conn
	case <-time.After(5 * time.Second):
		t.Fatal("the client never connected")
	}
	packet := packets.NewControlPacket(packets.PUBLISH)
	publish := packet.Content.(*packets.Publish)
	publish.Topic = topic
	publish.Payload = []byte(payload)
	if _, err := packet.WriteTo(conn); err != nil {
		t.Fatalf("deliver: %v", err)
	}
}

func dialBroker(t *testing.T, broker *fakeBroker, config MQTTConfig) *MQTTSession {
	t.Helper()
	session, err := DialMQTT(context.Background(), broker.address, config)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { _ = session.Close() })
	return session
}

func TestMQTTEndpoint(t *testing.T) {
	cases := []struct {
		address string
		want    string
		secure  bool
	}{
		{"mqtt://127.0.0.1:1883", "127.0.0.1:1883", false},
		{"tcp://broker:1883", "broker:1883", false},
		{"broker:1883", "broker:1883", false},
		// The default port follows the scheme, as the standard assigns them.
		{"broker", "broker:1883", false},
		{"mqtts://broker", "broker:8883", true},
		{"mqtts://broker:1884", "broker:1884", true},
		{"MQTTS://BROKER", "BROKER:8883", true},
		{"ssl://broker", "broker:8883", true},
		{"  mqtt://host/ ", "host:1883", false},
	}
	for _, testCase := range cases {
		endpoint, secure := mqttEndpoint(testCase.address)
		if endpoint != testCase.want || secure != testCase.secure {
			t.Errorf("mqttEndpoint(%q) = %q, %v; want %q, %v",
				testCase.address, endpoint, secure, testCase.want, testCase.secure)
		}
	}
}

func TestMQTTSessionReportsTheConnection(t *testing.T) {
	broker := startBroker(t)
	session := dialBroker(t, broker, MQTTConfig{ClientID: "lazyrest-test"})

	frame := nextFrame(t, session.Frames())
	if frame.Opcode != Event {
		t.Fatalf("opcode = %v, want event", frame.Opcode)
	}
	if packet, _ := frame.Attribute("packet"); packet != "connack" {
		t.Fatalf("packet = %q, want connack", packet)
	}
}

func TestMQTTSessionReportsEachSubscription(t *testing.T) {
	broker := startBroker(t)
	session := dialBroker(t, broker, MQTTConfig{
		Subscriptions: []MQTTSubscription{{Topic: "sensors/+/temp", QoS: 1}},
	})

	nextFrame(t, session.Frames()) // connack
	frame := nextFrame(t, session.Frames())
	if packet, _ := frame.Attribute("packet"); packet != "suback" {
		t.Fatalf("packet = %q, want suback", packet)
	}
	if topic, _ := frame.Attribute("topic"); topic != "sensors/+/temp" {
		t.Fatalf("topic = %q", topic)
	}
	if qos, _ := frame.Attribute("qos"); qos != "1" {
		t.Fatalf("qos = %q, want 1", qos)
	}
}

// A payload without the topic it arrived on hides the part that matters most.
func TestMQTTSessionReportsAMessageWithItsTopic(t *testing.T) {
	broker := startBroker(t)
	session := dialBroker(t, broker, MQTTConfig{
		Subscriptions: []MQTTSubscription{{Topic: "sensors/#"}},
	})
	nextFrame(t, session.Frames()) // connack
	nextFrame(t, session.Frames()) // suback

	broker.deliver(t, "sensors/1/temp", `{"celsius":21}`)

	frame := nextFrame(t, session.Frames())
	if string(frame.Payload) != `{"celsius":21}` {
		t.Fatalf("payload = %q", frame.Payload)
	}
	if topic, _ := frame.Attribute("topic"); topic != "sensors/1/temp" {
		t.Fatalf("topic = %q, want the topic it arrived on", topic)
	}
	if frame.Direction != Received {
		t.Fatalf("direction = %v, want Received", frame.Direction)
	}
}

func TestMQTTSessionPublishesToTheRequestTopic(t *testing.T) {
	broker := startBroker(t)
	session := dialBroker(t, broker, MQTTConfig{PublishTopic: "commands/reboot"})
	nextFrame(t, session.Frames()) // connack

	if err := session.Send(context.Background(), Frame{Payload: []byte("now")}); err != nil {
		t.Fatalf("send: %v", err)
	}

	frame := nextFrame(t, session.Frames())
	if frame.Direction != Sent {
		t.Fatalf("direction = %v, want Sent", frame.Direction)
	}
	if topic, _ := frame.Attribute("topic"); topic != "commands/reboot" {
		t.Fatalf("topic = %q", topic)
	}

	select {
	case published := <-broker.published:
		if published.Topic != "commands/reboot" || string(published.Payload) != "now" {
			t.Fatalf("broker received %q on %q", published.Payload, published.Topic)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("the broker never received the message")
	}
}

// A frame may name its own topic, which is how one session can publish to more
// than one place.
func TestMQTTSessionPublishesToATopicNamedByTheFrame(t *testing.T) {
	broker := startBroker(t)
	session := dialBroker(t, broker, MQTTConfig{PublishTopic: "default/topic"})
	nextFrame(t, session.Frames())

	err := session.Send(context.Background(), Frame{
		Payload:    []byte("hi"),
		Attributes: []Attribute{{Name: "topic", Value: "chosen/topic"}},
	})
	if err != nil {
		t.Fatalf("send: %v", err)
	}

	select {
	case published := <-broker.published:
		if published.Topic != "chosen/topic" {
			t.Fatalf("published to %q, want the topic the frame named", published.Topic)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("the broker never received the message")
	}
}

func TestMQTTSessionRefusesToPublishWithoutATopic(t *testing.T) {
	broker := startBroker(t)
	session := dialBroker(t, broker, MQTTConfig{})
	nextFrame(t, session.Frames())

	if err := session.Send(context.Background(), Frame{Payload: []byte("x")}); err == nil {
		t.Fatal("publishing with no topic anywhere returned no error")
	}
}

// The same invariant the other transports have: the dial bound must not reach
// the session.
func TestMQTTSessionOutlivesTheDialTimeout(t *testing.T) {
	broker := startBroker(t)
	session := dialBroker(t, broker, MQTTConfig{
		DialTimeout:   200 * time.Millisecond,
		Subscriptions: []MQTTSubscription{{Topic: "late/#"}},
	})
	nextFrame(t, session.Frames()) // connack
	nextFrame(t, session.Frames()) // suback

	time.Sleep(500 * time.Millisecond)
	broker.deliver(t, "late/message", "still here")

	if got := string(nextFrame(t, session.Frames()).Payload); got != "still here" {
		t.Fatalf("payload = %q", got)
	}
}

// Zero is what a broker assumes, and the row is narrow: saying it costs six
// characters and pushes the payload onto a line of its own.
func TestMQTTSessionOmitsTheDefaultQoS(t *testing.T) {
	broker := startBroker(t)
	session := dialBroker(t, broker, MQTTConfig{
		Subscriptions: []MQTTSubscription{{Topic: "plain/#"}, {Topic: "careful/#", QoS: 2}},
	})
	nextFrame(t, session.Frames()) // connack

	plain := nextFrame(t, session.Frames())
	if _, found := plain.Attribute("qos"); found {
		t.Errorf("a default quality of service was reported: %+v", plain.Attributes)
	}
	careful := nextFrame(t, session.Frames())
	if qos, _ := careful.Attribute("qos"); qos != "2" {
		t.Errorf("qos = %q, want 2 kept because it is not the default", qos)
	}

	broker.deliver(t, "plain/message", "body")
	message := nextFrame(t, session.Frames())
	if _, found := message.Attribute("qos"); found {
		t.Errorf("a default quality of service was reported on a message: %+v", message.Attributes)
	}
	if topic, _ := message.Attribute("topic"); topic != "plain/message" {
		t.Errorf("topic = %q", topic)
	}
}

// mqtts:// has to get through a certificate, and a development broker usually
// signs its own.
func TestMQTTSessionConnectsOverTLS(t *testing.T) {
	broker := startTLSBroker(t)
	session, err := DialMQTT(context.Background(), "mqtts://"+broker.address, MQTTConfig{
		InsecureSkipVerify: true,
		Subscriptions:      []MQTTSubscription{{Topic: "secure/#"}},
	})
	if err != nil {
		t.Fatalf("dial over TLS: %v", err)
	}
	t.Cleanup(func() { _ = session.Close() })

	nextFrame(t, session.Frames()) // connack
	nextFrame(t, session.Frames()) // suback
	broker.deliver(t, "secure/message", "over tls")

	if got := string(nextFrame(t, session.Frames()).Payload); got != "over tls" {
		t.Fatalf("payload = %q", got)
	}
}

// Without --insecure a certificate the session cannot verify must stop it,
// rather than being accepted quietly.
func TestMQTTSessionRefusesAnUnverifiableCertificate(t *testing.T) {
	broker := startTLSBroker(t)
	_, err := DialMQTT(context.Background(), "mqtts://"+broker.address, MQTTConfig{
		DialTimeout: 3 * time.Second,
	})
	if err == nil {
		t.Fatal("a self signed certificate was accepted without being asked for")
	}
}

// A plain broker reached over mqtts:// must fail rather than hang.
func TestMQTTSessionFailsWhenTheBrokerSpeaksNoTLS(t *testing.T) {
	broker := startBroker(t)
	_, err := DialMQTT(context.Background(), "mqtts://"+broker.address, MQTTConfig{
		DialTimeout:        3 * time.Second,
		InsecureSkipVerify: true,
	})
	if err == nil {
		t.Fatal("a plain broker answered an mqtts:// dial")
	}
}
