package stream

import (
	"net/http"
	"strconv"
	"strings"
)

// Header names a request uses to describe an MQTT session. They are ordinary
// headers so a request file needs no syntax of its own, and Subscribe repeats
// because http.Header keeps every value a name was given.
const (
	HeaderClientID  = "Client-Id"
	HeaderSubscribe = "Subscribe"
	HeaderTopic     = "Topic"
	HeaderUsername  = "Username"
	HeaderPassword  = "Password"
	HeaderKeepAlive = "Keep-Alive"
)

// MQTTConfigFromHeader reads what the request said about the session. It takes
// a header rather than a parsed request so the transport stays independent of
// the parser.
func MQTTConfigFromHeader(header http.Header) MQTTConfig {
	config := MQTTConfig{
		ClientID: header.Get(HeaderClientID),
		Username: header.Get(HeaderUsername),
		Password: header.Get(HeaderPassword),
	}
	if seconds, err := strconv.Atoi(strings.TrimSpace(header.Get(HeaderKeepAlive))); err == nil && seconds > 0 {
		config.KeepAlive = uint16(seconds)
	}
	for _, value := range header.Values(HeaderSubscribe) {
		topic, qos := topicAndQoS(value)
		if topic == "" {
			continue
		}
		config.Subscriptions = append(config.Subscriptions, MQTTSubscription{Topic: topic, QoS: qos})
	}
	config.PublishTopic, config.PublishQoS = topicAndQoS(header.Get(HeaderTopic))
	return config
}

// topicAndQoS reads `topic; qos=1`, the way an HTTP header carries a value and
// its parameters. A missing or unreadable quality of service is zero, which is
// what a broker assumes anyway.
func topicAndQoS(value string) (string, byte) {
	topic, parameters, _ := strings.Cut(value, ";")
	topic = strings.TrimSpace(topic)

	for _, parameter := range strings.Split(parameters, ";") {
		name, setting, found := strings.Cut(parameter, "=")
		if !found || !strings.EqualFold(strings.TrimSpace(name), "qos") {
			continue
		}
		level, err := strconv.Atoi(strings.TrimSpace(setting))
		if err != nil || level < 0 || level > 2 {
			continue
		}
		return topic, byte(level)
	}
	return topic, 0
}
