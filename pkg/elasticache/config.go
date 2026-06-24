package elasticache

import (
	"encoding/json"
	"os"
)

type Configuration struct {
	ElastiCacheConfig ElastiCacheConfig `json:"elasticache"`
}

type ElastiCacheConfig struct {
	Host                 string  `json:"host"`
	Port                 int     `json:"port"`
	Password             string  `json:"password"`
	Db                   int     `json:"db"`
	Options              Options `json:"options"`
	IsCluster            bool    `json:"isCluster"`
	ReadTimeout          int     `json:"readtimeout"`
	WriteTimeout         int     `json:"writetimeout"`
	InfoInterval         int     `json:"infointerval"`
	DialTimeout          int     `json:"dialtimeout"`
	CustomKeyPrefix      string  `json:"customkeyprefix"`
	PoolSize             int     `json:"poolsize"`
	IsKeysCommandEnabled bool    `json:"iskeyscommandenabled"`
	// Additional for chaos
	EnableCPUSpike  bool `json:"enableCpuSpike"`
	EnableLargeKey  bool `json:"enableLargeKey"`
	LargeKeySize    int  `json:"largeKeySize"` // in MB
	CPUSpikeWorkers int  `json:"cpuSpikeWorkers"`
	KeepLargeKey    bool `json:"keepLargeKey"`
}

type Options struct {
	Connections int `json:"connections"`
	Tls         Tls `json:"tls"`
}

type Tls struct {
	InsecureSkipVerify bool `json:"insecure"`
}

func LoadConfig(path string) Configuration {
	var config Configuration = Configuration{}
	raw, err := os.ReadFile(path)

	if err != nil {
		panic(err)
	}

	unmarshalError := json.Unmarshal(raw, &config)

	if unmarshalError != nil {
		panic(unmarshalError)
	}
	return config
}

func loadConfig(path *string) Configuration {
	return LoadConfig(*path)
}
