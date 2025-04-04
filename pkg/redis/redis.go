package redis

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

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

	rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := 0; i < config.RedisConfig.Options.Connections; i++ {
		time.Sleep(5 * time.Microsecond)
		wg.Add(1)
		go floodRedis(&wg, config, ctx)
	}

	wg.Add(1)
	go redisInfo(&wg, config, ctx)

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

	ticker := time.NewTicker(50 * time.Millisecond)

	var randomKey string

	if len(config.RedisConfig.CustomKeyPrefix) > 0 {
		randomKey = fmt.Sprintf("%s_%s%s", config.RedisConfig.CustomKeyPrefix, randSeq(5), "*")
	} else {
		randomKey = fmt.Sprintf("%s%s", randSeq(5), "*")
	}
	// simulate long running connections

	for {
		select {
		case <-ticker.C:
			if config.RedisConfig.IsKeysCommandEnabled {
				_, keyErr := client.Keys(ctx, randomKey).Result()
				log.Println(keyErr)
			} else {
				_, getErr := client.HGetAll(ctx, randomKey).Result()

				if getErr == redis.Nil || getErr == nil {
					continue
				}
				log.Println(getErr)
			}
		case <-ctx.Done():
			return
		}
	}
}

func redisInfo(wg *sync.WaitGroup, config Configuration, ctx context.Context) {
	defer wg.Done()

	client := getClient(config)
	defer client.Close()

	var Red = "\033[31m"
	const colorNone = "\033[0m"

	ticker := time.NewTicker(time.Duration(config.RedisConfig.InfoInterval) * time.Second)

	for {
		select {
		case <-ticker.C:
			info := client.Info(ctx, "clients", "cpu")
			fmt.Println("*****")
			fmt.Fprintf(os.Stdout, "%s %s", Red, info)
			fmt.Println("*****")
			fmt.Fprintf(os.Stdout, "%s", colorNone)
		case <-ctx.Done():
			return
		}
	}
}

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
