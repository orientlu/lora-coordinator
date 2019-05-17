package gateway

import (
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
	log "github.com/sirupsen/logrus"
)

const (
	// NotifyMacPrefixKey to get prefix from redis
	NotifyMacPrefixKey = "Backend.Gateway.NotifyTopicMacEventRedisPrefix"
	// NotifyMacKeyTempl prefix/gatewayID
	NotifyMacKeyTempl = "%s&%s"
	// NotifyMacValTempl mqtturl/updatetime
	NotifyMacValTempl = "%s&%s"
)
const updatePrefixTTL = 5 * time.Minute

// NotifyMacPrefixVal ..
var NotifyMacPrefixVal = ""
var updatePrefixtime = time.Now()

// GetGatewayMapMqttURL .. return the gateway connected mqtt broker
func GetGatewayMapMqttURL(p *redis.Pool, gatewayID []byte) (url string, err error) {
	c := p.Get()
	defer c.Close()

	if time.Now().Sub(updatePrefixtime) >= updatePrefixTTL {
		updatePrefixtime = time.Now()
		updateNotifyMapPrefix(p)
	}

	/// make key and get mqtturl/updatetime
	gwID := hex.EncodeToString(gatewayID)
	key := fmt.Sprintf(NotifyMacKeyTempl, NotifyMacPrefixVal, gwID)
	val, err := redis.String(c.Do("GET", key))
	if err != nil && err != redis.ErrNil {
		log.WithFields(log.Fields{
			"gatewayID": gwID,
			"err":       err,
		}).Errorf("coordinator-api/gateway: Get mqttUrl[gatewayID] error")
		return "", err
	}
	info := strings.Split(val, "&")
	if len(info) != len(strings.Split(NotifyMacValTempl, "&")) {
		return "", fmt.Errorf("coordinator-api/gateway: mqttUrl[gatewayID] value format error")
	}
	url = info[0]
	updatetime := info[1]
	log.WithFields(log.Fields{
		"gatewayId":  gwID,
		"mqtt":       url,
		"UpdateTime": updatetime,
	}).Trace("coordinator-api/gateway: Get mqttUrl[gatewayID]")
	return url, err
}

func updateNotifyMapPrefix(p *redis.Pool) {
	c := p.Get()
	defer c.Close()

	str, err := redis.String(c.Do("GET", NotifyMacPrefixKey))
	if err != nil && err != redis.ErrNil {
		log.WithFields(log.Fields{
			"key": NotifyMacPrefixKey,
			"err": err,
		}).Errorf("coordinator-api/gateway: Get key prefix error")
	}
	NotifyMacPrefixVal = str
	log.Infof("coordinator-api/gateway: update NotifyTopicMacEvent prexfix: %s", str)
}
