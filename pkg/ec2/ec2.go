package ec2

import (
	"context"
	"flag"
	"log"
	"math"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/rogarg19/chaos-nirvana/pkg/scenario"
)

type EC2Chaos struct{}

const DefaultMaxDiskFillMB = 512

func New() *EC2Chaos {
	return &EC2Chaos{}
}

func (*EC2Chaos) Start() {
	var configPath *string = flag.String("config", "config.json", "configuration file for chaos")
	if !flag.Parsed() {
		flag.Parse()
	}

	config := loadConfig(configPath)
	Run(config, 0)
}

func ConfigFromScenario(base Configuration, s scenario.Scenario) Configuration {
	switch s.Action {
	case scenario.ActionEC2HighCPU:
		base.EC2Config.EnableHighCPU = true
		base.EC2Config.CPUCores = s.Parameters.CPUCores
	case scenario.ActionEC2FullDisk:
		base.EC2Config.EnableFullDisk = true
		base.EC2Config.DiskFillPath = s.Parameters.DiskFillPath
	}
	return base
}

func ApplySafetyDefaults(config Configuration, maxDiskFillMB int, keepDiskFile bool) Configuration {
	if maxDiskFillMB <= 0 {
		maxDiskFillMB = DefaultMaxDiskFillMB
	}
	if config.EC2Config.MaxDiskFillMB <= 0 || config.EC2Config.MaxDiskFillMB > maxDiskFillMB {
		config.EC2Config.MaxDiskFillMB = maxDiskFillMB
	}
	config.EC2Config.KeepDiskFile = keepDiskFile || config.EC2Config.KeepDiskFile
	return config
}

func Run(config Configuration, duration time.Duration) {
	log.Printf("%+v", config)

	done := waitForStop(duration)

	var wg sync.WaitGroup

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if config.EC2Config.EnableHighCPU {
		cores := config.EC2Config.CPUCores
		if cores == 0 {
			cores = runtime.NumCPU()
		}
		for i := 0; i < cores; i++ {
			wg.Add(1)
			go highCPU(&wg, ctx)
		}
	}

	if config.EC2Config.EnableFullDisk {
		wg.Add(1)
		go fullDisk(&wg, config, ctx)
	}

	<-done

	cancel()

	log.Println("waiting for all child goroutines to exit gracefully...")
	wg.Wait()
	log.Println("all goroutines finished.")
}

func highCPU(wg *sync.WaitGroup, ctx context.Context) {
	defer wg.Done()

	log.Println("Starting high CPU simulation")

	for {
		select {
		case <-ctx.Done():
			return
		default:
			// CPU intensive calculation: compute pi approximation or fibonacci
			_ = math.Sqrt(123456789.0) * math.Sin(987654321.0)
			// Loop to keep CPU busy
			for i := 0; i < 1000000; i++ {
				math.Pow(float64(i), 2)
			}
		}
	}
}

func fullDisk(wg *sync.WaitGroup, config Configuration, ctx context.Context) {
	defer wg.Done()

	path := config.EC2Config.DiskFillPath
	if path == "" {
		path = "/tmp/chaos_fill"
	}

	maxBytes := int64(config.EC2Config.MaxDiskFillMB) * 1024 * 1024
	if maxBytes <= 0 {
		maxBytes = int64(DefaultMaxDiskFillMB) * 1024 * 1024
	}

	log.Printf("Starting full disk simulation at %s with limit %d MB", path, maxBytes/(1024*1024))

	// Create directory if needed
	if err := os.MkdirAll(path, 0755); err != nil {
		log.Printf("Failed to create disk fill directory %s: %v", path, err)
		return
	}

	file, err := os.CreateTemp(path, "chaos_*")
	if err != nil {
		log.Printf("Failed to create temp file: %v", err)
		return
	}
	if !config.EC2Config.KeepDiskFile {
		defer func() {
			if err := os.Remove(file.Name()); err != nil {
				log.Printf("Failed to clean up disk fill file %s: %v", file.Name(), err)
				return
			}
			log.Printf("Cleaned up disk fill file %s", file.Name())
		}()
	}
	defer file.Close()

	data := make([]byte, 1024*1024) // 1MB chunks
	for i := range data {
		data[i] = byte(i % 256)
	}

	var written int64
	for {
		select {
		case <-ctx.Done():
			return
		default:
			if written >= maxBytes {
				log.Printf("Reached max disk fill limit after writing %d MB", written/(1024*1024))
				return
			}
			chunk := data
			if remaining := maxBytes - written; remaining < int64(len(data)) {
				chunk = data[:int(remaining)]
			}
			n, err := file.Write(chunk)
			if err != nil {
				log.Printf("Disk full or error: %v", err)
				return
			}
			written += int64(n)
			file.Sync() // Force write to disk
		}
	}
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
