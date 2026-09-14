package stream

import (
	"net/http"
	"testing"
)

func TestMQTTConfigFromHeader(t *testing.T) {
	header := http.Header{}
	header.Set(HeaderClientID, "lazyrest-dev")
	header.Set(HeaderUsername, "sensor-reader")
	header.Set(HeaderPassword, "from-a-variable")
	header.Set(HeaderKeepAlive, "45")
	// http.Header keeps every value a name was given, which is how a request
	// subscribes to more than one topic.
	header.Add(HeaderSubscribe, "sensors/+/temp; qos=1")
	header.Add(HeaderSubscribe, "devices/#")
	header.Set(HeaderTopic, "commands/reboot; qos=2")

	config := MQTTConfigFromHeader(header)

	if config.ClientID != "lazyrest-dev" || config.Username != "sensor-reader" || config.Password != "from-a-variable" {
		t.Fatalf("credentials = %+v", config)
	}
	if config.KeepAlive != 45 {
		t.Errorf("keep alive = %d, want 45", config.KeepAlive)
	}
	if len(config.Subscriptions) != 2 {
		t.Fatalf("subscriptions = %+v, want two", config.Subscriptions)
	}
	if config.Subscriptions[0] != (MQTTSubscription{Topic: "sensors/+/temp", QoS: 1}) {
		t.Errorf("first subscription = %+v", config.Subscriptions[0])
	}
	if config.Subscriptions[1] != (MQTTSubscription{Topic: "devices/#", QoS: 0}) {
		t.Errorf("second subscription = %+v", config.Subscriptions[1])
	}
	if config.PublishTopic != "commands/reboot" || config.PublishQoS != 2 {
		t.Errorf("publish topic = %q qos %d", config.PublishTopic, config.PublishQoS)
	}
}

func TestTopicAndQoS(t *testing.T) {
	cases := []struct {
		in    string
		topic string
		qos   byte
	}{
		{"sensors/#", "sensors/#", 0},
		{"sensors/#; qos=1", "sensors/#", 1},
		{"  sensors/# ; QOS = 2 ", "sensors/#", 2},
		// A level a broker would reject is ignored rather than guessed at.
		{"sensors/#; qos=9", "sensors/#", 0},
		{"sensors/#; qos=oops", "sensors/#", 0},
		{"", "", 0},
	}
	for _, testCase := range cases {
		topic, qos := topicAndQoS(testCase.in)
		if topic != testCase.topic || qos != testCase.qos {
			t.Errorf("topicAndQoS(%q) = %q, %d; want %q, %d", testCase.in, topic, qos, testCase.topic, testCase.qos)
		}
	}
}

func TestMQTTConfigFromAnEmptyHeader(t *testing.T) {
	config := MQTTConfigFromHeader(http.Header{})
	if len(config.Subscriptions) != 0 || config.PublishTopic != "" || config.KeepAlive != 0 {
		t.Fatalf("an empty header produced %+v", config)
	}
}
