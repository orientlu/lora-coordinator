package config

import (
	"time"
)

// Config define the configuration structure
type Config struct {
	General struct {
		LogLevel  int  `mapstructure:"log_level"`
		LogReport bool `mapstructure:"log_report"`
	}
	Redis struct {
		URL         string        `mapstructure:"url"`
		MaxIdle     int           `mapstructure:"max_idle"`   // 维持空闲链接数
		MaxActive   int           `mapstructure:"max_active"` // 最大连接数
		IdleTimeout time.Duration `mapstructure:"idle_timeout"`
	}

	MQTT struct {
		Servers      []string
		Username     string
		Password     string
		QOS          uint8  `mapstructure:"qos"`
		CleanSession bool   `mapstructure:"clean_session"`
		ClientID     string `mapstructure:"client_id"`
		CACert       string `mapstructure:"ca_cert"`
		TLSCert      string `mapstructure:"tls_cert"`
		TLSKey       string `mapstructure:"tls_key"`
	} `mapstructure:"mqtt"`

	Backend struct {
		Gateway struct {
			NotifyTopicTemplate               string `mapstructure:"notify_topic_template"`
			NotifyTopicStorageCoroutineNumber int    `mapstructure:"notify_topic_storage_coroutine_number"`
			// complete topicName : NotifyTopicTemplate/NotifyTopicMacEvent
			NotifyTopicMacEvent             string `mapstructure:"notify_topic_mac_event"`
			NotifyTopicMacEventRedisPrefix  string `mapstructure:"notify_topic_mac_event_redis_prefix"`
			NotifyTopicMacEventRedisExpires int    `mapstructure:"notify_topic_mac_event_redis_expires"`
		}
	}
}

// C hold the global configuration
var C Config
