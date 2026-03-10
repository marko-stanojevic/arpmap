# Getting Started

## Prerequisites

- Linux or macOS environment with raw socket permission (`root` or `CAP_NET_RAW`), or
- Windows with native `SendARP()` support available through `iphlpapi.dll`

## Platform Notes

- Linux: `scan` and `find` use raw `AF_PACKET` sockets.
- macOS: `scan` and `find` use the BPF backend.
- Windows: `scan` and `find` use native `SendARP()` probes and do not require CGO.

## Quick Start

Install `arpmap` from a release artifact first, then run it from your shell.

Linux/macOS example:

```bash
tar -xzf arpmap_X.Y.Z_linux_amd64.tar.gz
sudo install arpmap /usr/local/bin/arpmap
arpmap --help
```

Ubuntu/Debian example:

```bash
wget https://github.com/marko-stanojevic/arpmap/releases/download/vX.Y.Z/arpmap_X.Y.Z_linux_amd64.deb
sudo apt install ./arpmap_X.Y.Z_linux_amd64.deb
arpmap --help
```

Windows PowerShell install example:

```powershell
Expand-Archive .\arpmap_X.Y.Z_windows_amd64.zip -DestinationPath .\arpmap
.\arpmap\arpmap.exe --help
```

Common usage examples:

```bash
# show commands
arpmap --help

# scan for active hosts
sudo arpmap scan --output devices.json

# find candidate free IPs
sudo arpmap find --count 10 --output free_ips.json

# scan a specific interface with debug output
sudo arpmap scan --interface eth0 --debug --attempts 1 --output devices.json

# find free IPs with a custom worker count
sudo arpmap find --interface eth0 --count 10 --workers 128 --output free_ips.json
```

Windows PowerShell usage examples:

```powershell
.\dist\arpmap.exe scan --interface "Wi-Fi" --output devices.json
.\dist\arpmap.exe scan --interface "Wi-Fi" --debug --workers 120 --attempts 1
.\dist\arpmap.exe find --interface "Wi-Fi" --count 10 --output free_ips.json
```

## Default Runtime Settings

- Default worker count is `256` on Linux/macOS and `64` on Windows.
- Default ARP attempts per target is `1`.
- Set `--workers` to override concurrency.
- Set `--attempts` to retry noisy or inconsistent targets.

## Typical Output Files

- `devices.json`: per-interface list of discovered `ip` and `mac`
- `free_ips.json`: per-interface list of non-responding host IPs

## Debug Output

With `--debug`, `arpmap` prints scan configuration, timing summaries, response counts, and small sample lists for responding and non-responding IPs. This is useful when tuning worker counts or troubleshooting Windows `SendARP()` latency.

Example:

```text
[INFO] Scanning interface Wi-Fi (192.168.1.0/24)
[DEBUG] workers=64 attempts=1 targets=254
[DEBUG] sample responding IPs: 192.168.1.1, 192.168.1.10, 192.168.1.20
[DEBUG] sample non-responding IPs: 192.168.1.101, 192.168.1.102, 192.168.1.103
[DEBUG] total_duration=2.14s responses=18 no_responses=236
```

## Next Steps

- [Architecture](architecture.md)
- [Contributing](../CONTRIBUTING.md)
