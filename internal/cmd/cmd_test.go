package cmd

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/marko-stanojevic/arpmap/internal/arp"
	"github.com/marko-stanojevic/arpmap/internal/iface"
	"github.com/marko-stanojevic/arpmap/internal/output"
)

func newTestApp() *app {
	return &app{
		resolveInterfaces: func(string) ([]iface.Info, error) { return nil, nil },
		scanNetwork:       func(iface.Info, ...arp.ScanOption) ([]output.Device, error) { return nil, nil },
		findFreeIPs:       func(iface.Info, int, ...arp.ScanOption) ([]string, error) { return nil, nil },
	}
}

func TestNewRootCmd_RegistersSubcommands(t *testing.T) {
	t.Parallel()

	root := newTestApp().newRootCmd()
	commands := root.Commands()
	if len(commands) != 2 {
		t.Fatalf("newRootCmd() command count = %d, want 2", len(commands))
	}

	got := map[string]bool{}
	for _, command := range commands {
		got[command.Name()] = true
	}
	if !got["scan"] || !got["find"] {
		t.Fatalf("newRootCmd() commands = %v, want scan and find", got)
	}
}

func TestNewScanCmd_DefaultFlags(t *testing.T) {
	t.Parallel()

	command := newTestApp().newScanCmd()

	workers, err := command.Flags().GetInt("workers")
	if err != nil {
		t.Fatalf("GetInt(workers) unexpected error: %v", err)
	}
	attempts, err := command.Flags().GetInt("attempts")
	if err != nil {
		t.Fatalf("GetInt(attempts) unexpected error: %v", err)
	}
	if workers != 0 {
		t.Fatalf("workers default = %d, want 0", workers)
	}
	if attempts != 1 {
		t.Fatalf("attempts default = %d, want 1", attempts)
	}
}

func TestNewFindCmd_DefaultFlags(t *testing.T) {
	t.Parallel()

	command := newTestApp().newFindCmd()

	count, err := command.Flags().GetInt("count")
	if err != nil {
		t.Fatalf("GetInt(count) unexpected error: %v", err)
	}
	attempts, err := command.Flags().GetInt("attempts")
	if err != nil {
		t.Fatalf("GetInt(attempts) unexpected error: %v", err)
	}
	if count != 0 {
		t.Fatalf("count default = %d, want 0", count)
	}
	if attempts != 1 {
		t.Fatalf("attempts default = %d, want 1", attempts)
	}
}

func TestRunScan_WritesResults(t *testing.T) {
	a := newTestApp()
	info := iface.Info{Name: "eth0", CIDRs: "192.168.1.0/24"}
	a.resolveInterfaces = func(name string) ([]iface.Info, error) {
		return []iface.Info{info}, nil
	}
	a.scanNetwork = func(info iface.Info, opts ...arp.ScanOption) ([]output.Device, error) {
		return []output.Device{{IP: "192.168.1.10", MAC: "aa:bb:cc:dd:ee:ff"}}, nil
	}

	tempDir := t.TempDir()
	options := scanOptions{
		Output:    filepath.Join(tempDir, "devices.json"),
		Interface: "eth0",
		Debug:     false,
		Workers:   0,
		Attempts:  1,
	}

	if err := a.runScan(a.newScanCmd(), nil, options); err != nil {
		t.Fatalf("runScan() unexpected error: %v", err)
	}

	data, err := os.ReadFile(options.Output)
	if err != nil {
		t.Fatalf("ReadFile() unexpected error: %v", err)
	}

	var result output.ScanResult
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Unmarshal() unexpected error: %v", err)
	}
	if len(result.Interfaces) != 1 {
		t.Fatalf("interfaces len = %d, want 1", len(result.Interfaces))
	}
	if len(result.Interfaces[0].Devices) != 1 {
		t.Fatalf("devices len = %d, want 1", len(result.Interfaces[0].Devices))
	}
}

func TestRunFind_AllFailuresReturnError(t *testing.T) {
	a := newTestApp()
	a.resolveInterfaces = func(name string) ([]iface.Info, error) {
		return []iface.Info{{Name: "eth0", CIDRs: "192.168.1.0/24"}}, nil
	}
	a.findFreeIPs = func(info iface.Info, max int, opts ...arp.ScanOption) ([]string, error) {
		return nil, errors.New("boom")
	}

	options := findOptions{
		Output:    filepath.Join(t.TempDir(), "free_ips.json"),
		Interface: "eth0",
		Count:     1,
		Debug:     false,
		Workers:   0,
		Attempts:  1,
	}

	err := a.runFind(a.newFindCmd(), nil, options)
	if err == nil {
		t.Fatal("runFind() error = nil, want non-nil")
	}
	if err.Error() != "all interface scans failed" {
		t.Fatalf("runFind() error = %q, want %q", err.Error(), "all interface scans failed")
	}
}
