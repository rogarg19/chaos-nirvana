package scenario

import "time"

type Service string

const (
	ServiceRedis       Service = "redis"
	ServiceElastiCache Service = "elasticache"
	ServiceEC2         Service = "ec2"
	ServiceKubernetes  Service = "kubernetes"
)

type Action string

const (
	ActionRedisConnectionFlood Action = "redis_connection_flood"
	ActionRedisCPUSpike        Action = "redis_cpu_spike"
	ActionEC2HighCPU           Action = "ec2_high_cpu"
	ActionEC2FullDisk          Action = "ec2_full_disk"
	ActionKubernetesKillPods   Action = "kubernetes_kill_pods"
)

type Scenario struct {
	Service      Service       `json:"service"`
	Action       Action        `json:"action"`
	Target       Target        `json:"target"`
	Parameters   Parameters    `json:"parameters"`
	Duration     time.Duration `json:"-"`
	DurationText string        `json:"duration,omitempty"`
	Source       string        `json:"source,omitempty"`
}

type Target struct {
	Namespace string `json:"namespace,omitempty"`
	Service   string `json:"service,omitempty"`
	Selector  string `json:"selector,omitempty"`
	Host      string `json:"host,omitempty"`
}

type Parameters struct {
	Percent          int    `json:"percent,omitempty"`
	Count            int    `json:"count,omitempty"`
	Connections      int    `json:"connections,omitempty"`
	CPUPercent       int    `json:"cpuPercent,omitempty"`
	CPUCores         int    `json:"cpuCores,omitempty"`
	LargeKeySizeMB   int    `json:"largeKeySizeMb,omitempty"`
	DiskFillPath     string `json:"diskFillPath,omitempty"`
	UseKeysCommand   bool   `json:"useKeysCommand,omitempty"`
	Cluster          bool   `json:"cluster,omitempty"`
	InfoIntervalSecs int    `json:"infoIntervalSecs,omitempty"`
}
