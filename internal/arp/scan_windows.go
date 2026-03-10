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
	windowsProbeAttempts    = 1
	windowsProbeRetryDelay  = 150 * time.Millisecond
	windowsResponseLogCap   = 3
	windowsNoResponseLogCap = 3
)

func scanWindows(info iface.Info, debug bool, workers int, attempts int, stopAtFreeCount int) ([]output.Device, error) {
	scanStart := time.Now()
	initStart := time.Now()
	if err := initializeSendARP(); err != nil {
		return nil, fmt.Errorf("initializing SendARP: %w", err)
	}
	initDuration := time.Since(initStart)

	targetStart := time.Now()
	targets := hostsFromNets(info.Nets)
	if len(targets) == 0 {
		return nil, nil
	}
	targetDuration := time.Since(targetStart)
	explicitWorkers := workers > 0
	if workers <= 0 {
		workers = windowsWorkerCount
	}
	if attempts <= 0 {
		attempts = defaultProbeAttempts
	}

	if debug {
		workerSource := "auto"
		if explicitWorkers {
			workerSource = "explicit"
		}
		fmt.Fprintf(os.Stderr, "[DEBUG] [windows] Scan started | mode=SendARP\n")
		fmt.Fprintf(os.Stderr, "[DEBUG] [windows] Startup timings | init=%v target_expansion=%v\n", initDuration, targetDuration)
		fmt.Fprintf(os.Stderr, "[DEBUG] [windows] Scan parameters | targets=%d workers=%d attempts=%d worker_source=%s\n", len(targets), workers, attempts, workerSource)
	}

	sem := make(chan struct{}, workers)
	var wg sync.WaitGroup
	var mu sync.Mutex
	devicesByIP := make(map[string]string, len(targets))
	errnoCounts := make(map[syscall.Errno]int)
	responseIPs := make([]string, 0, windowsResponseLogCap)
	noResponseIPs := make([]string, 0, windowsNoResponseLogCap)
	probedCount := 0
	responseCount := 0
	noResponseCount := 0
	errorCount := 0
	errCh := make(chan error, 1)
	stopCh := make(chan struct{}) // Signal to stop probing early
	shouldStop := func() bool {
		select {
		case <-stopCh:
			return true
		default:
			return false
		}
	}

	dispatchStart := time.Now()
	for _, ip := range targets {
		// Early exit check before starting new goroutine
		if stopAtFreeCount > 0 && noResponseCount >= stopAtFreeCount {
			close(stopCh)
		}
		if shouldStop() {
			break
		}

		target := cloneIP(ip)
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()

			// Check stop signal at the beginning
			if shouldStop() {
				return
			}

			mac, found, err := resolveMACWithSendARPRetries(target, attempts)

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
				freeCount := noResponseCount
				minInUseNeeded := len(targets) - stopAtFreeCount
				if stopAtFreeCount > 0 && debug && freeCount >= stopAtFreeCount {
					fmt.Fprintf(os.Stderr, "[DEBUG] [windows] Early-exit threshold reached | free=%d probed=%d in_use=%d\n", freeCount, probedCount, len(devicesByIP))
				}
				// Signal early exit if we have enough free IPs
				if stopAtFreeCount > 0 && freeCount >= stopAtFreeCount && len(devicesByIP) > minInUseNeeded {
					select {
					case <-stopCh:
					default:
						close(stopCh)
					}
				}
				if debug && len(noResponseIPs) < windowsNoResponseLogCap {
					noResponseIPs = append(noResponseIPs, target.String())
				}
				mu.Unlock()
				return
			}

			mu.Lock()
			devicesByIP[target.String()] = mac.String()
			responseCount++
			if debug && len(responseIPs) < windowsResponseLogCap {
				responseIPs = append(responseIPs, target.String())
			}
			mu.Unlock()
		}()
	}
	dispatchDuration := time.Since(dispatchStart)

	wg.Wait()
	close(errCh)
	if err, ok := <-errCh; ok {
		return nil, fmt.Errorf("scanning subnet hosts: %w", err)
	}

	if debug {
		totalDuration := time.Since(scanStart)
		fmt.Fprintf(os.Stderr, "[DEBUG] [windows] Timing summary | dispatch=%v total=%v\n", dispatchDuration, totalDuration)
		fmt.Fprintf(os.Stderr, "[DEBUG] [windows] Scan summary | probed=%d responded=%d no_response=%d errors=%d\n", probedCount, responseCount, noResponseCount, errorCount)
		if len(responseIPs) > 0 {
			fmt.Fprintf(os.Stderr, "[DEBUG] [windows] Sample IP addresses responding to ARP requests: %v\n", responseIPs)
		}
		if len(noResponseIPs) > 0 {
			fmt.Fprintf(os.Stderr, "[DEBUG] [windows] Sample IP addresses with no ARP response: %v\n", noResponseIPs)
		}
		if len(errnoCounts) > 0 {
			for errno, count := range errnoCounts {
				fmt.Fprintf(os.Stderr, "[DEBUG] [windows] SendARP error frequency | errno=%d count=%d\n", errno, count)
			}
		}
		fmt.Fprintf(os.Stderr, "[DEBUG] [windows] Discovery summary | unique_devices=%d\n", len(devicesByIP))
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
