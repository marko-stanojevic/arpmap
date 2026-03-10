# Go Development Guidelines

This repository is written in Go. When generating or suggesting code, follow these guidelines.

## General Principles
- Follow idiomatic Go practices.
- Keep code simple, readable, and maintainable.
- Prefer composition over inheritance.
- Avoid unnecessary abstractions.

## Formatting
- Code must comply with `gofmt` and `go vet`.
- Follow standard Go formatting conventions.
- Keep line length reasonable for readability.

## Project Structure
- Follow standard Go project layout.
- Packages should have clear, focused responsibilities.
- Avoid circular dependencies.
- Prefer small packages with clear APIs.

## Error Handling
- Always handle errors explicitly.
- Avoid ignoring returned errors.
- Use `fmt.Errorf("message: %w", err)` for wrapping errors.
- Do not panic except in truly unrecoverable situations.

## Logging
- Use structured logging where possible.
- Do not log sensitive information.
- Logging should provide actionable context.

## Concurrency
- Use goroutines responsibly.
- Prefer channels for coordination when appropriate.
- Avoid shared mutable state when possible.
- Use `context.Context` for cancellation and timeouts.

## Testing
- Generate table-driven tests where applicable.
- Prefer the standard `testing` package.
- Tests should be deterministic and independent.
- Mock external dependencies when necessary.

## Dependencies
- Prefer the Go standard library when possible.
- Avoid unnecessary third-party dependencies.
- Dependencies must be actively maintained.

## Documentation
- Exported functions and types must have Go-style comments.
- Comments should explain *why*, not *what*.

## Performance
- Avoid premature optimization.
- Prefer clear code over micro-optimizations.
- Benchmark before optimizing critical paths.