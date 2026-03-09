# GitHub Copilot Instructions

This repository is hostr, a hobby Go TUI application demonstrating system information display with [Bubble Tea](https://github.com/charmbracelet/bubbletea).

## Stack

- **Language**: Go 1.22+
- **TUI framework**: Bubble Tea (bubbletea) + Bubbles + Lipgloss
- **System metrics**: gopsutil/v3
- **Linter**: golangci-lint
- **Release**: GoReleaser

## Key Conventions

- Commands (`cmd/`) are thin wires; all logic lives in `internal/`
- Use `tea.Cmd` for all side effects (I/O, timers, data fetching)
- Wrap errors with `fmt.Errorf("context: %w", err)` at every boundary
- Every exported symbol needs a godoc comment
- Tests live next to the code they test (`foo_test.go` in the same directory)

## When Adding a New TUI Command

1. Create `cmd/<name>/main.go`
2. Add a model in `internal/ui/<name>/model.go`
3. Add a corresponding `model_test.go`
4. Wire a new entry in `.goreleaser.yml` under `builds:`

## When Adding New Metrics

1. Add the collector function in `internal/sysinfo/sysinfo.go`
2. Add fields to the `Info` struct
3. Update `internal/ui/model.go` to display the new fields
4. Write tests in `internal/sysinfo/sysinfo_test.go`

Refer to `AGENTS.md` for full coding guidelines.
