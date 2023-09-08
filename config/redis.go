package config

import (
	"context"
	"mofe64/playlistGen/util"
	"time"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client = ConnectRedis()

func ConnectRedis() *redis.Client {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tag := "CONNECT_REDIS_FUNC"
	opt, err := redis.ParseURL("redis://@localhost:6379/1?max_retries=2")
	if err != nil {
		util.ErrorLog.Fatalln(tag+": error parsing connection url", err)
	}
	client := redis.NewClient(opt)
	if err := client.Ping(ctx).Err(); err != nil {
		util.ErrorLog.Fatalln(tag+": error pinging redis connection", err)
	}
	return client
}
