//go:build !windows

package arp

import (
	"fmt"

	"github.com/marko-stanojevic/arpmap/internal/iface"
	"github.com/marko-stanojevic/arpmap/internal/output"
)

func scanWindows(_ iface.Info, _ bool, _ int, _ int) ([]output.Device, error) {
	return nil, fmt.Errorf("windows scan helper is unavailable on this platform")
}
