package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/marko-stanojevic/arpmap/internal/arp"
	"github.com/marko-stanojevic/arpmap/internal/iface"
	"github.com/marko-stanojevic/arpmap/internal/output"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func newTestApp() *app {
	return &app{
		resolveInterfaces: func(string) ([]iface.Info, error) { return nil, nil },
		scanNetwork:       func(iface.Info, ...arp.ScanOption) ([]output.Device, error) { return nil, nil },
		findFreeIPs:       func(iface.Info, int, ...arp.ScanOption) ([]string, error) { return nil, nil },
	}
}

type flagExpectation struct {
	name         string
	shorthand    string
	defaultValue string
}

func assertFlag(t *testing.T, flags *pflag.FlagSet, want flagExpectation) {
	t.Helper()

	flag := flags.Lookup(want.name)
	if flag == nil {
		t.Fatalf("flag %q not registered", want.name)
	}
	if flag.Shorthand != want.shorthand {
		t.Fatalf("flag %q shorthand = %q, want %q", want.name, flag.Shorthand, want.shorthand)
	}
	if flag.DefValue != want.defaultValue {
		t.Fatalf("flag %q default = %q, want %q", want.name, flag.DefValue, want.defaultValue)
	}
}

func applyScanOptions(opts []arp.ScanOption) arp.ScanConfig {
	cfg := arp.ScanConfig{}
	for _, opt := range opts {
		opt(&cfg)
	}
	return cfg
}

func executeHelp(t *testing.T, commandArgs []string, cmdFactory func() *app, build func(*app) *cobra.Command) string {
	t.Helper()

	app := cmdFactory()
	command := build(app)
	buffer := &bytes.Buffer{}
	command.SetOut(buffer)
	command.SetErr(buffer)
	command.SetArgs(commandArgs)

	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() unexpected error: %v", err)
	}

	return buffer.String()
}

func assertContainsAll(t *testing.T, got string, wants ...string) {
	t.Helper()

	for _, want := range wants {
		if !strings.Contains(got, want) {
			t.Fatalf("output missing %q\nfull output:\n%s", want, got)
		}
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

func TestRootCmd_HelpIncludesSubcommands(t *testing.T) {
	t.Parallel()

	help := executeHelp(t, []string{"--help"}, newTestApp, func(a *app) *cobra.Command {
		return a.newRootCmd()
	})

	assertContainsAll(t, help,
		"arpmap",
		"arpmap discovers devices on local subnets using ARP probes.",
		"scan",
		"find",
	)
}

func TestNewScanCmd_RegistersAllFlags(t *testing.T) {
	t.Parallel()

	command := newTestApp().newScanCmd()

	assertFlag(t, command.Flags(), flagExpectation{name: "interface", shorthand: "i", defaultValue: ""})
	assertFlag(t, command.Flags(), flagExpectation{name: "output", shorthand: "o", defaultValue: "devices.json"})
	assertFlag(t, command.Flags(), flagExpectation{name: "debug", shorthand: "", defaultValue: "false"})
	assertFlag(t, command.Flags(), flagExpectation{name: "workers", shorthand: "w", defaultValue: "0"})
	assertFlag(t, command.Flags(), flagExpectation{name: "attempts", shorthand: "a", defaultValue: "1"})
}

func TestNewFindCmd_RegistersAllFlags(t *testing.T) {
	t.Parallel()

	command := newTestApp().newFindCmd()

	assertFlag(t, command.Flags(), flagExpectation{name: "interface", shorthand: "i", defaultValue: ""})
	assertFlag(t, command.Flags(), flagExpectation{name: "output", shorthand: "o", defaultValue: "free_ips.json"})
	assertFlag(t, command.Flags(), flagExpectation{name: "count", shorthand: "c", defaultValue: "0"})
	assertFlag(t, command.Flags(), flagExpectation{name: "debug", shorthand: "", defaultValue: "false"})
	assertFlag(t, command.Flags(), flagExpectation{name: "workers", shorthand: "w", defaultValue: "0"})
	assertFlag(t, command.Flags(), flagExpectation{name: "attempts", shorthand: "a", defaultValue: "1"})
}

func TestScanCmd_HelpIncludesFlagsAndExamples(t *testing.T) {
	t.Parallel()

	help := executeHelp(t, []string{"--help"}, newTestApp, func(a *app) *cobra.Command {
		return a.newScanCmd()
	})

	assertContainsAll(t, help,
		"Sends ARP requests to every host in each subnet assigned to the",
		"arpmap scan --interface eth0 --output devices.json",
		"arpmap scan --output devices.json",
		"--interface",
		"--output",
		"--debug",
		"--workers",
		"--attempts",
	)
}

func TestFindCmd_HelpIncludesFlagsAndExamples(t *testing.T) {
	t.Parallel()

	help := executeHelp(t, []string{"--help"}, newTestApp, func(a *app) *cobra.Command {
		return a.newFindCmd()
	})

	assertContainsAll(t, help,
		"Scans the subnet(s) of the selected interface(s) via ARP and reports",
		"arpmap find --interface eth0 --count 10 --output free_ips.json",
		"arpmap find --output free_ips.json",
		"--interface",
		"--output",
		"--count",
		"--debug",
		"--workers",
		"--attempts",
	)
}

func TestScanCmd_Execute_UsesAllFlags(t *testing.T) {
	t.Parallel()

	a := newTestApp()
	info := iface.Info{Name: "eth0", CIDRs: "192.168.1.0/24"}
	var resolvedInterface string
	var capturedInfo iface.Info
	var capturedCfg arp.ScanConfig

	a.resolveInterfaces = func(name string) ([]iface.Info, error) {
		resolvedInterface = name
		return []iface.Info{info}, nil
	}
	a.scanNetwork = func(gotInfo iface.Info, opts ...arp.ScanOption) ([]output.Device, error) {
		capturedInfo = gotInfo
		capturedCfg = applyScanOptions(opts)
		return []output.Device{{IP: "192.168.1.10", MAC: "aa:bb:cc:dd:ee:ff"}}, nil
	}

	outputPath := filepath.Join(t.TempDir(), "devices.json")
	command := a.newScanCmd()
	command.SetArgs([]string{"--interface", "eth0", "--output", outputPath, "--debug", "--workers", "17", "--attempts", "3"})

	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() unexpected error: %v", err)
	}
	if resolvedInterface != "eth0" {
		t.Fatalf("resolveInterfaces() argument = %q, want %q", resolvedInterface, "eth0")
	}
	if !reflect.DeepEqual(capturedInfo, info) {
		t.Fatalf("scanNetwork() interface = %#v, want %#v", capturedInfo, info)
	}
	if !capturedCfg.Debug {
		t.Fatal("scanNetwork() debug = false, want true")
	}
	if capturedCfg.Workers != 17 {
		t.Fatalf("scanNetwork() workers = %d, want 17", capturedCfg.Workers)
	}
	if capturedCfg.Attempts != 3 {
		t.Fatalf("scanNetwork() attempts = %d, want 3", capturedCfg.Attempts)
	}

	data, err := os.ReadFile(outputPath)
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
}

func TestFindCmd_Execute_UsesAllFlags(t *testing.T) {
	t.Parallel()

	a := newTestApp()
	info := iface.Info{Name: "eth1", CIDRs: "10.0.0.0/24"}
	var resolvedInterface string
	var capturedInfo iface.Info
	var capturedCount int
	var capturedCfg arp.ScanConfig

	a.resolveInterfaces = func(name string) ([]iface.Info, error) {
		resolvedInterface = name
		return []iface.Info{info}, nil
	}
	a.findFreeIPs = func(gotInfo iface.Info, max int, opts ...arp.ScanOption) ([]string, error) {
		capturedInfo = gotInfo
		capturedCount = max
		capturedCfg = applyScanOptions(opts)
		return []string{"10.0.0.25", "10.0.0.26"}, nil
	}

	outputPath := filepath.Join(t.TempDir(), "free_ips.json")
	command := a.newFindCmd()
	command.SetArgs([]string{"--interface", "eth1", "--output", outputPath, "--count", "5", "--debug", "--workers", "9", "--attempts", "2"})

	if err := command.Execute(); err != nil {
		t.Fatalf("Execute() unexpected error: %v", err)
	}
	if resolvedInterface != "eth1" {
		t.Fatalf("resolveInterfaces() argument = %q, want %q", resolvedInterface, "eth1")
	}
	if !reflect.DeepEqual(capturedInfo, info) {
		t.Fatalf("findFreeIPs() interface = %#v, want %#v", capturedInfo, info)
	}
	if capturedCount != 5 {
		t.Fatalf("findFreeIPs() count = %d, want 5", capturedCount)
	}
	if !capturedCfg.Debug {
		t.Fatal("findFreeIPs() debug = false, want true")
	}
	if capturedCfg.Workers != 9 {
		t.Fatalf("findFreeIPs() workers = %d, want 9", capturedCfg.Workers)
	}
	if capturedCfg.Attempts != 2 {
		t.Fatalf("findFreeIPs() attempts = %d, want 2", capturedCfg.Attempts)
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("ReadFile() unexpected error: %v", err)
	}

	var result output.FindResult
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Unmarshal() unexpected error: %v", err)
	}
	if len(result.Interfaces) != 1 {
		t.Fatalf("interfaces len = %d, want 1", len(result.Interfaces))
	}
	if len(result.Interfaces[0].FreeIPs) != 2 {
		t.Fatalf("free IP count = %d, want 2", len(result.Interfaces[0].FreeIPs))
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
