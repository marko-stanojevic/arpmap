# arpmap

`arpmap` is a Go CLI for ARP-based host discovery on local IPv4 subnets.

It provides two commands:
- `scan`: discovers responding hosts and writes `IP -> MAC` mappings to JSON.
- `find`: reports candidate free IP addresses (addresses that did not respond).

## Requirements

- Go 1.22+
- Linux/macOS with permissions for raw sockets (`root` or `CAP_NET_RAW`)

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

- [Getting Started](docs/getting-started.md)
- [Development Guide](docs/development.md)
- [Architecture](docs/architecture.md)
- [CI/CD & Release Guide](docs/ci-cd.md)

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).
