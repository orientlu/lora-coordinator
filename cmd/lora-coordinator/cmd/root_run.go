package cmd

import (
	"net/http"
	// pprof
	_ "net/http/pprof"

	"os"
	"os/signal"
	"syscall"

	"github.com/gomodule/redigo/redis"
	"github.com/orientlu/lora-coordinator/internal/backend/gateway"
	"github.com/orientlu/lora-coordinator/internal/config"
	"github.com/orientlu/lora-coordinator/internal/mqtt"
	"github.com/orientlu/lora-coordinator/internal/storage"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// entry
func run(cmd *cobra.Command, args []string) error {
	log.Println("Start loRa-coordinator, Version: ", version)

	tasks := []func() error{
		setLog,
		setStorage,
		connectMqttBroker,
		startGatewaySaveServer,
	}

	for _, t := range tasks {
		if err := t(); err != nil {
			log.Fatal(err)
		}
	}

	if pprofSet {
		log.WithField("url", "127.0.0.1:"+pprofPort).Warning("running in pprof model")
		go http.ListenAndServe(":"+pprofPort, nil)
	}

	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	log.WithField("signal", <-sigChan).Info("signal received")

	exitChan := make(chan struct{})
	go func() {
		log.Warning("Stopping Server...")

		disconnectMqtt()

		stopGatewaySaveServer()

		exitChan <- struct{}{}
	}()
	select {
	case <-exitChan: // wait
	case s := <-sigChan:
		log.WithField("signal", s).Info("signal received again, stopping immediately")
	}

	return nil
}

func setLog() error {
	log.SetLevel(log.Level(uint8(config.C.General.LogLevel)))
	log.Println("set loglevel", viper.GetInt("general.log_level"))

	log.SetFormatter(&log.TextFormatter{
		//DisableColors: true,
		FullTimestamp: true,
	})
	log.SetReportCaller(config.C.General.LogReport)
	return nil
}

func setStorage() error {
	if err := storage.Setup(config.C); err != nil {
		log.WithError(err).Error("setup storage error")
		return errors.Wrap(err, "setup storage error")
	}

	redisConn := storage.RedisPool().Get()
	defer redisConn.Close()
	hello := "hello"
	if str, err := redis.String(redisConn.Do("SET", "lora-coordinator", hello)); err != nil {
		log.Errorf("root_run: redis set err, retrun: %s, err: %s", str, err)
	}
	str, err := redis.String(redisConn.Do("GET", "lora-coordinator"))
	if str != "hello" || err != nil {
		log.Errorf("root_run: redis get err, retrun: %s, err: %s", str, err)
	}

	return nil
}

func connectMqttBroker() error {

	// new backends and Add
	mqtt.AddBackend(gateway.GetBackend())

	if err := mqtt.Setup(config.C); err != nil {
		log.Println(err)
	}
	return nil
}

func disconnectMqtt() {
	mqtt.Close()
}

func startGatewaySaveServer() error {
	gateway.Start()
	return nil
}

func stopGatewaySaveServer() {
	gateway.Close()
}
