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

## Commit Examples

```bash
git commit -m "feat: add optional per-interface timeout"
git commit -m "fix: skip link-local subnets in resolver"
git commit -m "docs: update find command examples"
git commit -m "chore: align golangci-lint settings"
```

## Need Help?

- See [docs/getting-started.md](docs/getting-started.md)
- See [docs/development.md](docs/development.md)
- Open an issue for discussion before large changes

Be respectful and collaborative.
