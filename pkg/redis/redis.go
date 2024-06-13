package redis

import (
	"context"
	"flag"
	"log"
	"os"
	"sync"
	"time"
)

type RedisChaos struct{}

func New() *RedisChaos {
	return &RedisChaos{}
}

func (*RedisChaos) Start() {
	var configPath *string = flag.String("config", "config.json", "configuration file for chaos")

	config := loadConfig(configPath)

	var done = make(chan struct{}, 1)

	go func() {
		os.Stdin.Read(make([]byte, 1))
		close(done)
	}()

	var wg sync.WaitGroup

	//create a context that we cancel

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for i := 0; i < config.RedisConfig.Options.Connections; i++ {
		time.Sleep(50 * time.Millisecond)
		wg.Add(1)
		go floodRedis(&wg, config, ctx)
	}

	<-done

	//cancel the child goroutines
	cancel()

	log.Println("waiting for all child goroutines to exit gracefully...")
	wg.Wait()
	log.Println("all goroutines finished.")
}

func floodRedis(wg *sync.WaitGroup, config Configuration, ctx context.Context) {
	defer wg.Done()

	client := getRedisClient(config)
	defer client.Close()

	// simulate long running connections

	ticker := time.NewTicker(100 * time.Millisecond)

	for {
		select {
		case <-ticker.C:
			innerCtx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
			defer cancel()
			// _, err := client.Keys(innerCtx, "foo*").Result()
			val, err := client.Get(innerCtx, "foo*").Result()

			if err != nil {
				log.Println(err)
				continue
			}
			log.Println(val)
		case <-ctx.Done():
			return
		}
	}
}
