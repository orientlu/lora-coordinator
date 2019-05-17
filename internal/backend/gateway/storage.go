package gateway

import (
	"sync"

	gwapi "github.com/orientlu/lora-coordinator/api/gateway"
	"github.com/orientlu/lora-coordinator/internal/config"
	"github.com/orientlu/lora-coordinator/internal/storage"

	log "github.com/sirupsen/logrus"
)

var wg sync.WaitGroup

// Start server coroutine will read chan and save map to redis
func Start() {

	// save key prefix in redis for api
	gwapi.NotifyMacPrefixVal = config.C.Backend.Gateway.NotifyTopicMacEventRedisPrefix
	if err := storage.SaveKV(storage.RedisPool(),
		gwapi.NotifyMacPrefixKey,
		gwapi.NotifyMacPrefixVal); err != nil {
		log.Errorf("backend/gateway: set prefix error: %s", err)
	} else {
		log.Infoln("backend/gateway: set NotifyMacPrefixVal successful")
	}

	var coroutineNumber = config.C.Backend.Gateway.NotifyTopicStorageCoroutineNumber
	for i := 0; i < coroutineNumber; i++ {

		go func(id int) {
			wg.Add(1)
			defer wg.Done()

			for n := range GetNotifyMacChan() {
				if err := storage.SaveMapGatewayMqtt(storage.RedisPool(), n); err != nil {
					log.Errorf("backend/gateway: %s", err)
				}
			}
			log.Infof("backend/gateway: gatewaySaveServer stop, coroutineID : %d", id)

		}(i)

	}
	log.Infof("backend/gateway: start gatewaySaveServer, coroutineNumber : %d", coroutineNumber)
}

// Close wait all coroutine quit
// should called after /backend/gateway/gateway .Close()
func Close() {
	log.Info("backend/gateway: wait all coroutine exit")
	wg.Wait()
	log.Info("backend/gateway: all coroutine already exits")
}
