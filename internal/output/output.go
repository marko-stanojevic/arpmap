// Package output defines the JSON output structures for arpmap.

package output

// Device represents a discovered network host.
type Device struct {
	IP  string `json:"ip"`
	MAC string `json:"mac"`
}

// InterfaceResult holds scan results for one network interface.
type InterfaceResult struct {
	Interface string   `json:"interface"`
	Subnet    string   `json:"subnet"`
	Devices   []Device `json:"devices"`
}

// ScanResult is the top-level output for the scan command.
type ScanResult struct {
	Interfaces []InterfaceResult `json:"interfaces"`
}

// FreeInterfaceResult holds free-IP results for one network interface.
type FreeInterfaceResult struct {
	Interface string   `json:"interface"`
	Subnet    string   `json:"subnet"`
	FreeIPs   []string `json:"free_ips"`
}

// FindResult is the top-level output for the find command.
type FindResult struct {
	Interfaces []FreeInterfaceResult `json:"interfaces"`
}
