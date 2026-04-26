package nlp

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/rogarg19/chaos-nirvana/pkg/scenario"
)

var (
	percentPattern    = regexp.MustCompile(`(?i)(\d{1,3})\s*%`)
	podCountPattern   = regexp.MustCompile(`(?i)(?:kill|delete|terminate|restart)\s+(\d+)\s+pods?`)
	namespacePattern  = regexp.MustCompile(`(?i)(?:in|within)\s+(?:namespace|ns)\s+([a-z0-9._-]+)`)
	servicePattern    = regexp.MustCompile(`(?i)(?:service|svc)\s+([a-z0-9._-]+)`)
	selectorPattern   = regexp.MustCompile(`(?i)(?:selector|label selector)\s+([a-z0-9._/-]+=[a-z0-9._/-]+)`)
	durationPattern   = regexp.MustCompile(`(?i)(?:for|more than|over)\s+(\d+)\s*(second|seconds|sec|secs|minute|minutes|min|mins|hour|hours|hr|hrs)`)
	connectionPattern = regexp.MustCompile(`(?i)(\d+)\s+(?:connections|clients)`)
	coresPattern      = regexp.MustCompile(`(?i)(\d+)\s+(?:cores|cpu cores)`)
	pathPattern       = regexp.MustCompile(`(?i)(?:path|at|under)\s+(/[^\s]+)`)
)

func Parse(prompt string) (scenario.Scenario, error) {
	normalized := strings.ToLower(strings.TrimSpace(prompt))
	if normalized == "" {
		return scenario.Scenario{}, fmt.Errorf("prompt cannot be empty")
	}

	s := scenario.Scenario{
		Source: prompt,
		Parameters: scenario.Parameters{
			InfoIntervalSecs: 30,
		},
	}
	s.Duration = parseDuration(normalized)
	if s.Duration > 0 {
		s.DurationText = s.Duration.String()
	}

	switch {
	case mentionsAny(normalized, "kubernetes", "k8s", "pod", "pods"):
		return parseKubernetes(prompt, normalized, s)
	case mentionsAny(normalized, "elasticache"):
		s.Service = scenario.ServiceElastiCache
		return parseRedisLike(normalized, s)
	case mentionsAny(normalized, "redis"):
		s.Service = scenario.ServiceRedis
		return parseRedisLike(normalized, s)
	case mentionsAny(normalized, "ec2", "instance", "vm"):
		s.Service = scenario.ServiceEC2
		return parseEC2(normalized, s)
	default:
		return scenario.Scenario{}, fmt.Errorf("could not infer target service from prompt")
	}
}

func parseKubernetes(original, normalized string, s scenario.Scenario) (scenario.Scenario, error) {
	if !mentionsAny(normalized, "kill", "delete", "terminate", "restart") || !mentionsAny(normalized, "pod", "pods") {
		return scenario.Scenario{}, fmt.Errorf("kubernetes prompt must describe killing or deleting pods")
	}

	s.Service = scenario.ServiceKubernetes
	s.Action = scenario.ActionKubernetesKillPods
	s.Parameters.Count = parseInt(podCountPattern, normalized, 0)
	if s.Parameters.Count == 0 {
		s.Parameters.Percent = parsePercent(normalized, 100)
	}
	s.Target.Namespace = firstMatch(namespacePattern, normalized, "default")
	s.Target.Selector = firstMatch(selectorPattern, normalized, "")

	if svc := firstMatch(servicePattern, strings.ToLower(original), ""); svc != "" {
		s.Target.Service = svc
	}
	if s.Target.Selector == "" && s.Target.Service != "" {
		s.Target.Selector = "app=" + s.Target.Service
	}
	if s.Target.Selector == "" {
		return scenario.Scenario{}, fmt.Errorf("kubernetes pod scenario needs a service name or label selector")
	}
	return s, nil
}

func parseRedisLike(normalized string, s scenario.Scenario) (scenario.Scenario, error) {
	s.Parameters.Cluster = strings.Contains(normalized, "cluster")
	s.Parameters.CPUPercent = parsePercent(normalized, 0)
	s.Parameters.Connections = parseInt(connectionPattern, normalized, 10)

	if mentionsAny(normalized, "cpu", "load", "spike") {
		s.Action = scenario.ActionRedisCPUSpike
		if s.Parameters.CPUPercent == 0 {
			s.Parameters.CPUPercent = 90
		}
		if s.Parameters.Connections < 10 {
			s.Parameters.Connections = 10
		}
		return s, nil
	}

	if mentionsAny(normalized, "connection", "connections", "clients", "flood") {
		s.Action = scenario.ActionRedisConnectionFlood
		return s, nil
	}

	return scenario.Scenario{}, fmt.Errorf("redis prompt must describe CPU load/spike or connection flooding")
}

func parseEC2(normalized string, s scenario.Scenario) (scenario.Scenario, error) {
	switch {
	case mentionsAny(normalized, "cpu", "load", "spike"):
		s.Action = scenario.ActionEC2HighCPU
		s.Parameters.CPUCores = parseInt(coresPattern, normalized, 0)
		return s, nil
	case mentionsAny(normalized, "disk", "storage"):
		s.Action = scenario.ActionEC2FullDisk
		s.Parameters.DiskFillPath = firstMatch(pathPattern, normalized, "")
		return s, nil
	default:
		return scenario.Scenario{}, fmt.Errorf("ec2 prompt must describe CPU or disk chaos")
	}
}

func parsePercent(text string, fallback int) int {
	value := parseInt(percentPattern, text, fallback)
	if value < 0 {
		return 0
	}
	if value > 100 {
		return 100
	}
	return value
}

func parseDuration(text string) time.Duration {
	matches := durationPattern.FindStringSubmatch(text)
	if len(matches) != 3 {
		return 0
	}
	value, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0
	}
	switch matches[2] {
	case "second", "seconds", "sec", "secs":
		return time.Duration(value) * time.Second
	case "minute", "minutes", "min", "mins":
		return time.Duration(value) * time.Minute
	case "hour", "hours", "hr", "hrs":
		return time.Duration(value) * time.Hour
	default:
		return 0
	}
}

func parseInt(pattern *regexp.Regexp, text string, fallback int) int {
	matches := pattern.FindStringSubmatch(text)
	if len(matches) < 2 {
		return fallback
	}
	value, err := strconv.Atoi(matches[1])
	if err != nil {
		return fallback
	}
	return value
}

func firstMatch(pattern *regexp.Regexp, text string, fallback string) string {
	matches := pattern.FindStringSubmatch(text)
	if len(matches) < 2 {
		return fallback
	}
	return matches[1]
}

func mentionsAny(text string, words ...string) bool {
	for _, word := range words {
		if strings.Contains(text, word) {
			return true
		}
	}
	return false
}
