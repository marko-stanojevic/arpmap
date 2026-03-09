# CI/CD & Release Guide

## Workflows

### `ci.yml`

Triggers:
- push to `main` (when Go source/module files change)
- pull requests to `main`, `release/**`, `hotfix/**`
- manual dispatch

Pipeline:
1. `lint`: runs `.github/actions/go-lint`
2. `test`: runs `.github/actions/go-test` on Linux, macOS, and Windows
3. `build`: runs `.github/actions/go-build` and uploads debug artifacts

### `release.yml`

Trigger:
- manual dispatch (`workflow_dispatch`)

Pipeline:
1. run `.github/actions/go-release`
2. create/push release tag if missing
3. generate release notes
4. publish GitHub release with `dist/*` artifacts

## Local Validation Commands

```bash
go mod tidy
go build ./...
go vet ./...
go test ./... -v -race -cover
golangci-lint run ./...
```

## Versioning

Use semantic tags (`vMAJOR.MINOR.PATCH`) and conventional commits for clearer generated notes.

## Composite Actions

| Action | Purpose |
|--------|---------|
| `go-lint` | Run `golangci-lint` |
| `go-test` | Run tests with race detector and coverage |
| `go-build` | Build binaries/artifacts |
| `go-release` | Run release build and publish assets |
