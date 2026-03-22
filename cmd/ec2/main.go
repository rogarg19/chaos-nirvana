package main

import (
	"github.com/rogarg19/chaos-nirvana/pkg/ec2"
)

func main() {
	ec2ChaosClient := ec2.New()
	ec2ChaosClient.Start()
}