package redis

import (
	"encoding/json"
	"os"
)

type Configuration struct {
	RedisConfig RedisConfig `json:"redis"`
}

type RedisConfig struct {
	Host     string  `json:"host"`
	Port     int     `json:"port"`
	Password string  `json:"password"`
	Db       string  `json:"db"`
	Options  Options `json:"options"`
}

type Options struct {
	Connections int `json:"connections"`
	Tls         Tls `json:"tls"`
}

type Tls struct {
	InsecureSkipVerify bool `json:"insecure"`
}

func loadConfig(path *string) Configuration {
	var config Configuration = Configuration{}
	raw, err := os.ReadFile(*path)

	if err != nil {
		panic(err)
	}

	unmarshalError := json.Unmarshal(raw, &config)

	if unmarshalError != nil {
		panic(unmarshalError)
	}
	return config
}
