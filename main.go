package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/rogarg19/chaos-nirvana/pkg/ec2"
	"github.com/rogarg19/chaos-nirvana/pkg/elasticache"
	"github.com/rogarg19/chaos-nirvana/pkg/redis"
)

func main() {
	fmt.Println("Starting execution")
	testType := flag.String("type", "", "Specify test to be executed: redis, elasticache, ec2")

	flag.Parse()

	if *testType == "" {
		log.Fatal("Please specify a test type using -type flag")
	}

	switch *testType {
	case "redis":
		redisChaos := redis.New()
		redisChaos.Start()
	case "elasticache":
		elastiCacheChaos := elasticache.New()
		elastiCacheChaos.Start()
	case "ec2":
		ec2Chaos := ec2.New()
		ec2Chaos.Start()
	default:
		log.Fatalf("Unknown test type: %s", *testType)
	}
}
