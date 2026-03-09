# ⬡ hostr — Live System Information TUI

> A hobby terminal UI application demonstrating the power of TUIs for viewing real-time system metrics, built with Go and Bubble Tea

[![CI](https://github.com/marko-stanojevic/hostr/actions/workflows/ci.yml/badge.svg)](https://github.com/marko-stanojevic/hostr/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/marko-stanojevic/hostr)](go.mod)
[![License](https://img.shields.io/github/license/marko-stanojevic/hostr)](LICENSE)

---

## 💡 About hostr

hostr is a hobby project created to demonstrate the power and elegance of terminal user interfaces. It's a fast, interactive system information viewer built with Go and Bubble Tea.

**Why a TUI?**  
TUIs are lightweight, responsive, and work anywhere you have a terminal. No browser. No resource overhead. This project showcases how modern TUI frameworks make it easy to build beautiful, functional terminal applications.

---

## 🚀 Installation

### Pre-built Binaries

Download the latest release for your platform from [GitHub Releases](https://github.com/marko-stanojevic/hostr/releases).

```bash
# macOS (Apple Silicon)
curl -L https://github.com/marko-stanojevic/hostr/releases/download/v1.0.0/hostr_darwin_arm64.tar.gz | tar xz
./sysinfo

# Linux (x86_64)
curl -L https://github.com/marko-stanojevic/hostr/releases/download/v1.0.0/hostr_linux_amd64.tar.gz | tar xz
./sysinfo

# Windows (PowerShell)
$url = "https://github.com/marko-stanojevic/hostr/releases/download/v1.0.0/hostr_windows_amd64.zip"
Invoke-WebRequest -Uri $url -OutFile hostr.zip
Expand-Archive hostr.zip
.\sysinfo.exe
```

### From Source

Requires **Go 1.22+**.

```bash
git clone https://github.com/marko-stanojevic/hostr.git
cd hostr
go run ./cmd/sysinfo
```

## 🖥️ Using hostr

Launch the app with:

```bash
go run ./cmd/sysinfo
```

The TUI displays live system metrics organized by category:

```
⬡  System Info

╭─ System ──────────────────────────────╮
│  Hostname        myhost               │
│  OS / Arch       linux / amd64        │
│  Go Version      go1.22.3             │
│  Uptime          2d 4h 12m            │
╰───────────────────────────────────────╯
╭─ CPU ─────────────────────────────────╮
│  Model           Intel Core i7-...    │
│  Cores           8                    │
│  Usage           ████████░░░░  42.3%  │
╰───────────────────────────────────────╯
╭─ Memory ──────────────────────────────╮
│  Total           15.5 GB              │
│  Used            8.2 GB               │
│  Usage           █████████░░░  53.1%  │
╰───────────────────────────────────────╯
╭─ Disk ────────────────────────────────╮
│  Total           512.0 GB             │
│  Used            210.3 GB             │
│  Usage           █████░░░░░░░  41.1%  │
╰───────────────────────────────────────╯
Updated 14:22:07  •  r refresh  •  q quit
```

**Keyboard Shortcuts:**
- `r` — refresh metrics now
- `q` — quit

---

## ✨ Features

- **Live System Metrics** — CPU, memory, disk, hostname, OS, uptime, Go version
- **Auto-refresh** — updates every 3 seconds with manual refresh option
- **Cross-platform** — Linux, macOS, and Windows (no dependencies)
- **Fast & Lightweight** — Go binary, minimal resource usage
- **Modern TUI Demo** — showcases Bubble Tea framework capabilities with real-world use case

### Tech Stack

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) — TUI framework
- [Bubbles](https://github.com/charmbracelet/bubbles) — spinner and other UI components
- [Lipgloss](https://github.com/charmbracelet/lipgloss) — terminal styling
- [gopsutil](https://github.com/shirou/gopsutil) — system metric collection
- [GoReleaser](https://goreleaser.com/) — cross-platform binary builds
- [GitHub Actions](https://github.com/features/actions) — CI/CD automation

---

## 📂 Project Structure

```
hostr/
├── cmd/
│   └── sysinfo/
│       └── main.go              // Entry point
├── internal/
│   ├── cmd/
│   │   └── registry.go          // Command registry (extensible architecture)
│   ├── sysinfo/
│   │   ├── collector.go         // Collector interface for abstraction
│   │   ├── sysinfo.go           // Metric collection via gopsutil
│   │   └── sysinfo_test.go
│   └── ui/
│       ├── model.go             // Bubble Tea model, update, view
│       ├── model_test.go
│       └── styles.go            // Centralized style definitions
├── docs/
│   ├── getting-started.md
│   ├── development.md
│   ├── ci-cd.md
│   └── architecture.md
├── .devcontainer/
│   └── devcontainer.json
├── .github/
│   ├── actions/                 // Composite actions (go-lint, go-test, go-build, go-release)
│   ├── workflows/               // CI & release pipelines
│   └── copilot-instructions.md
├── .vscode/                     // Editor settings & tasks
├── .goreleaser.yml
├── .golangci.yml
├── AGENTS.md
├── CONTRIBUTING.md
└── go.mod
```

---

## 📖 Documentation

For developers:

- 🚀 [Getting Started](docs/getting-started.md) — setup and local development
- 🛠️ [Development Guide](docs/development.md) — project structure, conventions, testing
- 🔄 [CI/CD & Release Guide](docs/ci-cd.md) — GitHub Actions, versioning, releases
- 🏗️ [Architecture](docs/architecture.md) — design patterns and extensibility

---

## 🤝 Contributing

Contributions are welcome! See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

---

## 📄 License

This project is licensed under the MIT License — see [LICENSE](LICENSE) for details.

---

Built with ❤️ using [Bubble Tea](https://github.com/charmbracelet/bubbletea) by Charm.
