// Package sysinfo collects system information for display in the TUI.
package sysinfo

import (
	"fmt"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
)

// Info holds a snapshot of the current system state.
type Info struct {
	Hostname    string
	OS          string
	Arch        string
	GoVersion   string
	Uptime      string
	CPUModel    string
	CPUCores    int
	CPUUsage    float64
	MemTotal    string
	MemUsed     string
	MemPercent  float64
	DiskTotal   string
	DiskUsed    string
	DiskPercent float64
	CollectedAt time.Time
}

// Collect gathers a fresh snapshot of system information.
// It returns an Info value and any non-fatal error encountered.
func Collect() (Info, error) {
	info := Info{
		OS:          runtime.GOOS,
		Arch:        runtime.GOARCH,
		GoVersion:   runtime.Version(),
		CollectedAt: time.Now(),
	}

	// Host
	if h, err := host.Info(); err == nil {
		info.Hostname = h.Hostname
		info.Uptime = formatUptime(h.Uptime)
	}

	// CPU
	if cpus, err := cpu.Info(); err == nil && len(cpus) > 0 {
		info.CPUModel = cpus[0].ModelName
		info.CPUCores = int(cpus[0].Cores)
	}
	if percents, err := cpu.Percent(200*time.Millisecond, false); err == nil && len(percents) > 0 {
		info.CPUUsage = percents[0]
	}

	// Memory
	if vm, err := mem.VirtualMemory(); err == nil {
		info.MemTotal = formatBytes(vm.Total)
		info.MemUsed = formatBytes(vm.Used)
		info.MemPercent = vm.UsedPercent
	}

	// Disk (root / on Unix, C:\ on Windows)
	root := "/"
	if runtime.GOOS == "windows" {
		root = "C:\\"
	}
	if d, err := disk.Usage(root); err == nil {
		info.DiskTotal = formatBytes(d.Total)
		info.DiskUsed = formatBytes(d.Used)
		info.DiskPercent = d.UsedPercent
	}

	return info, nil
}

// formatBytes converts a byte count to a human-readable string.
func formatBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

// formatUptime converts seconds to a human-readable uptime string.
func formatUptime(seconds uint64) string {
	d := seconds / 86400
	h := (seconds % 86400) / 3600
	m := (seconds % 3600) / 60
	if d > 0 {
		return fmt.Sprintf("%dd %dh %dm", d, h, m)
	}
	return fmt.Sprintf("%dh %dm", h, m)
}
