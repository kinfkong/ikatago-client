package config

import (
	"log"

	"github.com/kinfkong/ikatago-client/utils"
	"github.com/spf13/viper"
)

var config *viper.Viper

// Init inits the config
func Init(configFile *string) {
	config = viper.New()

	if configFile != nil {
		config.SetConfigFile(*configFile)
	}

	config.SetDefault("world.url", utils.WorldURL)

	if configFile != nil {
		err := config.ReadInConfig()
		if err != nil {
			log.Fatal("error on parsing configuration file", err)
		}
	}
}

// GetConfig gets the configuration
func GetConfig() *viper.Viper {
	return config
}
