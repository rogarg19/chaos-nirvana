package elasticache

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
	Del(context.Context, ...string) *redis.IntCmd
	Keys(context.Context, string) *redis.StringSliceCmd
	HGetAll(ctx context.Context, key string) *redis.MapStringStringCmd
	Info(ctx context.Context, sections ...string) *redis.StringCmd
	Eval(ctx context.Context, script string, keys []string, args ...interface{}) *redis.Cmd
	Close() error
}

func getRedisClusterClient(config Configuration) *redis.ClusterClient {
	client := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:    []string{fmt.Sprintf("%s:%s", config.ElastiCacheConfig.Host, strconv.Itoa(config.ElastiCacheConfig.Port))},
		Password: config.ElastiCacheConfig.Password,
		TLSConfig: &tls.Config{
			InsecureSkipVerify: config.ElastiCacheConfig.Options.Tls.InsecureSkipVerify,
		},
		ReadTimeout:  time.Duration(config.ElastiCacheConfig.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(config.ElastiCacheConfig.WriteTimeout) * time.Second,
		DialTimeout:  time.Duration(config.ElastiCacheConfig.DialTimeout) * time.Second,
		PoolSize:     config.ElastiCacheConfig.PoolSize,
	})
	return client
}

func getRedisClient(config Configuration) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", config.ElastiCacheConfig.Host, strconv.Itoa(config.ElastiCacheConfig.Port)),
		Password:     config.ElastiCacheConfig.Password,
		DB:           config.ElastiCacheConfig.Db,
		ReadTimeout:  time.Duration(config.ElastiCacheConfig.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(config.ElastiCacheConfig.WriteTimeout) * time.Second,
		DialTimeout:  time.Duration(config.ElastiCacheConfig.DialTimeout) * time.Second,
		PoolSize:     config.ElastiCacheConfig.PoolSize,
	})

	client.Info(context.Background())

	return client
}

func getClient(config Configuration) redisClient {
	var client redisClient
	if config.ElastiCacheConfig.IsCluster {
		client = getRedisClusterClient(config)
	} else {
		client = getRedisClient(config)
	}
	return client
}
