//go:build windows

package arp

import (
	"errors"
	"fmt"
	"net"
	"os/exec"
	"regexp"
	"strings"
	"sync"

	"github.com/marko-stanojevic/arpmap/internal/iface"
	"github.com/marko-stanojevic/arpmap/internal/output"
)

var windowsARPLineRe = regexp.MustCompile(`(?i)\b((?:\d{1,3}\.){3}\d{1,3})\s+([0-9a-f]{2}(?:[-:][0-9a-f]{2}){5})\b`)

func scanWindows(info iface.Info) ([]output.Device, error) {
	targets := hostsFromNets(info.Nets)
	if len(targets) == 0 {
		return nil, nil
	}

	targetSet := make(map[string]struct{}, len(targets))
	for _, ip := range targets {
		targetSet[ip.String()] = struct{}{}
	}

	sem := make(chan struct{}, workerCount)
	var wg sync.WaitGroup
	errCh := make(chan error, 1)

	for _, ip := range targets {
		targetIP := ip.String()
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			if err := primeWindowsARPEntry(targetIP); err != nil {
				select {
				case errCh <- err:
				default:
				}
			}
		}()
	}

	wg.Wait()
	close(errCh)
	if err, ok := <-errCh; ok {
		return nil, fmt.Errorf("probing subnet hosts: %w", err)
	}

	arpOutput, err := exec.Command("arp", "-a").Output()
	if err != nil {
		return nil, fmt.Errorf("running arp -a: %w", err)
	}

	parsed := parseWindowsARPTable(string(arpOutput))
	devices := make([]output.Device, 0, len(parsed))
	seen := make(map[string]struct{}, len(parsed))
	for _, device := range parsed {
		if _, ok := targetSet[device.IP]; !ok {
			continue
		}
		if _, duplicate := seen[device.IP]; duplicate {
			continue
		}
		seen[device.IP] = struct{}{}
		devices = append(devices, device)
	}

	return devices, nil
}

func primeWindowsARPEntry(ip string) error {
	cmd := exec.Command("ping", "-n", "1", "-w", "250", ip)
	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return nil
		}
		return fmt.Errorf("ping %s: %w", ip, err)
	}
	return nil
}

func parseWindowsARPTable(raw string) []output.Device {
	lines := strings.Split(raw, "\n")
	devices := make([]output.Device, 0, len(lines))

	for _, line := range lines {
		matches := windowsARPLineRe.FindStringSubmatch(line)
		if len(matches) != 3 {
			continue
		}

		ip := net.ParseIP(matches[1]).To4()
		if ip == nil {
			continue
		}

		mac := strings.ToLower(strings.ReplaceAll(matches[2], "-", ":"))
		if _, err := net.ParseMAC(mac); err != nil {
			continue
		}

		devices = append(devices, output.Device{IP: ip.String(), MAC: mac})
	}

	return devices
}
