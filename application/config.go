package application

import (
	"github.com/spf13/viper"
)

type Config struct {
	ServerPort uint16
	RedisAddr  string
}

func LoadConfig() Config {
	viper.SetDefault("ServerPort", 3000)
	viper.SetDefault("RedisAddr", "localhost:6379")

	viper.AutomaticEnv()

	config := Config{
		ServerPort: uint16(viper.GetUint("ServerPort")),
		RedisAddr:  viper.GetString("RedisAddr"),
	}

	return config
}
