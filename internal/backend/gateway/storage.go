package gateway

import (
	"sync"

	"github.com/orientlu/lora-coordinator/internal/config"
	"github.com/orientlu/lora-coordinator/internal/storage"

	log "github.com/sirupsen/logrus"
)

var wg sync.WaitGroup

// Start server coroutine will read chan and save map to redis
func Start() {

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
