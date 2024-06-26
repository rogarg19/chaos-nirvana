package redis

import (
	"context"
	"crypto/tls"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type redisClient interface {
	Ping(context.Context) *redis.StatusCmd
	Get(context.Context, string) *redis.StringCmd
	Set(context.Context, string, interface{}, time.Duration) *redis.StatusCmd
	Keys(context.Context, string) *redis.StringSliceCmd
	HGetAll(ctx context.Context, key string) *redis.MapStringStringCmd
	Close() error
}

func getRedisClusterClient(config Configuration) *redis.ClusterClient {
	client := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:    []string{fmt.Sprintf("%s:%s", config.RedisConfig.Host, strconv.Itoa(config.RedisConfig.Port))},
		Password: "",
		TLSConfig: &tls.Config{
			InsecureSkipVerify: config.RedisConfig.Options.Tls.InsecureSkipVerify,
		},
		ReadTimeout:  time.Duration(config.RedisConfig.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(config.RedisConfig.WriteTimeout) * time.Second,
		DialTimeout:  time.Duration(config.RedisConfig.WriteTimeout) * time.Second,
		PoolSize:     config.RedisConfig.PoolSize,
	})
	return client
}

func getRedisClient(config Configuration) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", config.RedisConfig.Host, strconv.Itoa(config.RedisConfig.Port)),
		Password:     "",
		DB:           config.RedisConfig.Db,
		ReadTimeout:  time.Duration(config.RedisConfig.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(config.RedisConfig.WriteTimeout) * time.Second,
		DialTimeout:  time.Duration(config.RedisConfig.WriteTimeout) * time.Second,
		PoolSize:     config.RedisConfig.PoolSize,
	})

	return client
}

func getClient(config Configuration) redisClient {
	var client redisClient
	if config.RedisConfig.IsCluster {
		client = getRedisClusterClient(config)
	} else {
		client = getRedisClient(config)
	}
	return client
}
