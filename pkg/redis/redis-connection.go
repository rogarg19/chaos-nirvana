package redis

import (
	"crypto/tls"
	"fmt"
	"strconv"

	"github.com/redis/go-redis/v9"
)

func getRedisClient(config Configuration) *redis.ClusterClient {
	client := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:    []string{fmt.Sprintf("%s:%s", config.RedisConfig.Host, strconv.Itoa(config.RedisConfig.Port))},
		Password: "",
		TLSConfig: &tls.Config{
			InsecureSkipVerify: config.RedisConfig.Options.Tls.InsecureSkipVerify,
		},
	})

	return client
}
