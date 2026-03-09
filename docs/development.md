# Development Guide

## Adding a New TUI Command

1. Create the entry point:

   ```bash
   mkdir -p cmd/mycommand
   touch cmd/mycommand/main.go
   ```

2. Create the UI model in `internal/ui/mycommand/model.go`:

   ```go
   package mycommand

   import tea "github.com/charmbracelet/bubbletea"

   type Model struct{}

   func NewModel() Model { return Model{} }
   func (m Model) Init() tea.Cmd { return nil }
   func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return m, nil }
   func (m Model) View() string { return "Hello from mycommand!" }
   ```

3. Wire it in `cmd/mycommand/main.go`:

   ```go
   package main

   import (
       tea "github.com/charmbracelet/bubbletea"
       "github.com/marko-stanojevic/hostr/internal/ui/mycommand"
   )

   func main() {
       p := tea.NewProgram(mycommand.NewModel(), tea.WithAltScreen())
       p.Run()
   }
   ```

4. Add a corresponding `_test.go` file.

5. Add a new entry under `builds:` in `.goreleaser.yml`.

## Adding New System Metrics

1. Add fields to `internal/sysinfo/Info`:

   ```go
   type Info struct {
       // ...existing fields...
       NetworkRx string
       NetworkTx string
   }
   ```

2. Populate them in `Collect()`:

   ```go
   if counters, err := net.IOCounters(false); err == nil && len(counters) > 0 {
       info.NetworkRx = formatBytes(counters[0].BytesRecv)
       info.NetworkTx = formatBytes(counters[0].BytesSent)
   }
   ```

3. Update `internal/ui/model.go` to render the new fields.

4. Add tests in `internal/sysinfo/sysinfo_test.go`.

## Running Tests

```bash
# All tests with race detector
go test ./... -race

# With coverage
go test ./... -race -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Linting

```bash
golangci-lint run ./...
```

## Building a Release Locally

```bash
goreleaser release --snapshot --clean
# Binaries appear in dist/
```
