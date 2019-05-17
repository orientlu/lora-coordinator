package gateway_test

import (
	"testing"
	"time"

	api "github.com/orientlu/lora-coordinator/api/gateway"
	"github.com/orientlu/lora-coordinator/internal/config"
	"github.com/orientlu/lora-coordinator/internal/storage"

	log "github.com/sirupsen/logrus"
)

func BenchmarkGetGatewayMapMqttURL(b *testing.B) {
	prepareEnv()

	gatewayID := []byte{1, 2, 3, 4, 5, 6, 7, 8}
	mqttURL := "http://localhost:1883"

	testdata := &storage.MapGatewayMqtt{
		GateWayID:     gatewayID,
		MqttBrokerURL: mqttURL,
		UpdateTime:    time.Now().Local(),
		Expires:       60 * 10,
	}
	storage.SaveMapGatewayMqtt(storage.RedisPool(), testdata)

	for i := 0; i < b.N; i++ {
		url, err := api.GetGatewayMapMqttURL(storage.RedisPool(), gatewayID)
		if err != nil {
			log.WithError(err).Error("get url return error")
		}
		if url != mqttURL {
			log.Error("get error url")
		}
	}
	return
}

func prepareEnv() {
	config.C.Redis.MaxIdle = 10
	config.C.Redis.MaxActive = 100
	config.C.Redis.IdleTimeout = 5 * time.Minute
	config.C.Redis.URL = "redis://localhost:6379"
	if err := storage.Setup(config.C); err != nil {
		log.WithError(err).Error("setup storage error")
		return
	}
	conn := storage.RedisPool().Get()
	defer conn.Close()

	// save key prefix in redis for api
	config.C.Backend.Gateway.NotifyTopicMacEventRedisPrefix = "mac_event_test"
	api.NotifyMacPrefixVal = config.C.Backend.Gateway.NotifyTopicMacEventRedisPrefix
	if err := storage.SaveKV(storage.RedisPool(),
		api.NotifyMacPrefixKey,
		api.NotifyMacPrefixVal); err != nil {
		log.Errorf("set prefix error: %s", err)
		return
	}
}

func TestGetGatewayMapMqttURL(t *testing.T) {
	prepareEnv()
	gatewayID := []byte{1, 2, 3, 4, 5, 6, 7, 8}
	mqttURL1 := "http://localhost:1883"
	mqttURL2 := "http://xxxxx:1883"

	var tests = []struct {
		input        *storage.MapGatewayMqtt
		getGatewayID []byte
		wanURL       string
	}{
		{&storage.MapGatewayMqtt{
			GateWayID:     gatewayID,
			MqttBrokerURL: mqttURL1,
			UpdateTime:    time.Now().Local(),
			Expires:       2,
		}, gatewayID, mqttURL1},
		{&storage.MapGatewayMqtt{
			GateWayID:     gatewayID,
			MqttBrokerURL: mqttURL2,
			UpdateTime:    time.Now().Local(),
			Expires:       2,
		}, gatewayID, mqttURL2},
		{&storage.MapGatewayMqtt{ // no exist
			GateWayID:     []byte{1, 3, 4},
			MqttBrokerURL: mqttURL1,
			UpdateTime:    time.Now().Local(),
			Expires:       2,
		}, []byte{2, 2, 2, 2, 2, 2, 2, 2}, ""},
		{&storage.MapGatewayMqtt{
			GateWayID:     gatewayID,
			MqttBrokerURL: "",
			UpdateTime:    time.Now().Local(),
			Expires:       2,
		}, gatewayID, ""},
	}

	for _, test := range tests {
		storage.SaveMapGatewayMqtt(storage.RedisPool(), test.input)
		url, _ := api.GetGatewayMapMqttURL(storage.RedisPool(), test.getGatewayID)
		if url != test.wanURL {
			t.Errorf("get error url, get: %s, want: %s", url, test.wanURL)
		}
	}
}
