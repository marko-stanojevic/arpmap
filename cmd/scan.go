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
	scanInterface string
	scanOutput    string
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan the local network and output IP→MAC mappings",
	Long: `Sends ARP requests to every host in each subnet assigned to the
selected interface(s) and records the IP-to-MAC mappings that respond.

Examples:
  arpmap scan --interface eth0 --output devices.json
  arpmap scan --output devices.json`,
	RunE: runScan,
}

func init() {
	scanCmd.Flags().StringVarP(&scanInterface, "interface", "i", "", "Network interface to scan (default: all non-loopback interfaces)")
	scanCmd.Flags().StringVarP(&scanOutput, "output", "o", "devices.json", "Path to the output JSON file")
}

func runScan(cmd *cobra.Command, args []string) error {
	interfaces, err := iface.Resolve(scanInterface)
	if err != nil {
		return fmt.Errorf("resolving interfaces: %w", err)
	}

	result := output.ScanResult{
		Interfaces: []output.InterfaceResult{},
	}

	for _, ifc := range interfaces {
		fmt.Fprintf(os.Stderr, "[*] Scanning %s (%s) ...\n", ifc.Name, ifc.CIDRs)

		devices, err := arp.Scan(ifc)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[!] %s: %v\n", ifc.Name, err)
			continue
		}

		ifaceResult := output.InterfaceResult{
			Interface: ifc.Name,
			Subnet:    ifc.CIDRs,
			Devices:   devices,
		}
		result.Interfaces = append(result.Interfaces, ifaceResult)

		fmt.Fprintf(os.Stderr, "[+] %s: found %d device(s)\n", ifc.Name, len(devices))
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("marshalling JSON: %w", err)
	}

	if err := os.WriteFile(scanOutput, data, 0644); err != nil {
		return fmt.Errorf("writing output file: %w", err)
	}

	fmt.Fprintf(os.Stderr, "[+] Results written to %s\n", scanOutput)
	return nil
}
