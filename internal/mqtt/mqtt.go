package mqtt

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"math/rand"
	"sync"
	"time"

	"github.com/orientlu/lora-coordinator/internal/backend"
	"github.com/orientlu/lora-coordinator/internal/config"

	paho "github.com/eclipse/paho.mqtt.golang"
	log "github.com/sirupsen/logrus"
)

// Brokers hold all mqtt brokers client
var Brokers map[string]paho.Client = make(map[string]paho.Client)

var backends []backend.Backend = make([]backend.Backend, 0)

// Setup ..
func Setup(conf config.Config) error {
	log.Info("mqtt: setting up mqtt broker connection")

	if len(backends) == 0 {
		log.Warnln("You should call AddBackend before Setup")
	}

	rand.Seed(time.Now().Unix())
	id := rand.Intn(0xffffff)
	for _, url := range conf.MQTT.Server {

		go func(url string, idSufix int) {
			opts := paho.NewClientOptions()
			opts.AddBroker(url)
			opts.SetUsername(conf.MQTT.Username)
			opts.SetPassword(conf.MQTT.Password)
			opts.SetCleanSession(conf.MQTT.CleanSession)
			opts.SetClientID(fmt.Sprintf("%s%d", conf.MQTT.ClientID, idSufix))
			opts.SetOnConnectHandler(onConnected)
			opts.SetConnectionLostHandler(onConnectionLost)

			tlsconfig, err := newTLSConfig(conf.MQTT.CACert, conf.MQTT.TLSCert, conf.MQTT.TLSKey)
			if err != nil {
				log.WithError(err).WithFields(log.Fields{
					"ca_cert":  conf.MQTT.CACert,
					"tls_cert": conf.MQTT.TLSCert,
					"tls_key":  conf.MQTT.TLSKey,
				}).Fatal("mqtt: error loading mqtt certificate files")
			}
			if tlsconfig != nil {
				opts.SetTLSConfig(tlsconfig)
			}

			log.WithField("server", url).Info("mqtt: connecting to mqtt broker")
			Brokers[url] = paho.NewClient(opts)
			for {
				if token := Brokers[url].Connect(); token.Wait() && token.Error() != nil {
					log.Errorf("mqtt: connecting to mqtt broker [%s] failed, will retry in 2s: %s", url, token.Error())
					time.Sleep(2 * time.Second)
				} else {
					break
				}
			}
		}(url, id)

		id++
	}
	return nil
}

//AddBackend will called after mqtt brokers connected
func AddBackend(bs ...backend.Backend) {
	backends = append(backends, bs...)
}

// Close UnSubscribeTopic and Disconnect brokers
func Close() {
	log.Info("mqtt: closing...")
	var wg sync.WaitGroup

	// UnSubscribe all topic
	wg.Add(len(Brokers))
	for _, broker := range Brokers {

		go func(broker paho.Client) {
			defer wg.Done()
			for _, backend := range backends {
				backend.UnSubscribeTopic(broker)
			}
			broker.Disconnect(200)
		}(broker)

	}
	wg.Wait()

	// close all backend
	for _, b := range backends {
		b.Close()
	}

	log.Info("mqtt: closed")
}

func onConnected(c paho.Client) {
	opreader := c.OptionsReader()
	log.WithFields(log.Fields{
		"broker":    opreader.Servers()[0],
		"client_id": opreader.ClientID(),
	}).Info("mqtt: connected to mqtt server")

	for _, b := range backends {
		if err := b.SubscribeTopics(c); err != nil {
			log.WithError(err).Errorf("mqtt: subscribe error, backend %s", b.Type())
		}
	}
}

func onConnectionLost(c paho.Client, reason error) {
	opreader := c.OptionsReader()
	log.WithFields(log.Fields{
		"broker":    opreader.Servers()[0],
		"client_id": opreader.ClientID(),
	}).Error("mqtt: connected lost")
}

func newTLSConfig(cafile, certFile, certKeyFile string) (*tls.Config, error) {
	if cafile == "" && certFile == "" && certKeyFile == "" {
		return nil, nil
	}

	tlsConfig := &tls.Config{}

	// Import trusted certificates from CAfile.pem.
	if cafile != "" {
		cacert, err := ioutil.ReadFile(cafile)
		if err != nil {
			log.WithError(err).Error("mqtt: could not load ca certificate")
			return nil, err
		}
		certpool := x509.NewCertPool()
		certpool.AppendCertsFromPEM(cacert)

		tlsConfig.RootCAs = certpool // RootCAs = certs used to verify server cert.
	}

	// Import certificate and the key
	if certFile != "" && certKeyFile != "" {
		kp, err := tls.LoadX509KeyPair(certFile, certKeyFile)
		if err != nil {
			log.WithError(err).Error("mqtt: could not load mqtt tls key-pair")
			return nil, err
		}
		tlsConfig.Certificates = []tls.Certificate{kp}
	}

	return tlsConfig, nil
}
