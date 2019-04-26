package storage

import (
	"fmt"
	"sync"
	"time"

	"github.com/orientlu/lora-coordinator/internal/config"

	"github.com/gomodule/redigo/redis"
	log "github.com/sirupsen/logrus"
)

// redisPool holds Redis connection pool.
var redisPool *redis.Pool

const (
	redisDialWriteTimeout = time.Second
	redisDialReadTimeout  = time.Minute
	onBorrowPingInterval  = time.Minute
)

var setupOnce sync.Once

// Setup init redis pool
func Setup(conf config.Config) error {
	log.Info("storage: setting up Redis connection pool")

	setupOnce.Do(func() {
		redisPool = &redis.Pool{
			MaxIdle:     conf.Redis.MaxIdle,
			MaxActive:   conf.Redis.MaxActive,
			IdleTimeout: conf.Redis.IdleTimeout,

			Dial: func() (redis.Conn, error) {
				log.Warning("redis Dial")
				conn, err := redis.DialURL(conf.Redis.URL,
					redis.DialReadTimeout(redisDialReadTimeout),
					redis.DialWriteTimeout(redisDialWriteTimeout),
				)
				if err != nil {
					return nil, fmt.Errorf("storage/db: connect redis error: %s", err)
				}
				return conn, err
			},
			// check the health of an idle connection before the connection is returned to the app
			TestOnBorrow: func(c redis.Conn, t time.Time) error {
				if time.Now().Sub(t) < onBorrowPingInterval {
					return nil
				}

				_, err := c.Do("PING")
				if err != nil {
					return fmt.Errorf("storage/db: ping redis error: %s", err)
				}
				return nil
			},
		}
	})
	return nil
}

// RedisPool return  redis pool
func RedisPool() *redis.Pool {
	return redisPool
}
