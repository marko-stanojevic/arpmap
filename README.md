# arpmap

[![CI](https://github.com/marko-stanojevic/arpmap/actions/workflows/ci.yml/badge.svg)](https://github.com/marko-stanojevic/arpmap/actions/workflows/ci.yml)
[![Release](https://github.com/marko-stanojevic/arpmap/actions/workflows/release.yml/badge.svg)](https://github.com/marko-stanojevic/arpmap/actions/workflows/release.yml)
[![Latest Release](https://img.shields.io/github/v/release/marko-stanojevic/arpmap?display_name=tag)](https://github.com/marko-stanojevic/arpmap/releases/latest)
[![Coverage](https://img.shields.io/badge/Coverage-go%20test%20--race%20--cover-brightgreen)](https://github.com/marko-stanojevic/arpmap/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/badge/Go-1.22%2B-00ADD8?logo=go)](https://go.dev/)
[![OS Support](https://img.shields.io/badge/OS-Linux%20%7C%20macOS%20%7C%20Windows-blue)](#platform-support)

`arpmap` is a Go CLI for ARP-based host discovery on local IPv4 subnets.

It provides two commands:
- `scan`: discovers responding hosts and writes `IP -> MAC` mappings to JSON.
- `find`: reports candidate free IP addresses (addresses that did not respond).

## About

`arpmap` is designed for fast Layer-2 visibility on local networks where ICMP or higher-layer probes may be filtered.

- It sends ARP requests directly over raw sockets and listens for ARP replies.
- It scans every host address in each resolved IPv4 subnet on the selected interface(s).
- It is optimized for operational simplicity: JSON output, predictable behavior, and bounded concurrency.

## Requirements

- Go 1.22+
- Linux/macOS with permissions for raw sockets (`root` or `CAP_NET_RAW`)
- Windows scan/find is supported using `ping` + ARP table discovery (`arp -a`).

## Platform Support

- Linux: full support (scan/find with raw sockets)
- macOS: full support (scan/find with BPF backend)
- Windows: scan/find supported via `ping` priming + `arp -a` parsing

## Quick Start

```bash
git clone https://github.com/marko-stanojevic/arpmap.git
cd arpmap

# show CLI help
go run ./cmd --help

# scan all eligible interfaces
sudo go run ./cmd scan --output devices.json

# find free IPs (all interfaces, no limit)
sudo go run ./cmd find --output free_ips.json
```

## Commands

### `scan`

```bash
arpmap scan --interface eth0 --output devices.json
arpmap scan --output devices.json
```

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
```

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

## Technical Details

- Command wiring is in `internal/cmd` (Cobra); scan logic is in `internal/arp`.
- Interface/subnet resolution is handled by `internal/iface`.
- Raw socket backends are split by OS build tags:
	- Linux: `AF_PACKET` implementation
	- macOS: BPF implementation
	- Windows: explicit unsupported-runtime stub for clear failure behavior
- ARP replies are parsed from Ethernet frames and deduplicated as `ip -> mac`.

## Performance

`arpmap` favors fast fan-out with predictable completion behavior.

- Concurrency: up to 256 concurrent probe workers.
- Reply collection window: 2 seconds after probe send completion.
- Read polling deadline: 200ms.

Practical implication:

- Fastest possible scan completion is slightly above 2 seconds (plus send/parsing overhead).
- `/24` networks (254 hosts) are typically handled in one send wave due worker count.
- Actual runtime depends on interface speed, kernel scheduling, local ARP behavior, and privilege context.

To measure speed on your environment:

```bash
time sudo go run ./cmd scan --interface eth0 --output devices.json
```

## Build and Verify

```bash
go mod tidy
go build ./...
go vet ./...
go test ./... -v -race -cover
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

- [About](docs/about.md)
- [Getting Started](docs/getting-started.md)
- [Development Guide](docs/development.md)
- [Architecture](docs/architecture.md)
- [CI/CD & Release Guide](docs/ci-cd.md)

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).
