# CI/CD & Release Guide

## Workflows

| Workflow | Trigger | Purpose |
|----------|---------|---------|
| `ci.yml` | Push to `main`, PRs | Lint → Test (3 OS) → Build |
| `release.yml` | Push `v*.*.*` tag | Cross-compile + GitHub Release |

## Creating a Release

```bash
# Bump and tag
git tag v1.2.3
git push origin v1.2.3
```

GoReleaser will automatically:
- Cross-compile for Linux, macOS, and Windows (amd64 + arm64)
- Create archives (`.tar.gz` / `.zip`)
- Generate a `checksums.txt`
- Publish a GitHub Release with all artifacts

## Versioning

Tags follow [Semantic Versioning](https://semver.org/): `vMAJOR.MINOR.PATCH`.

Commit message conventions for the changelog:

| Prefix | Changelog section |
|--------|------------------|
| `feat:` | Features |
| `fix:` | Bug Fixes |
| `feat!:` / `fix!:` | Breaking Changes |
| `docs:`, `chore:`, `test:` | Excluded from changelog |

## Composite Actions

| Action | Purpose |
|--------|---------|
| `go-lint` | Runs `golangci-lint` |
| `go-test` | Runs `go test ./... -race -cover` |
| `go-build` | Debug build or cross-compile release binaries |
| `go-release` | Runs GoReleaser to publish a GitHub Release |

## Dependency Updates

Dependabot is configured to open weekly PRs for both Go module and GitHub Actions updates.
