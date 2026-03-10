# Development Guide

## Core Conventions

- Keep `cmd/main.go` as a thin entry point.
- Put all behavior in `internal/` packages.
- Wrap errors with context using `%w`.
- Add tests for non-trivial logic changes.
- Keep command side effects in command handlers, not package init logic.
- Prefer explicit construction over package globals and `init()` registration.

## Adding a New CLI Subcommand

1. Create a new file in `internal/cmd/` (for example `stats.go`).
2. Define command-local option structs for flags.
3. Add an `(*app).new...Cmd()` constructor that wires flags and delegates to an `(*app).run...(...)` method.
4. Register the new command from `(*app).newRootCmd()` in `internal/cmd/cmd.go`.
5. Add tests for flag defaults and command behavior.

Minimal pattern:

```go
type statsOptions struct {
   output string
}

func (a *app) newStatsCmd() *cobra.Command {
   opts := statsOptions{}

   cmd := &cobra.Command{
      Use:   "stats",
      Short: "Show scan statistics",
      RunE: func(cmd *cobra.Command, args []string) error {
         return a.runStats(cmd.Context(), opts)
      },
   }

   cmd.Flags().StringVarP(&opts.output, "output", "o", "stats.json", "Output JSON path")

   return cmd
}

func (a *app) runStats(ctx context.Context, opts statsOptions) error {
   _ = ctx
   _ = opts
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
- Keep the Windows `SendARP()` path behaviorally aligned with Linux/macOS where practical.
- Preserve bounded concurrency to avoid FD exhaustion.
- Preserve early-stop behavior for `find --count`.
- Keep retry behavior centralized so probe semantics stay consistent across platforms.
- Validate timeouts and reply parsing logic carefully.
- Avoid introducing blocking behavior in the read path.

## Local Quality Checks

```bash
go mod tidy
go build ./...
go vet ./...
golangci-lint run ./...
go test ./... -v -cover
```

Run the workspace test task when you want the full CI-style local pass, including `-race`:

```bash
go test ./... -v -race -cover
```

## Release Notes

Before release workflows:
- ensure command docs are current,
- ensure README examples reflect current flags and defaults,
- ensure JSON output changes are documented,
- ensure release asset filters still publish only archives and checksums,
- ensure CI checks are green.

## Current Command Shape

- `internal/cmd/cmd.go` owns the `app` dependency container and root command assembly.
- `internal/cmd/scan.go` and `internal/cmd/find.go` keep flags in command-local option structs.
- Tests in `internal/cmd/cmd_test.go` should cover both default flags and handler behavior.
