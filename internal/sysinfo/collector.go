package sysinfo

import "context"

// Collector defines the interface for collecting system metrics.
// This abstraction allows swapping different metric sources for testing
// or extending with custom collectors in the future.
type Collector interface {
	// Collect gathers a fresh snapshot of system information.
	// The context can be used to cancel long-running operations.
	Collect(ctx context.Context) (Info, error)
}

// DefaultCollector is the default implementation using gopsutil.
type DefaultCollector struct{}

// Collect gathers system information using the standard gopsutil approach.
func (c *DefaultCollector) Collect(ctx context.Context) (Info, error) {
	// Use global Collect function for now; can refactor if needed.
	// Context is accepted for future cancellation support.
	return Collect()
}

// NewCollector returns a new default collector.
func NewCollector() Collector {
	return &DefaultCollector{}
}
