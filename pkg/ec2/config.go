package ec2

import (
	"encoding/json"
	"os"
)

type Configuration struct {
	EC2Config EC2Config `json:"ec2"`
}

type EC2Config struct {
	EnableHighCPU  bool   `json:"enableHighCpu"`
	EnableFullDisk bool   `json:"enableFullDisk"`
	DiskFillPath   string `json:"diskFillPath"` // Path to fill disk, e.g., "/tmp/chaos"
	CPUCores       int    `json:"cpuCores"`     // Number of cores to utilize
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
