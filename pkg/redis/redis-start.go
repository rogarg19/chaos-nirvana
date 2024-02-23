package redis

import (
	"github.com/gargrohit2523/chaos-nirvana/pkg/config"
)

func Start() {
	config := config.LoadConfig()

	for i := 0; i < config.RedisConfig.Options.Connections; i++ {
		floodRedis()
	}
}

func floodRedis() {

}
