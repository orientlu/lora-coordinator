package storage

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/orientlu/lora-coordinator/api/gateway"

	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// MapGatewayMqtt ..
type MapGatewayMqtt struct {
	GateWayID     []byte
	MqttBrokerURL string
	UpdateTime    time.Time
	Expires       int // map expire time in second
}

// SaveMapGatewayMqtt store the struct to redis
func SaveMapGatewayMqtt(p *redis.Pool, m *MapGatewayMqtt) error {

	c := p.Get()
	defer c.Close()

	// formate
	key := fmt.Sprintf(gateway.NotifyMacKeyTempl, gateway.NotifyMacPrefixVal, hex.EncodeToString(m.GateWayID))
	val := fmt.Sprintf(gateway.NotifyMacValTempl, m.MqttBrokerURL, m.UpdateTime)

	str, err := redis.String(c.Do("SET", key, val, "EX", m.Expires))
	if err != nil {
		log.Errorf("storage/GatewayMqttt: redis set error %s, %s", str, err)
		return errors.Wrap(err, "set redis error")
	}
	log.WithFields(log.Fields{
		"key":       key,
		"val":       val,
		"expire(s)": m.Expires,
		"redis":     str,
	}).Trace("storage/GatewayMqtt: redis set val")
	return nil
}
