package config

import (
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

// App holds the application-wide configuration.
type App struct {
	ServiceName string      `mapstructure:"serviceName"`
	Environment string      `mapstructure:"appEnvironment"`
	HTTPServer  *HTTPServer `mapstructure:"httpServer"`
}

// HTTPServer holds the configuration for the HTTP server.
type HTTPServer struct {
	Host string `mapstructure:"httpHost"`
	Port string `mapstructure:"httpPort"`
}

// New returns the application-wide configuration.
func New() (*App, error) {
	viper.SetConfigName("config.yaml")
	viper.AddConfigPath("./config")
	viper.AutomaticEnv()
	viper.SetConfigType("yaml")

	var cfg App

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Watch watches for changes in the configuration file and updates the configuration accordingly.
// Stops watching if an error occurs while unmarshalling to avoid weird behavior.
func (x *App) Watch() {
	for {
		time.Sleep(10 * time.Second)
		viper.WatchConfig()

		if err := viper.Unmarshal(&x); err != nil {
			log.Error().Msgf("config: Error parsing config file, aborting... %s", err)
			return
		}
	}
}

// Addr returns the address of the configured HTTP server.
func (x *HTTPServer) Addr() string {
	return fmt.Sprintf("%s:%s", x.Host, x.Port)
}
