# Architecture

This document describes the current `arpmap` CLI architecture.

## Package Layout

```text
cmd/
└── main.go                 # process entrypoint, delegates to internal/cmd

internal/
├── cmd/                    # Cobra root + subcommands
│   ├── cmd.go              # root command and registration
│   ├── scan.go             # scan command
│   └── find.go             # find command
├── iface/
│   └── iface.go            # interface discovery + IPv4 subnet extraction
├── arp/
│   ├── arp.go              # ARP packet crafting, scan logic, free-IP logic
│   ├── scan_windows.go     # Windows scan via native SendARP()
│   ├── socket_linux.go     # AF_PACKET implementation
│   ├── socket_darwin.go    # BPF implementation
│   └── socket_windows.go   # Windows raw socket placeholder (unused)
└── output/
    └── output.go           # JSON response DTOs
```

## Runtime Flow

1. `cmd/main.go` calls `internal/cmd.Execute()`.
2. Root command dispatches to `scan` or `find`.
3. Command resolves interfaces via `iface.Resolve(...)`.
4. `arp.Scan(...)` sends ARP probes and collects replies.
5. `scan` emits discovered devices; `find` emits non-responding host IPs.
6. Results are serialized to JSON files using structs in `internal/output`.

## Key Design Decisions

### Thin entrypoint

`cmd/main.go` contains startup wiring only. All behavior lives in `internal/`.

### Platform-specific raw sockets

Raw socket code is split by build tags:
- Linux: `AF_PACKET`
- macOS: BPF
- Windows: scan path uses `SendARP()` from `iphlpapi.dll` (no raw socket dependency)

This keeps platform details isolated and avoids branching inside core scanner logic.

### Bounded concurrency

ARP requests are sent with a fixed worker limit to reduce the chance of file descriptor exhaustion and excessive network burst load.

Current defaults in `internal/arp/arp.go`:

- worker pool: `256`
- post-send reply window: `2s`
- read deadline poll interval: `200ms`

### Best-effort network discovery

`scan` and `find` iterate multiple interfaces. Failures on one interface are logged to stderr and processing continues for the rest.

## Error Handling

- Errors are wrapped with context (`fmt.Errorf("context: %w", err)`).
- Interface resolution errors are surfaced early.
- Per-interface scan failures do not abort the entire multi-interface run.

## Testing Focus Areas

When adding tests, prioritize:
- interface and subnet parsing edge cases,
- ARP packet serialization/parsing behavior,
- command-level output JSON shape and file writing behavior.
