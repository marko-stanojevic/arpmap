# Getting Started

## Prerequisites

- Go 1.22+
- Git
- `golangci-lint` (recommended)
- Linux or macOS environment with raw socket permission (`root` or `CAP_NET_RAW`)

## Quick Start

```bash
git clone https://github.com/marko-stanojevic/arpmap.git
cd arpmap

go mod tidy
go build ./...

# show commands
go run ./cmd --help

# scan for active hosts
sudo go run ./cmd scan --output devices.json

# find candidate free IPs
sudo go run ./cmd find --count 10 --output free_ips.json
```

## Verify Your Setup

```bash
go vet ./...
go test ./... -v -race -cover
```

## Typical Output Files

- `devices.json`: per-interface list of discovered `ip` and `mac`
- `free_ips.json`: per-interface list of non-responding host IPs

## Project Layout

```text
arpmap/
├── cmd/
│   └── main.go
├── internal/
│   ├── arp/      # ARP packet building, sending, reply collection
│   ├── cmd/      # Cobra command definitions (root/scan/find)
│   ├── iface/    # Interface and IPv4 subnet resolution
│   └── output/   # JSON output DTOs
└── docs/
```

## Next Steps

- [Development Guide](development.md)
- [Architecture](architecture.md)
- [CI/CD & Release Guide](ci-cd.md)
