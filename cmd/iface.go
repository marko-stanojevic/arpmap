// Package iface resolves network interfaces and their associated subnets.
package iface

import (
	"fmt"
	"net"
	"strings"
)

// Info holds a resolved interface and its associated IPv4 CIDRs.
type Info struct {
	Name  string
	Iface *net.Interface
	// CIDRs is a human-readable summary of all IPv4 networks on this interface.
	CIDRs  string
	// Nets holds the parsed IPv4 networks for iteration.
	Nets   []*net.IPNet
}

// Resolve returns interface info for the named interface, or all non-loopback
// interfaces with at least one IPv4 address when name is empty.
func Resolve(name string) ([]Info, error) {
	if name != "" {
		ifc, err := net.InterfaceByName(name)
		if err != nil {
			return nil, fmt.Errorf("interface %q not found: %w", name, err)
		}
		info, err := buildInfo(ifc)
		if err != nil {
			return nil, err
		}
		if info == nil {
			return nil, fmt.Errorf("interface %q has no usable IPv4 addresses", name)
		}
		return []Info{*info}, nil
	}

	all, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("listing interfaces: %w", err)
	}

	var result []Info
	for i := range all {
		ifc := all[i]
		// Skip loopback and down interfaces.
		if ifc.Flags&net.FlagLoopback != 0 || ifc.Flags&net.FlagUp == 0 {
			continue
		}
		info, err := buildInfo(&ifc)
		if err != nil || info == nil {
			continue
		}
		result = append(result, *info)
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("no usable interfaces found")
	}
	return result, nil
}

// buildInfo extracts IPv4 subnets from an interface. Returns nil when no IPv4
// address is found (caller should skip the interface).
func buildInfo(ifc *net.Interface) (*Info, error) {
	addrs, err := ifc.Addrs()
	if err != nil {
		return nil, fmt.Errorf("reading addresses for %s: %w", ifc.Name, err)
	}

	var nets []*net.IPNet
	var cidrs []string

	for _, addr := range addrs {
		var ipNet *net.IPNet
		switch v := addr.(type) {
		case *net.IPNet:
			ipNet = v
		case *net.IPAddr:
			ipNet = &net.IPNet{IP: v.IP, Mask: v.IP.DefaultMask()}
		default:
			continue
		}

		// Only IPv4, skip link-local (169.254.x.x).
		ip4 := ipNet.IP.To4()
		if ip4 == nil {
			continue
		}
		if ip4[0] == 169 && ip4[1] == 254 {
			continue
		}

		network := &net.IPNet{
			IP:   ip4.Mask(ipNet.Mask),
			Mask: ipNet.Mask,
		}
		nets = append(nets, network)
		cidrs = append(cidrs, network.String())
	}

	if len(nets) == 0 {
		return nil, nil
	}

	return &Info{
		Name:  ifc.Name,
		Iface: ifc,
		CIDRs: strings.Join(cidrs, ", "),
		Nets:  nets,
	}, nil
}
