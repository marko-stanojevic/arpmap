# Contributing to hostr

Thank you for your interest in contributing! We welcome bug fixes, improvements, and new ideas.

## Quick Start

1. **Fork and clone** the repository
2. **Create a branch**: `git checkout -b feat/your-feature`
3. **Make your changes**
4. **Run tests**: `go test ./... -race`
5. **Run lint**: `golangci-lint run ./...`
6. **Commit**: `git commit -m "feat: add your feature"`
7. **Push and open a Pull Request**

## What We Need

### Before You Start

- Check existing issues and PRs to avoid duplicates
- For significant changes, open an issue first to discuss the approach

### Code Requirements

✅ **Must have:**

- Tests pass (`go test ./... -race`)
- `go vet` and `golangci-lint` pass
- Tests for new packages or functions

✅ **Good to have:**

- Doc comments on all exported identifiers
- Clear commit messages following conventional commits
- Updated documentation if behaviour changes

### File Locations

- Command entry points → `cmd/<name>/main.go`
- Business logic → `internal/<pkg>/`
- Tests → same directory, `<file>_test.go`

## Commit Messages

```bash
git commit -m "feat: add CPU temperature display"    # New feature
git commit -m "fix: handle missing /proc/meminfo"    # Bug fix
git commit -m "docs: update getting-started guide"  # Docs only
git commit -m "chore: update golangci-lint config"  # Maintenance
```

## Need Help?

- 📖 See [Getting Started](docs/getting-started.md) for setup
- 📖 See [Development Guide](docs/development.md) for coding details
- 💬 Open an issue for questions

## Code of Conduct

Be respectful and collaborative. That's it.

---

**Thank you for contributing!** 🎉
