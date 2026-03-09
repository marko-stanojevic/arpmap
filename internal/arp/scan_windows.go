//go:build windows

package arp

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"os"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"github.com/marko-stanojevic/arpmap/internal/iface"
	"github.com/marko-stanojevic/arpmap/internal/output"
)

const (
	errorGenFailure      syscall.Errno = 31
	errorBadNetName      syscall.Errno = 67
	errorSemTimeout      syscall.Errno = 121
	errorNetworkUnreach  syscall.Errno = 1231
	errorHostUnreach     syscall.Errno = 1232
	errorElementNotFound syscall.Errno = 1168
)

var (
	iphlpapiDLL    = syscall.NewLazyDLL("iphlpapi.dll")
	sendARPProc    = iphlpapiDLL.NewProc("SendARP")
	initSendARPOne sync.Once
	initSendARPErr error
)

const (
	windowsWorkerCount      = 32
	windowsProbeAttempts    = 2
	windowsProbeRetryDelay  = 150 * time.Millisecond
	windowsNoResponseLogCap = 10
)

func scanWindows(info iface.Info, debug bool) ([]output.Device, error) {
	if err := initializeSendARP(); err != nil {
		return nil, fmt.Errorf("initializing SendARP: %w", err)
	}

	targets := hostsFromNets(info.Nets)
	if len(targets) == 0 {
		return nil, nil
	}

	if debug {
		fmt.Fprintf(os.Stderr, "[DEBUG] [windows] SendARP probing started\n")
		fmt.Fprintf(os.Stderr, "[DEBUG] [windows] Targets=%d, workers=%d, attempts=%d\n", len(targets), windowsWorkerCount, windowsProbeAttempts)
	}

	sem := make(chan struct{}, windowsWorkerCount)
	var wg sync.WaitGroup
	var mu sync.Mutex
	devicesByIP := make(map[string]string, len(targets))
	errnoCounts := make(map[syscall.Errno]int)
	noResponseIPs := make([]string, 0, windowsNoResponseLogCap)
	probedCount := 0
	responseCount := 0
	noResponseCount := 0
	errorCount := 0
	errCh := make(chan error, 1)

	for _, ip := range targets {
		target := cloneIP(ip)
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()

			mac, found, err := resolveMACWithSendARPRetries(target, windowsProbeAttempts)

			mu.Lock()
			probedCount++
			mu.Unlock()

			if err != nil {
				var errno syscall.Errno
				if errors.As(err, &errno) {
					mu.Lock()
					errnoCounts[errno]++
					mu.Unlock()
				}
				mu.Lock()
				errorCount++
				mu.Unlock()
				select {
				case errCh <- err:
				default:
				}
				return
			}
			if !found {
				mu.Lock()
				noResponseCount++
				if debug && len(noResponseIPs) < windowsNoResponseLogCap {
					noResponseIPs = append(noResponseIPs, target.String())
				}
				mu.Unlock()
				return
			}

			mu.Lock()
			devicesByIP[target.String()] = mac.String()
			responseCount++
			mu.Unlock()
		}()
	}

	wg.Wait()
	close(errCh)
	if err, ok := <-errCh; ok {
		return nil, fmt.Errorf("scanning subnet hosts: %w", err)
	}

	if debug {
		fmt.Fprintf(os.Stderr, "[DEBUG] [windows] Probed=%d, responded=%d, no-response=%d, errors=%d\n", probedCount, responseCount, noResponseCount, errorCount)
		if len(noResponseIPs) > 0 {
			fmt.Fprintf(os.Stderr, "[DEBUG] [windows] Sample no-response IPs: %v\n", noResponseIPs)
		}
		if len(errnoCounts) > 0 {
			for errno, count := range errnoCounts {
				fmt.Fprintf(os.Stderr, "[DEBUG] [windows] SendARP errno=%d count=%d\n", errno, count)
			}
		}
		fmt.Fprintf(os.Stderr, "[DEBUG] [windows] Unique devices discovered=%d\n", len(devicesByIP))
	}

	devices := make([]output.Device, 0, len(devicesByIP))
	for ip, mac := range devicesByIP {
		devices = append(devices, output.Device{IP: ip, MAC: mac})
	}

	return devices, nil
}

func resolveMACWithSendARPRetries(target net.IP, attempts int) (net.HardwareAddr, bool, error) {
	if attempts < 1 {
		attempts = 1
	}

	for attempt := 1; attempt <= attempts; attempt++ {
		mac, found, err := resolveMACWithSendARP(target)
		if err != nil {
			return nil, false, err
		}
		if found {
			return mac, true, nil
		}
		if attempt < attempts {
			time.Sleep(windowsProbeRetryDelay)
		}
	}

	return nil, false, nil
}

func initializeSendARP() error {
	initSendARPOne.Do(func() {
		initSendARPErr = sendARPProc.Find()
	})
	return initSendARPErr
}

func resolveMACWithSendARP(target net.IP) (net.HardwareAddr, bool, error) {
	ip4 := target.To4()
	if ip4 == nil {
		return nil, false, fmt.Errorf("target %q is not IPv4", target.String())
	}

	var mac [8]byte
	macLen := uint32(len(mac))
	r1, _, _ := sendARPProc.Call(
		uintptr(binary.LittleEndian.Uint32(ip4)),
		0,
		uintptr(unsafe.Pointer(&mac[0])),
		uintptr(unsafe.Pointer(&macLen)),
	)
	if r1 != 0 {
		errno := syscall.Errno(r1)
		if isNoARPResponseError(errno) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("SendARP %s: %w", target.String(), errno)
	}
	if macLen == 0 || macLen > uint32(len(mac)) {
		return nil, false, nil
	}

	resolved := make(net.HardwareAddr, macLen)
	copy(resolved, mac[:macLen])
	return resolved, true, nil
}

func isNoARPResponseError(errno syscall.Errno) bool {
	switch errno {
	case errorGenFailure, errorBadNetName, errorSemTimeout, errorNetworkUnreach, errorHostUnreach, errorElementNotFound:
		return true
	default:
		return false
	}
}
