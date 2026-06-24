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
	confirmed := flag.Bool("yes", false, "confirm execution of disruptive chaos actions")
	preflight := flag.Bool("preflight", true, "run target validation before supported disruptive actions")
	maxPods := flag.Int("max-pods", kubernetes.DefaultMaxPods, "maximum Kubernetes pods this run may delete")
	maxDiskMB := flag.Int("max-disk-mb", ec2.DefaultMaxDiskFillMB, "maximum disk space EC2 disk chaos may write")
	keepArtifacts := flag.Bool("keep-artifacts", false, "keep temporary chaos artifacts instead of cleaning them up")

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
		if err := printPreflight(context.Background(), parsed, *preflight, *maxPods); err != nil {
			log.Fatal(err)
		}
		if *dryRun {
			return
		}
		if err := requireConfirmation(parsed, *confirmed); err != nil {
			log.Fatal(err)
		}
		if err := executeScenario(context.Background(), parsed, *configPath, executionOptions{
			maxPods:       *maxPods,
			maxDiskMB:     *maxDiskMB,
			keepArtifacts: *keepArtifacts,
		}); err != nil {
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
		config := elasticache.LoadConfig(*configPath)
		config.ElastiCacheConfig.KeepLargeKey = *keepArtifacts
		if config.ElastiCacheConfig.EnableLargeKey && !*confirmed {
			log.Fatal("ElastiCache large key chaos requires -yes")
		}
		elasticache.Run(config, 0)
	case "ec2":
		config := ec2.ApplySafetyDefaults(ec2.LoadConfig(*configPath), *maxDiskMB, *keepArtifacts)
		if config.EC2Config.EnableFullDisk && !*confirmed {
			log.Fatal("EC2 full disk chaos requires -yes")
		}
		ec2.Run(config, 0)
	default:
		log.Fatalf("Unknown test type: %s", *testType)
	}
}

type executionOptions struct {
	maxPods       int
	maxDiskMB     int
	keepArtifacts bool
}

func printScenario(s scenario.Scenario) error {
	output, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	fmt.Printf("Parsed scenario:\n%s\n", output)
	return nil
}

func printPreflight(ctx context.Context, s scenario.Scenario, enabled bool, maxPods int) error {
	if !enabled || s.Service != scenario.ServiceKubernetes {
		return nil
	}
	plan, err := kubernetes.New().Plan(ctx, s, maxPods)
	if err != nil {
		return err
	}
	output, err := json.MarshalIndent(plan, "", "  ")
	if err != nil {
		return err
	}
	fmt.Printf("Kubernetes preflight:\n%s\n", output)
	return nil
}

func requireConfirmation(s scenario.Scenario, confirmed bool) error {
	if confirmed {
		return nil
	}
	switch s.Action {
	case scenario.ActionKubernetesKillPods, scenario.ActionEC2FullDisk:
		return fmt.Errorf("%s requires -yes", s.Action)
	default:
		return nil
	}
}

func executeScenario(ctx context.Context, s scenario.Scenario, configPath string, opts executionOptions) error {
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
		config = elasticache.ConfigFromScenario(config, s)
		config.ElastiCacheConfig.KeepLargeKey = opts.keepArtifacts
		elasticache.Run(config, s.Duration)
	case scenario.ServiceEC2:
		config := ec2.Configuration{}
		if fileExists(configPath) {
			config = ec2.LoadConfig(configPath)
		}
		config = ec2.ConfigFromScenario(config, s)
		config = ec2.ApplySafetyDefaults(config, opts.maxDiskMB, opts.keepArtifacts)
		ec2.Run(config, s.Duration)
	case scenario.ServiceKubernetes:
		return kubernetes.New().RunWithOptions(ctx, s, kubernetes.Options{MaxPods: opts.maxPods})
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
