# arpmap

[![CI](https://github.com/marko-stanojevic/arpmap/actions/workflows/ci.yml/badge.svg)](https://github.com/marko-stanojevic/arpmap/actions/workflows/ci.yml)
[![Release](https://github.com/marko-stanojevic/arpmap/actions/workflows/release.yml/badge.svg)](https://github.com/marko-stanojevic/arpmap/actions/workflows/release.yml)
[![Latest Release](https://img.shields.io/github/v/release/marko-stanojevic/arpmap?display_name=tag)](https://github.com/marko-stanojevic/arpmap/releases/latest)
[![Coverage](https://img.shields.io/badge/Coverage-go%20test%20--race%20--cover-brightgreen)](https://github.com/marko-stanojevic/arpmap/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/badge/Go-1.22%2B-00ADD8?logo=go)](https://go.dev/)
[![OS Support](https://img.shields.io/badge/OS-Linux%20%7C%20macOS%20%7C%20Windows-blue)](#platform-support)

`arpmap` is a cross-platform Go CLI for ARP-based host discovery on local IPv4 subnets.

It provides two primary workflows:
- `scan`: discover responding hosts and write `IP -> MAC` mappings to JSON.
- `find`: identify candidate free IP addresses that did not respond to ARP probes.

## About

`arpmap` is designed for fast Layer-2 visibility on networks where ICMP or higher-layer probes may be filtered, rate-limited, or disabled.

- It resolves active non-loopback interfaces and scans every IPv4 host address in each attached subnet.
- Linux and macOS use raw Ethernet capture backends for request fan-out and ARP reply collection.
- Windows uses the native `SendARP()` API and does not require CGO.
- Output is written as structured JSON for easy automation and post-processing.
- Worker count and probe attempts are configurable per command.
- Debug mode prints scan parameters, timing summaries, response metrics, and sample response/no-response addresses.

## Requirements

- Go 1.22+
- Linux/macOS with permissions for raw sockets (`root` or `CAP_NET_RAW`)
- Windows scan/find support via native `SendARP()` (`iphlpapi.dll`) without CGO

## Platform Support

- Linux: `scan` and `find` via raw `AF_PACKET` sockets
- macOS: `scan` and `find` via BPF
- Windows: `scan` and `find` via native `SendARP()` probes

## Installation

### Release binaries

Download the latest archives from the GitHub Releases page and unpack the archive for your platform.

### Build from source

```bash
git clone https://github.com/marko-stanojevic/arpmap.git
cd arpmap
go build -o dist/arpmap ./cmd
```

## Quick Start

```bash
# show CLI help
go run ./cmd --help

# scan all eligible interfaces
sudo go run ./cmd scan --output devices.json

# find free IPs (all interfaces, no limit)
sudo go run ./cmd find --output free_ips.json

# scan a specific interface with debug output
sudo go run ./cmd scan --interface eth0 --debug --output devices.json

# find 10 candidate free addresses with a custom worker count
sudo go run ./cmd find --interface eth0 --count 10 --workers 128 --output free_ips.json
```

Windows PowerShell examples:

```powershell
.\arpmap.exe scan --interface "Wi-Fi" --output devices.json
.\arpmap.exe scan --interface "Wi-Fi" --debug --workers 120 --attempts 1
.\arpmap.exe find --interface "Wi-Fi" --count 10 --output free_ips.json
```

## Commands

### `scan`

```bash
arpmap scan --interface eth0 --output devices.json
arpmap scan --output devices.json
arpmap scan --interface eth0 --debug --workers 128 --attempts 1
```

Important flags:

- `-i, --interface`: scan a specific interface by name
- `-o, --output`: output JSON path, default `devices.json`
- `--debug`: print timing, response metrics, and sampled response/no-response IPs
- `-w, --workers`: concurrent probe workers, `0` uses the platform default
- `-a, --attempts`: ARP probe attempts per target, default `1`

Output schema:

```json
{
	"interfaces": [
		{
			"interface": "eth0",
			"subnet": "192.168.1.0/24",
			"devices": [
				{ "ip": "192.168.1.10", "mac": "aa:bb:cc:dd:ee:ff" }
			]
		}
	]
}
```

### `find`

```bash
arpmap find --interface eth0 --count 10 --output free_ips.json
arpmap find --output free_ips.json
arpmap find --interface eth0 --count 10 --debug --attempts 1
```

Important flags:

- `-i, --interface`: scan a specific interface by name
- `-o, --output`: output JSON path, default `free_ips.json`
- `-c, --count`: maximum number of free IPs to return per subnet, `0` returns all
- `--debug`: print timing, response metrics, and sampled response/no-response IPs
- `-w, --workers`: concurrent probe workers, `0` uses the platform default
- `-a, --attempts`: ARP probe attempts per target, default `1`

Output schema:

```json
{
	"interfaces": [
		{
			"interface": "eth0",
			"subnet": "192.168.1.0/24",
			"free_ips": ["192.168.1.20", "192.168.1.21"]
		}
	]
}
```

`find` performs an ARP scan first, then reports addresses that did not respond. When `--count` is set, the scan can stop early once enough candidate free addresses have been identified.

## Debug Output

`--debug` is intended for operator troubleshooting and performance tuning. It prints scan settings, timing summaries, response counts, and sampled response/no-response IP addresses.

Example:

```text
[INFO] Scanning interface Wi-Fi (192.168.1.0/24)
[DEBUG] workers=64 attempts=1 targets=254
[DEBUG] sample responding IPs: 192.168.1.1, 192.168.1.10, 192.168.1.20
[DEBUG] sample non-responding IPs: 192.168.1.101, 192.168.1.102, 192.168.1.103
[DEBUG] total_duration=2.14s responses=18 no_responses=236
```

## Technical Details

- Command wiring is implemented with Cobra in `internal/cmd`.
- Interface and subnet resolution is handled by `internal/iface`.
- Scan and free-IP logic live in `internal/arp`.
- Output DTOs live in `internal/output`.
- ARP replies are deduplicated as `ip -> mac` before JSON is written.

Backends by platform:

- Linux: raw `AF_PACKET` socket with ARP filter attachment
- macOS: BPF device backend
- Windows: native `SendARP()` probing with per-target retry control

## Performance

`arpmap` favors fast fan-out with predictable completion behavior.

- Default workers:
  - Linux/macOS: `256`
  - Windows: `64`
- Default probe attempts: `1`
- Reply collection window for raw-socket backends: `2s` after probe dispatch completes
- Read polling deadline for raw-socket backends: `200ms`
- Retry spacing between repeated probes: `150ms`

Practical implication:

- Linux/macOS `/24` networks are typically handled in a single dispatch wave with default settings.
- Windows throughput depends more directly on `SendARP()` latency and worker count.
- Increasing `--workers` can reduce wall-clock time, but setting it too high may reduce stability on slower networks or hosts.
- Increasing `--attempts` may improve discovery on noisy networks at the cost of extra runtime.

To measure speed on your environment:

```bash
time sudo go run ./cmd scan --interface eth0 --output devices.json
```

## Build and Verify

```bash
go mod tidy
go build ./...
go vet ./...
golangci-lint run ./...
go test ./... -v -cover
```

## Project Layout

```text
arpmap/
├── cmd/
│   └── main.go
├── internal/
│   ├── arp/
│   ├── cmd/
│   ├── iface/
│   └── output/
├── docs/
├── .github/
├── AGENTS.md
├── CONTRIBUTING.md
└── go.mod
```

## Documentation

- [Getting Started](docs/getting-started.md)
- [Development Guide](docs/development.md)
- [Architecture](docs/architecture.md)
- [CI/CD & Release Guide](docs/ci-cd.md)

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).
