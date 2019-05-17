package gateway

import (
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/orientlu/lora-coordinator/internal/config"
	"github.com/orientlu/lora-coordinator/internal/storage"
	"github.com/brocaar/loraserver/api/gw"
	"github.com/golang/protobuf/proto"

	paho "github.com/eclipse/paho.mqtt.golang"
	log "github.com/sirupsen/logrus"
)

// Backend .. gateway
type Backend struct {
	// notifyMacChan storege msg wati to write into redis, notify event: mac
	notifyMacChan chan *storage.MapGatewayMqtt
	// coroutine should Add%Done self
	waitGroup   sync.WaitGroup
	backendType string
}

var gateway = Backend{
	notifyMacChan: make(chan *storage.MapGatewayMqtt, 20),
	backendType:   "gateway",
}

// GetBackend return backend gateway pointer
func GetBackend() *Backend {
	return &gateway
}

// Type return Backend's type
func (b *Backend) Type() string {
	return b.backendType
}

// NotifytHandler ..
func (b *Backend) NotifytHandler(client paho.Client, msg paho.Message) {
	reader := client.OptionsReader()
	server := reader.Servers()
	log.WithFields(log.Fields{
		"Topic":   msg.Topic(),
		"msg_len": len(msg.Payload()),
		"broker":  server[0],
	}).Trace("backen/gateway: handle mqtt msg")

	// topic: backenType/MsgType/EventType
	topic := strings.Replace(msg.Topic(), "/", " ", -1)
	topicSlice := strings.Fields(topic)
	if len(topicSlice) < 3 {
		log.Warningf("backen/gateway: bad topic: %s", msg.Topic())
		return
	}
	switch topicSlice[2] {
	case config.C.Backend.Gateway.NotifyTopicMacEvent:
		go handleNotifyMac(server[0].String(), msg.Payload())
	default:
		log.Warning("backen/gateway:unknow eventType")
	}
}

func handleNotifyMac(mqttURL string, payload []byte) {
	gateway.waitGroup.Add(1)
	defer gateway.waitGroup.Done()

	var stats gw.GatewayStats
	if err := proto.Unmarshal(payload, &stats); err != nil {
		log.Warningf("backen/gateway:unmarshl payload error %s", err)
		return
	}
	m := &storage.MapGatewayMqtt{
		GateWayID:     stats.GatewayId,
		MqttBrokerURL: mqttURL,
		UpdateTime:    time.Now().Local(),
		Expires:       config.C.Backend.Gateway.NotifyTopicMacEventRedisExpires,
	}
	gateway.notifyMacChan <- m

	log.WithFields(log.Fields{
		"GatewayId":     hex.EncodeToString(m.GateWayID),
		"MqttBrokerURL": m.MqttBrokerURL,
		"UpdateTime":    m.UpdateTime,
		"expires":       m.Expires,
	}).Trace("backen/gateway: send notifyMac -> notifyMacChan")
}

// GetNotifyMacChan ...
func GetNotifyMacChan() <-chan *storage.MapGatewayMqtt {
	return gateway.notifyMacChan
}

// SubscribeTopics ..
func (b *Backend) SubscribeTopics(client paho.Client) error {
	// Subscribe topic NotifyTopicTemplate
	for {
		log.WithFields(log.Fields{
			"topic": config.C.Backend.Gateway.NotifyTopicTemplate,
			"qos":   0,
		}).Info("backen/gateway: subscribing topic")
		if token := client.Subscribe(config.C.Backend.Gateway.NotifyTopicTemplate, 0, b.NotifytHandler); token.Wait() && token.Error() != nil {
			log.Error(token.Error(), "retry 1 second")
			time.Sleep(time.Second)
			continue
		}
		break
	}
	return nil
}

// UnSubscribeTopic ..
func (b *Backend) UnSubscribeTopic(client paho.Client) error {
	if token := client.Unsubscribe(config.C.Backend.Gateway.NotifyTopicTemplate); token.Wait() && token.Error() != nil {
		return fmt.Errorf("backend/gateway: unsubscribe topic %s, error: %s",
			config.C.Backend.Gateway.NotifyTopicTemplate, token.Error())
	}
	return nil
}

// Close  wait all backend coroutine quit and close chan
func (b *Backend) Close() {
	// wait all handler coroutine finish
	gateway.waitGroup.Wait()
	//  close all channel
	close(gateway.notifyMacChan)
}
