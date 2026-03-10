package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/marko-stanojevic/arpmap/internal/arp"
	"github.com/marko-stanojevic/arpmap/internal/output"
	"github.com/spf13/cobra"
)

type findOptions struct {
	Interface string
	Output    string
	Count     int
	Debug     bool
	Workers   int
	Attempts  int
}

func (a *app) newFindCmd() *cobra.Command {
	options := &findOptions{}
	findCmd := &cobra.Command{
		Use:   "find",
		Short: "Find free (unused) IP addresses in the local subnet(s)",
		Long: `Scans the subnet(s) of the selected interface(s) via ARP and reports
IP addresses that did not respond — i.e. addresses available for assignment.

Examples:
  arpmap find --interface eth0 --count 10 --output free_ips.json
  arpmap find --output free_ips.json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return a.runFind(cmd, args, *options)
		},
	}

	findCmd.Flags().StringVarP(&options.Interface, "interface", "i", "", "Network interface to scan (default: all non-loopback interfaces)")
	findCmd.Flags().StringVarP(&options.Output, "output", "o", "free_ips.json", "Path to the output JSON file")
	findCmd.Flags().IntVarP(&options.Count, "count", "c", 0, "Maximum number of free IPs to return per subnet (0 = all)")
	findCmd.Flags().BoolVar(&options.Debug, "debug", false, "Enable debug logging")
	findCmd.Flags().IntVarP(&options.Workers, "workers", "w", 0, "Number of concurrent probe workers (0 = platform default)")
	findCmd.Flags().IntVarP(&options.Attempts, "attempts", "a", 1, "Number of ARP probe attempts per target (default: 1)")
	return findCmd
}

func (a *app) runFind(cmd *cobra.Command, args []string, options findOptions) error {
	interfaces, err := a.resolveInterfaces(options.Interface)
	if err != nil {
		return fmt.Errorf("resolving interfaces: %w", err)
	}

	result := output.FindResult{
		Interfaces: []output.FreeInterfaceResult{},
	}

	failedInterfaces := 0
	successfulInterfaces := 0

	for _, ifc := range interfaces {
		fmt.Fprintf(os.Stderr, "[INFO] Starting free-IP discovery on interface=%s subnets=%s\n", ifc.Name, ifc.CIDRs)

		freeIPs, err := a.findFreeIPs(ifc, options.Count, arp.WithDebug(options.Debug), arp.WithWorkers(options.Workers), arp.WithAttempts(options.Attempts))
		if err != nil {
			fmt.Fprintf(os.Stderr, "[ERROR] Free-IP discovery failed on interface=%s: %v\n", ifc.Name, err)
			failedInterfaces++
			continue
		}
		successfulInterfaces++

		ifaceResult := output.FreeInterfaceResult{
			Interface: ifc.Name,
			Subnet:    ifc.CIDRs,
			FreeIPs:   freeIPs,
		}
		result.Interfaces = append(result.Interfaces, ifaceResult)

		fmt.Fprintf(os.Stderr, "[INFO] Completed free-IP discovery on interface=%s free_addresses=%d\n", ifc.Name, len(freeIPs))
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

	fmt.Fprintf(os.Stderr, "[INFO] Free-IP results written to %s\n", options.Output)
	return nil
}
