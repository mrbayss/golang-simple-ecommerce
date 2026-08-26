package config

import (
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

func NewRedis(config *viper.Viper) *redis.Client {
	redis := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.GetString("redis.host"), config.GetInt("redis.port")),
		Password: config.GetString("redis.password"),
		DB:       config.GetInt("redis.dbname"),
	})

	return redis
}
