package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/marko-stanojevic/arpmap/internal/arp"
	"github.com/marko-stanojevic/arpmap/internal/output"
	"github.com/spf13/cobra"
)

type scanOptions struct {
	Interface string
	Output    string
	Debug     bool
	Workers   int
	Attempts  int
}

func (a *app) newScanCmd() *cobra.Command {
	options := &scanOptions{}
	scanCmd := &cobra.Command{
		Use:   "scan",
		Short: "Scan the local network and output IP→MAC mappings",
		Long: `Sends ARP requests to every host in each subnet assigned to the
selected interface(s) and records the IP-to-MAC mappings that respond.

Examples:
  arpmap scan --interface eth0 --output devices.json
  arpmap scan --output devices.json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.runScan(cmd, args, *options)
		},
	}

	scanCmd.Flags().StringVarP(&options.Interface, "interface", "i", "", "Network interface to scan (default: all non-loopback interfaces)")
	scanCmd.Flags().StringVarP(&options.Output, "output", "o", "devices.json", "Path to the output JSON file")
	scanCmd.Flags().BoolVar(&options.Debug, "debug", false, "Enable debug logging")
	scanCmd.Flags().IntVarP(&options.Workers, "workers", "w", 0, "Number of concurrent probe workers (0 = platform default)")
	scanCmd.Flags().IntVarP(&options.Attempts, "attempts", "a", 1, "Number of ARP probe attempts per target (default: 1)")
	return scanCmd
}

func (a *app) runScan(cmd *cobra.Command, args []string, options scanOptions) error {
	interfaces, err := a.resolveInterfaces(options.Interface)
	if err != nil {
		return fmt.Errorf("resolving interfaces: %w", err)
	}

	result := output.ScanResult{
		Interfaces: []output.InterfaceResult{},
	}

	failedInterfaces := 0
	successfulInterfaces := 0

	for _, ifc := range interfaces {
		fmt.Fprintf(os.Stderr, "[INFO] Starting ARP scan on interface=%s subnets=%s\n", ifc.Name, ifc.CIDRs)

		devices, err := a.scanNetwork(ifc, arp.WithDebug(options.Debug), arp.WithWorkers(options.Workers), arp.WithAttempts(options.Attempts))
		if err != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] ARP scan failed on interface=%s: %v\n", ifc.Name, err)
			failedInterfaces++
			continue
		}
		successfulInterfaces++

		ifaceResult := output.InterfaceResult{
			Interface: ifc.Name,
			Subnet:    ifc.CIDRs,
			Devices:   devices,
		}
		result.Interfaces = append(result.Interfaces, ifaceResult)

		fmt.Fprintf(os.Stderr, "[INFO] Completed ARP scan on interface=%s discovered_devices=%d\n", ifc.Name, len(devices))
	}

	if successfulInterfaces == 0 {
		fmt.Fprintf(os.Stderr, "[ERROR] All interface scans failed (count=%d); output file was not created\n", failedInterfaces)
		return fmt.Errorf("all interface scans failed")
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("marshalling JSON: %w", err)
	}

	if err := os.WriteFile(options.Output, data, 0644); err != nil {
		return fmt.Errorf("writing output file: %w", err)
	}

	fmt.Fprintf(os.Stderr, "[INFO] Scan results written to %s\n", options.Output)
	return nil
}
