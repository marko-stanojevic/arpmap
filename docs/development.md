# Development Guide

## Core Conventions

- Keep `cmd/main.go` as a thin entry point.
- Put all behavior in `internal/` packages.
- Wrap errors with context using `%w`.
- Add tests for non-trivial logic changes.
- Keep command side effects in command handlers, not package init logic.

## Adding a New CLI Subcommand

1. Create a new file in `internal/cmd/` (for example `stats.go`).
2. Define a `cobra.Command` and flags.
3. Implement `RunE` with clear error wrapping.
4. Register the command in `internal/cmd/cmd.go` (`rootCmd.AddCommand(...)`).
5. Add tests for the new behavior (and fixtures if needed).

Minimal pattern:

```go
var statsCmd = &cobra.Command{
   Use:   "stats",
   Short: "Show scan statistics",
   RunE:  runStats,
}

func runStats(cmd *cobra.Command, args []string) error {
   // business logic
   return nil
}
```

## Extending Scan or Find Results

1. Update structures in `internal/output/output.go`.
2. Populate fields from `internal/cmd/scan.go` and/or `internal/cmd/find.go`.
3. Add tests to lock the JSON shape.
4. Update docs and examples.

## ARP Scanner Changes

When changing behavior in `internal/arp/`:
- Keep raw socket handling platform-safe (`socket_linux.go`, `socket_darwin.go`).
- Preserve bounded concurrency to avoid FD exhaustion.
- Validate timeouts and reply parsing logic carefully.
- Avoid introducing blocking behavior in the read path.

## Local Quality Checks

```bash
go mod tidy
go build ./...
go vet ./...
go test ./... -v -race -cover
golangci-lint run ./...
```

## Release Notes

Before release workflows:
- ensure command docs are current,
- ensure JSON output changes are documented,
- ensure CI checks are green.
