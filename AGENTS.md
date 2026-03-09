# AI Coding Agent Guidelines

This document provides guidance for AI coding assistants (Copilot, Claude, etc.) when generating code in this repository.  
All code should be **production-ready, maintainable, tested, and auditable**.

---

## Purpose

AI agents must generate code that:

- Preserves existing behavior and public APIs
- Follows idiomatic Go conventions (Effective Go, Go Code Review Comments)
- Includes proper testing, error handling, and documentation
- Is readable, secure, and auditable
- Can be integrated into CI/CD pipelines without modification

---

## Hard Rules (Must Follow)

- Never swallow errors silently — always handle or explicitly propagate them
- Do not use `panic` in library/internal packages; reserve it only for unrecoverable startup conditions in `main`
- Write a `_test.go` file for every new non-trivial package
- Do not add `//nolint` directives without an explanatory comment
- Never hardcode secrets, credentials, or environment-specific paths
- All exported identifiers must have a doc comment starting with the identifier name
- Avoid `init()` functions; prefer explicit initialization

---

## Coding Patterns and Conventions

### Package Layout

```
cmd/main.go               // Thin entry point; wires dependencies
internal/<pkg>/           // Business logic, unexported outside module
internal/<pkg>/<pkg>.go   // Primary file, same name as package
internal/<pkg>/<pkg>_test.go
```

### Error Handling

```go
result, err := doSomething()
if err != nil {
    return fmt.Errorf("doSomething: %w", err)
}
```

- Wrap errors with `%w` to preserve the chain
- Use sentinel errors (`var ErrNotFound = errors.New(...)`) for known failure modes
- Add context at each layer; avoid logging AND returning the same error

### Naming

- Receivers: short, consistent, lowercase (e.g. `c` for command struct, `i` for info type)
- Acronyms: `ID`, `URL`, `HTTP` (all caps)
- Avoid stuttering in package/type names

### Concurrency

- Document goroutine ownership and channel directions
- Prefer `context.Context` propagation over raw `time.Sleep`/`select` loops
- Use `errgroup` for fan-out operations

### CLI and Scanner Patterns

- Keep Cobra command wiring in `internal/cmd`
- Keep interface discovery logic in `internal/iface`
- Keep ARP packet/socket behavior isolated in `internal/arp`
- Keep output DTOs centralized in `internal/output`
- Avoid hidden side effects in `init()`; prefer explicit command registration

---

## Testing Guidelines

- Use the standard `testing` package; table-driven tests where practical
- Test file must be in the same package (`package foo_test` preferred for black-box)
- Mock external I/O via interfaces, not concrete types
- Aim for ≥ 80 % statement coverage
- Use `-race` in CI (`go test -race ./...`)

**Example Test Structure:**

```go
func TestResolve_ReturnsInterfaces(t *testing.T) {
    t.Parallel()

    interfaces, err := iface.Resolve("")
    if err != nil {
        t.Fatalf("Resolve() unexpected error: %v", err)
    }
    if len(interfaces) == 0 {
        t.Error("expected at least one resolved interface")
    }
}
```

---

## CI/CD and Quality

- `go vet` and `golangci-lint` must pass with zero errors
- Cross-compilation targets: `linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`, `windows/amd64`
- Semantic versioning via git tags (`v1.2.3`)
- GoReleaser handles cross-compilation and GitHub Release creation

---

## Commit Messages

Use conventional commit prefixes to control changelog sections:

```bash
git commit -m "feat: add subnet filter for scan"       # New feature
git commit -m "fix: skip invalid IPv4 masks"           # Bug fix
git commit -m "docs: update getting-started guide"     # Docs only
git commit -m "chore: bump golangci-lint to v1.60"     # Maintenance
git commit -m "feat!: rename output field for scan"    # Breaking change
```

---

## Quick Reference for AI Agents

- Wrap errors with `%w`; never swallow or `panic` in library code
- Export only what callers need; keep internals in `internal/`
- Every exported symbol needs a doc comment
- Write parallel table-driven tests; always call `t.Parallel()`
- Keep command parsing in `internal/cmd` and network logic in `internal/arp`
- Keep JSON output contracts in `internal/output` consistent with docs
- Run `go mod tidy` after adding or removing imports

---

## Example AI-Friendly Function Template

```go
// DoSomething performs X and returns Y.
// It returns an error if Z.
func DoSomething(ctx context.Context, input string) (Result, error) {
    if input == "" {
        return Result{}, fmt.Errorf("DoSomething: input must not be empty")
    }

    raw, err := fetchData(ctx, input)
    if err != nil {
        return Result{}, fmt.Errorf("DoSomething: fetch: %w", err)
    }

    return transform(raw), nil
}
```

---

✅ **Summary:**

Follow this document when generating code in this repository. This ensures:

- Consistent, idiomatic, and auditable Go
- Proper error propagation and context handling
- Testable code with clear ownership boundaries
- AI agents generate **safe and production-ready** code without human rework
