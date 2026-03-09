package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/marko-stanojevic/arpmap/arp"
	"github.com/marko-stanojevic/arpmap/iface"
	"github.com/marko-stanojevic/arpmap/output"
	"github.com/spf13/cobra"
)

var (
	findInterface string
	findOutput    string
	findCount     int
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
}

func runFind(cmd *cobra.Command, args []string) error {
	interfaces, err := iface.Resolve(findInterface)
	if err != nil {
		return fmt.Errorf("resolving interfaces: %w", err)
	}

	result := output.FindResult{
		Interfaces: []output.FreeInterfaceResult{},
	}

	for _, ifc := range interfaces {
		fmt.Fprintf(os.Stderr, "[*] Scanning %s (%s) ...\n", ifc.Name, ifc.CIDRs)

		freeIPs, err := arp.FindFree(ifc, findCount)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[!] %s: %v\n", ifc.Name, err)
			continue
		}

		ifaceResult := output.FreeInterfaceResult{
			Interface: ifc.Name,
			Subnet:    ifc.CIDRs,
			FreeIPs:   freeIPs,
		}
		result.Interfaces = append(result.Interfaces, ifaceResult)

		fmt.Fprintf(os.Stderr, "[+] %s: found %d free address(es)\n", ifc.Name, len(freeIPs))
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("marshalling JSON: %w", err)
	}

	if err := os.WriteFile(findOutput, data, 0644); err != nil {
		return fmt.Errorf("writing output file: %w", err)
	}

	fmt.Fprintf(os.Stderr, "[+] Results written to %s\n", findOutput)
	return nil
}
