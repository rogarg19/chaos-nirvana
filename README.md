# Chaos Nirvana

A comprehensive tool for chaos engineering simulations across various cloud services. This tool helps simulate failure scenarios to test system resilience and improve reliability.

## Features

- Simulate chaos in multiple cloud services
- Configurable chaos scenarios
- Real-time monitoring and logging
- Graceful shutdown handling
- JSON-based configuration
- Plain-English scenario parsing for common chaos commands
- Dry-run and preflight previews for safer execution
- Blast-radius limits for Kubernetes pod deletion and disk-fill chaos
- Cleanup for temporary disk files and injected large Redis keys

## Supported Services

### Redis
Simulates connection flooding and monitors Redis performance metrics.

### ElastiCache Redis
Extends Redis chaos with:
- CPU spike simulation (up to 100% CPU usage via intensive Lua scripts)
- Large key injection (configurable size to halt consumers under load)

### EC2 Instances
Simulates infrastructure-level chaos:
- High CPU usage across all cores
- Disk space exhaustion

### Kubernetes
Supports pod failure scenarios through `kubectl`.

## Installation

Ensure you have Go 1.19 or later installed.

```bash
git clone https://github.com/rogarg19/chaos-nirvana.git
cd chaos-nirvana
go mod tidy
go build -o chaos-nirvana main.go
```

## Usage

Run the tool with the appropriate service type and configuration:

```bash
./chaos-nirvana -type <service> -config <config_file>
```

Or run a plain-English scenario:

```bash
./chaos-nirvana -prompt "Kill 50% pods of service payments in namespace checkout" -dry-run
```

Preview the parsed scenario before executing it:

```bash
./chaos-nirvana -prompt "put load on redis cluster such that CPU usage spikes to 90% for more than 15 minutes" -dry-run
```

Execute disruptive scenarios only after confirming intent:

```bash
./chaos-nirvana -prompt "Kill 2 pods with selector app=payments in namespace checkout" -yes
```

Safety flags:

- `-yes`: required for disruptive actions such as Kubernetes pod deletion, EC2 disk fill, and ElastiCache large key injection.
- `-dry-run`: prints the parsed scenario and any supported preflight plan without executing.
- `-preflight`: validates supported targets before execution or dry-run. Enabled by default; use `-preflight=false` to skip.
- `-max-pods`: maximum Kubernetes pods a run may delete. Defaults to 5.
- `-max-disk-mb`: maximum data EC2 disk-fill chaos may write. Defaults to 512 MB.
- `-keep-artifacts`: keep temporary chaos artifacts instead of cleaning them up.

### Examples

#### Redis Chaos
```bash
./chaos-nirvana -type redis -config cmd/redis/config.json
```

#### ElastiCache Redis Chaos
```bash
./chaos-nirvana -type elasticache -config cmd/elasticache/config.json -yes
```

#### EC2 Chaos
```bash
./chaos-nirvana -type ec2 -config cmd/ec2/config.json -yes
```

#### Kubernetes Pod Chaos
```bash
./chaos-nirvana -prompt "Kill 50% pods of service payments in namespace checkout" -dry-run
./chaos-nirvana -prompt "Kill 50% pods of service payments in namespace checkout" -yes
```

By default, service prompts map to the selector `app=<service>`. You can provide an explicit selector:

```bash
./chaos-nirvana -prompt "Kill 2 pods with selector app=payments in namespace checkout"
```

The tool will start the chaos simulation and run until interrupted (Ctrl+C). It provides real-time logging and monitoring output.

## Configuration

Each service has its own configuration file in JSON format. Examples are provided in the `cmd/<service>/config.json` files.

### Redis Configuration
- `host`, `port`: Redis server details
- `connections`: Number of concurrent connections
- `isKeysCommandEnabled`: Whether to use KEYS command (expensive)
- Other Redis-specific settings

### ElastiCache Configuration
- Inherits Redis settings
- `enableCpuSpike`: Enable CPU spike simulation
- `enableLargeKey`: Enable large key injection
- `largeKeySize`: Size of large key in MB
- `keepLargeKey`: Keep the injected large key after the run instead of deleting it

### EC2 Configuration
- `enableHighCpu`: Enable high CPU simulation
- `enableFullDisk`: Enable disk filling simulation
- `diskFillPath`: Path to fill with data
- `cpuCores`: Number of CPU cores to utilize (0 for all)
- `maxDiskFillMb`: Maximum MB to write during disk-fill chaos
- `keepDiskFile`: Keep the generated disk-fill file after the run instead of deleting it

## Warning

⚠️ **This tool is designed for chaos engineering in controlled environments only. Running it in production can cause service disruptions, data loss, or system instability. Always test in isolated, non-production environments first.**

## Contributing

Contributions are welcome! Please submit issues and pull requests on GitHub.

## License

Licensed under the terms specified in LICENSE.
