<h1 align="center">arpmap</h1>

<p align="center"><strong>Cross-platform ARP network scanner for discovering active devices and finding free IP addresses on local IPv4 subnets.</strong></p>

<p align="center">
  <a href="https://github.com/marko-stanojevic/arpmap/actions/workflows/ci.yml"><img alt="CI" src="https://github.com/marko-stanojevic/arpmap/actions/workflows/ci.yml/badge.svg"></a>
  <a href="https://github.com/marko-stanojevic/arpmap/actions/workflows/release.yml"><img alt="Release" src="https://github.com/marko-stanojevic/arpmap/actions/workflows/release.yml/badge.svg"></a>
  <a href="https://github.com/marko-stanojevic/arpmap/releases/latest"><img alt="Latest Release" src="https://img.shields.io/github/v/release/marko-stanojevic/arpmap?display_name=tag"></a>
  <a href="https://github.com/marko-stanojevic/arpmap/actions/workflows/ci.yml"><img alt="Coverage" src="https://img.shields.io/badge/Coverage-go%20test%20--race%20--cover-brightgreen"></a>
  <a href="https://go.dev/"><img alt="Go Version" src="https://img.shields.io/badge/Go-1.22%2B-00ADD8?logo=go"></a>
  <a href="#platform-support"><img alt="OS Support" src="https://img.shields.io/badge/OS-Linux%20%7C%20macOS%20%7C%20Windows-blue"></a>
</p>

<p align="center">
  <a href="#quick-start">Quick Start</a> •
  <a href="#commands">Commands</a> •
  <a href="#installation">Installation</a> •
  <a href="#platform-support">Platform Support</a> •
  <a href="#documentation">Documentation</a>
</p>

<p align="center">
  <img src="docs/assets/arpmap-demo.gif" alt="arpmap demo">
</p>

> Built for fast Layer-2 visibility on networks where ping and higher-layer probes are filtered, rate-limited, or unavailable.

`arpmap` is a focused CLI for local IPv4 discovery and address planning.

| Workflow | Purpose | Output |
| --- | --- | --- |
| `scan` | Discover responding hosts on attached subnets | JSON with `IP -> MAC` mappings |
| `find` | Identify candidate free addresses that did not respond to ARP | JSON with available IPv4 addresses |

## Highlights

| Fast visibility | Automation-friendly | Cross-platform |
| --- | --- | --- |
| Scans every IPv4 host in each eligible local subnet | Writes structured JSON for scripts and tooling | Native support for Linux, macOS, and Windows |
| Useful when ICMP is blocked or unreliable | Debug output includes timing and response metrics | Windows support uses `SendARP()` without CGO |

## About

`arpmap` is designed for fast Layer-2 visibility on networks where ICMP or higher-layer probes may be filtered, rate-limited, or disabled.

- It resolves active non-loopback interfaces and scans every IPv4 host address in each attached subnet.
- Linux and macOS use raw Ethernet capture backends for request fan-out and ARP reply collection.
- Windows uses the native `SendARP()` API and does not require CGO.
- Output is written as structured JSON for easy automation and post-processing.
- Worker count and probe attempts are configurable per command.
- Debug mode prints scan parameters, timing summaries, response metrics, and sample response/no-response addresses.

## Requirements

- Linux/macOS with permissions for raw sockets (`root` or `CAP_NET_RAW`)
- Windows scan/find support via native `SendARP()` (`iphlpapi.dll`) without CGO

## Platform Support

- Linux: `scan` and `find` via raw `AF_PACKET` sockets
- macOS: `scan` and `find` via BPF
- Windows: `scan` and `find` via native `SendARP()` probes

## Installation

Pick the release artifact that matches your platform from the GitHub Releases page.

<table>
  <tr>
    <td valign="top" width="33%">
      <strong>Linux / macOS</strong><br>
      Archive install for direct use on your <code>PATH</code>.
      <pre lang="bash">tar -xzf arpmap_X.Y.Z_linux_amd64.tar.gz
sudo install arpmap /usr/local/bin/arpmap
arpmap --help</pre>
    </td>
    <td valign="top" width="33%">
      <strong>Windows</strong><br>
      Zip archive install for PowerShell environments.
      <pre lang="powershell">Expand-Archive .\arpmap_X.Y.Z_windows_amd64.zip -DestinationPath .\arpmap
.\arpmap\arpmap.exe --help</pre>
    </td>
    <td valign="top" width="33%">
      <strong>Ubuntu / Debian</strong><br>
      Native <code>.deb</code> package for apt-based systems.
      <pre lang="bash">wget https://github.com/marko-stanojevic/arpmap/releases/download/vX.Y.Z/arpmap_X.Y.Z_linux_amd64.deb
sudo apt install ./arpmap_X.Y.Z_linux_amd64.deb</pre>
    </td>
  </tr>
</table>

After installation, verify the CLI is available:

```bash
arpmap --help
```

## Quick Start

The commands below assume `arpmap` is already installed from a release artifact and available on your `PATH`.

```bash
# show CLI help
arpmap --help

# scan all eligible interfaces
sudo arpmap scan --output devices.json

# find free IPs (all interfaces, no limit)
sudo arpmap find --output free_ips.json

# scan a specific interface with debug output
sudo arpmap scan --interface eth0 --debug --output devices.json

# find 10 candidate free addresses with a custom worker count
sudo arpmap find --interface eth0 --count 10 --workers 128 --output free_ips.json
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

Sample output:

```json
{
  "interfaces": [
    {
      "interface": "eth0",
      "subnet": "192.168.1.0/24",
      "devices": [
        { "ip": "192.168.1.1", "mac": "dc:a6:32:00:11:02" },
        { "ip": "192.168.1.10", "mac": "48:21:0b:22:7f:31" },
        { "ip": "192.168.1.44", "mac": "84:3a:4b:10:45:99" }
      ]
    },
    {
      "interface": "wlan0",
      "subnet": "10.0.0.0/24",
      "devices": [
        { "ip": "10.0.0.1", "mac": "3c:52:82:2a:91:10" },
        { "ip": "10.0.0.23", "mac": "f0:2f:74:88:1d:6c" }
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

Sample output:

```json
{
  "interfaces": [
    {
      "interface": "eth0",
      "subnet": "192.168.1.0/24",
      "free_ips": [
        "192.168.1.120",
        "192.168.1.121",
        "192.168.1.122",
        "192.168.1.131"
      ]
    },
    {
      "interface": "wlan0",
      "subnet": "10.0.0.0/24",
      "free_ips": [
        "10.0.0.88",
        "10.0.0.91",
        "10.0.0.109"
      ]
    }
  ]
}
```

`find` performs an ARP scan first, then reports addresses that did not respond. When `--count` is set, the scan can stop early once enough candidate free addresses have been identified.

## Debug Output

`--debug` is intended for operator troubleshooting and performance tuning.

| Debug signal | What it tells you |
| --- | --- |
| Scan settings | Interface, subnet, worker count, attempts, and target volume |
| Timing summaries | Overall duration and backend read behavior |
| Response metrics | Responded targets, dispatch totals, and response rate |
| Sample addresses | A few responding and non-responding IPs for quick inspection |

Example:

```text
[INFO] Starting ARP scan on interface=eth0 subnets=192.168.1.0/24
[DEBUG] Scan parameters | targets=254 attempts=1
[DEBUG] Reader summary | reads=23 timeouts=11 unique_devices=18
[DEBUG] Scan completed | duration=2.14s dispatched=254 total_targets=254
[DEBUG] Response metrics | responded=18 dispatched=254 response_rate=7.1%
[DEBUG] Sample IP addresses responding to ARP requests: [192.168.1.1 192.168.1.10 192.168.1.44]
[DEBUG] Sample IP addresses with no ARP response: [192.168.1.120 192.168.1.121 192.168.1.122]
```

## Platform Implementation

Backends by platform:

- Linux: raw `AF_PACKET` socket with ARP filter attachment
- macOS: BPF device backend
- Windows: native `SendARP()` probing with per-target retry control

## Performance

`arpmap` favors fast fan-out with predictable completion behavior.

| Setting | Default |
| --- | --- |
| Workers on Linux/macOS | `256` |
| Workers on Windows | `64` |
| Probe attempts | `1` |
| Reply collection window | `2s` after dispatch completes |
| Read polling deadline | `200ms` |
| Retry spacing | `150ms` |

Practical guidance:

- Linux/macOS `/24` networks are usually handled in a single dispatch wave with the default worker count.
- Windows throughput depends more directly on `SendARP()` latency and the configured worker count.
- Increasing `--workers` can reduce runtime, but very high values may reduce stability on slower networks or hosts.
- Increasing `--attempts` can improve discovery on noisy networks, with a direct runtime tradeoff.

To measure speed on your environment:

```bash
time sudo arpmap scan --interface eth0 --output devices.json
```

## Packaging

- GitHub releases publish macOS/Linux archives, Windows zip archives, Debian `.deb` packages, and `checksums.txt`.
- Debian packages are generated with GoReleaser nFPM packaging.

## Documentation

- [Getting Started](docs/getting-started.md)
- [Architecture](docs/architecture.md)
- [CI/CD & Release Guide](docs/ci-cd.md)
- [Contributing](CONTRIBUTING.md)

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

Built with ❤️ by Marko Stanojevic
