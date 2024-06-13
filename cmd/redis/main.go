package main

import (
	"github.com/rogarg19/chaos-nirvana/pkg/redis"
)

func main() {
	redisChaosClient := redis.New()
	redisChaosClient.Start()
}
