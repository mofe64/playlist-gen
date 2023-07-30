package config

import (
	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client = ConnectRedis()

func ConnectRedis() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	return client
}
