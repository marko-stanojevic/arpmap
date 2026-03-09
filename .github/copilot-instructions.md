# GitHub Copilot Instructions

This repository is `arpmap`, a Go CLI for ARP-based host discovery and free-IP discovery on local IPv4 subnets.

## Stack

- **Language**: Go 1.22+
- **CLI framework**: Cobra
- **Networking**: raw sockets (`AF_PACKET` on Linux, BPF on macOS)
- **Linter**: golangci-lint
- **Release**: GoReleaser

## Key Conventions

- Commands (`cmd/`) are thin wires; all logic lives in `internal/`
- Keep command definitions and flags in `internal/cmd`
- Keep ARP internals and packet/socket logic in `internal/arp`
- Keep JSON DTOs in `internal/output`
- Wrap errors with `fmt.Errorf("context: %w", err)` at every boundary
- Every exported symbol needs a godoc comment
- Tests live next to the code they test (`foo_test.go` in the same directory)

## When Adding a New CLI Subcommand

1. Add `internal/cmd/<name>.go`
2. Define a `cobra.Command` and required flags
3. Implement `RunE` with wrapped errors
4. Register it in `internal/cmd/cmd.go`
5. Add tests in `internal/cmd` and/or relevant internal package

## When Changing Scan/Find Output

1. Update structures in `internal/output/output.go`
2. Populate new fields in `internal/cmd/scan.go` and/or `internal/cmd/find.go`
3. Update docs examples in `README.md` and `docs/`
4. Add tests to validate output shape and behavior

Refer to `AGENTS.md` for full coding guidelines.
