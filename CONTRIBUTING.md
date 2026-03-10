# Contributing to arpmap

Thanks for contributing. This project focuses on ARP-based subnet discovery and free-IP reporting.

## Quick Start

1. Fork and clone the repository.
2. Create a branch: `git checkout -b feat/your-feature`.
3. Install Go 1.22+ and make sure `go` is available on your `PATH`.
4. Install `golangci-lint` if you plan to run the full local validation set.
5. Implement your change.
6. Run checks:

```bash
go mod tidy
go build ./...
go vet ./...
go test ./... -v -race -cover
golangci-lint run ./...
```

1. Commit with a conventional message.
2. Open a pull request.

## Contribution Scope

Useful contributions include:

- ARP scan reliability and parsing improvements
- Interface/subnet resolution improvements
- Better command UX and validation
- Test coverage and docs updates

## Code Requirements

- Keep `cmd/main.go` thin; place behavior in `internal/` packages.
- Wrap errors using `%w` with context.
- Add tests for non-trivial behavior changes.
- Add/update godoc comments for exported identifiers.
- Update docs when command behavior or output changes.
- Prefer explicit construction over package globals and `init()` registration.

## Code Structure

- Keep command wiring in `internal/cmd` and reusable behavior in `internal/` packages.
- Keep command side effects in handlers, not in package initialization.
- Preserve behavioral alignment across Linux, macOS, and Windows implementations where practical.

## Adding a New CLI Subcommand

1. Create a new file in `internal/cmd/`.
2. Define a command-local options struct for flags.
3. Add an `(*app).new...Cmd()` constructor and delegate execution to an `(*app).run...(...)` method.
4. Register the command in `internal/cmd/cmd.go`.
5. Add tests for flag defaults, handler wiring, and output where relevant.

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
```

## Extending Output

1. Update structures in `internal/output/output.go`.
2. Populate fields from `internal/cmd/scan.go` and/or `internal/cmd/find.go`.
3. Add tests to lock the JSON shape.
4. Update README examples and any affected docs.

## ARP Backend Changes

When changing behavior in `internal/arp/`:

- Keep raw socket handling platform-safe in the platform-specific backend files.
- Preserve bounded concurrency to avoid FD exhaustion and unstable probe fan-out.
- Preserve early-stop behavior for `find --count`.
- Keep retry behavior centralized so probe semantics stay consistent across platforms.
- Validate timeout handling and reply parsing carefully.
- Avoid introducing blocking behavior in the read path.

## Local Validation

Run the full local check set before opening a PR:

```bash
go mod tidy
go build ./...
go vet ./...
go test ./... -v -race -cover
golangci-lint run ./...
```

Use the workspace tasks if you want the same commands through VS Code.

## CI/CD and Releases

### Workflows

- `ci.yml`: runs lint, test, and build jobs for pushes, pull requests, and manual dispatch.
- `release.yml`: runs the GoReleaser-based release flow from manual dispatch.

### Release Expectations

- Use semantic version tags in the form `vMAJOR.MINOR.PATCH`.
- Prefer conventional commits to keep generated release notes readable.
- Ensure README examples, command docs, and JSON output samples match the current behavior.
- Ensure release artifacts still include archives, Debian packages, and `checksums.txt`.

### Packaging Notes

- Debian packages are produced from `.goreleaser.yml` using nFPM.
- GitHub Releases are the primary distribution channel for archives and `.deb` packages.

## Commit Examples

```bash
git commit -m "feat: add optional per-interface timeout"
git commit -m "fix: skip link-local subnets in resolver"
git commit -m "docs: update find command examples"
git commit -m "chore: align golangci-lint settings"
```

## Need Help?

- See [docs/getting-started.md](docs/getting-started.md)
- See [docs/architecture.md](docs/architecture.md)
- Open an issue for discussion before large changes

Be respectful and collaborative.
