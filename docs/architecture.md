# Architecture

This document describes the structure and design patterns used in hostr.

## Package Layout

```
cmd/sysinfo/           # Entry point for sysinfo command
├── main.go           # Thin wire-up of dependencies

internal/
├── cmd/              # Command infrastructure & registry
│   └── registry.go   # Command registration for future extensibility
├── sysinfo/          # System metrics collection
│   ├── sysinfo.go    # Core Collect() function & data structures
│   ├── collector.go  # Collector interface for abstraction & testing
│   └── sysinfo_test.go
├── ui/               # Bubble Tea TUI components
│   ├── model.go      # Main TUI model, Update, View logic
│   ├── styles.go     # UI styling (separated for maintainability)
│   └── model_test.go
```

## Design Patterns

### 1. Dependency Injection

The `Model` holds a `sysinfo.Collector` interface, allowing different implementations:
- **Production**: `DefaultCollector` uses `gopsutil`
- **Testing**: Mock collectors can be injected for unit tests
- **Future**: Custom collectors can be added without changing `Model`

```go
type Model struct {
    collector sysinfo.Collector  // Interface, not concrete type
    // ...
}
```

### 2. Error Resilience

When metric collection fails, the model retains the last known good data:
- Previous metrics remain visible instead of blanking on error
- Error is logged but doesn't destroy the display
- User can manually refresh (press 'r') when ready

```go
case errMsg:
    // Keep previous data, just log the error
    m.err = msg
    m.loading = false
```

### 3. Command Registry

Future commands (e.g., `hostr disk`, `hostr process`) can be registered without modifying main:
```go
reg := cmd.NewRegistry()
reg.Register("sysinfo", sysinfo.NewCommand())
reg.Register("disk", disk.NewCommand())
cmd, _ := reg.Get("sysinfo")
```

### 4. Side Effects via Commands

Bubble Tea best practice: all async work (I/O, timers) lives in `tea.Cmd` closures:
```go
func (m Model) collectCmd() tea.Cmd {
    return func() tea.Msg {
        info, err := m.collector.Collect(tea.Background())
        // ...
    }
}
```

## Testing Strategy

### Unit Tests
- `sysinfo_test.go`: Tests `Collect()` logic, formatted outputs
- `model_test.go`: Tests `Update()` message handling

### Dependency Injection for Testing
Create a mock `Collector` that returns fixed data:
```go
type MockCollector struct {
    Data sysinfo.Info
}
func (m *MockCollector) Collect(ctx context.Context) (sysinfo.Info, error) {
    return m.Data, nil
}
```

Then pass it to `Model`:
```go
m := &Model{collector: &MockCollector{...}}
```

## Adding a New Command

1. **Create the command package** in `internal/<command>/`
2. **Implement the `cmd.Command` interface**:
   - `Name()` → display name
   - `Description()` → brief description
   - `Execute()` → runs the command
3. **Register in main** (or use discovery):
   ```go
   registry.Register("newcmd", newcmd.NewCommand())
   ```

## Code Style

Follows AGENTS.md guidelines:
- All public symbols have godoc comments
- Errors wrapped with context: `fmt.Errorf("action: %w", err)`
- No `panic()` in library code
- Tests co-located with code
- Imports organized: stdlib → third-party → local

## Future Improvements

- **Config system**: `internal/config/` for user settings
- **Plugins**: Interface-based metric collectors for extensibility
- **Multi-command CLI**: Main dispatcher supporting `hostr sysinfo | disk | net`
- **Performance metrics**: Expose metrics via HTTP for external use
