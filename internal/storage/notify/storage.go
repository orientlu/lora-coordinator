package notify

import (
	"sync"

	"github.com/orientlu/lora-coordinator/internal/backend/gateway"
	"github.com/orientlu/lora-coordinator/internal/config"
	"github.com/orientlu/lora-coordinator/internal/storage"

	"github.com/gomodule/redigo/redis"
	log "github.com/sirupsen/logrus"
)

var wg sync.WaitGroup

// Start ....
func Start() {

	var coroutineNumber = config.C.Backend.Gateway.NotifyTopicStorageCoroutineNumber
	for i := 0; i < coroutineNumber; i++ {

		go func(id int) {
			wg.Add(1)
			defer wg.Done()

			for n := range gateway.GetNotifyMacChan() {
				redisConn := storage.RedisPool().Get()
				str, err := redis.String(redisConn.Do("SET", n.Key, n.Val, "EX", n.Expires))
				if err != nil {
					log.Errorf("storage/notify: redis set error %s, %s", str, err)
				}
				log.WithFields(log.Fields{
					"key":       n.Key,
					"val":       n.Val,
					"expire(s)": n.Expires,
					"redis":     str,
				}).Trace("storage/notify:redis set val")

				redisConn.Close()
			}
			log.Infof("storage/notify: notify strage stop, coroutineID : %d", id)

		}(i)

	}
	log.Infof("storage/notify: start notify strage, coroutineNumber : %d", coroutineNumber)
}

// Close wait all coroutine quit
func Close() {
	log.Info("storage/notify: wait all coroutine exit")
	wg.Wait()
	log.Info("storage/notify: all coroutine already exits")
}
