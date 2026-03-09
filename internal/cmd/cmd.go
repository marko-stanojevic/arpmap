// Package cmd contains all CLI command definitions for arpmap.
package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "arpmap",
	Short: "ARP-based local network scanner",
	Long: `arpmap discovers devices on local subnets using ARP probes.

It operates at Layer 2 and requires raw socket privileges (run as root
or with CAP_NET_RAW capability).`,
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(scanCmd)
	rootCmd.AddCommand(findCmd)
}
