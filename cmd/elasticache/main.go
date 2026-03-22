package main

import (
	"github.com/rogarg19/chaos-nirvana/pkg/elasticache"
)

func main() {
	elastiCacheChaosClient := elasticache.New()
	elastiCacheChaosClient.Start()
}