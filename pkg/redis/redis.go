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

	log.Printf("%+v", config)

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
		time.Sleep(5 * time.Microsecond)
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

	client := getClient(config)
	defer client.Close()

	ticker := time.NewTicker(10 * time.Microsecond)

	innerCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// simulate long running connections

	for {
		select {
		case <-ticker.C:
			_, err := client.Keys(innerCtx, "foo*").Result()

			if err != nil {
				log.Println(err)
			}
		case <-ctx.Done():
			return
		}
	}
}
