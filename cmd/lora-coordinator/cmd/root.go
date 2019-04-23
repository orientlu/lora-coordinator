package cmd

import (
	"strings"
	"time"

	"github.com/orientlu/lora-coordinator/internal/config"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var version string

var pprofSet bool
var pprofPort string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "lora-coordinator",
	Short: "As coordinator",
	Long: `LoRa coordinator maintain the mapping of gateway and mqtt broker
	so that lora server can select which mqtt broker to publich downlik topic`,
	RunE: run, // root_run.go
}

// Execute execute the root command, call by main.main
func Execute(v string) {
	version = v
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// defien flags
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "set the path of configuration file")
	rootCmd.PersistentFlags().Int("log-level", 4, "trace=6, debug=5, info=4, error=2, fatal=1, panic=0")
	rootCmd.PersistentFlags().Bool("log-report", false, "show called method detial in log field for debug")
	rootCmd.Flags().BoolVar(&pprofSet, "pprof", false, "enable pprof, default url http://127.0.0.1:9876/debug/pprof/")
	rootCmd.Flags().StringVar(&pprofPort, "pprof-port", "9876", "url: http://127.0.0.1:PORT/debug/pprof/")

	// bind viper config
	viper.BindPFlag("general.log_level", rootCmd.PersistentFlags().Lookup("log-level"))
	viper.BindPFlag("general.log_report", rootCmd.PersistentFlags().Lookup("log-report"))

	// set defaut config
	viper.SetDefault("general.log_level", 4)
	viper.SetDefault("general.log_report", false)
	viper.SetDefault("redis.url", "redis://localhost:6379")
	viper.SetDefault("redis.max_idle", 10)
	viper.SetDefault("redis.max_active", 100)
	viper.SetDefault("redis.idle_timeout", 5*time.Minute)
	viper.SetDefault("mqtt.server", []string{"tcp://localhost:1883"})
	viper.SetDefault("mqtt.clean_session", true)
	viper.SetDefault("mqtt.client_id", "coordinator-")
	viper.SetDefault("backend.gateway.notify_topic_template", "gateway/notify/+")
	viper.SetDefault("backend.gateway.notify_topic_mac_event", "mac")
	viper.SetDefault("backend.gateway.notify_topic_mac_event_redis_prefix", "mac_event_")
	viper.SetDefault("backend.gateway.notify_topic_mac_event_redis_expires", 30)
	viper.SetDefault("backend.gateway.notify_topic_storage_coroutine_number", 2)

	// add new command
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(configCmd)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		// Search config in . home /etc/lora-coordinator directory with name ".lora-coordinator" (without extension).
		// lora-coordinator.json lora-coordinator.toml ..
		viper.SetConfigName("lora-coordinator")
		viper.AddConfigPath(".")
		viper.AddConfigPath("$HOME")
		viper.AddConfigPath("/etc/lora-coordinator")
	}

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		switch err.(type) {
		case viper.ConfigFileNotFoundError:
			log.Warning("No configuration file found, using default.")
		default:
			log.WithError(err).Fatal("read configuration file error")
		}
	} else {
		log.Println("Using config file:", viper.ConfigFileUsed())
	}

	// read in environment variables that match
	viper.SetEnvPrefix("LORA_COORDINATOR")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	if err := viper.Unmarshal(&config.C); err != nil {
		log.WithError(err).Fatal("unmarshal config error")
	}
}
