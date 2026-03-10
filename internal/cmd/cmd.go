// Package cmd contains all CLI command definitions for arpmap.
package cmd

import (
	"github.com/marko-stanojevic/arpmap/internal/arp"
	"github.com/marko-stanojevic/arpmap/internal/iface"
	"github.com/marko-stanojevic/arpmap/internal/output"
	"github.com/spf13/cobra"
)

type app struct {
	resolveInterfaces func(string) ([]iface.Info, error)
	scanNetwork       func(iface.Info, ...arp.ScanOption) ([]output.Device, error)
	findFreeIPs       func(iface.Info, int, ...arp.ScanOption) ([]string, error)
}

// Execute runs the root command.
func Execute() error {
	return newApp().newRootCmd().Execute()
}

func newApp() *app {
	return &app{
		resolveInterfaces: iface.Resolve,
		scanNetwork:       scanAdapter,
		findFreeIPs:       arp.FindFree,
	}
}

func (a *app) newRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "arpmap",
		Short: "ARP-based local network scanner",
		Long: `arpmap discovers devices on local subnets using ARP probes.

It operates at Layer 2 and requires raw socket privileges (run as root
or with CAP_NET_RAW capability).`,
	}

	rootCmd.AddCommand(a.newScanCmd())
	rootCmd.AddCommand(a.newFindCmd())
	return rootCmd
}

func scanAdapter(info iface.Info, opts ...arp.ScanOption) ([]output.Device, error) {
	return arp.Scan(info, opts...)
}
