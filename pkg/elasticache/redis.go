package elasticache

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

type ElastiCacheChaos struct{}

func New() *ElastiCacheChaos {
	return &ElastiCacheChaos{}
}

func (*ElastiCacheChaos) Start() {
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

	for i := 0; i < config.ElastiCacheConfig.Options.Connections; i++ {
		time.Sleep(5 * time.Microsecond)
		wg.Add(1)
		go floodElastiCache(&wg, config, ctx)
	}

	wg.Add(1)
	go elasticacheInfo(&wg, config, ctx)

	// Additional chaos
	if config.ElastiCacheConfig.EnableLargeKey {
		wg.Add(1)
		go injectLargeKey(&wg, config, ctx)
	}

	if config.ElastiCacheConfig.EnableCPUSpike {
		for i := 0; i < 5; i++ { // Run multiple CPU spike goroutines
			wg.Add(1)
			go cpuSpike(&wg, config, ctx)
		}
	}

	<-done

	//cancel the child goroutines
	cancel()

	log.Println("waiting for all child goroutines to exit gracefully...")
	wg.Wait()
	log.Println("all goroutines finished.")
}

func floodElastiCache(wg *sync.WaitGroup, config Configuration, ctx context.Context) {
	defer wg.Done()

	client := getClient(config)
	defer client.Close()

	ticker := time.NewTicker(50 * time.Millisecond)

	var randomKey string

	if len(config.ElastiCacheConfig.CustomKeyPrefix) > 0 {
		randomKey = fmt.Sprintf("%s_%s%s", config.ElastiCacheConfig.CustomKeyPrefix, randSeq(5), "*")
	} else {
		randomKey = fmt.Sprintf("%s%s", randSeq(5), "*")
	}
	// simulate long running connections

	for {
		select {
		case <-ticker.C:
			if config.ElastiCacheConfig.IsKeysCommandEnabled {
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

func elasticacheInfo(wg *sync.WaitGroup, config Configuration, ctx context.Context) {
	defer wg.Done()

	client := getClient(config)
	defer client.Close()

	var Red = "\033[31m"
	const colorNone = "\033[0m"

	ticker := time.NewTicker(time.Duration(config.ElastiCacheConfig.InfoInterval) * time.Second)

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

func injectLargeKey(wg *sync.WaitGroup, config Configuration, ctx context.Context) {
	defer wg.Done()

	client := getClient(config)
	defer client.Close()

	// Create a large value, e.g., 10MB or as configured
	size := config.ElastiCacheConfig.LargeKeySize * 1024 * 1024 // MB to bytes
	if size == 0 {
		size = 10 * 1024 * 1024 // default 10MB
	}
	largeValue := strings.Repeat("A", size)

	key := "chaos_large_key"

	err := client.Set(ctx, key, largeValue, 0).Err()
	if err != nil {
		log.Printf("Failed to set large key: %v", err)
		return
	}

	log.Printf("Injected large key '%s' with size %d MB", key, config.ElastiCacheConfig.LargeKeySize)

	// Keep it alive, perhaps periodically update or just hold
	ticker := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-ticker.C:
			// Optionally update the key to keep it active
			client.Set(ctx, key, largeValue, 0)
		case <-ctx.Done():
			return
		}
	}
}

func cpuSpike(wg *sync.WaitGroup, config Configuration, ctx context.Context) {
	defer wg.Done()

	client := getClient(config)
	defer client.Close()

	// CPU intensive Lua script: calculate fibonacci or something
	script := `
	local function fib(n)
		if n <= 1 then return n end
		return fib(n-1) + fib(n-2)
	end
	return fib(35)  -- Adjust for CPU intensity
	`

	ticker := time.NewTicker(100 * time.Millisecond)

	for {
		select {
		case <-ticker.C:
			_, err := client.Eval(ctx, script, []string{}, nil).Result()
			if err != nil {
				log.Printf("CPU spike eval error: %v", err)
			}
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