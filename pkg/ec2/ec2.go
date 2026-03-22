package ec2

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"runtime"
	"sync"
	"time"
)

type EC2Chaos struct{}

func New() *EC2Chaos {
	return &EC2Chaos{}
}

func (*EC2Chaos) Start() {
	var configPath *string = flag.String("config", "config.json", "configuration file for chaos")

	config := loadConfig(configPath)

	log.Printf("%+v", config)

	var done = make(chan struct{}, 1)

	go func() {
		os.Stdin.Read(make([]byte, 1))
		close(done)
	}()

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

	log.Printf("Starting full disk simulation at %s", path)

	// Create directory if needed
	os.MkdirAll(path, 0755)

	file, err := os.CreateTemp(path, "chaos_*")
	if err != nil {
		log.Printf("Failed to create temp file: %v", err)
		return
	}
	defer file.Close()

	data := make([]byte, 1024*1024) // 1MB chunks
	for i := range data {
		data[i] = byte(i % 256)
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
			_, err := file.Write(data)
			if err != nil {
				log.Printf("Disk full or error: %v", err)
				return
			}
			file.Sync() // Force write to disk
		}
	}
}