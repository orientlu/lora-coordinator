package backend

import (
	paho "github.com/eclipse/paho.mqtt.golang"
)

// Backend interface
type Backend interface {
	// SubscribeTopics called by mqtt connected, to subscribe backend topic
	SubscribeTopics(client paho.Client) error

	// UnSubscribeTopic
	UnSubscribeTopic(client paho.Client) error

	// Type backend type
	Type() string

	// Close  wait all backend coroutine quit and close chan
	Close()
}
