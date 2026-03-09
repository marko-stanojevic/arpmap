package sysinfo_test

import (
	"strings"
	"testing"

	"github.com/marko-stanojevic/hostr/internal/sysinfo"
)

func TestCollect_ReturnsInfo(t *testing.T) {
	t.Parallel()

	info, err := sysinfo.Collect()
	if err != nil {
		t.Fatalf("Collect() returned unexpected error: %v", err)
	}

	if info.OS == "" {
		t.Error("expected OS to be non-empty")
	}
	if info.Arch == "" {
		t.Error("expected Arch to be non-empty")
	}
	if !strings.HasPrefix(info.GoVersion, "go") {
		t.Errorf("expected GoVersion to start with 'go', got %q", info.GoVersion)
	}
	if info.CollectedAt.IsZero() {
		t.Error("expected CollectedAt to be set")
	}
}

func TestCollect_MemoryPercent(t *testing.T) {
	t.Parallel()

	info, err := sysinfo.Collect()
	if err != nil {
		t.Fatalf("Collect() returned unexpected error: %v", err)
	}

	if info.MemPercent < 0 || info.MemPercent > 100 {
		t.Errorf("MemPercent out of range [0,100]: %f", info.MemPercent)
	}
}

func TestCollect_DiskPercent(t *testing.T) {
	t.Parallel()

	info, err := sysinfo.Collect()
	if err != nil {
		t.Fatalf("Collect() returned unexpected error: %v", err)
	}

	if info.DiskPercent < 0 || info.DiskPercent > 100 {
		t.Errorf("DiskPercent out of range [0,100]: %f", info.DiskPercent)
	}
}

func TestCollect_CPUUsage(t *testing.T) {
	t.Parallel()

	info, err := sysinfo.Collect()
	if err != nil {
		t.Fatalf("Collect() returned unexpected error: %v", err)
	}

	if info.CPUUsage < 0 || info.CPUUsage > 100 {
		t.Errorf("CPUUsage out of range [0,100]: %f", info.CPUUsage)
	}
}
