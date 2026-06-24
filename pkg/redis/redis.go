package redis

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rogarg19/chaos-nirvana/pkg/scenario"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

type RedisChaos struct{}

func New() *RedisChaos {
	return &RedisChaos{}
}

func (*RedisChaos) Start() {
	var configPath *string = flag.String("config", "config.json", "configuration file for chaos")
	if !flag.Parsed() {
		flag.Parse()
	}

	config := loadConfig(configPath)
	Run(config, 0)
}

func ConfigFromScenario(base Configuration, s scenario.Scenario) Configuration {
	if base.RedisConfig.Host == "" {
		base.RedisConfig.Host = "localhost"
	}
	if base.RedisConfig.Port == 0 {
		base.RedisConfig.Port = 6379
	}
	if base.RedisConfig.ReadTimeout == 0 {
		base.RedisConfig.ReadTimeout = 5
	}
	if base.RedisConfig.WriteTimeout == 0 {
		base.RedisConfig.WriteTimeout = 5
	}
	if base.RedisConfig.DialTimeout == 0 {
		base.RedisConfig.DialTimeout = 5
	}
	if base.RedisConfig.InfoInterval == 0 {
		base.RedisConfig.InfoInterval = 30
	}
	if base.RedisConfig.Options.Connections == 0 {
		base.RedisConfig.Options.Connections = 10
	}

	if s.Parameters.Connections > 0 {
		base.RedisConfig.Options.Connections = s.Parameters.Connections
	}
	base.RedisConfig.IsCluster = s.Parameters.Cluster
	if s.Target.Host != "" {
		base.RedisConfig.Host = s.Target.Host
	}
	if s.Parameters.UseKeysCommand {
		base.RedisConfig.IsKeysCommandEnabled = true
	}
	if s.Action == scenario.ActionRedisCPUSpike {
		base.RedisConfig.EnableCPUSpike = true
		base.RedisConfig.CPUSpikeWorkers = cpuSpikeWorkers(s.Parameters.CPUPercent)
	}
	return base
}

func Run(config Configuration, duration time.Duration) {
	log.Printf("%+v", config)

	done := waitForStop(duration)

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

	if config.RedisConfig.EnableCPUSpike {
		workers := config.RedisConfig.CPUSpikeWorkers
		if workers == 0 {
			workers = 5
		}
		for i := 0; i < workers; i++ {
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

func cpuSpikeWorkers(cpuPercent int) int {
	if cpuPercent <= 0 {
		return 5
	}
	workers := (cpuPercent + 19) / 20
	if workers < 1 {
		return 1
	}
	if workers > 10 {
		return 10
	}
	return workers
}

func cpuSpike(wg *sync.WaitGroup, config Configuration, ctx context.Context) {
	defer wg.Done()

	client := getClient(config)
	defer client.Close()

	script := `
	local function fib(n)
		if n <= 1 then return n end
		return fib(n-1) + fib(n-2)
	end
	return fib(35)
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

func waitForStop(duration time.Duration) <-chan struct{} {
	done := make(chan struct{})
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		defer signal.Stop(stop)
		defer close(done)
		if duration > 0 {
			timer := time.NewTimer(duration)
			defer timer.Stop()
			select {
			case <-timer.C:
			case <-stop:
			}
			return
		}

		stdinDone := make(chan struct{}, 1)
		go func() {
			_, _ = os.Stdin.Read(make([]byte, 1))
			stdinDone <- struct{}{}
		}()
		select {
		case <-stdinDone:
		case <-stop:
		}
	}()

	return done
}
