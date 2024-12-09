package config

import (
	"time"

	"github.com/spf13/viper"
)

func setupConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	viper.AddConfigPath("/etc/bb/")

	viper.SetDefault("push.telegram.token", "")
	viper.SetDefault("push.telegram.timeout", time.Duration(5)*time.Second)
}

func LoadConfig() {
	setupConfig()
	viper.SafeWriteConfig()
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
}
