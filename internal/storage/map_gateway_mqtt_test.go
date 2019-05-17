package storage_test

import (
	"testing"
	"time"

	api "github.com/orientlu/lora-coordinator/api/gateway"
	"github.com/orientlu/lora-coordinator/internal/config"
	"github.com/orientlu/lora-coordinator/internal/storage"

	log "github.com/sirupsen/logrus"
)

func BenchmarkGetGatewayMapMqttURL(b *testing.B) {

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

	gatewayID := []byte{1, 2, 3, 4, 5, 6, 7, 8}
	mqttURL := "http://localhost:1883"

	testdata := &storage.MapGatewayMqtt{
		GateWayID:     gatewayID,
		MqttBrokerURL: mqttURL,
		UpdateTime:    time.Now().Local(),
		Expires:       60 * 10,
	}

	for i := 0; i < b.N; i++ {
		storage.SaveMapGatewayMqtt(storage.RedisPool(), testdata)
	}
	return
}
