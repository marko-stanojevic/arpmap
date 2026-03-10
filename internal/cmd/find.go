package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/marko-stanojevic/arpmap/internal/arp"
	"github.com/marko-stanojevic/arpmap/internal/iface"
	"github.com/marko-stanojevic/arpmap/internal/output"
	"github.com/spf13/cobra"
)

var (
	findInterface string
	findOutput    string
	findCount     int
	findDebug     bool
	findWorkers   int
	findAttempts  int
)

var findCmd = &cobra.Command{
	Use:   "find",
	Short: "Find free (unused) IP addresses in the local subnet(s)",
	Long: `Scans the subnet(s) of the selected interface(s) via ARP and reports
IP addresses that did not respond — i.e. addresses available for assignment.

Examples:
  arpmap find --interface eth0 --count 10 --output free_ips.json
  arpmap find --output free_ips.json`,
	RunE: runFind,
}

func init() {
	findCmd.Flags().StringVarP(&findInterface, "interface", "i", "", "Network interface to scan (default: all non-loopback interfaces)")
	findCmd.Flags().StringVarP(&findOutput, "output", "o", "free_ips.json", "Path to the output JSON file")
	findCmd.Flags().IntVarP(&findCount, "count", "c", 0, "Maximum number of free IPs to return per subnet (0 = all)")
	findCmd.Flags().BoolVar(&findDebug, "debug", false, "Enable debug logging")
	findCmd.Flags().IntVarP(&findWorkers, "workers", "w", 0, "Number of concurrent probe workers (0 = platform default)")
	findCmd.Flags().IntVarP(&findAttempts, "attempts", "a", 1, "Number of ARP probe attempts per target (default: 1)")
}

func runFind(cmd *cobra.Command, args []string) error {
	interfaces, err := iface.Resolve(findInterface)
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

		freeIPs, err := arp.FindFree(ifc, findCount, arp.WithDebug(findDebug), arp.WithWorkers(findWorkers), arp.WithAttempts(findAttempts))
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

	if err := os.WriteFile(findOutput, data, 0644); err != nil {
		return fmt.Errorf("writing output file: %w", err)
	}

	fmt.Fprintf(os.Stderr, "[INFO] Free-IP results written to %s\n", findOutput)
	return nil
}
