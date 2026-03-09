# Getting Started

## Prerequisites

- **Go 1.22+** — [download](https://go.dev/dl/)
- **Git**
- **Visual Studio Code** with the Go extension (recommended)
- **golangci-lint** — `brew install golangci-lint` or see [install docs](https://golangci-lint.run/welcome/install/)

Optional:
- **Docker / Rancher Desktop** for the devcontainer
- **goreleaser** for local release builds

## Quick Start

```bash
# 1. Clone the repository
git clone https://github.com/marko-stanojevic/hostr.git
cd hostr

# 2. Download dependencies
go mod download

# 3. Run the sysinfo TUI
go run ./cmd/sysinfo

# 4. Run tests
go test ./... -race

# 5. Run the linter
golangci-lint run ./...
```

## Keyboard Controls (sysinfo)

| Key | Action |
|-----|--------|
| `r` | Refresh stats immediately |
| `q` / `Ctrl+C` | Quit |

## Project Layout

```
hostr/
├── cmd/
│   └── sysinfo/main.go       # Entry point for the sysinfo command
├── internal/
│   ├── sysinfo/              # System metric collection
│   │   ├── sysinfo.go
│   │   └── sysinfo_test.go
│   └── ui/                   # Bubble Tea model + view
│       ├── model.go
│       └── model_test.go
├── docs/                     # Documentation
├── .github/                  # CI/CD workflows and templates
├── .devcontainer/            # Dev container config
├── .vscode/                  # Editor settings
├── .goreleaser.yml           # Cross-compilation & release
└── .golangci.yml             # Linter config
```

## Next Steps

- Read the [Development Guide](development.md) for adding new commands and metrics
- Read the [CI/CD Guide](ci-cd.md) for release and versioning workflows
- Read the [Architecture Guide](architecture.md) for design patterns and extensibility
