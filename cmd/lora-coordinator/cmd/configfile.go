package cmd

import (
	"os"
	"text/template"

	"github.com/orientlu/lora-coordinator/internal/config"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const configTemplate = `
## general setting
[general]
# trace=6, debug=5, info=4, warning=3, error=2, fatal=1, panic=0
log_level={{ .General.LogLevel }}

# ture/false, show detailed log for debug
log_report={{ .General.LogReport }}


## redis setting
# Please note that Redis 2.6.0+ is required.
[redis]
# Redis url (e.g. redis://user:password@hostname/0)
#	 redis://:password@hostname:port/db_number
# For more information about the Redis URL format, see:
# https://www.iana.org/assignments/uri-schemes/prov/redi
url ="{{ .Redis.URL }}"

# Max idle connections in the pool. dial frequency
max_idle={{ .Redis.MaxIdle }}

# Max active connections in the pool.
max_active={{ .Redis.MaxActive }}

# Close connections after remaining idle for this duration. If the value
# is zero, then idle connections are not closed. You should set
# the timeout to a value less than the server's timeout.
idle_timeout="{{ .Redis.IdleTimeout }}"

## Mqtt setting
[mqtt]
# Server list of mqtt broker
# (e.g. ["scheme://host:port", "..."] where scheme is tcp, ssl or ws
servers=[{{ range $i, $url := .MQTT.Servers }}
	"{{ $url }}",{{end}}
]

# Connect with the given password (optional)
password="{{ .MQTT.Password }}"

# 0: at most once
# 1: at least once
# 2: exactly once
# Note: an increase of this value will decrease the performance.
# For more information: https://www.hivemq.com/blog/mqtt-essentials-part-6-mqtt-quality-of-service-levels
qos={{ .MQTT.QOS }}

# Set the "clean session" flag in the connect message when this client
# connects to an MQTT broker. By setting this flag you are indicating
# that no messages saved by the broker for this client should be delivered.
clean_session={{ .MQTT.CleanSession }}

# Set the client id to be used by this client when connecting to the MQTT
# broker. A client id must be no longer than 23 characters. When left blank,
# a random id will be generated. This requires clean_session=true.
client_id="{{ .MQTT.ClientID }}"

#CA certificate file (optional)
# Use this when setting up a secure connection (when server uses ssl://...)
# but the certificate used by the server is not trusted by any CA certificate
# on the server (e.g. when self generated).
ca_cert="{{ .MQTT.CACert }}"

# TLS certificate file (optional)
tls_cert="{{ .MQTT.TLSCert }}"

# TLS key file (optional)
tls_key="{{ .MQTT.TLSKey }}"


## Backend setting
[backend]
 ### Gateway setting
 [backend.gateway]
 # default gateway/notify/+
 notify_topic_template="{{ .Backend.Gateway.NotifyTopicTemplate }}"

 # start how many croutine to rev chan and set to redis
 notify_topic_storage_coroutine_number={{ .Backend.Gateway.NotifyTopicStorageCoroutineNumber }}

 # sub topic of NotifyTopicTemplate mac : gateway/notify/mac
 # mqttUrl <-> gatewayId, so we can know gateway connect which mqtt broker now
 notify_topic_mac_event="{{ .Backend.Gateway.NotifyTopicMacEvent }}"

 # set key_name prefix in redis, (e.g. thePrefix/[gatewayID]:MqttUrl/updatetime)
 notify_topic_mac_event_redis_prefix="{{ .Backend.Gateway.NotifyTopicMacEventRedisPrefix }}"

 # set the map mqttUrl <-> gatewayID expire time(second) in redis
 notify_topic_mac_event_redis_expires={{ .Backend.Gateway.NotifyTopicMacEventRedisExpires }}
`

var configCmd = &cobra.Command{
	Use:   "configfile",
	Short: "Print the loRa coordinator configuration file",
	RunE: func(cmd *cobra.Command, args []string) error {
		t := template.Must(template.New("config").Parse(configTemplate))
		if err := t.Execute(os.Stdout, &config.C); err != nil {
			log.Println(err)
			return errors.Wrap(err, "excute config template error")
		}
		return nil
	},
}
