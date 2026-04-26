package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/rogarg19/chaos-nirvana/pkg/ec2"
	"github.com/rogarg19/chaos-nirvana/pkg/elasticache"
	"github.com/rogarg19/chaos-nirvana/pkg/kubernetes"
	"github.com/rogarg19/chaos-nirvana/pkg/nlp"
	"github.com/rogarg19/chaos-nirvana/pkg/redis"
	"github.com/rogarg19/chaos-nirvana/pkg/scenario"
)

func main() {
	fmt.Println("Starting execution")
	testType := flag.String("type", "", "Specify test to be executed: redis, elasticache, ec2")
	configPath := flag.String("config", "config.json", "configuration file for chaos")
	prompt := flag.String("prompt", "", "plain English chaos scenario to execute")
	dryRun := flag.Bool("dry-run", false, "parse and print the scenario without executing it")

	flag.Parse()

	if *prompt != "" {
		parsed, err := nlp.Parse(*prompt)
		if err != nil {
			log.Fatal(err)
		}
		if *testType != "" {
			parsed.Service = scenario.Service(*testType)
		}
		if err := printScenario(parsed); err != nil {
			log.Fatal(err)
		}
		if *dryRun {
			return
		}
		if err := executeScenario(context.Background(), parsed, *configPath); err != nil {
			log.Fatal(err)
		}
		return
	}

	if *testType == "" {
		log.Fatal("Please specify a test type using -type flag or a scenario using -prompt")
	}

	switch *testType {
	case "redis":
		redis.Run(redis.LoadConfig(*configPath), 0)
	case "elasticache":
		elasticache.Run(elasticache.LoadConfig(*configPath), 0)
	case "ec2":
		ec2.Run(ec2.LoadConfig(*configPath), 0)
	default:
		log.Fatalf("Unknown test type: %s", *testType)
	}
}

func printScenario(s scenario.Scenario) error {
	output, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	fmt.Printf("Parsed scenario:\n%s\n", output)
	return nil
}

func executeScenario(ctx context.Context, s scenario.Scenario, configPath string) error {
	switch s.Service {
	case scenario.ServiceRedis:
		config := redis.Configuration{}
		if fileExists(configPath) {
			config = redis.LoadConfig(configPath)
		}
		redis.Run(redis.ConfigFromScenario(config, s), s.Duration)
	case scenario.ServiceElastiCache:
		config := elasticache.Configuration{}
		if fileExists(configPath) {
			config = elasticache.LoadConfig(configPath)
		}
		elasticache.Run(elasticache.ConfigFromScenario(config, s), s.Duration)
	case scenario.ServiceEC2:
		config := ec2.Configuration{}
		if fileExists(configPath) {
			config = ec2.LoadConfig(configPath)
		}
		ec2.Run(ec2.ConfigFromScenario(config, s), s.Duration)
	case scenario.ServiceKubernetes:
		return kubernetes.New().Run(ctx, s)
	default:
		return fmt.Errorf("unsupported service %q", s.Service)
	}
	return nil
}

func fileExists(path string) bool {
	if path == "" {
		return false
	}
	_, err := os.Stat(path)
	return err == nil
}
